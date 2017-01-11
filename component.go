package blueprint

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/nstogner/kit/env"
)

var (
	TmplDir = "/tmpl"
	GoPath  = env.MustString("GOPATH")
)

const (
	KindAPI  = "api"
	KindCron = "cron"
	KindDB   = "db"
	LangGo   = "go"
)

type Component interface {
	Scaffold() error
	Validate() []error
	Identity() Identifier
	Dependencies() []Identifier
}

// ParseIdentifier parses a string of expected form "name.variation.kind" into
// an Identifier struct.
func ParseIdentifier(s string) (*Identifier, error) {
	split := strings.Split(s, ".")
	err := errors.New("invalid identifier: expected form: name.variation.kind")
	if len(split) != 3 {
		return nil, err
	}
	return &Identifier{
		Name:      split[0],
		Variation: split[1],
		Kind:      split[2],
	}, nil
}

// ParseIdentifier parses a string slice of expected form "name.variation.kind"
// into a slice of Identifier structs.
func ParseIdentifierSlice(s []string) ([]Identifier, error) {
	ids := make([]Identifier, len(s))
	for i, str := range s {
		id, err := ParseIdentifier(str)
		if err != nil {
			return nil, err
		}
		ids[i] = *id
	}
	return ids, nil
}

type Identifier struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Variation string `json:"variation"`
}

func (idt Identifier) Host() string {
	return fmt.Sprintf("%s-%s-%s", idt.Name, idt.Variation, idt.Kind)
}

func (idt Identifier) String() string {
	return fmt.Sprintf("%s.%s.%s", idt.Name, idt.Variation, idt.Kind)
}

func (idt Identifier) projectDir() string {
	return filepath.Join(idt.Variation, "src", idt.Kind, idt.Name)
}

func (idt Identifier) templateDir() string {
	return filepath.Join(TmplDir, idt.Variation, idt.Kind)
}

func (idt Identifier) protoFilepath() string {
	return filepath.Join("proto", idt.String()+".proto")
}

func (idt Identifier) kubeFilepath() string {
	return filepath.Join("kube", idt.String()+"."+KubeFileType)
}

type UnmarshalFunc func([]byte, interface{}) error

func Decode(r io.Reader, u UnmarshalFunc) (Component, error) {
	btys, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read input: %s", err)
	}

	var idt Identifier
	if err := u(btys, &idt); err != nil {
		return nil, fmt.Errorf("unable to decode component: %s", err)
	}

	switch idt.Kind {
	case KindAPI:
		var api API
		if err := u(btys, &api); err != nil {
			return nil, fmt.Errorf("unable to decode as %s component: %s", KindAPI, err)
		}
		return api, nil
	case KindCron:
		var cron Cron
		if err := u(btys, &cron); err != nil {
			return nil, fmt.Errorf("unable to decode as %s component: %s", KindCron, err)
		}
		return cron, nil
	case KindDB:
		var db DB
		if err := u(btys, &db); err != nil {
			return nil, fmt.Errorf("unable to decode as %s component: %s", KindCron, err)
		}
		return db, nil
	default:
		return nil, fmt.Errorf("component type not recognized: '%s'", idt.Kind)
	}
}

func validateDependency(d *Identifier) error {
	if d == nil {
		return nil
	}

	// Ensure kubernetes file exists for dependency
	fp := d.kubeFilepath()
	_, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return fmt.Errorf("dependency does not appear to exist: missing kubernetes file: '%s'", fp)
	}
	if err != nil {
		return fmt.Errorf("unable to stat dependency's kubernetes file '%s': %s", fp, err)
	}

	return nil
}

func validateLang(lang string) error {
	switch lang {
	case LangGo:
		return nil
	default:
		return fmt.Errorf("language not recognized: %s", lang)
	}
}

func validateProtoFile(f string) error {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		return fmt.Errorf("missing expected api protobuf definition file: %s", f)
	}
	if err != nil {
		return fmt.Errorf("unable to stat api protobuf definition file: %s", err)
	}
	return nil
}

func validateProjectDir(d string) error {
	// Ensure target directory does not exist
	_, err := os.Stat(d)
	if err == nil {
		// Exists
		return fmt.Errorf("conflict: component path '%s' already exists", d)
	} else {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to stat project directory '%s': %s", d, err)
		}
	}
	return nil
}

func validateKubeFile(f string) error {
	// Ensure kube file does not exist
	_, err := os.Stat(f)
	if err == nil {
		// Exists
		return fmt.Errorf("conflict: kubernetes file '%s' already exists", f)
	} else {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to stat kubernetes file '%s': %s", f, err)
		}
	}
	return nil
}
