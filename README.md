# Valet

Valet is a tool for Kubernetes operators and developers, focused on solving three related use cases:
* How do I write automated end-to-end tests for Kubernetes applications?
* How do I maintain documentation for workflows related to using my application on Kubernetes?
* How do I automatically deploy many applications and ensure the state of an entire dev or production Kubernetes cluster?

With Valet, you can write a single config and use valet to run it as an end to end test, produce a documented user guide, 
or wire up to GitOps to automatically deploy the configuration against a real cluster. 

## Installing and Running

The easiest way to use Valet is to run from source:
`git clone https://github.com/solo-io/valet && cd valet`
`make build TAGGED_VERSION=$(git describe --tags)`

Now move `_output/valet` to somewhere in your bin, or create an alias: `alias valet="$(pwd)/_output/valet"`

To get a jump start with some known configuration, clone the valet-config repo:
`git clone https://github.com/solo-io/valet-config && cd valet-config`
`valet ensure -f clusters/demos/petclinic-minikube.yaml`

## Applications

Valet provides a simple declarative API for **applications**, which consist of a list of resources such as 
**manifests**, **helm charts**, **templates**, **patches**, **secrets**, **namespaces**, and other **applications**. 
Given a set of user **values** (key-value pairs that help with rendering), and **flags** (to toggle on optional resources),
an application can be rendered and ensured on a cluster. 

Once an application is defined, it can be deployed automatically against the current Kube context 
using `valet ensure application -f path/to/application`. Valet is idempotent, so running ensure several times in a 
row should not affect the health of the application. 

Click [here](cli/internal/ensure/resource/application/README.md) for more details on the applications API. 

## Workflows

Valet also provides an API for **workflows**, which consist of a set of steps such as **installing or uninstalling applications**, 
**applying or deleting manifests**, validate the health of an application with **curl commands** and **conditions**, 
set up **dns entries**, or run other **workflows**.

Like applications, workflows in valet are rendered with a set of user **values** and **flags** that determine how 
the resources are rendered and which optional resources/steps are used. 

Once a workflow is defined, it can be executed automatically using `valet ensure -f path/to/workflow.yaml`. Like applications, 
workflows are intended to be idempotent so that they can be ensured regardless of the starting state of the cluster. 

Click [here](cli/internal/ensure/resource/workflow/README.md) for more details on the workflow API. 

## Clusters

When running `valet ensure -f path/to/workflow.yaml`, the file can contain a `cluster` section to first point the local 
Kube context to a specific cluster before executing the workflow. If the cluster doesn't exist, Valet will create it. 
Valet currently supports **Minikube**, **EKS**, and **GKE** Kubernetes clusters. 

If a cluster is not specified in the call to ensure, the workflow will be executed against the current Kube context. 

Click [here](cli/internal/ensure/resource/cluster/README.md) for more details on the workflow API. 

## Running Valet in CI 

Valet publishes a container on release to the `quay.io/solo-io/valet` Docker repo, with all the dependencies
pre-loaded. This makes it easy to invoke in a CI system, such as Google Cloud Build 
(see cloudbuild.yaml for complete example):

```yaml
steps:
  - name: 'quay.io/solo-io/valet:0.1.5'
    args: ['ensure', '-f', 'cli/internal/ensure/test/fixtures/all.yaml', '--gke-cluster-name', 'valet-$SHORT_SHA']
    id: 'valet-ensure-valet'
    env:
      - 'HELM_HOME=/root/.helm'
    secretEnv: ['GITHUB_TOKEN', 'LICENSE_KEY', 'AWS_ACCESS_KEY_ID', 'AWS_SECRET_ACCESS_KEY']
    waitFor: ['valet-build', 'valet-ensure-cluster']
```

### GitOps

Valet helps with setting up a GitOps story for your cluster config. One option is managing the valet config in a repo, 
and then running `valet ensure` in CI to ensure the config. 

In some cases, it may be preferable to store the rendered applications rather than the valet config in a GitOps repo. 
Valet supports a `dry-run` flag for this reason, and is designed to be flexible but not strongly opinionated about how 
to assemble applications and workflows. It may be preferable to use some of the great tools from the community, including
**Helm**, **Kustomize**, or weave **Flux** for synchronizing resources on a cluster; even then, Valet could be a useful
additional tool to help with automation and documentation. 



## Teardown

When you are done using a config, Valet can clean it up with `valet teardown -f path/to/config.yaml`. 
This can be useful if the cluster gets into a bad state, or if it is no longer useful. **If the config
contains a cluster definition, the cluster will be destroyed.**

For tearing down specific applications, run `valet teardown application -f path/to/application.yaml`.  

## Config

In order to save a variable that will always get loaded into the environment when valet runs, run:
`valet config set FOO=bar`. 

This writes out this value to a global config file in `$HOME/.valet/global.yaml`. This file can be edited 
directly. 

It is important to provide a few common environment variables to simplify usage of valet, especially with Solo products:
* `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`: These are necessary for valet to update Route53 DNS, and
are also used in the standard Gloo application and demos. 
* `GOOGLE_APPLICATION_CREDENTIALS`: This is necessary to manage Kubernetes clusters in Google Cloud
* `LICENSE_KEY`: This is necessary for installing Enterprise Gloo
* `GITHUB_TOKEN`: This is necessary for deploying Service Mesh Hub

## Build (beta)

Valet offers the ability to build artifacts from any repo. 
It currently supports building Go binaries, tagging and pushing docker containers, 
and producing Helm charts and manifests. Valet typically determines the artifacts to 
build by reading a file called `artifacts.yaml` in the root of the repo. An example 
is included in [this repo](artifacts.yaml) for building Valet itself. 

Valet artifacts are always stored locally in the `_artifacts` directory at the root of the repo. 

Valet artifacts are versioned, with the version provided at runtime. 
This version may be a semver version (i.e. 1.2.3), a SHA, or any other version. 
Typically, when running locally, you would use git to determine a meaningful, 
unique, and stable version:

`valet build -v $(git describe --tags --dirty | cut -c 2-)`

If your repo does not have any tags, then you may use this instead:

`valet build -v $(git describe --always --dirty)`

Artifacts can be tagged for upload to the google storage valet bucket 
by adding `upload: true` to the binary or helm chart. 
(NOTE: this requires google storage writer permissions on the valet bucket.) 
