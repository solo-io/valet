# Applications

An **Application** in Valet is a declarative configuration that specifies a set of resources that can be rendered
into an exact Kubernetes manifest, given a set of (possibly empty) user **values** and **flags**. 

## Commands

An application is defined in a config file, and can be ensured with: `valet ensure application -f path/to/application.yaml`.

To uninstall an application, run `valet teardown application -f path/to/application.yaml`.

An application manifest can be rendered as a dry run with the `--dry-run` flag on `valet ensure application`.  

## Example

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

## Values

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

## Flags

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

## Application Resources

Valet defines a `Resource` interface that requires specifying an `Ensure` and `Teardown` function. For application resources,
they must also be fully renderable without deploying any resources to a cluster, to enable `dry-run`. 

Valet currently supports the following types of application resources: **yaml manifests**, **helm charts**, 
**secrets**, **templates**, **patches**, **namespaces**, and **applications**. 

Valet will render resources for an application in order. When rendering, it will pass user values and flags to each 
resource and, depending on the resource, may update certain fields on the resource based on the values.  

### Manifests

Valet supports applying a manifest as an application resource. 

```yaml
name: manifest-example
resources:
  - path: path/to/manifest.yaml
``` 

This reads the manifest into `stdin` and then runs `kubectl apply -f -`. 

### Helm Charts

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

### Secrets

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

### Templates

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

### Patches

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

### Namespaces

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

### Applications

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