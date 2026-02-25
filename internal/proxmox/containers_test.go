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

func TestListContainers_success(t *testing.T) {
	t.Parallel()

	want := []Container{
		{VMID: 200, Name: "ct-web", Status: "running"},
		{VMID: 201, Name: "ct-db", Status: "stopped"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/lxc" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListContainers(context.Background(), "pve1")
	if err != nil {
		t.Fatalf("ListContainers: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d containers, want %d", len(got), len(want))
	}
	for i, ct := range got {
		if ct.VMID != want[i].VMID || ct.Name != want[i].Name || ct.Status != want[i].Status {
			t.Errorf("container[%d]: got %+v, want %+v", i, ct, want[i])
		}
	}
}

func TestListContainers_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListContainers(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetContainerStatus_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{"status": "running", "vmid": float64(200)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/lxc/200/status/current" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetContainerStatus(context.Background(), "pve1", 200)
	if err != nil {
		t.Fatalf("GetContainerStatus: %v", err)
	}
	if got["status"] != "running" {
		t.Errorf("status: got %v, want running", got["status"])
	}
}

const ctUPID = "UPID:pve1:000015E3:00000000:60F4B3A7:vzstart:200:root@pam:"

func ctLifecycleServer(t *testing.T, expectPath string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != expectPath {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, ctUPID))
	}))
}

func TestStartContainer_success(t *testing.T) {
	t.Parallel()

	srv := ctLifecycleServer(t, "/nodes/pve1/lxc/200/status/start")
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).StartContainer(context.Background(), "pve1", 200)
	if err != nil {
		t.Fatalf("StartContainer: %v", err)
	}
	if upid != ctUPID {
		t.Errorf("upid: got %q, want %q", upid, ctUPID)
	}
}

func TestStopContainer_success(t *testing.T) {
	t.Parallel()

	srv := ctLifecycleServer(t, "/nodes/pve1/lxc/200/status/stop")
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).StopContainer(context.Background(), "pve1", 200)
	if err != nil {
		t.Fatalf("StopContainer: %v", err)
	}
	if upid != ctUPID {
		t.Errorf("upid: got %q, want %q", upid, ctUPID)
	}
}

func ctErrorServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "CT is locked", http.StatusInternalServerError)
	}))
}

func TestGetContainerStatus_apiError(t *testing.T) {
	t.Parallel()
	srv := ctErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).GetContainerStatus(context.Background(), "pve1", 200)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestStartContainer_apiError(t *testing.T) {
	t.Parallel()
	srv := ctErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).StartContainer(context.Background(), "pve1", 200)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestStopContainer_apiError(t *testing.T) {
	t.Parallel()
	srv := ctErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).StopContainer(context.Background(), "pve1", 200)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestShutdownContainer_success(t *testing.T) {
	t.Parallel()

	srv := ctLifecycleServer(t, "/nodes/pve1/lxc/200/status/shutdown")
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).ShutdownContainer(context.Background(), "pve1", 200)
	if err != nil {
		t.Fatalf("ShutdownContainer: %v", err)
	}
	if upid != ctUPID {
		t.Errorf("upid: got %q, want %q", upid, ctUPID)
	}
}

func TestShutdownContainer_apiError(t *testing.T) {
	t.Parallel()
	srv := ctErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).ShutdownContainer(context.Background(), "pve1", 200)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestRebootContainer_success(t *testing.T) {
	t.Parallel()

	srv := ctLifecycleServer(t, "/nodes/pve1/lxc/200/status/reboot")
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).RebootContainer(context.Background(), "pve1", 200)
	if err != nil {
		t.Fatalf("RebootContainer: %v", err)
	}
	if upid != ctUPID {
		t.Errorf("upid: got %q, want %q", upid, ctUPID)
	}
}

func TestRebootContainer_apiError(t *testing.T) {
	t.Parallel()
	srv := ctErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).RebootContainer(context.Background(), "pve1", 200)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestDeleteContainer_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/lxc/200" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, ctUPID))
	}))
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).DeleteContainer(context.Background(), "pve1", 200, false)
	if err != nil {
		t.Fatalf("DeleteContainer: %v", err)
	}
	if upid != ctUPID {
		t.Errorf("upid: got %q, want %q", upid, ctUPID)
	}
}

func TestDeleteContainer_purge(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/lxc/200" || r.URL.Query().Get("purge") != "1" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, ctUPID))
	}))
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).DeleteContainer(context.Background(), "pve1", 200, true)
	if err != nil {
		t.Fatalf("DeleteContainer purge: %v", err)
	}
	if upid != ctUPID {
		t.Errorf("upid: got %q, want %q", upid, ctUPID)
	}
}

func TestDeleteContainer_apiError(t *testing.T) {
	t.Parallel()
	srv := ctErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).DeleteContainer(context.Background(), "pve1", 200, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestCreateContainer_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/lxc" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, ctUPID))
	}))
	defer srv.Close()

	req := CreateContainerRequest{
		VMID:       300,
		OSTemplate: "local:vztmpl/debian-12-standard_12.7-1_amd64.tar.zst",
		Hostname:   "test-ct",
		Memory:     512,
		RootFS:     "local-lvm:8",
	}
	upid, err := newTestClient(t, srv.URL).CreateContainer(context.Background(), "pve1", &req)
	if err != nil {
		t.Fatalf("CreateContainer: %v", err)
	}
	if upid != ctUPID {
		t.Errorf("upid: got %q, want %q", upid, ctUPID)
	}
}

func TestCreateContainer_apiError(t *testing.T) {
	t.Parallel()
	srv := ctErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).CreateContainer(context.Background(), "pve1", &CreateContainerRequest{
		VMID:       300,
		OSTemplate: "local:vztmpl/debian-12-standard_12.7-1_amd64.tar.zst",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestCloneContainer_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/lxc/200/clone" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, ctUPID))
	}))
	defer srv.Close()

	req := CloneContainerRequest{NewID: 301, Hostname: "cloned-ct"}
	upid, err := newTestClient(t, srv.URL).CloneContainer(context.Background(), "pve1", 200, &req)
	if err != nil {
		t.Fatalf("CloneContainer: %v", err)
	}
	if upid != ctUPID {
		t.Errorf("upid: got %q, want %q", upid, ctUPID)
	}
}

func TestCloneContainer_apiError(t *testing.T) {
	t.Parallel()
	srv := ctErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).CloneContainer(context.Background(), "pve1", 200, &CloneContainerRequest{NewID: 301})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestGetContainerConfig_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{"hostname": "influxdb", "memory": float64(2048), "cores": float64(2)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/lxc/100/config" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetContainerConfig(context.Background(), "pve1", 100)
	if err != nil {
		t.Fatalf("GetContainerConfig: %v", err)
	}
	if got["hostname"] != "influxdb" {
		t.Errorf("hostname: got %v, want influxdb", got["hostname"])
	}
}

func TestGetContainerConfig_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetContainerConfig(context.Background(), "pve1", 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSetContainerConfig_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "want PUT", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/lxc/200/config" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	onboot := 1
	swap := 512
	req := SetContainerConfigRequest{Memory: 1024, Swap: &swap, OnBoot: &onboot}
	if err := newTestClient(t, srv.URL).SetContainerConfig(context.Background(), "pve1", 200, &req); err != nil {
		t.Fatalf("SetContainerConfig: %v", err)
	}
}

func TestSetContainerConfig_apiError(t *testing.T) {
	t.Parallel()
	srv := ctErrorServer(t)
	defer srv.Close()
	req := SetContainerConfigRequest{Memory: 512}
	err := newTestClient(t, srv.URL).SetContainerConfig(context.Background(), "pve1", 200, &req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestSetContainerConfig_omitempty(t *testing.T) {
	t.Parallel()

	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "want PUT", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/lxc/200/config" {
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

	// Only set Memory — hostname, swap, onboot, description must be absent from body.
	req := SetContainerConfigRequest{Memory: 512}
	if err := newTestClient(t, srv.URL).SetContainerConfig(context.Background(), "pve1", 200, &req); err != nil {
		t.Fatalf("SetContainerConfig: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	for _, key := range []string{"hostname", "swap", "onboot", "description"} {
		if _, present := decoded[key]; present {
			t.Errorf("field %q should be omitted but was present in request body", key)
		}
	}
	if decoded["memory"] == nil {
		t.Error("field \"memory\" should be present but was absent")
	}
}

func TestResizeContainerDisk_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "want PUT", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/lxc/200/resize" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, ctUPID))
	}))
	defer srv.Close()

	req := ResizeDiskRequest{Disk: "rootfs", Size: "+5G"}
	upid, err := newTestClient(t, srv.URL).ResizeContainerDisk(context.Background(), "pve1", 200, &req)
	if err != nil {
		t.Fatalf("ResizeContainerDisk: %v", err)
	}
	if upid != ctUPID {
		t.Errorf("upid: got %q, want %q", upid, ctUPID)
	}
}

func TestResizeContainerDisk_apiError(t *testing.T) {
	t.Parallel()
	srv := ctErrorServer(t)
	defer srv.Close()
	req := ResizeDiskRequest{Disk: "rootfs", Size: "+5G"}
	_, err := newTestClient(t, srv.URL).ResizeContainerDisk(context.Background(), "pve1", 200, &req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestMigrateContainer_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/lxc/200/migrate" {
			http.NotFound(w, r)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusInternalServerError)
			return
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if payload["target"] != "pve2" {
			http.Error(w, "wrong target", http.StatusBadRequest)
			return
		}
		if payload["restart"] != float64(1) {
			http.Error(w, "restart should be 1", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, ctUPID))
	}))
	defer srv.Close()

	restart := 1
	req := MigrateContainerRequest{Target: "pve2", Restart: &restart}
	upid, err := newTestClient(t, srv.URL).MigrateContainer(context.Background(), "pve1", 200, &req)
	if err != nil {
		t.Fatalf("MigrateContainer: %v", err)
	}
	if upid != ctUPID {
		t.Errorf("upid: got %q, want %q", upid, ctUPID)
	}
}

func TestMigrateContainer_apiError(t *testing.T) {
	t.Parallel()
	srv := ctErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).MigrateContainer(context.Background(), "pve1", 200, &MigrateContainerRequest{Target: "pve2"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}
