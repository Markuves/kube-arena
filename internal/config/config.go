package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ImageSpec describes a Docker image to build and load into the KIND cluster.
type ImageSpec struct {
	Name       string `yaml:"name"`       // image tag, e.g. local/k8s-monitor:latest
	Dockerfile string `yaml:"dockerfile"` // path relative to the config YAML file
	Context    string `yaml:"context"`    // build context path relative to the config YAML file
}

// KindConfig representa el esquema mínimo para example.yaml (infraestructura KIND + Terraform).
type KindConfig struct {
	KindConfigPath string            `yaml:"kindConfigPath"`
	ClusterName    string            `yaml:"clusterName"`
	TerraformDir   string            `yaml:"terraformDir,omitempty"`
	Variables      map[string]string `yaml:"variables,omitempty"`
	Images         []ImageSpec       `yaml:"images,omitempty"`
	Manifests      []string          `yaml:"manifests,omitempty"`
}

// TestConfig representa el esquema mínimo para test-example.yaml (pruebas ClusterLoader2).
type TestConfig struct {
	ClusterName    string `yaml:"clusterName"`
	KubeconfigPath string `yaml:"kubeconfigPath,omitempty"`
	TestConfigPath string `yaml:"testConfigPath"`
	Repeat         int    `yaml:"repeat,omitempty"`
	TimeoutSeconds int    `yaml:"timeoutSeconds,omitempty"`
}

// SchedulerConfig representa la configuración para desplegar un scheduler personalizado.
type SchedulerConfig struct {
	ClusterName    string `yaml:"clusterName"`
	KubeconfigPath string `yaml:"kubeconfigPath,omitempty"`
	// SchedulerDir es la carpeta que contiene el código/Dockerfile del scheduler.
	SchedulerDir string `yaml:"schedulerDir"`
	// Image es el nombre de la imagen que se construirá para el scheduler.
	Image string `yaml:"image"`
	// ManifestPath es la ruta (relativa al YAML) al manifest de Kubernetes que usa esa imagen.
	ManifestPath string `yaml:"manifestPath"`
}

// LoadKindConfig carga y valida la configuración KIND/Terraform desde un YAML.
func LoadKindConfig(path string) (*KindConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo: %w", err)
	}

	var cfg KindConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("no se pudo parsear YAML: %w", err)
	}

	if cfg.ClusterName == "" {
		return nil, fmt.Errorf("clusterName es obligatorio")
	}

	return &cfg, nil
}

// LoadTestConfig carga y valida la configuración de pruebas desde un YAML.
func LoadTestConfig(path string) (*TestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo: %w", err)
	}

	var cfg TestConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("no se pudo parsear YAML: %w", err)
	}

	if cfg.TestConfigPath == "" {
		return nil, fmt.Errorf("testConfigPath es obligatorio")
	}
	if cfg.ClusterName == "" {
		return nil, fmt.Errorf("clusterName es obligatorio")
	}

	if cfg.Repeat <= 0 {
		cfg.Repeat = 1
	}

	return &cfg, nil
}

// LoadSchedulerConfig carga la configuración del scheduler personalizado.
func LoadSchedulerConfig(path string) (*SchedulerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo: %w", err)
	}

	var cfg SchedulerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("no se pudo parsear YAML: %w", err)
	}

	if cfg.ClusterName == "" {
		return nil, fmt.Errorf("clusterName es obligatorio")
	}
	if cfg.SchedulerDir == "" {
		return nil, fmt.Errorf("schedulerDir es obligatorio")
	}
	if cfg.Image == "" {
		return nil, fmt.Errorf("image es obligatorio")
	}
	if cfg.ManifestPath == "" {
		return nil, fmt.Errorf("manifestPath es obligatorio")
	}

	return &cfg, nil
}

