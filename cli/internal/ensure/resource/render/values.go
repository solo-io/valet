package render

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/stringutils"
	"os"
	"reflect"
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

	ValetField  = "valet"
	TemplateTag = "template"
	DefaultTag  = "default"
	KeyTag      = "key"
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

func (v Values) RenderFields(input interface{}) error {
	structVal := reflect.ValueOf(input).Elem()
	structType := reflect.TypeOf(input).Elem()
	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		fieldValue := structVal.Field(i)
		if fieldValue.Kind() == reflect.String {
			rendered := fieldValue.String()
			valetTags := strings.Split(fieldType.Tag.Get(ValetField), ",")
			if stringutils.ContainsString(TemplateTag, valetTags) {
				loaded, err := LoadTemplate(rendered, v)
				if err != nil {
					return err
				}
				rendered = loaded
			}
			if rendered == "" {
				key := getTagValue(valetTags, KeyTag)
				if key != "" && v.ContainsKey(key) {
					val, err := v.GetValue(key)
					if err != nil {
						return err
					}
					rendered = val
				}
			}
			if rendered == "" {
				rendered = getTagValue(valetTags, DefaultTag)
			}
			fieldValue.SetString(rendered)
		}
	}
	return nil
}

func getTagValue(fieldTags []string, tag string) string {
	prefix := fmt.Sprintf("%s=", tag)
	for _, fieldTag := range fieldTags {
		if strings.HasPrefix(fieldTag, prefix) {
			return strings.TrimPrefix(fieldTag, prefix)
		}
	}
	return ""
}

type Flags []string

func (f Flags) ToString() string {
	return fmt.Sprintf("[%s]", strings.Join(f, ", "))
}
