package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ks-arena",
	Short: "ks-arena: entorno de pruebas de Kubernetes con KIND, Terraform y ClusterLoader2",
}

// Execute ejecuta el comando raíz.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

