//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const defaultImage = "ks-arena:it"

func requireDocker(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker no está disponible en el PATH")
	}
}

func runCmd(t *testing.T, timeout time.Duration, dir, name string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("timeout ejecutando %s %s\nOUT:\n%s", name, strings.Join(args, " "), string(out))
	}
	if err != nil {
		t.Fatalf("falló %s %s: %v\nOUT:\n%s", name, strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

func runCmdAllowFail(t *testing.T, timeout time.Duration, dir, name string, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), fmt.Errorf("timeout ejecutando %s %s", name, strings.Join(args, " "))
	}
	return string(out), err
}

func ensureImage(t *testing.T, repoRoot, image string) {
	t.Helper()
	// Si no existe localmente, la construimos.
	cmd := exec.Command("docker", "image", "inspect", image)
	if err := cmd.Run(); err == nil {
		return
	}
	runCmd(t, 30*time.Minute, repoRoot, "docker", "build", "-t", image, ".")
}

func buildCLI(t *testing.T, repoRoot string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "ks-arena")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	runCmd(t, 2*time.Minute, repoRoot, "go", "build", "-o", bin, "./cmd/ks-arena")
	return bin
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func TestKsArena_KindAndTest_FromYAML(t *testing.T) {
	requireDocker(t)

	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// Este test corre desde integration/, subimos a la raíz del repo.
	repoRoot = filepath.Dir(repoRoot)

	image := os.Getenv("KS_ARENA_TEST_IMAGE")
	if image == "" {
		image = defaultImage
	}

	ensureImage(t, repoRoot, image)
	cli := buildCLI(t, repoRoot)

	workspace := t.TempDir()
	clusterName := fmt.Sprintf("ks-arena-it-%d", time.Now().UnixNano())

	// Terraform mínimo (no-op).
	writeFile(t, filepath.Join(workspace, "terraform", "main.tf"), `
terraform {
  required_version = ">= 1.3.0"
}
`)

	// YAML de kind (infra).
	writeFile(t, filepath.Join(workspace, "example.yaml"), fmt.Sprintf(`
clusterName: %s
terraformDir: terraform
variables:
  example: "1"
`, clusterName))

	// Config mínimo de ClusterLoader2 (smoke).
	writeFile(t, filepath.Join(workspace, "cl2-smoke.yaml"), `
name: smoke
namespace:
  number: 1
steps:
- name: WaitShort
  measurements:
  - Identifier: Wait
    Method: Sleep
    Params:
      duration: 1s
`)

	// YAML de test.
	writeFile(t, filepath.Join(workspace, "test-example.yaml"), fmt.Sprintf(`
clusterName: %s
testConfigPath: cl2-smoke.yaml
repeat: 1
timeoutSeconds: 300
`, clusterName))

	// Ejecuta kind + terraform dentro del runner.
	out, err := runCmdAllowFail(t, 20*time.Minute, workspace, cli, "kind", "--image", image, "example.yaml")
	if err != nil {
		// En algunos hosts (cgroups/rootless/entornos restringidos) KIND puede fallar.
		// En esos casos, este test no puede validar el flujo end-to-end.
		if strings.Contains(out, "kubelet is not healthy") || strings.Contains(out, "required cgroups disabled") {
			t.Skipf("KIND no pudo inicializar el cluster en este host/entorno. Salida:\n%s", out)
		}
		t.Fatalf("ks-arena kind falló: %v\nOUT:\n%s", err, out)
	}

	// Ejecuta clusterloader2 con test config dentro del runner.
	runCmd(t, 10*time.Minute, workspace, cli, "test", "--image", image, "test-example.yaml")

	// Cleanup del cluster KIND usando las herramientas dentro de la imagen runner.
	// Esto evita requerir kind en el host.
	runCmd(t, 5*time.Minute, workspace, "docker",
		"run", "--rm", "--privileged",
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		image,
		"bash", "-lc", fmt.Sprintf("kind delete cluster --name %q", clusterName),
	)
}

