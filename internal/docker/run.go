package docker

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
)

var execCommand = exec.Command

type RunOptions struct {
	Image             string
	Workdir           string // host directory to mount at /workspace
	Args              []string
	Privileged        bool
	MountDockerSocket bool
	// HostNetwork runs the container in the host network namespace so it can
	// reach KIND API servers bound to 127.0.0.1 on the host.
	HostNetwork bool
	Stdout      io.Writer
	Stderr      io.Writer
}

func Run(opts RunOptions) error {
	if opts.Image == "" {
		return fmt.Errorf("image es obligatorio")
	}
	if opts.Workdir == "" {
		return fmt.Errorf("workdir es obligatorio")
	}
	absWorkdir, err := filepath.Abs(opts.Workdir)
	if err != nil {
		return fmt.Errorf("no se pudo resolver workdir absoluto: %w", err)
	}

	dockerArgs := []string{"run", "--rm"}
	if opts.Privileged {
		dockerArgs = append(dockerArgs, "--privileged")
	}
	if opts.MountDockerSocket {
		dockerArgs = append(dockerArgs, "-v", "/var/run/docker.sock:/var/run/docker.sock")
	}
	if opts.HostNetwork {
		dockerArgs = append(dockerArgs, "--network=host")
	}

	dockerArgs = append(dockerArgs,
		"-v", fmt.Sprintf("%s:/workspace", absWorkdir),
		"-w", "/workspace",
		opts.Image,
	)
	dockerArgs = append(dockerArgs, opts.Args...)

	cmd := execCommand("docker", dockerArgs...)
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker run falló: %w", err)
	}
	return nil
}

