package cmd

import (
	"fmt"

	"github.com/perfect-panel/server/pkg/updater"
	"github.com/spf13/cobra"
)

var (
	checkOnly bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates and update PPanel to the latest version",
	Long: `Check for available updates from GitHub releases and automatically
update the PPanel binary to the latest version.

Examples:
  # Check for updates only
  ppanel-server update --check

  # Update to the latest version
  ppanel-server update`,
	Run: func(cmd *cobra.Command, args []string) {
		u := updater.NewUpdater()

		if checkOnly {
			checkForUpdates(u)
			return
		}

		performUpdate(u)
	},
}

func init() {
	updateCmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Check for updates without applying them")
}

func checkForUpdates(u *updater.Updater) {
	fmt.Println("Checking for updates...")

	release, hasUpdate, err := u.CheckForUpdates()
	if err != nil {
		fmt.Printf("Error checking for updates: %v\n", err)
		return
	}

	if !hasUpdate {
		fmt.Println("You are already running the latest version!")
		return
	}

	fmt.Printf("\nNew version available!\n")
	fmt.Printf("Current version: %s\n", u.CurrentVersion)
	fmt.Printf("Latest version:  %s\n", release.TagName)
	fmt.Printf("\nRelease notes:\n%s\n", release.Body)
	fmt.Printf("\nTo update, run: ppanel-server update\n")
}

func performUpdate(u *updater.Updater) {
	fmt.Println("Starting update process...")

	if err := u.Update(); err != nil {
		fmt.Printf("Update failed: %v\n", err)
		return
	}

	fmt.Println("\nUpdate completed successfully!")
	fmt.Println("Please restart the application to use the new version.")
}
