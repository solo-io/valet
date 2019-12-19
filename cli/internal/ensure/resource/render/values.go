package render

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/stringutils"
	cmd_runner "github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	VersionKey    = "Version"
	NamespaceKey  = "Namespace"
	DomainKey     = "Domain"
	HostedZoneKey = "HostedZone"
	PathKey       = "Path"
	NameKey       = "Name"

	EnvPrefix      = "env:"
	TemplatePrefix = "template:"
	KeyPrefix      = "key:"
	CmdPrefix      = "cmd:"
	FilePrefix     = "file:"

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

func (v Values) DeepCopy() Values {
	values := make(map[string]string)
	for k, val := range v {
		values[k] = val
	}
	return values
}

func (v Values) Load(tmpl string, runner cmd_runner.Runner) (string, error) {
	parsed, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}
	out := strings.Builder{}
	vals, err := v.Render(runner)
	if err != nil {
		return "", err
	}
	err = parsed.Execute(&out, vals)
	return out.String(), err
}

func (v Values) Render(runner cmd_runner.Runner) (map[string]interface{}, error) {
	vals := make(map[string]interface{})
	for k := range v {
		v, err := v.GetValue(k, runner)
		if err != nil {
			return nil, err
		}
		vals[k] = v
	}
	return vals, nil
}

func (v Values) ContainsKey(key string) bool {
	if v == nil {
		return false
	}
	_, ok := v[key]
	return ok
}

func (v Values) GetValue(key string, runner cmd_runner.Runner) (string, error) {
	val, ok := v[key]
	if !ok {
		return "", ValueNotFoundError(key)
	}
	if strings.HasPrefix(val, KeyPrefix) {
		key := strings.TrimPrefix(val, KeyPrefix)
		return v.GetValue(key, runner)
	} else if strings.HasPrefix(val, TemplatePrefix) {
		tmpl := strings.TrimPrefix(val, TemplatePrefix)
		otherVals := v.DeepCopy()
		delete(otherVals, key)
		return LoadTemplate(tmpl, otherVals, runner)
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
			return runner.Output(context.TODO(), cmd)
		default:
			cmd := &cmd_runner.Command{
				Name: splitCmd[0],
				Args: splitCmd[1:],
			}
			return runner.Output(context.TODO(), cmd)
		}
	} else if strings.HasPrefix(val, FilePrefix) {
		fileString := strings.TrimPrefix(val, FilePrefix)
		otherVals := v.DeepCopy()
		delete(otherVals, key)
		path, err := LoadTemplate(fileString, otherVals, runner)
		if err != nil {
			return "", err
		}
		byt, err := ioutil.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(byt), nil
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

func (v Values) RenderFields(input interface{}, runner cmd_runner.Runner) error {
	structVal := reflect.ValueOf(input).Elem()
	structType := reflect.TypeOf(input).Elem()
	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		valetTags := strings.Split(fieldType.Tag.Get(ValetField), ",")
		fieldValue := structVal.Field(i)
		if fieldValue.Kind() == reflect.String {
			rendered := fieldValue.String()

			if key := getTagValue(valetTags, KeyTag); key != "" && rendered == "" && v.ContainsKey(key) {
				val, err := v.GetValue(key, runner)
				if err != nil {
					return err
				}
				rendered = val
			}

			if defaultValue := getTagValue(valetTags, DefaultTag); defaultValue != "" && rendered == "" {
				rendered = defaultValue
			}

			if stringutils.ContainsString(TemplateTag, valetTags) {
				loaded, err := LoadTemplate(rendered, v, runner)
				if err != nil {
					return err
				}
				rendered = loaded
			}

			fieldValue.SetString(rendered)
		} else if fieldValue.Kind() == reflect.Int {
			if fieldValue.Int() == 0 {
				rendered := getTagValue(valetTags, DefaultTag)
				if rendered == "" {
					continue
				}
				val, err := strconv.Atoi(rendered)
				if err != nil {
					return err
				}
				fieldValue.SetInt(int64(val))
			}
		} else if fieldValue.Kind() == reflect.Struct {
			if err := v.RenderFields(fieldValue.Addr().Interface(), runner); err != nil {
				return err
			}
		} else if fieldValue.Kind() == reflect.Ptr {
			originalValue := fieldValue.Elem()
			if !originalValue.IsValid() || originalValue.Kind() != reflect.Struct {
				continue
			}
			if err := v.RenderFields(fieldValue.Interface(), runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v Values) RenderValues(runner cmd_runner.Runner) (map[string]interface{}, error) {
	vals := make(map[string]interface{})
	for k := range v {
		v, err := v.GetValue(k, runner)
		if err != nil {
			return nil, err
		}
		vals[k] = v
	}
	return vals, nil
}

func (v Values) RenderStringValues(runner cmd_runner.Runner) (map[string]string, error) {
	vals := make(map[string]string)
	for k := range v {
		v, err := v.GetValue(k, runner)
		if err != nil {
			return nil, err
		}
		vals[k] = v
	}
	return vals, nil
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
