/*
Copyright © 2023 linx sulinke1133@gmail.com
*/
package cmd

import (
	"github.com/gookit/color"

	"github.com/spf13/cobra"
)

const (
	version = "v0.1.8"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version of k cli",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		color.Println(version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
