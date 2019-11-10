package render

type InputParams struct {
	Values     Values
	Flags      Flags
	Step       bool
	Registries map[string]Registry
}

func (i *InputParams) LoadFile(registryName, path string) (string, error) {
	registry, err := i.GetRegistry(registryName)
	if err != nil {
		return "", err
	}
	return registry.LoadFile(path)
}

func (i *InputParams) DeepCopy() InputParams {
	var flags []string
	flags = append(flags, i.Flags...)
	values := make(map[string]string)
	for k, v := range i.Values {
		values[k] = v
	}
	return InputParams{
		Flags:      flags,
		Values:     values,
		Step:       i.Step,
		Registries: i.Registries,
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
		return &LocalRegistry{}, nil
	}
	return nil, UnknownRegistryError(name)
}

func (i *InputParams) SetRegistry(name string, registry Registry) {
	if i.Registries == nil {
		i.Registries = make(map[string]Registry)
	}
	i.Registries[name] = registry
}
