/*
Copyright © 2023 linx sulinke1133@gmail.com
*/
package cmd

import (
	"github.com/linxlib/kapi/cmd/k/utils/daemon"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run the program of KAPI with hot reload",
	Long:  `this command will execute 'go run <entry file>' and restart it when file changed.`,
	Run: func(cmd *cobra.Command, args []string) {
		daemon.Run(args...)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
