package load

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
)

func Path(relative, absolute string) string {
	if os.Getenv("ENV") == "TEST" {
		_, filename, _, _ := runtime.Caller(1)
		return filepath.Join(path.Dir(filename), relative)
	}
	dir, _ := os.Getwd()
	return dir + absolute
}
