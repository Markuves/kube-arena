//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunnerImage_HasRequiredTools(t *testing.T) {
	requireDocker(t)

	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot = filepath.Dir(repoRoot)

	image := os.Getenv("KS_ARENA_TEST_IMAGE")
	if image == "" {
		image = defaultImage
	}
	ensureImage(t, repoRoot, image)

	// Verifica que el runner tenga herramientas base instaladas.
	runCmd(t, 2*time.Minute, repoRoot, "docker", "run", "--rm", "--entrypoint", "bash", image, "-lc",
		"kind version && kubectl version --client=true && terraform version && (clusterloader2 --help >/dev/null 2>&1 || true)",
	)
}

