package version

import "fmt"

var (
	Version = "dev"
	Commit  = "unknown"
	Ref     = "unknown"
)

func Info() string {
	return fmt.Sprintf("ralph version=%s commit=%s ref=%s", Version, Commit, Ref)
}
