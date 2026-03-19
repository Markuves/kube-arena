package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/turing/kube-arena/internal/config"
	"github.com/turing/kube-arena/internal/docker"
)

var image string

var kindCmd = &cobra.Command{
	Use:   "kind <config.yaml>",
	Short: "Orquesta el contenedor runner para crear KIND + cargar imágenes + aplicar manifests",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		yamlPath := args[0]
		cfg, err := config.LoadKindConfig(yamlPath)
		if err != nil {
			return fmt.Errorf("error cargando configuración KIND desde %s: %w", yamlPath, err)
		}

		absYaml, err := filepath.Abs(yamlPath)
		if err != nil {
			return fmt.Errorf("no se pudo resolver path absoluto del YAML: %w", err)
		}
		yamlDir := filepath.Dir(absYaml)
		yamlBase := filepath.Base(absYaml)

		// Build each image on the host Docker daemon before creating the cluster.
		// The runner will then load them into KIND via `kind load docker-image`.
		for _, img := range cfg.Images {
			absContext := filepath.Join(yamlDir, img.Context)
			absDf := filepath.Join(yamlDir, img.Dockerfile)
			fmt.Fprintf(os.Stdout, ">>> Building image %s ...\n", img.Name)
			buildCmd := exec.Command("docker", "build", "-t", img.Name, "-f", absDf, absContext)
			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr
			if err := buildCmd.Run(); err != nil {
				return fmt.Errorf("docker build para %s falló: %w", img.Name, err)
			}
		}

		runnerArgs := []string{
			"kind",
			"--cluster-name", cfg.ClusterName,
			"--config-yaml", yamlBase,
		}
		if cfg.KindConfigPath != "" {
			runnerArgs = append(runnerArgs, "--kind-config", cfg.KindConfigPath)
		}
		if cfg.TerraformDir != "" {
			runnerArgs = append(runnerArgs, "--terraform-dir", cfg.TerraformDir)
			for k, v := range cfg.Variables {
				runnerArgs = append(runnerArgs, "--var", fmt.Sprintf("%s=%s", k, v))
			}
		}
		for _, img := range cfg.Images {
			runnerArgs = append(runnerArgs, "--image", img.Name)
		}
		for _, manifest := range cfg.Manifests {
			runnerArgs = append(runnerArgs, "--manifest", manifest)
		}

		return docker.Run(docker.RunOptions{
			Image:             image,
			Workdir:           yamlDir,
			Privileged:        true,
			MountDockerSocket: true,
			HostNetwork:       true,
			Args:              runnerArgs,
			Stdout:            os.Stdout,
			Stderr:            os.Stderr,
		})
	},
}

func init() {
	kindCmd.Flags().StringVar(&image, "image", "ghcr.io/turing/kube-arena/ks-arena:latest", "Imagen runner a utilizar")
	rootCmd.AddCommand(kindCmd)
}
