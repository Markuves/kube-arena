package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/turing/kube-arena/internal/config"
	"github.com/turing/kube-arena/internal/docker"
)

var verbose bool
var testImage string

var testCmd = &cobra.Command{
	Use:   "test <test-config.yaml>",
	Short: "Orquesta el contenedor runner para ejecutar ClusterLoader2",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		yamlPath := args[0]
		cfg, err := config.LoadTestConfig(yamlPath)
		if err != nil {
			return fmt.Errorf("error cargando configuración de test desde %s: %w", yamlPath, err)
		}

		absYaml, err := filepath.Abs(yamlPath)
		if err != nil {
			return fmt.Errorf("no se pudo resolver path absoluto del YAML: %w", err)
		}
		yamlDir := filepath.Dir(absYaml)
		yamlBase := filepath.Base(absYaml)

		runnerArgs := []string{
			"test",
			"--cluster-name", cfg.ClusterName,
			"--config-yaml", yamlBase,
			"--test-config", cfg.TestConfigPath,
			"--repeat", fmt.Sprintf("%d", cfg.Repeat),
		}
		if cfg.KubeconfigPath != "" {
			runnerArgs = append(runnerArgs, "--kubeconfig", cfg.KubeconfigPath)
		}
		if cfg.TimeoutSeconds > 0 {
			runnerArgs = append(runnerArgs, "--timeout-seconds", fmt.Sprintf("%d", cfg.TimeoutSeconds))
		}
		if verbose {
			runnerArgs = append(runnerArgs, "--verbose")
		}

		return docker.Run(docker.RunOptions{
			Image:      testImage,
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
	testCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Muestra la salida detallada de ClusterLoader2")
	testCmd.Flags().StringVar(&testImage, "image", "ghcr.io/turing/kube-arena/ks-arena:latest", "Imagen runner a utilizar")
	rootCmd.AddCommand(testCmd)
}

