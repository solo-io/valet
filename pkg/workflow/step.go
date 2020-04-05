package workflow

import (
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/render"
	"github.com/solo-io/valet/pkg/step/helm"
	"github.com/solo-io/valet/pkg/step/kubectl"
	"github.com/solo-io/valet/pkg/step/validation"
	"reflect"
)

// A workflow.Step is a container for an api.Step implementation.
// Exactly one of the member pointers should be non-nil.
// This makes it easy to serialize and deserialize a workflow as yaml
type Step struct {
	Apply            *kubectl.Apply         `json:"apply,omitempty"`
	CreateSecret     *kubectl.CreateSecret  `json:"createSecret,omitempty"`
	InstallHelmChart *helm.InstallHelmChart `json:"installHelmChart,omitempty"`

	Curl        *validation.Curl `json:"curl,omitempty"`
	WaitForPods *validation.WaitForPods `json:"waitForPods,omitempty"`

	Values render.Values `json:"values,omitempty"`
	// Optional, used for identifying a specific step in a docs ref
	Id string `json:"id,omitempty"`
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

func (s *Step) WithId(id string) *Step {
	s.Id = id
	return s
}

func ApplyWithId(id, path string) *Step {
	return &Step{
		Id: id,
		Apply: &kubectl.Apply{
			Path: path,
		},
	}
}

func Apply(path string) *Step {
	return &Step{
		Apply: &kubectl.Apply{
			Path: path,
		},
	}
}

func WaitForPods(namespace string) *Step {
	return &Step{
		WaitForPods: &validation.WaitForPods{
			Namespace: namespace,
		},
	}
}