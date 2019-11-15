# Values

A **Value** in valet is a key value pair which can be defined in quite a few ways, and is accessible during more parts 
of a valet configuration to aid in customization.

## Explination

**Values** are a map (string -> string) that most commonly contains values that are string constants, as demonstrated above:

```yaml
values:
  VirtualServiceName: smh
  Namespace: sm-marketplace
  UpstreamName: sm-marketplace-smm-apiserver-8080
  UpstreamNamespace: gloo-system
```

Conventionally, `values` are written in CamelCase, as demonstrated here. 

An application may specify certain `values` are **required**, to help validate an input when trying to render an application.

`Values` work in order of precedence.

1) user values passed in via CLI or via file.
2) default values on the workflow/application. 
3) localized values to a resource.

## Semantics

As shown above, `values` can be a simple set of (string -> string), but valet allows for much more customization. 

```yaml
values:
  EnvExample: "env:EXAMPLE_ONE" # This populates the value by reading this environment variable
  CmdExample: "cmd:minikube ip" # This populates the value by running this command
  KeyExample: "key:EnvExample" # This creates a value that is an alias for another key
  TemplateExample: "template:{{ .KeyExample }}" # This creates a value by executing a go template using the other values
  FileExample: "file:$HOME/a/file/on/my/{{ .FileName }}" # This executes the template, expands the env, and then gets the content of the file 
``` 

These are the 5 special keywords that can be prefixed to `values` to get special behavior from valet.
They are the following:

* `env:`
    * this prefix tells valet to expand the value by looking it up as an environment variable
* `cmd:`
    * this prefix tells valet to execute the cmd and place stdout in place of the value
* `key:`
    * this prefix tells valet to replace this value with that of another one, identified by the key provided
* `template:`
    * this prefix tells valet to template out the string using the other `values` as inputs
* `file:`
    * this prefix is a combination of the others, with additional capability. It will expand env vars, template the value, 
    and then retrieve the contents of the file which is pointed to. That file contents will then be used as the value.

## Tags

Valet takes advantage of a go concept called struct tags. These are used when evaluating the struct `values`. They can be
accessed similarly to `json` tags.
```go
type ExampleResource struct {
	InnerValue string `yaml:"innerValue" valet:"key=Value"`
}
```

The available tags are:

* default
    * this tag sets a default value
* key
    * this tag sets the value of the field by getting the value in `values` using the provided key
* template
    * This tag specifies that the value of this struct should be templated using the available `values`

As these tags use the string related functions, they must be strings in order to function properly.

These tags are executed in the order listed above. If there is no value in the field, then the default gets added.
If there is still no default value, and there is a key present, a value is looked up for that key. After that the string
is templated. This has the potential for nested templating, but that is not recommended as other struct fields may break
if passed templates.