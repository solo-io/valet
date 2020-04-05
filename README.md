# Valet

Valet makes it easy to define a **workflow**, consisting of a set of steps that deploy, configure, and test applications
on Kubernetes. Valet enables you to use the same exact workflow definitions for testing, demos, and documentation.

Valet workflows are typically stored in code or yaml in a git repository. Valet workflows may have secrets and templates 
that are populated at runtime using `Values`, which can be defined from environment variables, files, user input, or other 
sources. Locally, you may add variables to your global valet config that will be loaded into the environment at runtime. 

It is easy to write a new type of step, and more step types will be introduced over time as needs arise. 

## Use Cases

Valet has a few intended use cases:
* Simplify writing automation in Go that simulate real Kubernetes workflows for deploying, configuring, and testing products
* Simplify setting up and distributing complex product demos
* Improving the quality of - and preventing regressions in - documented user workflows

### Writing workflows in Go

Valet workflows can be written in Go to automate workflows and end-to-end tests. Here is an example of a workflow 
that starts with installing a helm chart, deploying a few manifests, and testing a service with curl: 

```go
installGloo := &workflow.Step{
        InstallHelmChart: &helm.InstallHelmChart{
            ReleaseName: "gloo",
            ReleaseUri:  "https://storage.googleapis.com/solo-public-helm/charts/gloo-1.3.17.tgz",
            Namespace:   "gloo-system",
            WaitForPods: true,
        },
    }
}
initialCurl := &workflow.Step{
    Curl: &check.Curl{
        Service:    gatewayProxy(),
        Path:       "/",
        StatusCode: 200,
    },
}
workflow := &workflow.Workflow{
    Steps: []*workflow.Step{
        installGloo,
        // Part 1: Deploy the monolith
        workflow.Apply("petclinic.yaml"),
        workflow.WaitForPods("default"),
        workflow.Apply("vs-1.yaml"),
        initialCurl,
        // ...
    },
}
err := workflow.Run(workflow.DefaultContext(), nil)
// Utilizing gomega for test assertions
Expect(err).To(BeNil())
```

A more complete example is available [here](test/e2e/gloo-petclinic/petclinic_test.go).

### Automating and distributing demo workflows

Valet workflows can be written in or serialized to yaml. The workflow above is available in yaml form 
[here](test/e2e/gloo-petclinic/workflow.yaml).

Use the `valet` command line tool to run a yaml workflow (`valet run -f workflow.yaml`).

### Improving workflow documentation

The `valet` command line tool includes a `gen-docs` command that can read a template (markdown) file and output 
a markdown file with the templates replaced by content from valet workflows. 

Here is an example of a docs template:
```
{{%valet 
workflow: workflow.yaml
step: deploy-monolith
%}}
```

`valet gen-docs ...` will replace that with documentation from the step that is referenced. In this example, the 
replacement is:
```
kubectl apply -f petclinic.yaml
```

This helps write workflows that match your automated tests, by using the workflow as the source of truth. Docs 
can also include flags that may result in a different output. For instance, we could have written this into the template
instead:
```
{{%valet 
workflow: workflow.yaml
step: deploy-monolith
flags:
  - YamlOnly
%}}
```

When generating docs for a step that applies a manifest, the `YamlOnly` flag tells the renderer to just show the
raw yaml, rather than the kubectl command. 

For a complete example, check out the template [here](test/e2e/gloo-petclinic/template.md), which was rendered into 
[this](test/e2e/gloo-petclinic/README.md)

## Values

A **Value** in valet is a key value pair which can be defined in quite a few ways, and is accessible during more parts 
of a valet configuration to aid in customization.

### Explanation

**Values** are a map (string -> string) that most commonly contains values that are string constants, as demonstrated below:

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

### Value prefixes

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
    * files in this context have one limitation compared to file refs in the rest of valet. Any file path here
    has to be relative to the root directory in which valet is run. This is issue is being tracked
    [here](https://github.com/solo-io/valet/issues/122)

### Tags

Workflow steps are represented as structs in Go, and some struct fields have special valet tags:
```go
type ExampleResource struct {
	InnerValue string `json:"innerValue" valet:"key=Value"`
}
```

Prior to execution, a workflow is rendered, including updating the value of struct fields based on their tags and 
the runtime values available. The available tags are:

* default
    * this tag sets a default value
* key
    * this tag sets the value of the field by getting the value in `values` using the provided key
* template
    * This tag specifies that the value of this struct should be templated using the available `values`

These tags are executed in the order listed above. If there is no value in the field, then the default gets added.
If there is still no default value, and there is a key present, a value is looked up for that key. After that the string
is templated. This has the potential for nested templating, but that is not recommended as other struct fields may break
if passed templates.

## Global Config

In order to save a variable that will always get loaded into the environment when valet runs, run:
`valet config set FOO=bar`. 

This writes out this value to a global config file in `$HOME/.valet/global.yaml`. This file can be edited 
directly. To use a different global config location, set `--global-config-path`. 

Often, this is a good place to store environment variables for things like credentials, so they can be left 
out of the workflow. In CI workflows, the environment variable can be provided in the preferred way depending 
on the CI tool. 