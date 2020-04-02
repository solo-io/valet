package workflow

import (
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/render"
	"github.com/solo-io/valet/pkg/step"
	"github.com/solo-io/valet/pkg/step/helm"
	"reflect"
)

// A workflow.Step is a container for an api.Step implementation.
// Exactly one of the member fields should be non-nil.
// This makes it easy to serialize and deserialize a workflow as yaml
type Step struct {
	Apply            *step.Apply            `json:"apply"`
	InstallHelmChart *helm.InstallHelmChart `json:"installHelmChart"`

	Values render.Values `json:"values"`
}

// Return the actual pointer to an api.Step implementation.
func (k *Step) Get() api.Step {
	var val interface{}
	structVal := reflect.ValueOf(k).Elem()
	structType := reflect.TypeOf(k).Elem()
	for i := 0; i < structType.NumField(); i++ {
		fieldValue := structVal.Field(i)
		if fieldValue.Kind() == reflect.Ptr {
			if !fieldValue.IsNil() {
				val = fieldValue.Interface()
				break
			}
		}
	}

	if val == nil {
		return nil
	}

	return val.(api.Step)
}
