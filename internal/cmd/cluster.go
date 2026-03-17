package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/turing/kube-arena/internal/config"
	"github.com/turing/kube-arena/internal/docker"
)

var clusterImage string

var downCmd = &cobra.Command{
	Use:   "down <config.yaml>",
	Short: "Elimina el cluster KIND definido en el YAML",
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

		runnerArgs := []string{
			"down",
			"--cluster-name", cfg.ClusterName,
		}

		return docker.Run(docker.RunOptions{
			Image:             clusterImage,
			Workdir:           yamlDir,
			Privileged:        true,
			MountDockerSocket: true,
			Args:              runnerArgs,
			Stdout:            os.Stdout,
			Stderr:            os.Stderr,
		})
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset <config.yaml>",
	Short: "Recrea el cluster KIND (down + kind)",
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

		// Primero down.
		if err := docker.Run(docker.RunOptions{
			Image:             clusterImage,
			Workdir:           yamlDir,
			Privileged:        true,
			MountDockerSocket: true,
			Args: []string{
				"down",
				"--cluster-name", cfg.ClusterName,
			},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "warning: fallo al eliminar cluster (continuando con recreate): %v\n", err)
		}

		// Luego crear de nuevo (equivalente a kind <yaml>).
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
			Image:             clusterImage,
			Workdir:           yamlDir,
			Privileged:        true,
			MountDockerSocket: true,
			Args:              runnerArgs,
			Stdout:            os.Stdout,
			Stderr:            os.Stderr,
		})
	},
}

func init() {
	downCmd.Flags().StringVar(&clusterImage, "image", "ghcr.io/turing/kube-arena/ks-arena:latest", "Imagen runner a utilizar")
	resetCmd.Flags().StringVar(&clusterImage, "image", "ghcr.io/turing/kube-arena/ks-arena:latest", "Imagen runner a utilizar")
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(resetCmd)
}

