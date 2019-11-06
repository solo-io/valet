# Valet

Valet is a tool for Kubernetes operators to easily assemble and manage many applications across a 
Kubernetes cluster. Today, it is still painful and time consuming for operators to manage clusters, despite
powerful tools in the community:
* `Kubectl` gives you the ability to view or manage any resource on a cluster, among many other things
* `Helm` gives you the ability to package complex applications into a single, configurable chart, and deploy those charts
* `Kustomize` gives the the ability to patch a Kubernetes manifest before deploying to adjust it for the particular deployment environment

Valet introduces the concept of a declarative **Application**, consisting of a list of resources 
(namespaces, secrets, helm charts, manifests, templates, patches, and other applications). Given a set of 
user values, Valet can render an application into an exact Kubernetes manifest. Valet can **ensure**
applications on a cluster, dry run to output the rendered manifest without touching a cluster, and 
**teardown** (uninstall) the application. 

Valet also introduces the concept of a **Workflow**, consisting of a set of steps that are run in order on a cluster. 
Steps could include installing or uninstalling an **Application**, setting up a DNS entry, curling an endpoint, 
waiting for a Kubernetes resource to meet a condition, or running another **Workflow**.  

With **Applications** and **Workflows**, Valet enables Kubernetes operators to:
* Ensure the state of an entire cluster, with a simple declarative configuration, with native GitOps support
* Easily customize an application by bundling extra resources and custom values, including secrets
* Easily automate demo and test workflows, or replicate real environments in a staging cluster
* Organize DevOps across many engineering teams, without having to write custom helm charts or automation 

## Applications

An **Application** in Valet is a declarative configuration that specifies a set of resources that can be rendered
into an exact Kubernetes manifest, given a set of (possibly empty) user **values** and **flags**. 

Here's an example application configuration for the service-mesh-hub application:

```yaml
name: service-mesh-hub
requiredValues:
  - Version
values:
  Namespace: sm-marketplace
resources:
  - namespace:
      name: sm-marketplace
  - secret:
      name: github-token
      entries:
        token:
          envVar: GITHUB_TOKEN
  - helmChart:
      repoUrl: "https://storage.googleapis.com/sm-marketplace-helm/"
      chartName: sm-marketplace
      repoName: sm-marketplace
      set:
        - "namespace.create=true"
  - application:
      path: registry/apps/gloo-app.yaml
    values:
      VirtualServiceName: smh
      Namespace: sm-marketplace
      UpstreamName: sm-marketplace-smm-apiserver-8080
      UpstreamNamespace: gloo-system
    flags:
      - gloo-app
```

This is an application that consists of a namespace, a secret, a helm chart, and another application that deploys a Gloo
virtual service to create an ingress route to the application.

Let's look section by section: 

```yaml
name: service-mesh-hub
```

Currently applications have a name field, that is only used for convenience. In the future, this may be used to identify 
applications in a registry. 

```yaml
requiredValues:
  - Version
```

This application requires a value called `Version` passed in. Values are passed through the application to all the resources, 
and in this case the version is used to determine which helm chart tag to use. By calling the value required, `Valet` can 
return a useful error to the user when it is missing. 

```yaml
values:
  Namespace: sm-marketplace
```

By default, this application uses `sm-marketplace` as the namespace value. 

```yaml
resources:
  - namespace:
      name: sm-marketplace
  - secret:
      name: github-token
      entries:
        token:
          envVar: GITHUB_TOKEN
```

Before the helm chart is applied, service-mesh-hub wants a github token to be present in the install namespace. Here, we 
create that by grabbing the token from an environment variable. 

```yaml
  - helmChart:
      repoUrl: "https://storage.googleapis.com/sm-marketplace-helm/"
      chartName: sm-marketplace
      repoName: sm-marketplace
      set:
        - "namespace.create=true"
```

After the namespace and token are created, the helm chart is applied. In this case, there is only one value provided via a set command;
a values file could be provided instead. 
```yaml
  - application:
      path: registry/apps/gloo-app.yaml
    values:
      VirtualServiceName: smh
      Namespace: sm-marketplace
      UpstreamName: sm-marketplace-smm-apiserver-8080
      UpstreamNamespace: gloo-system
    flags:
      - gloo-app
```

The last resource in the application is a Gloo virtual service to create a route to the service mesh hub UI. The gloo-app.yaml
is a generic application that is reused for other products, so specific values are provided to construct the service mesh hub 
virtual service. This resource has a flag `gloo-app`, which must be provided or the resource will be omitted during rendering. 

### Values

A user can provide two types of inputs when specifying a workflow, cluster config, or ensuring a specific application: 
**values** and **flags**. 

**Values** are a map (string -> string) that most commonly contains values that are string constants, as demonstrated above:

```yaml
values:
  VirtualServiceName: smh
  Namespace: sm-marketplace
  UpstreamName: sm-marketplace-smm-apiserver-8080
  UpstreamNamespace: gloo-system
```

However, there are a few prefixes that enable specifying other types of values:

```yaml
values:
  EnvExample: "env:EXAMPLE_ONE" # This populates the value by reading this environment variable
  CmdExample: "cmd:minikube ip" # This populates the value by running this command
  KeyExample: "key:EnvExample" # This creates a value that is an alias for another key
  TemplateExample: "template:{{ .KeyExample }}" # This creates a value by executing a go template using the other values
```

Conventionally, values are written in CamelCase, as demonstrated here. 

An application may specify certain values are **required**, to help validate an input when trying to render an application. 

### Flags

**Flags** are string labels that can be associated with certain resources in an application. If a flag on a resource is not
provided by a user, that resource is omitted during rendering.  

```yaml
  - application:
      path: registry/apps/gloo-app.yaml
    values:
      VirtualServiceName: smh
      Namespace: sm-marketplace
      UpstreamName: sm-marketplace-smm-apiserver-8080
      UpstreamNamespace: gloo-system
    flags:
      - gloo-app
```

In this example, this resource (an application) will only be rendered if the `gloo-app` flag is provided. 

### Application Resources

Valet defines a `Resource` interface that requires specifying an `Ensure` and `Teardown` function. For application resources,
they must also be fully renderable without deploying any resources to a cluster, to enable `dry-run`. 

Valet currently supports the following types of application resources: **yaml manifests**, **helm charts**, 
**secrets**, **templates**, **patches**, **namespaces**, and **applications**. 

Valet will render resources for an application in order. When rendering, it will pass user values and flags to each 
resource and, depending on the resource, may update certain fields on the resource based on the values.  

#### Manifests

Valet supports applying a manifest as an application resource. 

```yaml
name: manifest-example
resources:
  - path: path/to/manifest.yaml
``` 

This reads the manifest into `stdin` and then runs `kubectl apply -f -`. 

#### Helm Charts

Valet supports rendering a helm chart with custom values, deploying it, and waiting for all of the pods to become ready. 

```yaml
name: helm-chart-example
resources:
  - helmChart:
      repoUrl: "https://storage.googleapis.com/gloo-ee-helm/"
      chartName: gloo-ee
      repoName: gloo-ee
      version: 0.20.4
      namespace: gloo-system
      set:
        - "gloo.namespace.create=true"
        - "grafana.defaultInstallationEnabled=false"
        - "gloo.gatewayProxies.gatewayProxyV2.readConfig=true"
      setEnv:
        license_key: LICENSE_KEY
```

This will download the helm chart, untar it, run `helm template` to render the manifest with the provided values, 
and then apply it. Currently this doesn't support a custom values file, but that can be easily added. 

Instead of defining here, provide `Version` and `Namespace` as values to specify them elsewhere. This can dramatically
reduce the amount of work to maintain a cluster configuration over many upgrades. 

Valet will wait for all the pods in the namespace are ready before finishing ensuring the helm chart resource. 

#### Secrets

Often, when deploying an application, it is necessary to populate some secrets. Valet provides a flexible approach to
defining secrets to support GitOps workflows where secrets often can't (or shouldn't) be committed to a repo.

```yaml
name: secret-example
resources:
  - secret:
      name: aws-creds
      namespace: example
      entries:
        aws_access_key_id:
          envVar: AWS_ACCESS_KEY_ID
        aws_secret_access_key:
          envVar: AWS_SECRET_ACCESS_KEY
  - secret:
      name: gcloud-example
      namespace: example
      entries:
        private-key:
          gcloudKmsEncryptedFile:
            ciphertextFile: cluster/approval-bot/private-key.enc
            gcloudProject: solo-corp
            keyring: build
            key: buildkey
        constant:
          file: path/to/file.ext
``` 

In this example, two secrets are created. The first is a set of aws credentials, pulled from the environment. The 
second is an example using two other ways to populate secret values with file contents. Valet supports
decrypting files using gcloud kms decryption. Note that valet needs to be configured with the proper google credentials. 
Valet also supports reading an unencrypted file's contents as a secret value. 

In the future, this can be extended to support other secret providers. It was important for Valet to not rely on 
the cluster to decrypt the secret, since the cluster may be created at the same time the secret is created -- making it impossible
to save the encrypted secret in a repo for GitOps. However, valet could be used with something like SealedSecrets for 
secret management, and rely on cluster creation and secret encryption to be handled elsewhere. 

#### Templates

In the effort to support reuse of configuration, valet supports a simple form of templating similar to Helm. A valet 
template is a single yaml file that contains a manifest, with certain values extracted and referenced with go template
`{{ .VariableName }}` syntax. Then, the resource or application can provide the value for that variable, and valet 
will render the template first before applying it. 

An application may define one or more templates as resources:

```yaml
name: template-example
resources:
  - template: 
      path: path/to/template.yaml
```

Consider the following template file:

```yaml
apiVersion: certmanager.k8s.io/v1alpha1
kind: Certificate
metadata:
  name: {{ .Domain }}
  namespace: {{ .Namespace }}
spec:
  secretName: {{ .Domain }}
  dnsNames:
    - {{ .Domain }}
  acme:
    config:
      - dns01:
          provider: route53
        domains:
          - {{ .Domain }}
  issuerRef:
    name: letsencrypt-dns-prod
    kind: ClusterIssuer
```

This is a template for a certificate that extracts `Domain` and `Namespace` variables. These can be provided to the application 
in a few ways:

1. As `values` on the template directly

```yaml
name: template-example
resources:
  - template: 
      path: path/to/template.yaml
      values: 
        Domain: foo
      	Namespace: "env:NAMESPACE"
```

2. As `values` on the application

```yaml
name: template-example
values:
  Domain: foo
  Namespace: bar
resources:
  - template: 
      path: path/to/template.yaml
```

3. As inherited values from the parent application or config

When relying on inherited values, it is useful to specify those values as required by the application:

```yaml
name: template-example
requiredValues: 
  - Domain
  - Namespace
resources:
  - template: 
      path: path/to/template.yaml
```

#### Patches

Typically it is desirable to deploy the correct resource first in a declarative way, rather than rely on deploying 
an incorrect resource first, then fixing it. However, this isn't always realistic, and valet provides a mechanism to 
expose `kubectl patch` using patch resources.

Consider this patch, which adds a volume to a proxy container with istio certs:

```yaml
spec:
  template:
    spec:
      containers:
        - name: gateway-proxy-v2
          volumeMounts:
            - mountPath: /etc/certs/
              name: istio-certs
              readOnly: true
      volumes:
        - name: istio-certs
          secret:
            defaultMode: 420
            optional: true
            secretName: istio.default
```

To add this patch as a resource in an application, use:

```yaml
name: patch-example
resources:
  - patch:
      path: path/to/patch.yaml
      kubeType: deployment
      name: gateway-proxy-v2
      namespace: gloo-system
      patchType: strategic
```

Note that the semantics of patching can be complex, and `strategic` is not always a desirable patch type. The goal here
was to expose the Kubernetes semantics, for more background check out [these docs](https://kubernetes.io/docs/tasks/run-application/update-api-object-kubectl-patch/).

A patch can omit the `name` and `namespace` field if `Name` and `Namespace` are provided as values instead. 

#### Namespaces

It may be desirable to create namespaces as part of managing an application, or to update existing namespaces with a label. 
Valet supports defining namespaces just like other types of resources:

```yaml
name: ns-example
resources:
  - namespace:
      name: ns
      labels:
        istio-injection: enabled
```

This ensures there's a namespace called `ns` with the istio injection label. Valet will use `kubectl apply` to create this, 
so that it will be upserted. 

#### Applications

Valet naturally supports nesting applications as resources inside an application. This makes it very easy to extend 
applications and maximize the re-usability of configuration. An application is referenced just as it is in the valet
config:

```yaml
name: application-dependency-example
resources:
  - application:
      path: path/to/other/app.yaml
      values:
        Foo: bar
``` 

Values can be provided when specifying the resource, or they can be inherited, or both. Just as before, inherited values 
override values defined on the resource, to simplify things for the application assembler. 

## Workflows

Valet supports defining **Workflows**, which are like **Applications** but with slightly different semantics. 
**Workflows** consist of a list of **Steps**, that are performed in order. Unlike **Applications**, **Steps** 
in a **Workflow** may depend on side effects of previous steps, and thus may not be executed as a meaningful dry-run
without affecting the cluster. 

For instance, after installing Gloo it may be desirable to set up a DNS entry for the Gloo proxy service. However, 
this requires (a) modifying another system outside Kubernetes, and (b) waiting for the service to be assigned an 
external IP. 

Like **Applications**, users can provide **Values** and **Flags** to workflows to help influence how the steps are 
executed. 

Here's an example workflow that installs Gloo Enterprise with a DNS entry to the proxy, and (if a flag is provided)
a virtual service to create a route to the Gloo UI:

```yaml
requiredValues:
  - Version
  - DnsHostedZone
  - DnsDomain
  - GlooUiDomain
values:
  Namespace: gloo-system
steps:
  - install:
      path: registry/apps/gloo-enterprise.yaml
      values:
        UpstreamName: gloo-system-apiserver-ui-8080
        UpstreamNamespace: "key:Namespace"
        VirtualServiceName: glooui
        Domain: "key:GlooUiDomain"
      flags:
        - glooui-app
  - dnsEntry:
      service:
        namespace: gloo-system
        name: gateway-proxy-v2
    values:
      HostedZone: "key:DnsHostedZone"
      Domain: "key:DnsDomain"
```

### Steps

Valet supports workflows that include **installing and uninstalling applications**, setting up **dns entries**, 
waiting for a **condition** to be met on a Kubernetes resource, waiting for a **curl** request to return an 
expected status, or running another **workflow**. 

#### Installing applications

A common workflow step is installing (or uninstalling) an application. Here's an example of referencing an application, 
providing values and flags. 

```yaml
steps:
  - install:
      path: registry/apps/gloo-enterprise.yaml
      values:
        UpstreamName: gloo-system-apiserver-ui-8080
        UpstreamNamespace: "key:Namespace"
        VirtualServiceName: glooui
        Domain: "key:GlooUiDomain"
      flags:
        - gloo-app
```

#### DNS entries

A common use case when setting up clusters is to define DNS mappings from new domain names to IP addressed exposed by
services in the cluster. Valet currently supports this with Amazon Route53 DNS. 

```yaml
steps: 
  - dnsEntry:
      domain: example.my.hosted.zone
      hostedZone: my.hosted.zone.
      service:
        namespace: gloo-system
        name: gateway-proxy-v2
```

This creates a DNS mapping for `example.my.hosted.zone` to the Gloo proxy. `domain` is not required - it can instead 
be provided by a value:

```yaml
application: dns-example
requiredValues:
  - Domain
resources: 
  - dnsEntry:
      hostedZone: my.hosted.zone.
      service:
        namespace: gloo-system
        name: gateway-proxy-v2
```

Note that the environment variables for AWS creds must be set: `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`. If 
the domain doesn't end with the hosted zone (minus the trailing "."), the AWS APIs will return an error. 

Valet uses a TTL of 30 seconds and these DNS may need to be manually cleaned up when finished. 

#### Conditions

Sometimes, it is necessary to tell valet to wait until a certain condition is met before continuing to deploy 
resources for an application. This condition was modeled as a resource to precisely indicate when the wait should
occur. For instance, to wait for a field on a CRD to change, use:

```yaml
steps:
  - condition:
      type: MyCustomType
      name: example
      namespace: ns
      jsonpath: '{.spec.my.status}'
      value: OK
      timeout: 240s
``` 

This waits until the `spec.my.status` field on the specified resource matches the value `OK`, or fails after 240s. 

#### Curl

It may be desirable to create a workflow that involves changing some resources and then curling an endpoint to 
check for an expected result. A curl command could look like:

```yaml
steps:
  - curl:
      path: /
      host: example.com
      headers:
        Authorization: "token foo"
      statusCode: 200
      service:
        name: gateway-proxy-v2
        namespace: gloo-system
```

The host and service ref are both configurable to support cases where, for instance, you are trying to test 
a service by routing to the IP directly, but the routing rules are configured for a particular domain. This example
curls the Gloo proxy IP, but with the host header set, to match on a virtual service for the "example.com" domain. 

#### Workflows

Workflows may reference other workflows, loaded via a path just like applications:  

```yaml
  - workflow:
      path: registry/workflows/deploy-smh-app.yaml
      values:
        ApplicationPath: registry/apps/istio.yaml
        ApplicationName: istio-demo
```

## Clusters

Putting it all together, Valet enables managing configurations for entire **Clusters** in a single config. 

Valet currently supports managing Minikube and GKE clusters, through the `cluster` section of the config. 
If no cluster is defined in the valet config, valet will ensure the applications against the current kube config. 

### GKE

The example above ensures a GKE cluster:
```yaml
cluster:
  gke:
    name: glooe-petclinic-demo
    location: us-central1-a
    project: solo-test-236622
```

In order to manage a GKE cluster, valet needs the following permissions on the project:
* Kubernetes Engine Cluster Admin
* Cloud Build Service Account (when running valet in cloud build)
* Service Account User (when running valet in cloud build)

To provide credentials, you can use the `GOOGLE_APPLICATION_CREDENTIALS` environment variable, for instance:
`GOOGLE_APPLICATION_CREDENTIALS=path/to/google/creds.json valet ensure -f path/to/config.yaml`

See below for configuring this and other variables globally for your local client. 

### Minikube

To ensure a Minikube cluster instead, use:
```yaml
cluster:
  minikube: {}
```

By default, this will start a new Minikube cluster with 8192MB RAM, 4 CPUs, and Kube version v1.13.0. 

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

## Setting your local Kubernetes context

Run `valet set-context -f path/to/config.yaml` to set your local Kubernetes context to the cluster in the config file. This will 
not ensure or teardown any resources. 

## Teardown

When you are done using a config, Valet can clean it up with `valet teardown -f path/to/config.yaml`. 
This can be useful if the cluster gets into a bad state, or if it is no longer useful. **If the config
contains a cluster definition, the cluster will be destroyed.**

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

## Installing and Running

The easiest way to use Valet is to run from source:
`git clone https://github.com/solo-io/valet && cd valet`
`make build TAGGED_VERSION=$(git describe --tags)`

Now move `_output/valet` to somewhere in your bin, or create an alias: `alias valet="$(pwd)/_output/valet"`

To get a jump start with some known configuration, clone the valet-config repo:
`git clone https://github.com/solo-io/valet-config && cd valet-config`
`valet ensure -f clusters/demos/petclinic-minikube.yaml`
