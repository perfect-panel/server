package cmd

import (
	"fmt"

	"github.com/perfect-panel/ppanel-server/internal/config"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "PPanel version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[PPanel version] " + config.Version)
	},
}
