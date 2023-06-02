/*
Copyright © 2023 Yuzuki Mimura

*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "mcontainer",
	Short:   "micro container runtime",
	Long:    "mcontainer is a simple container runtime implementation",
	RunE:    rootMain,
	Version: "0.0.1",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func rootMain(cmd *cobra.Command, args []string) error {
	return nil
}
