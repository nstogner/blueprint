package blueprint

import "fmt"

type API struct {
	Identifier

	DB *Identifier `json:"db"`
}

func (cmp API) Identity() Identifier {
	return cmp.Identifier
}

func (cmp API) Dependencies() []Identifier {
	ds := make([]Identifier, 0)
	if cmp.DB != nil {
		ds = append(ds, *cmp.DB)
	}
	return ds
}

func (cmp API) Validate() []error {
	return nonNilErrs(
		validateLang(cmp.Variation),
		validateProjectDir(cmp.projectDir()),
		validateKubeFile(cmp.kubeFilepath()),
		validateProtoFile(cmp.protoFilepath()),
		validateDependency(cmp.DB),
	)
}

func (cmp API) Scaffold() error {
	if err := deployTemplate(cmp); err != nil {
		return fmt.Errorf("unable to deploy templates: %s", err)
	}

	if err := compileProtos("proto", cmp.protoFilepath(), cmp.projectDir()); err != nil {
		return fmt.Errorf("unable to compile protos: %s", err)
	}

	if err := deployKubeFile(cmp.kubeFilepath(),
		[]interface{}{
			kubeService(cmp.Identifier, 8080),
			kubeDeploymentAPI(cmp.Name),
		},
	); err != nil {
		return fmt.Errorf("unable to deploy kubernetes files: %s", err)
	}

	if err := runImports(cmp.projectDir()); err != nil {
		return fmt.Errorf("unable to run goimports for all go files: %s", err)
	}

	return nil
}
