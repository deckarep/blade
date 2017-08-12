package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version outputs the version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("verison is 0.0.1")
	},
}
