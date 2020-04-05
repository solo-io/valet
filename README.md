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

## Values

TODO

## Global Config

TODO