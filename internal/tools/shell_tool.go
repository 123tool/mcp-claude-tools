package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

type ProcessInfo struct {
	Cmd       *exec.Cmd
	OutBuffer *bytes.Buffer
	ErrBuffer *bytes.Buffer
	StartTime time.Time
}

type ShellManager struct {
	mu        sync.RWMutex
	processes map[string]*ProcessInfo
}

func NewShellManager() *ShellManager {
	return &ShellManager{
		processes: make(map[string]*ProcessInfo),
	}
}

// RunBash mengeksekusi perintah dengan batas waktu atau di latar belakang
func (sm *ShellManager) RunBash(command string, timeoutSec int, runInBg bool, bgID string) (string, error) {
	ctx := context.Background()
	var cancel context.CancelFunc

	if !runInBg && timeoutSec > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	var stdout, stderr bytes.Buffer

	if runInBg {
		if bgID == "" {
			return "", fmt.Errorf("bg_id wajib diisi jika berjalan di latar belakang (run_in_bg=true)")
		}
		
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		
		err := cmd.Start()
		if err != nil {
			return "", err
		}

		sm.mu.Lock()
		sm.processes[bgID] = &ProcessInfo{
			Cmd:       cmd,
			OutBuffer: &stdout,
			ErrBuffer: &stderr,
			StartTime: time.Now(),
		}
		sm.mu.Unlock()

		return fmt.Sprintf("Proses latar belakang berhasil dimulai dengan ID: %s", bgID), nil
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	output := stdout.String() + stderr.String()
	if err != nil {
		return output, fmt.Errorf("perintah keluar dengan error: %v", err)
	}

	return output, nil
}

// GetBashOutput mengambil output dari proses latar belakang yang sedang/sudah berjalan
func (sm *ShellManager) GetBashOutput(bgID string) (string, string, bool) {
	sm.mu.RLock()
	proc, exists := sm.processes[bgID]
	sm.mu.RUnlock()

	if !exists {
		return "", "Proses tidak ditemukan", false
	}

	running := proc.Cmd.ProcessState == nil || !proc.Cmd.ProcessState.Exited()
	return proc.OutBuffer.String(), proc.ErrBuffer.String(), running
}

// KillShell menghentikan proses latar belakang secara paksa
func (sm *ShellManager) KillShell(bgID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	proc, exists := sm.processes[bgID]
	if !exists {
		return fmt.Errorf("proses dengan ID %s tidak ditemukan", bgID)
	}

	if proc.Cmd.Process != nil {
		err := proc.Cmd.Process.Kill()
		if err != nil {
			return err
		}
	}

	delete(sm.processes, bgID)
	return nil
}
