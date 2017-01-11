package blueprint

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/imports"
)

// runImports recursively runs goimports on all git files in a directory.
func runImports(projectDir string) error {
	return filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("unable to walk directory: %s", err)
		}

		if isGoFile(info) {
			src, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("unable to read file: %s", err)
			}

			newSrc, err := imports.Process(path, src, nil)
			if err != nil {
				return fmt.Errorf("unable to run goimports on file: %s", err)
			}

			if err := ioutil.WriteFile(path, newSrc, 0644); err != nil {
				return fmt.Errorf("unable to write file: %s", err)
			}
		}

		return nil
	})
}

func isGoFile(f os.FileInfo) bool {
	// ignore non-Go files
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}
