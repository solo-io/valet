# Workflows

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

## Steps

Valet supports workflows that include **installing and uninstalling applications**, applying a **patch** to a resource,
setting up **dns entries**, waiting for a **condition** to be met on a Kubernetes resource, waiting for a 
**curl** request to return an expected status, or running another **workflow**. 

### Installing applications

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

### DNS entries

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

### Helm3 Deploys

In 0.5.0, Valet introduces the Helm3 Deploy step in a workflow. This is intended to replace the `helmChart` 
application resource type. 

The following config is an example of a workflow to deploy an application with `helm3Deploy`:

```yaml
steps:
- helm3Deploy:
    releaseName: gloo
    namespace: gloo-system
    releaseUri: "https://storage.googleapis.com/gloo-ee-helm/charts/gloo-ee-1.2.6.tgz"
    set:
      "license_key": "env:LICENSE_KEY"
    valuesFiles:
      - custom-values.yaml
```

This is implemented using Helm3 as a library, and does not require Helm installed locally. 

### Patches

Typically it is desirable to deploy the correct resource first in a declarative way, rather than rely on deploying 
an incorrect resource first, then fixing it. However, this isn't always realistic, and valet provides a mechanism to 
expose `kubectl patch` using patch steps in a workflow. 

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
steps:
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

### Conditions

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

### Curl

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

### Workflows

Workflows may reference other workflows, loaded via a path just like applications:  

```yaml
  - workflow:
      path: registry/workflows/deploy-smh-app.yaml
      values:
        ApplicationPath: registry/apps/istio.yaml
        ApplicationName: istio-demo
```

### Restarting Pods

A workflow may have a step that requires restarting a pod, for instance to restart a proxy after 
updating it's bootstrap configuration. To restart a pods, select a namespace and optional a label 
selector: 

```yaml
steps:
- restartPods:
    namespace: gloo-system
    labels:
      - "gloo=gateway-proxy"
```

If the namespace isn't specified, Valet will look for a namespace in the `Namespace` value. 
