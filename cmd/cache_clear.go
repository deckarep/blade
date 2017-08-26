package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	cacheCmd.AddCommand(clearCmd)
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear will destroy the cache, Blade will rebuild it on next run.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run: clearing the cache")
	},
}
