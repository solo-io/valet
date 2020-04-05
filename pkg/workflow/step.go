package workflow

import (
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/render"
	"github.com/solo-io/valet/pkg/step/aws"
	"github.com/solo-io/valet/pkg/step/check"
	"github.com/solo-io/valet/pkg/step/cluster"
	"github.com/solo-io/valet/pkg/step/helm"
	"github.com/solo-io/valet/pkg/step/kubectl"
	"reflect"
)

// A workflow.Step is a container for an api.Step implementation.
// Exactly one of the member pointers should be non-nil.
// This makes it easy to serialize and deserialize a workflow as yaml
type Step struct {
	DnsEntry         *aws.DnsEntry          `json:"dnsEntry,omitempty"`
	Condition        *check.Condition       `json:"condition,omitempty"`
	Curl             *check.Curl            `json:"curl,omitempty"`
	WaitForPods      *check.WaitForPods     `json:"waitForPods,omitempty"`
	EnsureCluster    *cluster.EnsureCluster `json:"ensureCluster,omitempty"`
	Apply            *kubectl.Apply         `json:"apply,omitempty"`
	CreateSecret     *kubectl.CreateSecret  `json:"createSecret,omitempty"`
	Patch            *kubectl.Patch         `json:"patch,omitempty"`
	InstallHelmChart *helm.InstallHelmChart `json:"installHelmChart,omitempty"`

	Values render.Values `json:"values,omitempty"`
	// Optional, used for identifying a specific step in a docs ref
	Id string `json:"id,omitempty"`
}

// Return the actual pointer to an api.Step implementation.
func (s *Step) Get() api.Step {
	var val interface{}
	structVal := reflect.ValueOf(s).Elem()
	structType := reflect.TypeOf(s).Elem()
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

func Apply(path string) *Step {
	return &Step{
		Apply: &kubectl.Apply{
			Path: path,
		},
	}
}

func WaitForPods(namespace string) *Step {
	return &Step{
		WaitForPods: &check.WaitForPods{
			Namespace: namespace,
		},
	}
}
