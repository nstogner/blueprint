package blueprint

import (
	"bytes"
	"fmt"
	"os/exec"
)

func sysExec(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	var b bytes.Buffer
	cmd.Stderr = &b

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("protoc failed: %s: %s", err, b.String())
	}

	return nil
}

func keys(m map[string]bool) []string {
	var ks []string
	for k, _ := range m {
		ks = append(ks, k)
	}
	return ks
}

func nonNilErrs(errs ...error) []error {
	nne := make([]error, 0)
	for _, e := range errs {
		if e != nil {
			nne = append(nne, e)
		}
	}
	return nne
}
