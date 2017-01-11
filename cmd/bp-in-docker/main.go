package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nstogner/blueprint"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("missing command argument")
		os.Exit(1)
	}

	cmd, ok := cmds[os.Args[1]]
	if !ok {
		fatal("command not found", os.Args[1])
	}
	cmd()
}

var cmds = map[string]func(){
	"new": func() {
		cmp, err := blueprint.Decode(os.Stdin, json.Unmarshal)
		if err != nil {
			fatal("unable to decode component", err)
		}

		if errs := cmp.Validate(); len(errs) != 0 {
			fatal("component validation failed", combineErrs(errs))
		}

		if err := cmp.Scaffold(); err != nil {
			fatal("unable to scaffold", err)
		}
	},
}

func fatal(msg string, err interface{}) {
	fmt.Printf("%s: %s\n", msg, err)
	os.Exit(1)
}

func combineErrs(errs []error) string {
	str := "\n"
	for _, e := range errs {
		str = str + fmt.Sprintf(" - %s\n", e)
	}
	return str
}
