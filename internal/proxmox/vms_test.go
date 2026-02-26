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

func TestListVMs_success(t *testing.T) {
	t.Parallel()

	want := []VM{
		{VMID: 100, Name: "web01", Status: "running"},
		{VMID: 101, Name: "db01", Status: "stopped"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/qemu" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListVMs(context.Background(), "pve1")
	if err != nil {
		t.Fatalf("ListVMs: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d VMs, want %d", len(got), len(want))
	}
	for i, vm := range got {
		if vm.VMID != want[i].VMID || vm.Name != want[i].Name || vm.Status != want[i].Status {
			t.Errorf("vm[%d]: got %+v, want %+v", i, vm, want[i])
		}
	}
}

func TestListVMs_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListVMs(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetVMStatus_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{"status": "running", "vmid": float64(100)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/qemu/100/status/current" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetVMStatus(context.Background(), "pve1", 100)
	if err != nil {
		t.Fatalf("GetVMStatus: %v", err)
	}
	if got["status"] != "running" {
		t.Errorf("status: got %v, want running", got["status"])
	}
}

const testUPID = "UPID:pve1:000015E3:00000000:60F4B3A7:qmstart:100:root@pam:"

func vmLifecycleServer(t *testing.T, expectPath string) *httptest.Server {
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
		_, _ = w.Write(jsonEnvelope(t, testUPID))
	}))
}

func TestStartVM_success(t *testing.T) {
	t.Parallel()

	srv := vmLifecycleServer(t, "/nodes/pve1/qemu/100/status/start")
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).StartVM(context.Background(), "pve1", 100)
	if err != nil {
		t.Fatalf("StartVM: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func TestStopVM_success(t *testing.T) {
	t.Parallel()

	srv := vmLifecycleServer(t, "/nodes/pve1/qemu/100/status/stop")
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).StopVM(context.Background(), "pve1", 100)
	if err != nil {
		t.Fatalf("StopVM: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func TestShutdownVM_success(t *testing.T) {
	t.Parallel()

	srv := vmLifecycleServer(t, "/nodes/pve1/qemu/100/status/shutdown")
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).ShutdownVM(context.Background(), "pve1", 100)
	if err != nil {
		t.Fatalf("ShutdownVM: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func vmErrorServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "VM is locked", http.StatusInternalServerError)
	}))
}

func TestGetVMStatus_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).GetVMStatus(context.Background(), "pve1", 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestStartVM_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).StartVM(context.Background(), "pve1", 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestStopVM_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).StopVM(context.Background(), "pve1", 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestShutdownVM_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).ShutdownVM(context.Background(), "pve1", 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestRebootVM_success(t *testing.T) {
	t.Parallel()

	srv := vmLifecycleServer(t, "/nodes/pve1/qemu/100/status/reboot")
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).RebootVM(context.Background(), "pve1", 100)
	if err != nil {
		t.Fatalf("RebootVM: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func TestRebootVM_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).RebootVM(context.Background(), "pve1", 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestSuspendVM_success(t *testing.T) {
	t.Parallel()

	srv := vmLifecycleServer(t, "/nodes/pve1/qemu/100/status/suspend")
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).SuspendVM(context.Background(), "pve1", 100)
	if err != nil {
		t.Fatalf("SuspendVM: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func TestSuspendVM_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).SuspendVM(context.Background(), "pve1", 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestResumeVM_success(t *testing.T) {
	t.Parallel()

	srv := vmLifecycleServer(t, "/nodes/pve1/qemu/100/status/resume")
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).ResumeVM(context.Background(), "pve1", 100)
	if err != nil {
		t.Fatalf("ResumeVM: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func TestResumeVM_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).ResumeVM(context.Background(), "pve1", 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestDeleteVM_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/qemu/100" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, testUPID))
	}))
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).DeleteVM(context.Background(), "pve1", 100, false)
	if err != nil {
		t.Fatalf("DeleteVM: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func TestDeleteVM_purge(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/qemu/100" || r.URL.Query().Get("purge") != "1" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, testUPID))
	}))
	defer srv.Close()

	upid, err := newTestClient(t, srv.URL).DeleteVM(context.Background(), "pve1", 100, true)
	if err != nil {
		t.Fatalf("DeleteVM purge: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func TestDeleteVM_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).DeleteVM(context.Background(), "pve1", 100, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestCreateVM_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/qemu" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, testUPID))
	}))
	defer srv.Close()

	req := CreateVMRequest{VMID: 200, Name: "test-vm", Memory: 512, Cores: 1}
	upid, err := newTestClient(t, srv.URL).CreateVM(context.Background(), "pve1", &req)
	if err != nil {
		t.Fatalf("CreateVM: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func TestCreateVM_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).CreateVM(context.Background(), "pve1", &CreateVMRequest{VMID: 200})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestCloneVM_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/qemu/100/clone" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, testUPID))
	}))
	defer srv.Close()

	req := CloneVMRequest{NewID: 201, Name: "cloned-vm"}
	upid, err := newTestClient(t, srv.URL).CloneVM(context.Background(), "pve1", 100, &req)
	if err != nil {
		t.Fatalf("CloneVM: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func TestCloneVM_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).CloneVM(context.Background(), "pve1", 100, &CloneVMRequest{NewID: 201})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestGetVMConfig_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{"cores": float64(4), "memory": float64(4096), "name": "downloads"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve3/qemu/108/config" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetVMConfig(context.Background(), "pve3", 108)
	if err != nil {
		t.Fatalf("GetVMConfig: %v", err)
	}
	if got["name"] != "downloads" {
		t.Errorf("name: got %v, want downloads", got["name"])
	}
}

func TestGetVMConfig_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetVMConfig(context.Background(), "pve1", 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSetVMConfig_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "want PUT", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/qemu/100/config" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	onboot := 1
	req := SetVMConfigRequest{Memory: 2048, Cores: 4, OnBoot: &onboot}
	if err := newTestClient(t, srv.URL).SetVMConfig(context.Background(), "pve1", 100, &req); err != nil {
		t.Fatalf("SetVMConfig: %v", err)
	}
}

func TestSetVMConfig_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	req := SetVMConfigRequest{Memory: 1024}
	err := newTestClient(t, srv.URL).SetVMConfig(context.Background(), "pve1", 100, &req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestSetVMConfig_omitempty(t *testing.T) {
	t.Parallel()

	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "want PUT", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/qemu/100/config" {
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

	// Only set Memory — name, cores, onboot, description must be absent from body.
	req := SetVMConfigRequest{Memory: 512}
	if err := newTestClient(t, srv.URL).SetVMConfig(context.Background(), "pve1", 100, &req); err != nil {
		t.Fatalf("SetVMConfig: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	for _, key := range []string{"name", "cores", "onboot", "description"} {
		if _, present := decoded[key]; present {
			t.Errorf("field %q should be omitted but was present in request body", key)
		}
	}
	if decoded["memory"] == nil {
		t.Error("field \"memory\" should be present but was absent")
	}
}

func TestResizeVMDisk_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "want PUT", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/qemu/100/resize" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, testUPID))
	}))
	defer srv.Close()

	req := ResizeDiskRequest{Disk: "scsi0", Size: "+10G"}
	upid, err := newTestClient(t, srv.URL).ResizeVMDisk(context.Background(), "pve1", 100, &req)
	if err != nil {
		t.Fatalf("ResizeVMDisk: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func TestResizeVMDisk_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	req := ResizeDiskRequest{Disk: "scsi0", Size: "+10G"}
	_, err := newTestClient(t, srv.URL).ResizeVMDisk(context.Background(), "pve1", 100, &req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

func TestMigrateVM_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/qemu/100/migrate" {
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
		if payload["online"] != float64(1) {
			http.Error(w, "online should be 1", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, testUPID))
	}))
	defer srv.Close()

	online := 1
	req := MigrateVMRequest{Target: "pve2", Online: &online}
	upid, err := newTestClient(t, srv.URL).MigrateVM(context.Background(), "pve1", 100, &req)
	if err != nil {
		t.Fatalf("MigrateVM: %v", err)
	}
	if upid != testUPID {
		t.Errorf("upid: got %q, want %q", upid, testUPID)
	}
}

func TestMigrateVM_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	_, err := newTestClient(t, srv.URL).MigrateVM(context.Background(), "pve1", 100, &MigrateVMRequest{Target: "pve2"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

const testRestoreVMUPID = "UPID:pve1:000015E3:00000000:60F4B3A7:qmrestore:100:root@pam:"

func TestRestoreVM_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/qemu" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if payload["archive"] != "local:backup/vzdump-qemu-100-2024_01_01-00_00_00.vma.zst" {
			http.Error(w, "wrong archive", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, testRestoreVMUPID))
	}))
	defer srv.Close()

	req := RestoreVMRequest{
		VMID:    100,
		Archive: "local:backup/vzdump-qemu-100-2024_01_01-00_00_00.vma.zst",
		Storage: "local-lvm",
	}
	upid, err := newTestClient(t, srv.URL).RestoreVM(context.Background(), "pve1", &req)
	if err != nil {
		t.Fatalf("RestoreVM: %v", err)
	}
	if upid != testRestoreVMUPID {
		t.Errorf("upid: got %q, want %q", upid, testRestoreVMUPID)
	}
}

func TestRestoreVM_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	req := RestoreVMRequest{VMID: 100, Archive: "local:backup/vzdump-qemu-100-2024_01_01-00_00_00.vma.zst"}
	_, err := newTestClient(t, srv.URL).RestoreVM(context.Background(), "pve1", &req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}
