package render

import (
	"github.com/solo-io/valet/cli/internal/ensure/client"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type InputParams struct {
	Values        Values
	Flags         Flags
	Step          bool
	Registries    map[string]Registry
	CommandRunner cmd.Runner
	IngressClient client.IngressClient
	DnsClient     client.AwsDnsClient
}

func (i *InputParams) LoadFile(registryName, path string) (string, error) {
	registry, err := i.GetRegistry(registryName)
	if err != nil {
		return "", err
	}
	path, err = LoadTemplate(path, i.Values, i.Runner())
	if err != nil {
		return "", err
	}
	return registry.LoadFile(path)
}

func (i *InputParams) DeepCopy() InputParams {
	var flags []string
	flags = append(flags, i.Flags...)
	values := i.Values.DeepCopy()
	registries := make(map[string]Registry)
	for k, v := range i.Registries {
		registries[k] = v
	}
	return InputParams{
		Flags:         flags,
		Values:        values,
		Step:          i.Step,
		Registries:    registries,
		CommandRunner: i.CommandRunner,
		IngressClient: i.IngressClient,
	}
}

func (i *InputParams) MergeValues(values Values) InputParams {
	output := i.DeepCopy()
	for k, v := range values {
		if !output.Values.ContainsKey(k) {
			output.Values[k] = v
		}
	}
	return output
}

func (i *InputParams) MergeFlags(flags Flags) InputParams {
	output := i.DeepCopy()
	for _, flag := range flags {
		found := false
		for _, existingFlag := range output.Flags {
			if flag == existingFlag {
				found = true
				break
			}
		}
		if !found {
			output.Flags = append(output.Flags, flag)
		}
	}
	return output
}

func (i *InputParams) GetRegistry(name string) (Registry, error) {
	if i.Registries == nil {
		i.Registries = make(map[string]Registry)
	}
	if reg, ok := i.Registries[name]; ok {
		return reg, nil
	}
	if name == DefaultRegistry {
		return &DirectoryRegistry{}, nil
	}
	return nil, UnknownRegistryError(name)
}

func (i *InputParams) SetRegistry(name string, registry Registry) {
	if i.Registries == nil {
		i.Registries = make(map[string]Registry)
	}
	i.Registries[name] = registry
}

func (i *InputParams) Runner() cmd.Runner {
	if i.CommandRunner == nil {
		return cmd.DefaultCommandRunner()
	}
	return i.CommandRunner
}

func (i *InputParams) GetIngressClient() client.IngressClient {
	if i.IngressClient == nil {
		return &client.KubeIngressClient{}
	}
	return i.IngressClient
}

func (i *InputParams) GetDnsClient() (client.AwsDnsClient, error) {
	if i.DnsClient == nil {
		return client.NewAwsDnsClient()
	}
	return i.DnsClient, nil
}

func (i *InputParams) RenderFields(input interface{}) error {
	return i.Values.RenderFields(input, i.Runner())
}
