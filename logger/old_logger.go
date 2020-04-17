package logger

import (
	"fmt"
	"os"
	"path"

	"github.com/sonatype-nexus-community/nancy/types"
)

// GetLogFileLocation will return the location on disk of the log file
func GetLogFileLocation() (result string) {
	result, _ = os.UserHomeDir()
	err := os.MkdirAll(path.Join(result, types.OssIndexDirName), os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	result = path.Join(result, types.OssIndexDirName, DefaultLogFile)
	return
}
