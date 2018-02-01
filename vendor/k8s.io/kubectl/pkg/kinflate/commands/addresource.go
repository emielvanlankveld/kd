/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"io"

	"github.com/ghodss/yaml"

	"errors"

	"github.com/spf13/cobra"
	manifest "k8s.io/kubectl/pkg/apis/manifest/v1alpha1"
	"k8s.io/kubectl/pkg/kinflate/util/fs"

	"fmt"

	"k8s.io/kubectl/pkg/kinflate"
)

type addResourceOptions struct {
	resourceFilePath string
}

// NewCmdAddResource adds the name of a file containing a resource to the manifest.
func NewCmdAddResource(out, errOut io.Writer, fsys fs.FileSystem) *cobra.Command {
	var o addResourceOptions

	cmd := &cobra.Command{
		Use:   "addresource",
		Short: "Add the name of a file containing a resource to the manifest.",
		Long:  "Add the name of a file containing a resource to the manifest.",
		Example: `
		addresource {filepath}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.RunAddResource(out, errOut, fsys)
		},
	}
	return cmd
}

// Validate validates addResource command.
func (o *addResourceOptions) Validate(args []string) error {
	if len(args) != 1 {
		return errors.New("must specify a resource file")
	}
	o.resourceFilePath = args[0]
	return nil
}

// Complete completes addResource command.
func (o *addResourceOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// RunAddResource runs addResource command (do real work).
func (o *addResourceOptions) RunAddResource(out, errOut io.Writer, fsys fs.FileSystem) error {
	_, err := fsys.Stat(o.resourceFilePath)
	if err != nil {
		return err
	}

	content, err := fsys.ReadFile(kinflate.KubeManifestFileName)
	if err != nil {
		return err
	}

	// TODO: Refactor to a common location you guys!
	// See pkg/kinflate/util.go:loadManifestPkg
	var m manifest.Manifest
	err = yaml.Unmarshal(content, &m)
	if err != nil {
		return err
	}

	if stringInSlice(o.resourceFilePath, m.Resources) {
		return fmt.Errorf("resource %s already in manifest", o.resourceFilePath)
	}

	m.Resources = append(m.Resources, o.resourceFilePath)

	bytes, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	err = fsys.WriteFile(kinflate.KubeManifestFileName, bytes)
	if err != nil {
		return err
	}
	return nil
}
