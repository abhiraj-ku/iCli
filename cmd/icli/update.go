package main

import (
	"fmt"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update iCli to the latest version",
	Long:  "Checks the GitHub release repository for updates and updates the iCli binary in place.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Checking for updates... (Current version: %s)\n", version)

		// init the updater to point to release repo
		updater, err := selfupdate.NewUpdater(selfupdate.Config{})
		if err != nil {
			return fmt.Errorf("failed to create updater: %w", err)
		}

		latest, err := updater.UpdateSelf(cmd.Context(), version, selfupdate.NewRepositorySlug("abhiraj-ku", "iCli"))
		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		if latest.Version() == version {
			fmt.Println("✓ You are already on the latest version.")
		} else {
			fmt.Printf("Successfully updated to %s!\n", latest.Version())
			fmt.Println("Release Notes:\n", latest.ReleaseNotes)
		}

		return nil
	},
}
