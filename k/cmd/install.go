package cmd

import (
	"github.com/gogf/gf/container/garray"
	"github.com/gogf/gf/container/gset"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"runtime"
	"strings"
)

// installFolderPath contains installFolderPath-related data,
// such as path, writable, binaryFilePath, and installed.
type installFolderPath struct {
	path           string
	writable       bool
	binaryFilePath string
	installed      bool
}

func Install() {
	// Ask where to install.
	paths := getInstallPathsData()
	if len(paths) <= 0 {
		_log.Printf("未找到可安装路径, 你可以手动复制 k 到环境变量PATH中的路径.")
		return
	}
	_log.Printf("找到多个可安装位置(来自 $PATH): ")
	_log.Printf("  %2s | %8s | %9s | %s", "Id", "可写入", "已安装", "路径")

	// Print all paths status and determine the default selectedID value.
	var (
		selectedID = -1
		pathSet    = gset.NewStrSet() // Used for repeated items filtering.
	)
	for id, aPath := range paths {
		if !pathSet.AddIfNotExist(aPath.path) {
			continue
		}
		_log.Printf("  %2d | %8t | %9t | %s", id, aPath.writable, aPath.installed, aPath.path)
		if selectedID == -1 {
			// Use the previously installed path as the most priority choice.
			if aPath.installed {
				selectedID = id
			}
		}
	}
	// If there's no previously installed path, use the first writable path.
	if selectedID == -1 {
		// Order by choosing priority.
		commonPaths := garray.NewStrArrayFrom(g.SliceStr{
			`/usr/local/bin`,
			`/usr/bin`,
			`/usr/sbin`,
			`C:\Windows`,
			`C:\Windows\system32`,
			`C:\Go\bin`,
			`C:\Program Files`,
			`C:\Program Files (x86)`,
		})
		// Check the common installation directories.
		commonPaths.Iterator(func(k int, v string) bool {
			for id, aPath := range paths {
				if strings.EqualFold(aPath.path, v) {
					selectedID = id
					return false
				}
			}
			return true
		})
		if selectedID == -1 {
			selectedID = 0
		}
	}

	// Get input and update selectedID.
	input := gcmd.Scanf("请选择一个安装位置 [默认 %d]: ", selectedID)
	if input != "" {
		selectedID = gconv.Int(input)
	}

	// Check if out of range.
	if selectedID >= len(paths) || selectedID < 0 {
		_log.Printf("无效的安装位置 Id: %d", selectedID)
		return
	}

	// Get selected destination path.
	dstPath := paths[selectedID]
	// Install the new binary.
	err := gfile.CopyFile(gfile.SelfPath(), dstPath.binaryFilePath)
	if err != nil {
		_log.Printf("安装 k 到 '%s' 失败: %v", dstPath.path, err)
		_log.Printf("您可以手动安装 k, 复制到路径: %s", dstPath.path)
	} else {
		_log.Printf("k 已成功安装到: %s", dstPath.path)
	}

	// Uninstall the old binary.
	for _, aPath := range paths {
		// Do not delete myself.
		if aPath.binaryFilePath != "" &&
			aPath.binaryFilePath != dstPath.binaryFilePath &&
			gfile.SelfPath() != aPath.binaryFilePath {
			gfile.Remove(aPath.binaryFilePath)
		}
	}
}

// IsInstalled returns whether the binary is installed.
func IsInstalled() bool {
	paths := getInstallPathsData()
	for _, aPath := range paths {
		if aPath.installed {
			return true
		}
	}
	return false
}

// GetInstallPathsData returns the installation paths data for the binary.
func getInstallPathsData() []installFolderPath {
	var folderPaths []installFolderPath
	// Pre generate binaryFileName.
	binaryFileName := "k" + gfile.Ext(gfile.SelfPath())
	switch runtime.GOOS {
	case "darwin":
		darwinInstallationCheckPaths := []string{"/usr/local/bin"}
		for _, v := range darwinInstallationCheckPaths {
			folderPaths = checkPathAndAppendToInstallFolderPath(
				folderPaths, v, binaryFileName,
			)
		}
		fallthrough

	default:
		// $GOPATH/bin
		gopath := gfile.Join(runtime.GOROOT(), "bin")
		folderPaths = checkPathAndAppendToInstallFolderPath(
			folderPaths, gopath, binaryFileName,
		)
		// Search and find the writable directory path.
		envPath := genv.Get("PATH", genv.Get("Path"))
		if gstr.Contains(envPath, ";") {
			for _, v := range gstr.SplitAndTrim(envPath, ";") {
				folderPaths = checkPathAndAppendToInstallFolderPath(
					folderPaths, v, binaryFileName,
				)
			}
		} else if gstr.Contains(envPath, ":") {
			for _, v := range gstr.SplitAndTrim(envPath, ":") {
				folderPaths = checkPathAndAppendToInstallFolderPath(
					folderPaths, v, binaryFileName,
				)
			}
		} else if envPath != "" {
			folderPaths = checkPathAndAppendToInstallFolderPath(
				folderPaths, envPath, binaryFileName,
			)
		} else {
			folderPaths = checkPathAndAppendToInstallFolderPath(
				folderPaths, "/usr/local/bin", binaryFileName,
			)
		}
	}
	return folderPaths
}

// checkPathAndAppendToInstallFolderPath checks if `path` is writable and already installed.
// It adds the `path` to `folderPaths` if it is writable or already installed, or else it ignores the `path`.
func checkPathAndAppendToInstallFolderPath(folderPaths []installFolderPath, path string, binaryFileName string) []installFolderPath {
	var (
		binaryFilePath = gfile.Join(path, binaryFileName)
		writable       = gfile.IsWritable(path)
		installed      = isInstalled(binaryFilePath)
	)
	if !writable && !installed {
		return folderPaths
	}
	return append(
		folderPaths,
		installFolderPath{
			path:           path,
			writable:       writable,
			binaryFilePath: binaryFilePath,
			installed:      installed,
		})
}

// Check if this gf binary path exists.
func isInstalled(path string) bool {
	return gfile.Exists(path)
}
