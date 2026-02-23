package proxmox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const taskUPID = "UPID:pve1:000015E3:00000000:60F4B3A7:qmstart:100:root@pam:"

func TestGetTaskStatus_success(t *testing.T) {
	t.Parallel()

	want := TaskStatus{
		UPID:       taskUPID,
		Node:       "pve1",
		Status:     "stopped",
		ExitStatus: "OK",
		Type:       "qmstart",
		User:       "root@pam",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectPath := "/nodes/pve1/tasks/" + taskUPID + "/status"
		if r.URL.Path != expectPath {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetTaskStatus(context.Background(), "pve1", taskUPID)
	if err != nil {
		t.Fatalf("GetTaskStatus: %v", err)
	}
	if got.Status != "stopped" {
		t.Errorf("status: got %q, want stopped", got.Status)
	}
	if got.ExitStatus != "OK" {
		t.Errorf("exitstatus: got %q, want OK", got.ExitStatus)
	}
	if got.Node != "pve1" {
		t.Errorf("node: got %q, want pve1", got.Node)
	}
}

func TestGetTaskStatus_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetTaskStatus(context.Background(), "pve1", "nosuchupid")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
