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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	manifest "k8s.io/kubectl/pkg/apis/manifest/v1alpha1"
	"k8s.io/kubectl/pkg/kinflate/configmapandsecret"
	"k8s.io/kubectl/pkg/kinflate/hash"
	kutil "k8s.io/kubectl/pkg/kinflate/util"
)

func populateMap(m map[kutil.GroupVersionKindName]*unstructured.Unstructured, obj *unstructured.Unstructured, newName string) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	oldName := accessor.GetName()
	gvk := obj.GetObjectKind().GroupVersionKind()
	gvkn := kutil.GroupVersionKindName{GVK: gvk, Name: oldName}

	if _, found := m[gvkn]; found {
		return fmt.Errorf("cannot use a duplicate name %q for %s", oldName, gvk)
	}
	accessor.SetName(newName)
	m[gvkn] = obj
	return nil
}

func populateConfigMapAndSecretMap(manifest *manifest.Manifest, m map[kutil.GroupVersionKindName]*unstructured.Unstructured) error {
	for _, cm := range manifest.Configmaps {
		unstructuredConfigMap, nameWithHash, err := makeConfigmapAndGenerateName(cm)
		if err != nil {
			return err
		}
		err = populateMap(m, unstructuredConfigMap, nameWithHash)
		if err != nil {
			return err
		}
	}

	for _, secret := range manifest.Secrets {
		unstructuredSecret, nameWithHash, err := makeSecretAndGenerateName(secret)
		if err != nil {
			return err
		}
		err = populateMap(m, unstructuredSecret, nameWithHash)
		if err != nil {
			return err
		}
	}
	return nil
}

func makeConfigmapAndGenerateName(cm manifest.ConfigMap) (*unstructured.Unstructured, string, error) {
	corev1CM, err := makeConfigMap(cm)
	if err != nil {
		return nil, "", err
	}
	h, err := hash.ConfigMapHash(corev1CM)
	if err != nil {
		return nil, "", err
	}
	nameWithHash := fmt.Sprintf("%s-%s", corev1CM.GetName(), h)
	unstructuredCM, err := objectToUnstructured(corev1CM)
	return unstructuredCM, nameWithHash, err
}

func makeSecretAndGenerateName(secret manifest.Secret) (*unstructured.Unstructured, string, error) {
	corev1Secret, err := makeSecret(secret)
	if err != nil {
		return nil, "", err
	}
	h, err := hash.SecretHash(corev1Secret)
	if err != nil {
		return nil, "", err
	}
	nameWithHash := fmt.Sprintf("%s-%s", corev1Secret.GetName(), h)
	unstructuredCM, err := objectToUnstructured(corev1Secret)
	return unstructuredCM, nameWithHash, err
}

func objectToUnstructured(in runtime.Object) (*unstructured.Unstructured, error) {
	marshaled, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out unstructured.Unstructured
	err = out.UnmarshalJSON(marshaled)
	return &out, err
}

func makeConfigMap(cm manifest.ConfigMap) (*corev1.ConfigMap, error) {
	corev1cm := &corev1.ConfigMap{}
	corev1cm.APIVersion = "v1"
	corev1cm.Kind = "ConfigMap"
	corev1cm.Name = cm.NamePrefix
	corev1cm.Data = map[string]string{}
	var err error
	switch cm.Type {
	case "env":
		err = configmapandsecret.HandleConfigMapFromEnvFileSource(corev1cm, cm.EnvSource)
	case "file":
		err = configmapandsecret.HandleConfigMapFromFileSources(corev1cm, cm.FileSources)
	case "literal":
		err = configmapandsecret.HandleConfigMapFromLiteralSources(corev1cm, cm.LiteralSources)
	default:
		err = fmt.Errorf("unknown type of configmap: %v", cm.Type)
	}
	return corev1cm, err
}

func makeSecret(secret manifest.Secret) (*corev1.Secret, error) {
	corev1secret := &corev1.Secret{}
	corev1secret.APIVersion = "v1"
	corev1secret.Kind = "Secret"
	corev1secret.Name = secret.NamePrefix
	corev1secret.Type = corev1.SecretTypeOpaque
	corev1secret.Data = map[string][]byte{}
	var err error
	switch secret.Type {
	case "tls":
		if err = validateTLS(secret.TLS.CertFile, secret.TLS.KeyFile); err != nil {
			return nil, err
		}
		tlsCrt, err := ioutil.ReadFile(secret.TLS.CertFile)
		if err != nil {
			return nil, err
		}
		tlsKey, err := ioutil.ReadFile(secret.TLS.KeyFile)
		if err != nil {
			return nil, err
		}
		corev1secret.Type = corev1.SecretTypeTLS
		corev1secret.Data[corev1.TLSCertKey] = []byte(tlsCrt)
		corev1secret.Data[corev1.TLSPrivateKeyKey] = []byte(tlsKey)
	case "env":
		err = configmapandsecret.HandleFromEnvFileSource(corev1secret, secret.EnvSource)
	case "file":
		err = configmapandsecret.HandleFromFileSources(corev1secret, secret.FileSources)
	case "literal":
		err = configmapandsecret.HandleFromLiteralSources(corev1secret, secret.LiteralSources)
	default:
		err = fmt.Errorf("unknown type of secret: %v", secret.Type)
	}
	return corev1secret, err
}

func validateTLS(cert, key string) error {
	if len(key) == 0 {
		return fmt.Errorf("key must be specified")
	}
	if len(cert) == 0 {
		return fmt.Errorf("certificate must be specified")
	}
	if _, err := tls.LoadX509KeyPair(cert, key); err != nil {
		return fmt.Errorf("failed to load key pair %v", err)
	}
	return nil
}
