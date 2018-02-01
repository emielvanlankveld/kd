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
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// MapTransformer contains a map string->string and path configs
// The map will be applied to the fields specified in path configs.
type MapTransformer struct {
	m           map[string]string
	pathConfigs []PathConfig
}

var _ Transformer = &MapTransformer{}

// NewDefaultingLabelsMapTransformer construct a MapTransformer with defaultLabelsPathConfigs.
func NewDefaultingLabelsMapTransformer(m map[string]string) (Transformer, error) {
	return NewMapTransformer(defaultLabelsPathConfigs, m)
}

// NewDefaultingAnnotationsMapTransformer construct a MapTransformer with defaultAnnotationsPathConfigs.
func NewDefaultingAnnotationsMapTransformer(m map[string]string) (Transformer, error) {
	return NewMapTransformer(defaultAnnotationsPathConfigs, m)
}

// NewMapTransformer construct a MapTransformer.
func NewMapTransformer(pc []PathConfig, m map[string]string) (Transformer, error) {
	if m == nil {
		return nil, nil
	}
	if pc == nil {
		return nil, errors.New("pathConfigs is not expected to be nil")
	}
	return &MapTransformer{pathConfigs: pc, m: m}, nil
}

// Transform apply each <key, value> pair in the MapTransformer to the
// fields specified in MapTransformer.
func (o *MapTransformer) Transform(m map[GroupVersionKindName]*unstructured.Unstructured) error {
	for gvkn := range m {
		obj := m[gvkn]
		objMap := obj.UnstructuredContent()
		for _, path := range o.pathConfigs {
			if !SelectByGVK(gvkn.GVK, path.GroupVersionKind) {
				continue
			}
			err := mutateField(objMap, path.Path, path.CreateIfNotPresent, o.addMap)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *MapTransformer) addMap(in interface{}) (interface{}, error) {
	m, ok := in.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%#v is expectd to be %T", in, m)
	}
	for k, v := range o.m {
		m[k] = v
	}
	return m, nil
}
