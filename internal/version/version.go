package version

var (
	Version   string = "dev"
	GitCommit string = "unknown"
	BuildTime string = "unknown"
)

func GetVersion() string {
	return Version
}

func GetGitCommit() string {
	return GitCommit
}

func GetBuildTime() string {
	return BuildTime
}

func GetFullVersion() string {
	return Version + " (commit: " + GitCommit + ", built: " + BuildTime + ")"
}
