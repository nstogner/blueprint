package blueprint

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/nstogner/kit/log"
)

// deployTemplate copies a template directory for a given component
// and runs all .tmpl files through a templater.
func deployTemplate(cmp Component) error {
	if err := os.MkdirAll(cmp.Identity().projectDir(), 0755); err != nil {
		return fmt.Errorf("unable to make directories: %s", err)
	}

	// NOTE: Use system exec here b/c there is no easy way to do this in go
	source := cmp.Identity().templateDir() + "/"
	if err := sysExec("rsync", "-a", source, cmp.Identity().projectDir()); err != nil {
		return fmt.Errorf("unable to rsync directory '%s' -> '%s': %s", source, cmp.Identity().projectDir(), err)
	}

	if err := executeTemplates(cmp); err != nil {
		return fmt.Errorf("unable to execute templates: %s", err)
	}

	return nil
}

// executeTemplates recursively executes all .tmpl files in a dir
func executeTemplates(cmp Component) error {
	var dir = cmp.Identity().projectDir()

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("unable to walk directory: %s", err)
		}

		// Remove all placeholder files (used to keep commit template directory structure
		// to git repo)
		if info.Name() == "placeholder" && !info.IsDir() {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("unable to remove placeholder file: %s", err)
			}
			return nil
		}

		if strings.HasSuffix(path, ".tmpl") {
			tgt := strings.TrimSuffix(path, ".tmpl")

			log.Debug("processing template",
				"template", path, "target", tgt)

			tmplContent, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("unable to read file: %s", err)
			}

			tmpl, err := template.New("").Funcs(template.FuncMap{
				"title": strings.Title,
			}).Parse(string(tmplContent))
			if err != nil {
				return fmt.Errorf("unable to parse template file '%s': %s", info.Name(), err)
			}

			f, err := os.Create(tgt)
			if err != nil {
				return fmt.Errorf("unable to create executed template destination file: %s", err)
			}

			var dbDep *Identifier
			var hasDbDep bool
			for _, d := range cmp.Dependencies() {
				if d.Kind == KindDB {
					dbDep = &d
					hasDbDep = true
				}
			}

			params := struct {
				Const  map[string]interface{}
				Comp   Component
				CompID string
				Deps   []Identifier
				// Helpers so that Deps doesnt have to be filtered in the template
				DBDep    *Identifier
				HasDBDep bool
			}{
				Const: map[string]interface{}{
					"KindAPI":     KindAPI,
					"KindCron":    KindCron,
					"KindDB":      KindDB,
					"LangGo":      LangGo,
					"DBTypeMySQL": DBTypeMySQL,
				},
				Comp:     cmp,
				CompID:   cmp.Identity().String(),
				Deps:     cmp.Dependencies(),
				DBDep:    dbDep,
				HasDBDep: hasDbDep,
			}
			if err := tmpl.Execute(f, params); err != nil {
				return fmt.Errorf("unable to execute template: %s", err)
			}
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("unable to remove executed template: %s", err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
