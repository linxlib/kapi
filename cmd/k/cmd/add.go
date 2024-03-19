/*
Copyright © 2023 linx sulinke1133@gmail.com
*/
package cmd

import (
	"fmt"
	template2 "github.com/linxlib/kapi/cmd/k/template"
	"github.com/linxlib/kapi/cmd/k/utils"
	"github.com/linxlib/logs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/user"
	"reflect"
	"strconv"
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

func ifNotEmpty(v string, f string) string {
	if v != "" {
		if strings.Contains(f, "%") {
			return fmt.Sprintf(f, v)
		}
		return f
	}
	return ""
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "add controller",
	Long:  "add controller and method",
	Example: `
k add controller -n Hello
k add controller -n Hello -m World
k add controller -n Hello -m World -b MyBody -r MyResult`,
	Run: func(cmd *cobra.Command, args []string) {
		controllerName := cmd.Flag("name").Value.String()
		methodName := cmd.Flag("method").Value.String()
		body := cmd.Flag("request").Value.String()
		result := cmd.Flag("response").Value.String()
		action := strings.ToUpper(cmd.Flag("action").Value.String())
		switch action {
		case "GET", "POST", "PUT", "DELETE", "PATCH":
		default:
			action = "GET"
			logs.Warn("flag action not provided use GET instead")
		}
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
				"Body":           ifNotEmpty(body, ", req *%s"),
				"RESP":           ifNotEmpty(result, "\n// @RESP %s"),
				"BodyStruct":     ifNotEmpty(body, "type %s struct {}\n"),
				"ResultStruct":   ifNotEmpty(result, "type %s struct {}\n\n"),
				"ACTION":         ifNotEmpty(action, action),
			})
			if err != nil {
				logs.Error(err)
				return
			}
			logs.Infof("method api/controllers/%sController.go => %s created!", controllerName, methodName)
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
				"BaseRoute":      strings.ToLower(controllerName),
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
				"Body":           ifNotEmpty(body, ", req *%s"),
				"RESP":           ifNotEmpty(result, "\n// @RESP %s"),
				"BodyStruct":     ifNotEmpty(body, "type %s struct {}\n"),
				"ResultStruct":   ifNotEmpty(result, "type %s struct {}\n\n"),
				"ACTION":         ifNotEmpty(action, action),
			})
			if err != nil {
				logs.Error(err)
				return
			}
			logs.Infof("api/controllers/%sController.go => %s created! \nadd kapi.RegisterRouter(new(controllers.%sController)) to main.go to enable it.", controllerName, methodName, controllerName)
		}

	},
}

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "add or set config",
	Long:    "add or set config",
	Example: `k add config <key> <value>`,
	Run: func(cmd *cobra.Command, args []string) {
		switch len(args) {
		case 1: // just return value of key
			logs.Infof("config[%s]=%+v", args[0], viper.Get(args[0]))
		case 2: // write config key's value
			i := viper.Get(args[0])
			rt := reflect.TypeOf(i)
			kind := rt.Kind()

			logs.Warnf("modify config[%s]=%+v", args[0], args[1])
			fmt.Print("input y to confirm(y/n):")
			var y string
			_, _ = fmt.Scan(&y)
			if strings.ToLower(y) != "y" || strings.ToLower(y) != "yes" {
				logs.Infoln("canceled~!")
				return
			}
			switch kind {
			case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
				v, _ := strconv.Atoi(args[1])
				viper.Set(args[0], v)
			case reflect.Float32, reflect.Float64:
				v, _ := strconv.ParseFloat(args[1], 64)
				viper.Set(args[0], v)
			case reflect.Bool:
				v, _ := strconv.ParseBool(args[1])
				viper.Set(args[0], v)
			default:
				viper.Set(args[0], args[1])
			}

			err := viper.WriteConfig()
			if err != nil {
				logs.Error(err)
				return
			}
		}

	},
}
var dockerCmd = &cobra.Command{
	Use:     "docker",
	Short:   "create something for docker",
	Long:    "create something for docker",
	Example: `k add docker`,
	//Run: func(cmd *cobra.Command, args []string) {
	//
	//},
}

var dockerFileCmd = &cobra.Command{
	Use:     "file",
	Short:   "create Dockerfile",
	Long:    "create Dockerfile",
	Example: `k add docker file`,
	Run: func(cmd *cobra.Command, args []string) {
		if utils.FileIsExist("Dockerfile") || utils.FileIsExist("dockerfile") {
			logs.Error("Dockerfile exists")
			return
		}
		fs, err := utils.T.ParseFS(template2.Files, "files/*")
		if err != nil {
			logs.Error(err)
			return
		}
		w, err := os.OpenFile("Dockerfile", os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			logs.Error(err)
			return
		}
		u, _ := user.Current()
		appName := utils.GetMod("go.mod")
		appNameIndex := strings.LastIndex(appName, "/")
		appName = appName[appNameIndex+1:]
		viper.SetDefault("server.port", 2023)
		err = fs.ExecuteTemplate(w, "Dockerfile.tmpl", map[string]any{
			"MAINTAINER": u.Name,
			"APPNAME":    appName,
			"PORT":       viper.Get("server.port"),
		})
		if err != nil {
			logs.Error(err)
		}
		logs.Info("Dockerfile created!")
	},
}

var dockerComposeCmd = &cobra.Command{
	Use:     "compose",
	Short:   "create docker-compose.yml",
	Long:    "create docker-compose.yml",
	Example: `k add docker compose`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 0 {
			logs.Error("specify a docker image url")
			return
		}
		if utils.FileIsExist("docker-compose.yml") || utils.FileIsExist("docker-compose.yaml") {
			logs.Error("docker-compose.yaml exists")
			return
		}
		fs, err := utils.T.ParseFS(template2.Files, "files/*")
		if err != nil {
			logs.Error(err)
			return
		}
		w, err := os.OpenFile("docker-compose.yaml", os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			logs.Error(err)
			return
		}
		appName := utils.GetMod("go.mod")
		appNameIndex := strings.LastIndex(appName, "/")
		appName = appName[appNameIndex+1:]
		viper.SetDefault("server.docVer", "latest")
		viper.SetDefault("server.docName", appName)
		viper.SetDefault("server.docDesc", appName)
		viper.SetDefault("server.basePath", "")
		viper.SetDefault("server.port", 2023)
		err = fs.ExecuteTemplate(w, "docker-compose.tmpl", map[string]any{
			"IMAGE":    args[0],
			"VERSION":  viper.GetString("server.docVer"),
			"DOCNAME":  viper.GetString("server.docName"),
			"DOCDESC":  viper.GetString("server.docDesc"),
			"BASEPATH": viper.GetString("server.basePath"),
			"APPNAME":  appName,
			"PORT":     viper.GetInt("server.port"),
		})
		if err != nil {
			logs.Error(err)
			return
		}
		logs.Info("docker-compose.yaml created!")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(controllerCmd)
	addCmd.AddCommand(configCmd)
	dockerCmd.AddCommand(dockerFileCmd)
	dockerCmd.AddCommand(dockerComposeCmd)
	addCmd.AddCommand(dockerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	controllerCmd.Flags().StringP("name", "n", "", "controller name. required ")
	controllerCmd.Flags().StringP("action", "a", "", "http method")
	controllerCmd.Flags().StringP("method", "m", "", "method name. if controller exist, this command will append method at the end of file")
	controllerCmd.Flags().StringP("request", "b", "", "request Body struct name")
	controllerCmd.Flags().StringP("response", "r", "", "response Body struct name")
}
