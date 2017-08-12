package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(clearCmd)
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "cache does operations against the Blade database file.",
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear will destroy the cache, Blade will rebuild it on next run.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run: clearing the cache")
	},
}
