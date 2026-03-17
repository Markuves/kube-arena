package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/turing/kube-arena/internal/config"
	"github.com/turing/kube-arena/internal/docker"
)

var image string

var kindCmd = &cobra.Command{
	Use:   "kind <config.yaml>",
	Short: "Orquesta el contenedor runner para crear KIND + Terraform",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		yamlPath := args[0]
		cfg, err := config.LoadKindConfig(yamlPath)
		if err != nil {
			return fmt.Errorf("error cargando configuración KIND/Terraform desde %s: %w", yamlPath, err)
		}

		absYaml, err := filepath.Abs(yamlPath)
		if err != nil {
			return fmt.Errorf("no se pudo resolver path absoluto del YAML: %w", err)
		}
		yamlDir := filepath.Dir(absYaml)
		yamlBase := filepath.Base(absYaml)

		runnerArgs := []string{
			"kind",
			"--cluster-name", cfg.ClusterName,
			"--terraform-dir", cfg.TerraformDir,
			"--config-yaml", yamlBase,
		}
		if cfg.KindConfigPath != "" {
			runnerArgs = append(runnerArgs, "--kind-config", cfg.KindConfigPath)
		}
		for k, v := range cfg.Variables {
			runnerArgs = append(runnerArgs, "--var", fmt.Sprintf("%s=%s", k, v))
		}

		return docker.Run(docker.RunOptions{
			Image:      image,
			Workdir:    yamlDir,
			Privileged: true,
			MountDockerSocket: true,
			Args:       runnerArgs,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
		})
	},
}

func init() {
	kindCmd.Flags().StringVar(&image, "image", "ghcr.io/turing/kube-arena/ks-arena:latest", "Imagen runner a utilizar")
	rootCmd.AddCommand(kindCmd)
}

