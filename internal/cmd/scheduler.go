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

var schedulerRunnerImage string

var schedulerCmd = &cobra.Command{
	Use:   "scheduler <scheduler-config.yaml>",
	Short: "Compila/empaca un scheduler personalizado y lo despliega en el cluster KIND",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		yamlPath := args[0]
		cfg, err := config.LoadSchedulerConfig(yamlPath)
		if err != nil {
			return fmt.Errorf("error cargando configuración de scheduler desde %s: %w", yamlPath, err)
		}

		absYaml, err := filepath.Abs(yamlPath)
		if err != nil {
			return fmt.Errorf("no se pudo resolver path absoluto del YAML: %w", err)
		}
		yamlDir := filepath.Dir(absYaml)

		// 1) Construir la imagen del scheduler a partir de schedulerDir (debe contener Dockerfile).
		schedDir := filepath.Join(yamlDir, cfg.SchedulerDir)
		if _, err := os.Stat(schedDir); err != nil {
			return fmt.Errorf("schedulerDir no existe: %w", err)
		}

		buildCmd := exec.Command("docker", "build", "-t", cfg.Image, ".")
		buildCmd.Dir = schedDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("docker build del scheduler falló: %w", err)
		}

		// 2) Aplicar el manifest en el cluster KIND usando el runner (la imagen del scheduler ya está en el daemon de Docker del host).
		runnerArgs := []string{
			"scheduler",
			"--cluster-name", cfg.ClusterName,
			"--manifest", cfg.ManifestPath,
		}
		if cfg.KubeconfigPath != "" {
			runnerArgs = append(runnerArgs, "--kubeconfig", cfg.KubeconfigPath)
		}

		return docker.Run(docker.RunOptions{
			Image:             schedulerRunnerImage,
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
	schedulerCmd.Flags().StringVar(&schedulerRunnerImage, "image", "ghcr.io/turing/kube-arena/ks-arena:latest", "Imagen runner a utilizar")
	rootCmd.AddCommand(schedulerCmd)
}

