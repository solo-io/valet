# Valet

Valet is a tool for Kubernetes operators to easily assemble and manage many applications across a 
Kubernetes cluster. Today, it is still painful and time consuming for operators to manage clusters, despite
powerful tools in the community:
* `Kubectl` gives you the ability to view or manage any resource on a cluster, among many other things
* `Helm` gives you the ability to package complex applications into a single, configurable chart, and deploy those charts
* `Kustomize` gives the the ability to patch a Kubernetes manifest before deploying to adjust it for the particular deployment environment

`Valet` extends these tools and ideas, giving Kubernetes operators the ability to:
* Manage many applications and resources, potentially for an entire cluster, with a simple declarative configuration
* Easily customize an application by bundling a set of resources, such as helm charts, manifests, patches, secrets, or other applications
* Organize DevOps across many engineering teams, without having to write custom helm charts or automation 
* Ensure the state of a production cluster as part of a GitOps workflow 

Valet was also written with the developer in mind: developers can easily deploy a cluster configuration to 
Minikube or a dev cluster to replicate a realistic Kubernetes environment. This makes it a powerful tool 
for sharing new complex demos or reproducing bugs that only appear in a production environment. 

## Valet example

To see the power of Valet, consider this example of a valet configuration:

```yaml
cluster:
  gke:
    name: glooe-petclinic-demo
    location: us-central1-a
    project: solo-test-236622
applications:
  - path: registry/apps/cert-manager.yaml
  - path: registry/apps/gloo-enterprise-app.yaml
    values:
      Domain: valet-glooui-example.corp.solo.io
      Version: 0.20.4
  - path: registry/apps/petclinic-app.yaml
    values:
      Domain: valet-petclinic-example.corp.solo.io
```

This configuration specifies a GKE cluster that should have several applications deployed: cert manager, gloo-enterprise 0.20.4, and 
the petclinic demo application. Valet can ensure this configuration with `valet ensure -f path/to/config.yaml`, causing it to 
create the cluster if necessary and make sure all the applications are deployed. 

### Example app configuration: gloo-enterprise-app

Let's look at `gloo-enterprise-app.yaml`, one of the applications referenced in the valet config:

```yaml
name: gloo-enterprise-app
requiredValues:
  - Domain
  - Version
values:
  Namespace: gloo-system
  UpstreamName: gloo-system-apiserver-ui-8080
  UpstreamNamespace: gloo-system
  VirtualServiceName: glooui
resources:
  - application:
      path: registry/apps/gloo-enterprise.yaml
  - application:
      path: registry/apps/gloo-app.yaml
```

At the top, this application requires `Domain` and `Version` variables. These are provided above in the cluster config. 
The rest of the values needed by the resources for this application are provided at the top. This application contains 
two resources, which are each applications themselves. 

#### gloo-enterprise

First is `gloo-enterprise.yaml`:

```yaml
name: gloo-enterprise
requiredValues:
  - Version
values:
  Namespace: gloo-system
resources:
  - helmChart:
      repoUrl: "https://storage.googleapis.com/gloo-ee-helm/"
      chartName: gloo-ee
      repoName: gloo-ee
      set:
        - "gloo.namespace.create=true"
        - "prometheus.enabled=false"
        - "grafana.defaultInstallationEnabled=false"
        - "gloo.gatewayProxies.gatewayProxyV2.readConfig=true"
      setEnv:
        license_key: LICENSE_KEY
```

This is an application that deploys a single resource -- the enterprise Gloo Helm chart, with several values explicitly 
provided, or set based on environment variables. 

#### gloo-app

The other application is called `gloo-app.yaml` which bundles the resources necessary to expose a Kubernetes service
in Gloo behind a domain registered with AWS Route53 and with a valid certificate. 

```yaml
# This is a wrapper for the resources (virtual services, certificate, and dns entry) that are necessary to
# configure a simple application in Gloo

name: gloo-app
requiredValues:
  - Domain
  - Namespace
  - VirtualServiceName
  - UpstreamName
  - UpstreamNamespace
resources:
  - template:
      path: registry/templates/cert.yaml
  - dnsEntry:
      hostedZone: corp.solo.io.
      service:
        namespace: gloo-system
        name: gateway-proxy-v2
  - template:
      path: registry/templates/virtual-service-ssl.yaml
```

In this case, we are trying to deploy a `gloo-app` to expose the Gloo UI via the domain provided in the 
valet config: `valet-glooui-example.corp.solo.io`.

## Clusters

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

## Applications

As we saw above, an application in Valet is a bundle of a bunch of resources. A valet config can have one 
or more applications:

```yaml
applications:
  - path: path/to/app.yaml
    values:
      Foo: bar
  - path: https://url/to/app.yaml
  - path: ...
```

All paths in Valet can be either relative paths from the valet working directory, or URLs. Here, the path points
to a file containing yaml that implements the Application API. In addition to the path, the valet config may 
provide values that will be passed into the application and any resources that need values from the application. 

### Basic Application API

An application has a few required fields:
```yaml
name: app-name
values: {...}
requiredValues: [...]
resources: [...]
```

The **name** currently is just a convenient name for the application. In the future, this may become used to identify
an application inside of a registry. 

The **values** are a map of key-value string pairs that will be provided to resources in the application. If values 
were provided to the application, they will be merged with the values defined on the application, with the inherited 
values taking precedence. This allows the person assembling applications to override values as necessary. 

The **requiredValues** are a map of string keys that must be present in the **values** or inherited values. When ensuring
a valet config, an error will be returned if a required value is not provided. 

The **resources** are a list of application resources that will be ensured in order. 

### Application Resources

Valet defines a `Resource` interface that requires specifying an `Ensure` and `Teardown` function. Anything that can implement
these functions could be considered a resource in Valet, and it currently supports: **yaml manifests**, **helm charts**, 
**secrets**, **templates**, **patches**, **namespaces**, **DNS entries**, **conditions**, and **applications**. Valet will ensure the 
resources for an application in order.

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

1. As `values` or `envValues` on the template directly

```yaml
name: template-example
resources:
  - template: 
      path: path/to/template.yaml
      values: 
        Domain: foo
      envValues:
        Namespace: BAR # reads environment variable BAR
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

#### DNS entries

A common use case when setting up clusters is to define DNS mappings from new domain names to IP addressed exposed by
services in the cluster. Valet currently supports this with Amazon Route53 DNS. 

```yaml
application: dns-example
resources: 
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
name: condition-example
resources:
  - condition:
      type: MyCustomType
      name: example
      namespace: ns
      jsonpath: '{.spec.my.status}'
      value: OK
      timeout: 240s
``` 

This waits until the `spec.my.status` field on the specified resource matches the value `OK`, or fails after 240s. 

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
