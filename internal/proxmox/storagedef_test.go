package proxmox

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListStorages_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"storage": "local", "type": "dir", "content": "iso,vztmpl,backup"},
		{"storage": "nfs-backups", "type": "nfs", "server": "pbs.storage.invalid"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storage" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListStorages(context.Background(), "")
	if err != nil {
		t.Fatalf("ListStorages: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d storages, want %d", len(got), len(want))
	}
	if got[0]["storage"] != "local" {
		t.Errorf("storage[0]: got %v, want local", got[0]["storage"])
	}
}

func TestListStorages_withTypeFilter(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"storage": "nfs-backups", "type": "nfs", "server": "pbs.storage.invalid"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storage" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("type") != "nfs" {
			http.Error(w, "want type=nfs query param", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListStorages(context.Background(), "nfs")
	if err != nil {
		t.Fatalf("ListStorages with type filter: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d storages, want 1", len(got))
	}
	if got[0]["type"] != "nfs" {
		t.Errorf("type: got %v, want nfs", got[0]["type"])
	}
}

func TestListStorages_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListStorages(context.Background(), "")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetStorage_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{
		"storage":   "pbs-store",
		"type":      "pbs",
		"server":    "pbs.storage.invalid",
		"datastore": "pbs-ds",
		"username":  "backup@pbs",
		"content":   "backup",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storage/pbs-store" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetStorage(context.Background(), "pbs-store")
	if err != nil {
		t.Fatalf("GetStorage: %v", err)
	}
	if got["storage"] != "pbs-store" {
		t.Errorf("storage: got %v, want pbs-store", got["storage"])
	}
	if got["type"] != "pbs" {
		t.Errorf("type: got %v, want pbs", got["type"])
	}
}

func TestGetStorage_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetStorage(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAddStorage_success(t *testing.T) {
	t.Parallel()

	req := &AddStorageRequest{
		Storage:   "pbs-store",
		Type:      "pbs",
		Server:    "pbs.storage.invalid",
		Datastore: "pbs-ds",
		Username:  "backup@pbs",
		Content:   "backup",
	}
	want := map[string]any{
		"storage":   "pbs-store",
		"type":      "pbs",
		"server":    "pbs.storage.invalid",
		"datastore": "pbs-ds",
		"username":  "backup@pbs",
		"content":   "backup",
	}
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/storage" {
			http.NotFound(w, r)
			return
		}
		var err error
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).AddStorage(context.Background(), req)
	if err != nil {
		t.Fatalf("AddStorage: %v", err)
	}
	if got["storage"] != "pbs-store" {
		t.Errorf("storage: got %v, want pbs-store", got["storage"])
	}

	// Verify request body contains expected fields.
	var decoded map[string]any
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	for _, key := range []string{"storage", "type", "server", "datastore", "username", "content"} {
		if decoded[key] == nil {
			t.Errorf("field %q should be present in request body but was absent", key)
		}
	}
	// Omitted fields must not appear in the body.
	for _, key := range []string{"path", "export", "password", "fingerprint", "nodes"} {
		if _, present := decoded[key]; present {
			t.Errorf("field %q should be omitted (omitempty) but was present in request body", key)
		}
	}
}

func TestAddStorage_apiError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "storage already exists", http.StatusBadRequest)
	}))
	defer srv.Close()

	req := &AddStorageRequest{Storage: "dupe", Type: "dir", Path: "/mnt/dupe"}
	_, err := newTestClient(t, srv.URL).AddStorage(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAddStorage_nilRequest(t *testing.T) {
	t.Parallel()

	_, err := newTestClient(t, "http://unused.invalid").AddStorage(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request, got nil")
	}
}

func TestUpdateStorage_success(t *testing.T) {
	t.Parallel()

	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/storage/pbs-store" {
			http.NotFound(w, r)
			return
		}
		var err error
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	// Only set Content — all other fields must be absent (omitempty).
	req := &UpdateStorageRequest{Content: "backup,iso"}
	if err := newTestClient(t, srv.URL).UpdateStorage(context.Background(), "pbs-store", req); err != nil {
		t.Fatalf("UpdateStorage: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if decoded["content"] != "backup,iso" {
		t.Errorf("content: got %v, want backup,iso", decoded["content"])
	}
	// Fields not set must not appear in the body.
	for _, key := range []string{"server", "export", "path", "datastore", "username", "password", "fingerprint", "nodes"} {
		if _, present := decoded[key]; present {
			t.Errorf("field %q should be omitted (omitempty) but was present in request body", key)
		}
	}
}

func TestUpdateStorage_apiError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid content type", http.StatusBadRequest)
	}))
	defer srv.Close()

	req := &UpdateStorageRequest{Content: "invalid"}
	if err := newTestClient(t, srv.URL).UpdateStorage(context.Background(), "pbs-store", req); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateStorage_nilRequest(t *testing.T) {
	t.Parallel()

	if err := newTestClient(t, "http://unused.invalid").UpdateStorage(context.Background(), "pbs-store", nil); err == nil {
		t.Fatal("expected error for nil request, got nil")
	}
}

func TestRemoveStorage_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/storage/pbs-store" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	if err := newTestClient(t, srv.URL).RemoveStorage(context.Background(), "pbs-store"); err != nil {
		t.Fatalf("RemoveStorage: %v", err)
	}
}

func TestRemoveStorage_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	if err := newTestClient(t, srv.URL).RemoveStorage(context.Background(), "missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
