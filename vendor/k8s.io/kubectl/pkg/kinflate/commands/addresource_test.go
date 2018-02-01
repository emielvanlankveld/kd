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
	"bytes"
	"os"
	"testing"

	"strings"

	"k8s.io/kubectl/pkg/kinflate"
	"k8s.io/kubectl/pkg/kinflate/util/fs"
)

const (
	// This should be in manifestTemplate.
	resourceKnownToBeInManifest = "deployment.yaml"
	resourceFileName            = "myWonderfulResource.yaml"
	resourceFileContent         = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func TestAddResourceHappyPath(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(resourceFileName, []byte(resourceFileContent))
	fakeFS.WriteFile(kinflate.KubeManifestFileName, []byte(manifestTemplate))

	cmd := NewCmdAddResource(buf, os.Stderr, fakeFS)
	args := []string{resourceFileName}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := fakeFS.ReadFile(kinflate.KubeManifestFileName)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), resourceFileName) {
		t.Errorf("expected resource name in manifest")
	}
}

func TestAddResourceAlreadyThere(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(resourceKnownToBeInManifest, []byte(resourceFileContent))
	fakeFS.WriteFile(kinflate.KubeManifestFileName, []byte(manifestTemplate))
	cmd := NewCmdAddResource(buf, os.Stderr, fakeFS)
	args := []string{resourceKnownToBeInManifest}
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected already there problem")
	}
	if err.Error() != "resource "+resourceKnownToBeInManifest+" already in manifest" {
		t.Errorf("unexpected error %v", err)
	}
}

func TestAddResourceNoArgs(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	fakeFS := fs.MakeFakeFS()

	cmd := NewCmdAddResource(buf, os.Stderr, fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify a resource file" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
