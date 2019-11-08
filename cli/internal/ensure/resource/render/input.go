package render

type InputParams struct {
	Values Values
	Flags  Flags
	Step   bool
}

func (i *InputParams) DeepCopy() InputParams {
	var flags []string
	flags = append(flags, i.Flags...)
	values := make(map[string]string)
	for k, v := range i.Values {
		values[k] = v
	}
	return InputParams{
		Flags:  flags,
		Values: values,
		Step:   i.Step,
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
