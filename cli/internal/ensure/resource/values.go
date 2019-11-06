package resource

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/solo-io/go-utils/errors"
	cmd_runner "github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	VersionKey    = "Version"
	NamespaceKey  = "Namespace"
	DomainKey     = "Domain"
	HostedZoneKey = "HostedZone"

	EnvPrefix      = "env:"
	TemplatePrefix = "template:"
	KeyPrefix      = "key:"
	CmdPrefix      = "cmd:"
)

var (
	ValueNotFoundError = func(key string) error {
		return errors.Errorf("Value %s not provided", key)
	}
	RequiredValueNotProvidedError = func(key string) error {
		return errors.Errorf("Required value %s not found", key)
	}
)

type Values map[string]string

func MergeValues(merge, with Values) Values {
	if with == nil {
		with = make(map[string]string)
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
	if strings.HasPrefix(val, KeyPrefix) {
		key := strings.TrimPrefix(val, KeyPrefix)
		return v.GetValue(key)
	} else if strings.HasPrefix(val, TemplatePrefix) {
		template := strings.TrimPrefix(val, TemplatePrefix)
		return LoadTemplate(template, v)
	} else if strings.HasPrefix(val, EnvPrefix) {
		env := strings.TrimPrefix(val, EnvPrefix)
		return os.Getenv(env), nil
	} else if strings.HasPrefix(val, CmdPrefix) {
		cmdString := strings.TrimPrefix(val, CmdPrefix)
		splitCmd := strings.Split(cmdString, " ")
		switch len(splitCmd) {
		case 0:
			return "", nil
		case 1:
			cmd := &cmd_runner.Command{
				Name: splitCmd[0],
			}
			return cmd.Output(context.TODO())
		default:
			cmd := &cmd_runner.Command{
				Name: splitCmd[0],
				Args: splitCmd[1:],
			}
			return cmd.Output(context.TODO())
		}
	} else {
		return val, nil
	}
}

func (v Values) ToString() string {
	var entries []string
	for k, v := range v {
		entries = append(entries, fmt.Sprintf("%s=%s", k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(entries, ", "))
}
