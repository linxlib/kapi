/*
Copyright © 2023 linx slk1133@qq.com
*/
package cmd

import (
	"fmt"
	template2 "github.com/linxlib/kapi/cmd/k/template"
	"github.com/linxlib/kapi/cmd/k/utils"
	"github.com/linxlib/logs"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add some codes into kapi project",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add called")
	},
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "add controller",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		controllerName := cmd.Flag("name").Value.String()
		methodName := cmd.Flag("method").Value.String()
		if controllerName == "" {
			logs.Error("please specify controller name with -n or --name=?")
			return
		}
		controllerFilePath := fmt.Sprintf("api/controllers/%sController.go", controllerName)
		if utils.FileIsExist(controllerFilePath) {
			if methodName == "" {
				logs.Error("please specify method name with -m or --method=?")
				return
			}
			w, err := os.OpenFile(controllerFilePath, os.O_APPEND|os.O_WRONLY, os.ModePerm)
			if err != nil {
				logs.Error(err)
				return
			}
			fs, err := utils.T.ParseFS(template2.Files, "files/*")
			if err != nil {
				logs.Error(err)
				return
			}
			err = fs.ExecuteTemplate(w, "Method.tmpl", map[string]any{
				"MethodName":     methodName,
				"ControllerName": controllerName,
				"ControllerID":   strings.ToLower(string(controllerName[0])),
				"ToLower": func(ori string) string {
					return strings.ToLower(ori)
				},
			})
			if err != nil {
				logs.Error(err)
				return
			}
		} else {
			fs, err := utils.T.ParseFS(template2.Files, "files/*")
			if err != nil {
				logs.Error(err)
				return
			}
			w, err := os.OpenFile(controllerFilePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
			if err != nil {
				logs.Error(err)
				return
			}
			err = fs.ExecuteTemplate(w, "Controller.tmpl", map[string]any{
				"ControllerName": controllerName,
			})
			if err != nil {
				logs.Error(err)
				return
			}

			if methodName == "" {
				logs.Infof("api/controllers/%sController.go created!", controllerName)
				return
			}

			err = fs.ExecuteTemplate(w, "Method.tmpl", map[string]any{
				"MethodName":     methodName,
				"ControllerName": controllerName,
				"ControllerID":   strings.ToLower(string(controllerName[0])),
			})
			if err != nil {
				logs.Error(err)
				return
			}
			logs.Infof("api/controllers/%sController.go:7 => %s created! \nadd kapi.RegisterRouter(new(controllers.%sController)) to main.go to enable it.", controllerName, methodName, controllerName)
		}

	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(controllerCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	controllerCmd.Flags().StringP("name", "n", "", "controller name. required ")
	controllerCmd.Flags().StringP("method", "m", "", "method name. if controller exist, this command will append method at the end of file")
}
