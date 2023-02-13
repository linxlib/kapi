/*
Copyright © 2023 linx slk1133@qq.com
*/
package cmd

import (
	template2 "github.com/linxlib/kapi/cmd/k/template"
	"github.com/linxlib/kapi/cmd/k/utils"
	"github.com/linxlib/logs"
	"github.com/spf13/cobra"
	"os"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize KAPI project directory",
	Long:  `this command will create some directories on the current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logs.Error("you should specify a module name, only one parameter is supported. eg. k init github.com/linxlib/kapi/example")
			return
		}
		name := args[0]
		if utils.FileIsExist("go.mod") || utils.FileIsExist("go.sum") {
			logs.Errorf("file %s exist, run this command in an empty directory.", "go.mod/go.sum")
			return
		}
		err := utils.RunCommand("go", "mod", "init", name)
		if err != nil {
			return
		}
		err = utils.BuildDir("api/controllers/")
		if err != nil {
			logs.Error(err)
			return
		}
		fs, err := utils.T.ParseFS(template2.Files, "files/*")
		if err != nil {
			logs.Error(err)
			return
		}
		w, err := os.OpenFile("api/controllers/HealthController.go", os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			logs.Error(err)
			return
		}
		err = fs.ExecuteTemplate(w, "HealthController.tmpl", nil)
		if err != nil {
			logs.Error(err)
			return
		}
		w, err = os.OpenFile("main.go", os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			logs.Error(err)
			return
		}
		err = fs.ExecuteTemplate(w, "main.tmpl", map[string]any{
			"ModName": name,
		})
		if err != nil {
			logs.Error(err)
			return
		}
		utils.RunGoTidy()
		logs.Info("complete~!")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
