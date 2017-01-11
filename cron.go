package blueprint

import "fmt"

type Cron struct {
	Identifier

	Schedule string `json:"schedule"`
}

func (cmp Cron) Identity() Identifier {
	return cmp.Identifier
}

func (cmp Cron) Dependencies() []Identifier {
	return []Identifier{}
}

func (cmp Cron) Validate() []error {
	return nonNilErrs(
		validateLang(cmp.Variation),
		validateProjectDir(cmp.projectDir()),
		validateKubeFile(cmp.kubeFilepath()),
	)
	// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
	//
	// Licensed under the Apache License, Version 2.0 (the "License");
	// you may not use this file except in compliance with the License.
	// You may obtain a copy of the License at
	//
	//     http://www.apache.org/licenses/LICENSE-2.0
	//
	// Unless required by applicable law or agreed to in writing, software
	// distributed under the License is distributed on an "AS IS" BASIS,
	// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	// See the License for the specific language governing permissions and
	// limitations under the License.

}

func (cmp Cron) Scaffold() error {
	if err := deployTemplate(cmp); err != nil {
		return fmt.Errorf("unable to deploy templates: %s", err)
	}

	if err := deployKubeFile(cmp.kubeFilepath(),
		[]interface{}{
			// kubeCronJob(cmp.Name, cmp.Schedule),
			kubeScheduledJob(cmp.Name, cmp.Schedule),
		},
	); err != nil {
		return fmt.Errorf("unable to deploy kubernetes files: %s", err)
	}

	if err := runImports(cmp.projectDir()); err != nil {
		return fmt.Errorf("unable to run goimports for all go files: %s", err)
	}

	return nil
}
