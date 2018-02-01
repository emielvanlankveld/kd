/*
Copyright 2018 The Kubernetes Authors.

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

package kinflate

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	manifest "k8s.io/kubectl/pkg/apis/manifest/v1alpha1"
	kutil "k8s.io/kubectl/pkg/kinflate/util"
)

func makeEnvConfigMap(name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			"DB_USERNAME": "admin",
			"DB_PASSWORD": "somepw",
		},
	}
}

func makeUnstructuredEnvConfigMap(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":              name,
				"creationTimestamp": nil,
			},
			"data": map[string]interface{}{
				"DB_USERNAME": "admin",
				"DB_PASSWORD": "somepw",
			},
		},
	}
}

func makeFileConfigMap(name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			"app-init.ini": `FOO=bar
BAR=baz
`,
		},
	}
}

func makeLiteralConfigMap(name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			"a": "x",
			"b": "y",
		},
	}
}

func makeTLSSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string][]byte{
			corev1.TLSCertKey: []byte(`-----BEGIN CERTIFICATE-----
MIIB0zCCAX2gAwIBAgIJAI/M7BYjwB+uMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTIwOTEyMjE1MjAyWhcNMTUwOTEyMjE1MjAyWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANLJ
hPHhITqQbPklG3ibCVxwGMRfp/v4XqhfdQHdcVfHap6NQ5Wok/4xIA+ui35/MmNa
rtNuC+BdZ1tMuVCPFZcCAwEAAaNQME4wHQYDVR0OBBYEFJvKs8RfJaXTH08W+SGv
zQyKn0H8MB8GA1UdIwQYMBaAFJvKs8RfJaXTH08W+SGvzQyKn0H8MAwGA1UdEwQF
MAMBAf8wDQYJKoZIhvcNAQEFBQADQQBJlffJHybjDGxRMqaRmDhX0+6v02TUKZsW
r5QuVbpQhH6u+0UgcW0jp9QwpxoPTLTWGXEWBBBurxFwiCBhkQ+V
-----END CERTIFICATE-----
`),
			corev1.TLSPrivateKeyKey: []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBANLJhPHhITqQbPklG3ibCVxwGMRfp/v4XqhfdQHdcVfHap6NQ5Wo
k/4xIA+ui35/MmNartNuC+BdZ1tMuVCPFZcCAwEAAQJAEJ2N+zsR0Xn8/Q6twa4G
6OB1M1WO+k+ztnX/1SvNeWu8D6GImtupLTYgjZcHufykj09jiHmjHx8u8ZZB/o1N
MQIhAPW+eyZo7ay3lMz1V01WVjNKK9QSn1MJlb06h/LuYv9FAiEA25WPedKgVyCW
SmUwbPw8fnTcpqDWE3yTO3vKcebqMSsCIBF3UmVue8YU3jybC3NxuXq3wNm34R8T
xVLHwDXh/6NJAiEAl2oHGGLz64BuAfjKrqwz7qMYr9HCLIe/YsoWq/olzScCIQDi
D2lWusoe2/nEqfDVVWGWlyJ7yOmqaVm/iNUN9B2N2g==
-----END RSA PRIVATE KEY-----
`),
		},
		Type: corev1.SecretTypeTLS,
	}
}

func makeEnvSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string][]byte{
			"DB_USERNAME": []byte("admin"),
			"DB_PASSWORD": []byte("somepw"),
		},
		Type: corev1.SecretTypeOpaque,
	}
}

func makeUnstructuredEnvSecret(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":              name,
				"creationTimestamp": nil,
			},
			"type": string(corev1.SecretTypeOpaque),
			"data": map[string]interface{}{
				"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
				"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("somepw")),
			},
		},
	}
}

func makeFileSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string][]byte{
			"app-init.ini": []byte(`FOO=bar
BAR=baz
`),
		},
		Type: corev1.SecretTypeOpaque,
	}
}

func makeLiteralSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string][]byte{
			"a": []byte("x"),
			"b": []byte("y"),
		},
		Type: corev1.SecretTypeOpaque,
	}
}

func TestConstructConfigMap(t *testing.T) {
	type testCase struct {
		description string
		input       manifest.ConfigMap
		expected    *corev1.ConfigMap
	}

	testCases := []testCase{
		{
			description: "construct config map from env",
			input: manifest.ConfigMap{
				Type:       "env",
				NamePrefix: "envConfigMap",
				Generic: manifest.Generic{
					EnvSource: "examples/simple/instances/exampleinstance/configmap/app.env",
				},
			},
			expected: makeEnvConfigMap("envConfigMap"),
		},
		{
			description: "construct config map from file",
			input: manifest.ConfigMap{
				Type:       "file",
				NamePrefix: "fileConfigMap",
				Generic: manifest.Generic{
					FileSources: []string{"examples/simple/instances/exampleinstance/configmap/app-init.ini"},
				},
			},
			expected: makeFileConfigMap("fileConfigMap"),
		},
		{
			description: "construct config map from literal",
			input: manifest.ConfigMap{
				Type:       "literal",
				NamePrefix: "literalConfigMap",
				Generic: manifest.Generic{
					LiteralSources: []string{"a=x", "b=y"},
				},
			},
			expected: makeLiteralConfigMap("literalConfigMap"),
		},
	}

	for _, tc := range testCases {
		cm, err := makeConfigMap(tc.input)
		if err != nil {
			t.Fatalf("unepxected error: %v", err)
		}
		if !reflect.DeepEqual(*cm, *tc.expected) {
			t.Fatalf("in testcase: %q updated:\n%#v\ndoesn't match expected:\n%#v\n", tc.description, *cm, tc.expected)
		}
	}
}

func TestConstructSecret(t *testing.T) {
	type testCase struct {
		description string
		input       manifest.Secret
		expected    *corev1.Secret
	}

	testCases := []testCase{
		{
			description: "construct secret from tls",
			input: manifest.Secret{
				Type:       "tls",
				NamePrefix: "tlsSecret",
				TLS: &manifest.TLS{
					CertFile: "examples/simple/instances/exampleinstance/secret/tls.cert",
					KeyFile:  "examples/simple/instances/exampleinstance/secret/tls.key",
				},
			},
			expected: makeTLSSecret("tlsSecret"),
		},
		{
			description: "construct secret from env",
			input: manifest.Secret{
				Type:       "env",
				NamePrefix: "envSecret",
				Generic: manifest.Generic{
					EnvSource: "examples/simple/instances/exampleinstance/configmap/app.env",
				},
			},
			expected: makeEnvSecret("envSecret"),
		},
		{
			description: "construct secret from file",
			input: manifest.Secret{
				Type:       "file",
				NamePrefix: "fileSecret",
				Generic: manifest.Generic{
					FileSources: []string{"examples/simple/instances/exampleinstance/configmap/app-init.ini"},
				},
			},
			expected: makeFileSecret("fileSecret"),
		},
		{
			description: "construct secret from literal",
			input: manifest.Secret{
				Type:       "literal",
				NamePrefix: "literalSecret",
				Generic: manifest.Generic{
					LiteralSources: []string{"a=x", "b=y"},
				},
			},
			expected: makeLiteralSecret("literalSecret"),
		},
	}

	for _, tc := range testCases {
		cm, err := makeSecret(tc.input)
		if err != nil {
			t.Fatalf("unepxected error: %v", err)
		}
		if !reflect.DeepEqual(*cm, *tc.expected) {
			t.Fatalf("in testcase: %q updated:\n%#v\ndoesn't match expected:\n%#v\n", tc.description, *cm, tc.expected)
		}
	}
}

func TestObjectConvertToUnstructured(t *testing.T) {
	type testCase struct {
		description string
		input       *corev1.ConfigMap
		expected    *unstructured.Unstructured
	}

	testCases := []testCase{
		{
			description: "convert config map",
			input:       makeEnvConfigMap("envConfigMap"),
			expected:    makeUnstructuredEnvConfigMap("envConfigMap"),
		},
		{
			description: "convert secret",
			input:       makeEnvConfigMap("envSecret"),
			expected:    makeUnstructuredEnvConfigMap("envSecret"),
		},
	}
	for _, tc := range testCases {
		actual, err := objectToUnstructured(tc.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Fatalf("%#v\ndoesn't match expected\n%#v\n", actual, tc.expected)
		}
	}
}

func TestPopulateMap(t *testing.T) {
	expectedMap := map[kutil.GroupVersionKindName]*unstructured.Unstructured{
		{
			GVK: schema.GroupVersionKind{
				Version: "v1",
				Kind:    "ConfigMap",
			},
			Name: "envConfigMap",
		}: makeUnstructuredEnvConfigMap("newNameConfigMap"),
		{
			GVK: schema.GroupVersionKind{
				Version: "v1",
				Kind:    "Secret",
			},
			Name: "envSecret",
		}: makeUnstructuredEnvSecret("newNameSecret"),
	}

	m := map[kutil.GroupVersionKindName]*unstructured.Unstructured{}
	err := populateMap(m, makeUnstructuredEnvConfigMap("envConfigMap"), "newNameConfigMap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = populateMap(m, makeUnstructuredEnvSecret("envSecret"), "newNameSecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(m, expectedMap) {
		t.Fatalf("%#v\ndoesn't match expected\n%#v\n", m, expectedMap)
	}

	err = populateMap(m, makeUnstructuredEnvSecret("envSecret"), "newNameSecret")
	if err == nil || !strings.Contains(err.Error(), "duplicate name") {
		t.Fatalf("expected error to contain %q, but got: %v", "duplicate name", err)
	}
}

func TestPopulateMapOfConfigMapAndSecret(t *testing.T) {
	m := map[kutil.GroupVersionKindName]*unstructured.Unstructured{}
	manifest := &manifest.Manifest{
		Configmaps: []manifest.ConfigMap{
			{
				Type:       "env",
				NamePrefix: "envConfigMap",
				Generic: manifest.Generic{
					EnvSource: "examples/simple/instances/exampleinstance/configmap/app.env",
				},
			},
		},
		Secrets: []manifest.Secret{
			{
				Type:       "env",
				NamePrefix: "envSecret",
				Generic: manifest.Generic{
					EnvSource: "examples/simple/instances/exampleinstance/configmap/app.env",
				},
			},
		},
	}
	expectedMap := map[kutil.GroupVersionKindName]*unstructured.Unstructured{
		{
			GVK: schema.GroupVersionKind{
				Version: "v1",
				Kind:    "ConfigMap",
			},
			Name: "envConfigMap",
		}: makeUnstructuredEnvConfigMap("envConfigMap-d2c89bt4kk"),
		{
			GVK: schema.GroupVersionKind{
				Version: "v1",
				Kind:    "Secret",
			},
			Name: "envSecret",
		}: makeUnstructuredEnvSecret("envSecret-684h2mm268"),
	}
	err := populateConfigMapAndSecretMap(manifest, m)
	if err != nil {
		t.Fatalf("unexpected erorr: %v", err)
	}
	if !reflect.DeepEqual(m, expectedMap) {
		t.Fatalf("%#v\ndoesn't match expected\n%#v\n", m, expectedMap)
	}
}
