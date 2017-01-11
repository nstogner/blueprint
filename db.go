package blueprint

import "fmt"

var (
	DBTypeMySQL = "mysql"
)

type DB struct {
	Identifier
}

func (cmp DB) Identity() Identifier {
	return cmp.Identifier
}

func (cmp DB) Dependencies() []Identifier {
	return []Identifier{}
}

func (cmp DB) Validate() []error {
	errs := make([]error, 0)

	if err := validateKubeFile(cmp.kubeFilepath()); err != nil {
		errs = append(errs, err)
	}

	switch cmp.Variation {
	case DBTypeMySQL:
	default:
		errs = append(errs, fmt.Errorf("db type not recognized: '%s'", cmp.Variation))
	}

	return errs
}

func (cmp DB) Scaffold() error {
	if err := deployKubeFile(cmp.kubeFilepath(),
		[]interface{}{
			kubeService(cmp.Identifier, 3306),
			mustKubeStatefulSetDB(cmp.Name, cmp.Variation),
		},
	); err != nil {
		return fmt.Errorf("unable to deploy kubernetes files: %s", err)
	}

	return nil
}
