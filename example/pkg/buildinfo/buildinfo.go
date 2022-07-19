package buildinfo

import "fmt"

var (
	Version     = ""
	GitCommitId = ""
	ProjectName = ""
	BuildTime   = ""
	Branch      = ""
	GoVersion   = ""
	OsArch      = ""
	Mode        = ""
	Banner      = `ProjectName: %s
GoVersion: %s
OS/Arch: %s
GitCommitId: %s
Version: %s
BuildTime: %s
Branch: %s
Mode: %s
`
)

func Print() {
	fmt.Printf(Banner, ProjectName, GoVersion, OsArch, GitCommitId, Version, BuildTime, Branch, Mode)
}

func IsProd() bool {
	return Mode == "prod"
}
