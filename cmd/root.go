package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
}

var rootCmd = &cobra.Command{
	Use:   "PPanel",
	Short: "PPanel is a modern multi-user agent panel.",
	Long: `[ PPanel is a pure, professional, and perfect open-source proxy panel tool, designed to be your ideal choice for learning and practical use.]
[ Simple and easy to operate.]`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("args:", args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
