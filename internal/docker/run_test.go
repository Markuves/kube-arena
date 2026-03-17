package docker

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRun_BuildsExpectedDockerArgs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("paths differ on windows")
	}

	var gotName string
	var gotArgs []string

	oldExec := execCommand
	t.Cleanup(func() { execCommand = oldExec })

	execCommand = func(name string, args ...string) *exec.Cmd {
		gotName = name
		gotArgs = append([]string{}, args...)
		return exec.Command("true")
	}

	var stdout, stderr bytes.Buffer
	workdir := t.TempDir()

	err := Run(RunOptions{
		Image:             "ks-arena:local",
		Workdir:           workdir,
		Args:              []string{"kind", "--cluster-name", "c1"},
		Privileged:        true,
		MountDockerSocket: true,
		Stdout:            &stdout,
		Stderr:            &stderr,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if gotName != "docker" {
		t.Fatalf("expected command docker, got %q", gotName)
	}

	abs, _ := filepath.Abs(workdir)
	wantPrefix := []string{
		"run", "--rm",
		"--privileged",
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", abs + ":/workspace",
		"-w", "/workspace",
		"ks-arena:local",
		"kind", "--cluster-name", "c1",
	}

	if strings.Join(gotArgs, "\n") != strings.Join(wantPrefix, "\n") {
		t.Fatalf("unexpected docker args:\nGOT : %#v\nWANT: %#v", gotArgs, wantPrefix)
	}
}

func TestRun_ValidatesRequiredFields(t *testing.T) {
	err := Run(RunOptions{
		Workdir: t.TempDir(),
	})
	if err == nil || !strings.Contains(err.Error(), "image es obligatorio") {
		t.Fatalf("expected image required error, got %v", err)
	}

	err = Run(RunOptions{
		Image: "x",
	})
	if err == nil || !strings.Contains(err.Error(), "workdir es obligatorio") {
		t.Fatalf("expected workdir required error, got %v", err)
	}
}

func TestRun_AbsWorkdirError(t *testing.T) {
	// filepath.Abs can fail if CWD is missing; simulate by chdir to a removed dir.
	oldWd, err := os.Getwd()
	if err != nil {
		t.Skip("cannot get wd")
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Skip("cannot chdir")
	}
	if err := os.RemoveAll(tmp); err != nil {
		t.Skip("cannot remove tmp dir")
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	err = Run(RunOptions{Image: "x", Workdir: "."})
	if err == nil {
		t.Fatalf("expected error")
	}
}

