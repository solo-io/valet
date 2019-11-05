package resource

import (
	"fmt"
	"github.com/solo-io/go-utils/errors"
	"os"
	"strings"
)

const (
	VersionKey    = "Version"
	NamespaceKey  = "Namespace"
	DomainKey     = "Domain"
	HostedZoneKey = "HostedZone"
)

var (
	ValueNotFoundError = func(key string) error {
		return errors.Errorf("Value %s not provided", key)
	}
	InvalidValueError  = errors.Errorf("Invalid value")
	RequiredValueNotProvidedError = func(key string) error {
		return errors.Errorf("Required value %s not found", key)
	}
)

type InputValue struct {
	Value    string `yaml:"value"`
	Env      string `yaml:"env"`
	Key      string `yaml:"key"`
	Template string `yaml:"template"`
}

func (i InputValue) ToString() string {
	if i.Value != "" {
		return i.Value
	}
	if i.Env != "" {
		return fmt.Sprintf("env:%s", i.Env)
	}
	if i.Key != "" {
		return fmt.Sprintf("key:%s", i.Key)
	}
	if i.Template != "" {
		return fmt.Sprintf("template:%s", i.Template)
	}
	return ""
}

type Values map[string]InputValue

func MergeValues(merge, with Values) Values {
	if with == nil {
		with = make(map[string]InputValue)
	}
	for k, v := range merge {
		with[k] = v
	}
	return with
}

func (v Values) ContainsKey(key string) bool {
	if v == nil {
		return false
	}
	_, ok := v[key]
	return ok
}

func (v Values) GetValue(key string) (string, error) {
	val, ok := v[key]
	if !ok {
		return "", ValueNotFoundError(key)
	}
	if val.Value != "" {
		return val.Value, nil
	}
	if val.Key != "" {
		return v.GetValue(val.Key)
	}
	if val.Env != "" {
		return os.Getenv(val.Env), nil
	}
	if val.Template != "" {
		return LoadTemplate(val.Template, v)
	}
	return "", InvalidValueError
}

func (v Values) ToString() string {
	var entries []string
	for k, v := range v {
		entries = append(entries, fmt.Sprintf("%s=%s", k, v.ToString()))
	}
	return fmt.Sprintf("{%s}", strings.Join(entries, ", "))
}
