package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempYAML(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(p, []byte(contents), 0o600); err != nil {
		t.Fatalf("write temp yaml: %v", err)
	}
	return p
}

func TestLoadKindConfig_RequiresFields(t *testing.T) {
	p := writeTempYAML(t, `
clusterName: ""
terraformDir: ""
`)
	_, err := LoadKindConfig(p)
	if err == nil || !strings.Contains(err.Error(), "clusterName es obligatorio") {
		t.Fatalf("expected clusterName required error, got %v", err)
	}

	p = writeTempYAML(t, `
clusterName: c1
terraformDir: ""
`)
	_, err = LoadKindConfig(p)
	if err == nil || !strings.Contains(err.Error(), "terraformDir es obligatorio") {
		t.Fatalf("expected terraformDir required error, got %v", err)
	}
}

func TestLoadKindConfig_LoadsOptionalFields(t *testing.T) {
	p := writeTempYAML(t, `
kindConfigPath: kind-config.yaml
clusterName: c1
terraformDir: terraform
variables:
  a: "1"
  b: "two"
`)
	cfg, err := LoadKindConfig(p)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg.KindConfigPath != "kind-config.yaml" {
		t.Fatalf("kindConfigPath mismatch: %q", cfg.KindConfigPath)
	}
	if cfg.Variables["a"] != "1" || cfg.Variables["b"] != "two" {
		t.Fatalf("variables mismatch: %#v", cfg.Variables)
	}
}

func TestLoadTestConfig_DefaultRepeat(t *testing.T) {
	p := writeTempYAML(t, `
clusterName: c1
testConfigPath: clusterloader2.yaml
repeat: 0
`)
	cfg, err := LoadTestConfig(p)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg.Repeat != 1 {
		t.Fatalf("expected repeat default 1, got %d", cfg.Repeat)
	}
}

func TestLoadTestConfig_RequiresFields(t *testing.T) {
	p := writeTempYAML(t, `
clusterName: ""
testConfigPath: ""
`)
	_, err := LoadTestConfig(p)
	if err == nil || !strings.Contains(err.Error(), "testConfigPath es obligatorio") {
		t.Fatalf("expected testConfigPath required error, got %v", err)
	}
}

func TestLoadSchedulerConfig_RequiresFields(t *testing.T) {
	p := writeTempYAML(t, `
clusterName: ""
schedulerDir: ""
image: ""
manifestPath: ""
`)
	_, err := LoadSchedulerConfig(p)
	if err == nil || !strings.Contains(err.Error(), "clusterName es obligatorio") {
		t.Fatalf("expected clusterName required error, got %v", err)
	}

	p = writeTempYAML(t, `
clusterName: c1
schedulerDir: ""
image: ""
manifestPath: ""
`)
	_, err = LoadSchedulerConfig(p)
	if err == nil || !strings.Contains(err.Error(), "schedulerDir es obligatorio") {
		t.Fatalf("expected schedulerDir required error, got %v", err)
	}

	p = writeTempYAML(t, `
clusterName: c1
schedulerDir: dir
image: ""
manifestPath: ""
`)
	_, err = LoadSchedulerConfig(p)
	if err == nil || !strings.Contains(err.Error(), "image es obligatorio") {
		t.Fatalf("expected image required error, got %v", err)
	}

	p = writeTempYAML(t, `
clusterName: c1
schedulerDir: dir
image: img
manifestPath: ""
`)
	_, err = LoadSchedulerConfig(p)
	if err == nil || !strings.Contains(err.Error(), "manifestPath es obligatorio") {
		t.Fatalf("expected manifestPath required error, got %v", err)
	}
}


