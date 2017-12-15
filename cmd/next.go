package cmd

import (
	"fmt"

	"github.com/joshuathompson/baton/api"
	"github.com/spf13/cobra"
)

func skipToNext(cmd *cobra.Command, args []string) {
	err := api.SkipToNext(&options)

	if err != nil {
		fmt.Printf("Failed to skip to next track\n")
	} else {
		fmt.Printf("Skipped to next track\n")
	}
}

func init() {
	rootCmd.AddCommand(skipToNextCmd)

	skipToNextCmd.Flags().StringVarP(&options.DeviceID, "device", "d", "", "id of the device this command is targeting")
}

var skipToNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Skip to next track",
	Long:  `Skip to next track`,
	Run:   skipToNext,
}
