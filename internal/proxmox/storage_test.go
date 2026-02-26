package proxmox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const storageDeleteUPID = "UPID:pve1:000015E3:00000000:60F4B3A7:imgdel:local:root@pam:"

func TestListStorageContent_success(t *testing.T) {
	t.Parallel()

	want := []StorageContent{
		{VolID: "local:iso/debian-12.iso", Content: "iso", Format: "iso", Size: 1234567890},
		{VolID: "local:vztmpl/debian-12-standard.tar.zst", Content: "vztmpl", Format: "tar.zst", Size: 987654321},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/storage/local/content" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListStorageContent(context.Background(), "pve1", "local", "")
	if err != nil {
		t.Fatalf("ListStorageContent: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d items, want %d", len(got), len(want))
	}
	for i, item := range got {
		if item.VolID != want[i].VolID || item.Content != want[i].Content {
			t.Errorf("item[%d]: got %+v, want %+v", i, item, want[i])
		}
	}
}

func TestListStorageContent_withContentFilter(t *testing.T) {
	t.Parallel()

	want := []StorageContent{
		{VolID: "local:iso/debian-12.iso", Content: "iso", Format: "iso", Size: 1234567890},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/storage/local/content" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("content") != "iso" {
			http.Error(w, "want content=iso query param", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListStorageContent(context.Background(), "pve1", "local", "iso")
	if err != nil {
		t.Fatalf("ListStorageContent with filter: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d items, want 1", len(got))
	}
	if got[0].VolID != want[0].VolID {
		t.Errorf("volid: got %q, want %q", got[0].VolID, want[0].VolID)
	}
}

func TestListStorageContent_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListStorageContent(context.Background(), "missing", "local", "")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetStorageContentInfo_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{
		"volid":   "local:iso/debian-12.iso",
		"content": "iso",
		"size":    float64(1234567890),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/nodes/pve1/storage/local/content/local:iso%2Fdebian-12.iso" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetStorageContentInfo(context.Background(), "pve1", "local", "local:iso/debian-12.iso")
	if err != nil {
		t.Fatalf("GetStorageContentInfo: %v", err)
	}
	if got["volid"] != "local:iso/debian-12.iso" {
		t.Errorf("volid: got %v, want local:iso/debian-12.iso", got["volid"])
	}
}

func TestGetStorageContentInfo_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetStorageContentInfo(context.Background(), "pve1", "local", "local:iso/missing.iso")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteStorageContent_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "want DELETE", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.EscapedPath() != "/nodes/pve1/storage/local/content/local:iso%2Fdebian-12.iso" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, storageDeleteUPID))
	}))
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).DeleteStorageContent(context.Background(), "pve1", "local", "local:iso/debian-12.iso")
	if err != nil {
		t.Fatalf("DeleteStorageContent: %v", err)
	}
	if upid != storageDeleteUPID {
		t.Errorf("upid: got %q, want %q", upid, storageDeleteUPID)
	}
}

func TestDeleteStorageContent_apiError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "storage error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).DeleteStorageContent(context.Background(), "pve1", "local", "local:iso/debian-12.iso")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}
