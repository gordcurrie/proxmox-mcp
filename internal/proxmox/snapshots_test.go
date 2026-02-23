package proxmox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const snapUPID = "UPID:pve1:000015E3:00000000:60F4B3A7:qmsnapshot:100:root@pam:"

func snapshotServer(t *testing.T, method, expectPath string, responseData any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != expectPath {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, responseData))
	}))
}

func snapshotErrorServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "snapshot error", http.StatusInternalServerError)
	}))
}

// VM snapshot tests

func TestListVMSnapshots_success(t *testing.T) {
	t.Parallel()

	want := []Snapshot{
		{Name: "snap1", Description: "first snapshot"},
		{Name: "snap2", Parent: "snap1"},
	}
	srv := snapshotServer(t, http.MethodGet, "/nodes/pve1/qemu/100/snapshot", want)
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListVMSnapshots(context.Background(), "pve1", 100)
	if err != nil {
		t.Fatalf("ListVMSnapshots: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d snapshots, want %d", len(got), len(want))
	}
	if got[0].Name != "snap1" {
		t.Errorf("got name %q, want %q", got[0].Name, "snap1")
	}
}

func TestListVMSnapshots_apiError(t *testing.T) {
	t.Parallel()
	srv := snapshotErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).ListVMSnapshots(context.Background(), "pve1", 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestListVMSnapshots_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListVMSnapshots(context.Background(), "missing", 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateVMSnapshot_success(t *testing.T) {
	t.Parallel()

	srv := snapshotServer(t, http.MethodPost, "/nodes/pve1/qemu/100/snapshot", snapUPID)
	defer srv.Close()

	req := CreateVMSnapshotRequest{Snapname: "snap1", Description: "test snap"}
	upid, err := newTestClient(t, srv.URL).CreateVMSnapshot(context.Background(), "pve1", 100, req)
	if err != nil {
		t.Fatalf("CreateVMSnapshot: %v", err)
	}
	if upid != snapUPID {
		t.Errorf("upid: got %q, want %q", upid, snapUPID)
	}
}

func TestCreateVMSnapshot_apiError(t *testing.T) {
	t.Parallel()
	srv := snapshotErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).CreateVMSnapshot(context.Background(), "pve1", 100, CreateVMSnapshotRequest{Snapname: "s"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestRollbackVMSnapshot_success(t *testing.T) {
	t.Parallel()

	srv := snapshotServer(t, http.MethodPost, "/nodes/pve1/qemu/100/snapshot/snap1/rollback", snapUPID)
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).RollbackVMSnapshot(context.Background(), "pve1", 100, "snap1")
	if err != nil {
		t.Fatalf("RollbackVMSnapshot: %v", err)
	}
	if upid != snapUPID {
		t.Errorf("upid: got %q, want %q", upid, snapUPID)
	}
}

func TestRollbackVMSnapshot_apiError(t *testing.T) {
	t.Parallel()
	srv := snapshotErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).RollbackVMSnapshot(context.Background(), "pve1", 100, "snap1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestDeleteVMSnapshot_success(t *testing.T) {
	t.Parallel()

	srv := snapshotServer(t, http.MethodDelete, "/nodes/pve1/qemu/100/snapshot/snap1", snapUPID)
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).DeleteVMSnapshot(context.Background(), "pve1", 100, "snap1")
	if err != nil {
		t.Fatalf("DeleteVMSnapshot: %v", err)
	}
	if upid != snapUPID {
		t.Errorf("upid: got %q, want %q", upid, snapUPID)
	}
}

func TestDeleteVMSnapshot_apiError(t *testing.T) {
	t.Parallel()
	srv := snapshotErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).DeleteVMSnapshot(context.Background(), "pve1", 100, "snap1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

// Container snapshot tests

const ctSnapUPID = "UPID:pve1:000015E3:00000000:60F4B3A7:vzsnap:200:root@pam:"

func TestListContainerSnapshots_success(t *testing.T) {
	t.Parallel()

	want := []Snapshot{
		{Name: "ctsnap1", Description: "container snapshot"},
	}
	srv := snapshotServer(t, http.MethodGet, "/nodes/pve1/lxc/200/snapshot", want)
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListContainerSnapshots(context.Background(), "pve1", 200)
	if err != nil {
		t.Fatalf("ListContainerSnapshots: %v", err)
	}
	if len(got) != 1 || got[0].Name != "ctsnap1" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestListContainerSnapshots_apiError(t *testing.T) {
	t.Parallel()
	srv := snapshotErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).ListContainerSnapshots(context.Background(), "pve1", 200)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestListContainerSnapshots_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListContainerSnapshots(context.Background(), "missing", 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateContainerSnapshot_success(t *testing.T) {
	t.Parallel()

	srv := snapshotServer(t, http.MethodPost, "/nodes/pve1/lxc/200/snapshot", ctSnapUPID)
	defer srv.Close()

	req := CreateContainerSnapshotRequest{Snapname: "ctsnap1", Description: "test"}
	upid, err := newTestClient(t, srv.URL).CreateContainerSnapshot(context.Background(), "pve1", 200, req)
	if err != nil {
		t.Fatalf("CreateContainerSnapshot: %v", err)
	}
	if upid != ctSnapUPID {
		t.Errorf("upid: got %q, want %q", upid, ctSnapUPID)
	}
}

func TestCreateContainerSnapshot_apiError(t *testing.T) {
	t.Parallel()
	srv := snapshotErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).CreateContainerSnapshot(context.Background(), "pve1", 200, CreateContainerSnapshotRequest{Snapname: "s"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestRollbackContainerSnapshot_success(t *testing.T) {
	t.Parallel()

	srv := snapshotServer(t, http.MethodPost, "/nodes/pve1/lxc/200/snapshot/ctsnap1/rollback", ctSnapUPID)
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).RollbackContainerSnapshot(context.Background(), "pve1", 200, "ctsnap1")
	if err != nil {
		t.Fatalf("RollbackContainerSnapshot: %v", err)
	}
	if upid != ctSnapUPID {
		t.Errorf("upid: got %q, want %q", upid, ctSnapUPID)
	}
}

func TestRollbackContainerSnapshot_apiError(t *testing.T) {
	t.Parallel()
	srv := snapshotErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).RollbackContainerSnapshot(context.Background(), "pve1", 200, "ctsnap1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestDeleteContainerSnapshot_success(t *testing.T) {
	t.Parallel()

	srv := snapshotServer(t, http.MethodDelete, "/nodes/pve1/lxc/200/snapshot/ctsnap1", ctSnapUPID)
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).DeleteContainerSnapshot(context.Background(), "pve1", 200, "ctsnap1")
	if err != nil {
		t.Fatalf("DeleteContainerSnapshot: %v", err)
	}
	if upid != ctSnapUPID {
		t.Errorf("upid: got %q, want %q", upid, ctSnapUPID)
	}
}

func TestDeleteContainerSnapshot_apiError(t *testing.T) {
	t.Parallel()
	srv := snapshotErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).DeleteContainerSnapshot(context.Background(), "pve1", 200, "ctsnap1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}
