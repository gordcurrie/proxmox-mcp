package proxmox

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const backupUPID = "UPID:pve1:000015E3:00000000:60F4B3A7:vzdump:100:root@pam:"

func TestCreateBackup_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/vzdump" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			http.Error(w, "want non-empty body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, backupUPID))
	}))
	defer srv.Close()

	req := &CreateBackupRequest{
		VMID:     100,
		Storage:  "local",
		Mode:     "snapshot",
		Compress: "zstd",
	}
	upid, err := newTestClient(t, srv.URL).CreateBackup(context.Background(), "pve1", req)
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	if upid != backupUPID {
		t.Errorf("upid: got %q, want %q", upid, backupUPID)
	}
}

func TestCreateBackup_apiError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "backup error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	req := &CreateBackupRequest{VMID: 100}
	_, err := newTestClient(t, srv.URL).CreateBackup(context.Background(), "pve1", req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestCreateBackup_nilReq(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, "http://localhost")
	_, err := c.CreateBackup(context.Background(), "pve1", nil)
	if err == nil {
		t.Fatal("expected error for nil req, got nil")
	}
}

func TestListBackups_success(t *testing.T) {
	t.Parallel()

	want := []StorageContent{
		{VolID: "local:backup/vzdump-qemu-100-2024_01_01.vma.zst", Content: "backup", Format: "vma.zst", Size: 5368709120},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/storage/local/content" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("content") != "backup" {
			http.Error(w, "want content=backup query param", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListBackups(context.Background(), "pve1", "local")
	if err != nil {
		t.Fatalf("ListBackups: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d backups, want 1", len(got))
	}
	if got[0].VolID != want[0].VolID {
		t.Errorf("volid: got %q, want %q", got[0].VolID, want[0].VolID)
	}
}

func TestListBackups_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListBackups(context.Background(), "missing", "local")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
