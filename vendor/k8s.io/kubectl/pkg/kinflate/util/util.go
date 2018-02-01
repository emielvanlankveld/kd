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

package util

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/ghodss/yaml"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

// GroupVersionKindName contains GroupVersionKind and original name of the resource.
type GroupVersionKindName struct {
	// GroupVersionKind of the resource.
	GVK schema.GroupVersionKind
	// original name of the resource before transformation.
	Name string
}

// Decode decodes a list of objects in byte array format.
// Decoded object will be inserted in `into` if it's not nil. Otherwise, it will
// construct a new map and return it.
func Decode(in []byte, into map[GroupVersionKindName]*unstructured.Unstructured) (map[GroupVersionKindName]*unstructured.Unstructured, error) {
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(in), 1024)
	objs := []*unstructured.Unstructured{}

	var err error
	for {
		var out unstructured.Unstructured
		err = decoder.Decode(&out)
		if err != nil {
			break
		}
		objs = append(objs, &out)
	}
	if err != io.EOF {
		return nil, err
	}

	if into == nil {
		into = map[GroupVersionKindName]*unstructured.Unstructured{}
	}
	for i := range objs {
		metaAccessor, err := meta.Accessor(objs[i])
		if err != nil {
			return nil, err
		}
		name := metaAccessor.GetName()
		typeAccessor, err := meta.TypeAccessor(objs[i])
		if err != nil {
			return nil, err
		}
		apiVersion := typeAccessor.GetAPIVersion()
		kind := typeAccessor.GetKind()
		gv, err := schema.ParseGroupVersion(apiVersion)
		if err != nil {
			return nil, err
		}
		gvk := gv.WithKind(kind)
		gvkn := GroupVersionKindName{
			GVK:  gvk,
			Name: name,
		}
		into[gvkn] = objs[i]
	}
	return into, nil
}

// Encode encodes the map `in` and output the encoded objects separated by `---`.
func Encode(in map[GroupVersionKindName]*unstructured.Unstructured) ([]byte, error) {
	gvknList := []GroupVersionKindName{}
	for gvkn := range in {
		gvknList = append(gvknList, gvkn)
	}
	sort.Sort(ByGVKN(gvknList))

	firstObj := true
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, gvkn := range gvknList {
		obj := in[gvkn]
		out, err := yaml.Marshal(obj)
		if err != nil {
			return nil, err
		}
		if !firstObj {
			_, err = buf.WriteString("---\n")
			if err != nil {
				return nil, err
			}
		}
		_, err = buf.Write(out)
		if err != nil {
			return nil, err
		}
		firstObj = false
	}
	return buf.Bytes(), nil
}

// SelectByGVK returns true if `selector` selects `in`; otherwise, false.
// If `selector` and `in` are the same, return true.
// If `selector` is nil, it is considered as a wildcard and always return true.
// e.g. selector <Group: "", Version: "", Kind: "Deployemt"> CAN select
// <Group: "extensions", Version: "v1beta1", Kind: "Deployemt">.
// selector <Group: "apps", Version: "", Kind: "Deployemt"> CANNOT select
// <Group: "extensions", Version: "v1beta1", Kind: "Deployemt">.
func SelectByGVK(in schema.GroupVersionKind, selector *schema.GroupVersionKind) bool {
	if selector == nil {
		return true
	}
	if len(selector.Group) > 0 {
		if in.Group != selector.Group {
			return false
		}
	}
	if len(selector.Version) > 0 {
		if in.Version != selector.Version {
			return false
		}
	}
	if len(selector.Kind) > 0 {
		if in.Kind != selector.Kind {
			return false
		}
	}
	return true
}

type mutateFunc func(interface{}) (interface{}, error)

func mutateField(m map[string]interface{}, pathToField []string, createIfNotPresent bool, fns ...mutateFunc) error {
	if len(pathToField) == 0 {
		return nil
	}

	_, found := m[pathToField[0]]
	if !found {
		if !createIfNotPresent {
			return nil
		}
		m[pathToField[0]] = map[string]interface{}{}
	}

	if len(pathToField) == 1 {
		var err error
		for _, fn := range fns {
			m[pathToField[0]], err = fn(m[pathToField[0]])
			if err != nil {
				return err
			}
		}
		return nil
	}

	v := m[pathToField[0]]
	newPathToField := pathToField[1:]
	switch typedV := v.(type) {
	case map[string]interface{}:
		return mutateField(typedV, newPathToField, createIfNotPresent, fns...)
	case []interface{}:
		for i := range typedV {
			item := typedV[i]
			typedItem, ok := item.(map[string]interface{})
			if !ok {
				return fmt.Errorf("%#v is expectd to be %T", item, typedItem)
			}
			err := mutateField(typedItem, newPathToField, createIfNotPresent, fns...)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("%#v is not expected to be a primitive type", typedV)
	}
}
