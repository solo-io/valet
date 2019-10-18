# Valet

Valet is a tool that automates building and deploying applications to Kubernetes. 
Valet can be used in the dev environment, in CI, or as part of a GitOps deployment workflow. 

## Usage

There are three primary commands for interacting with Valet: **ensure**, **teardown**, and **build**. 
There is also a **config** command for setting global configuration. 

### Ensure 

Valet offers the ability to set up Kubernetes clusters, and deploy a set of applications, which may 
consist of a set of namespaces, Helm charts, manifests, and secrets in an order defined by a the 
ensure config.

#### Ensure an application against the current kube context

To use Valet to ensure a resources on a Kubernetes cluster, run: `valet ensure -f config.yaml`

For example, if we were running from the root of the repo: `valet ensure -f cli/internal/ensure/test/fixtures/applications/smh.yaml`
```yaml
applications:
  - name: service-mesh-hub
    resources:
      - helmChart:
          repoUrl: "https://storage.googleapis.com/sm-marketplace-helm/"
          chartName: sm-marketplace
          repoName: sm-marketplace
          version: 0.3.9
          namespace: sm-marketplace
          set:
            - "namespace.create=true"
``` 

This command installs an application called **service-mesh-hub** with a single resource: a 
helm chart from a public Helm repo, with a single "set" override. 
 
Let's try a more complicated example: `valet ensure -f cli/internal/ensure/test/fixtures/applications/petclinic.yaml`
```yaml
gloo:
  version: 0.20.7
applications:
  - name: petclinic
    resources:
      - path: "https://raw.githubusercontent.com/sololabs/demos/b523571c66057a5591bce22ad896729f1fee662b/petclinic_demo/petclinic.yaml"
      - path: "https://raw.githubusercontent.com/sololabs/demos/b523571c66057a5591bce22ad896729f1fee662b/petclinic_demo/petclinic-vets.yaml"
      - path: "https://raw.githubusercontent.com/sololabs/demos/b523571c66057a5591bce22ad896729f1fee662b/petclinic_demo/petclinic-db.yaml"
      - path: "https://raw.githubusercontent.com/sololabs/demos/b523571c66057a5591bce22ad896729f1fee662b/petclinic_demo/petclinic-virtual-service.yaml"
      - secret:
          name: aws-creds
          namespace: gloo-system
          entries:
            secret_access_key:
              envVar: AWS_SECRET_ACCESS_KEY
            access_key_id:
              envVar: AWS_ACCESS_KEY_ID
      # relative path assumes valet is run from the root of the repo
      - path: "cli/internal/ensure/test/fixtures/upstreams/aws.yaml"
 ```
 
Here, Valet deployed Gloo, and then a single application called petclinic. It applied 4 manifests, 
then created a secret and a Gloo upstream referencing that secret. 

#### Secret management

Valet currently supports 3 ways to inject secret values into applications: **environment variables**, **files**, and **gcloud kms encrypted files**. 

Here's an example config that creates secrets with all 3 types of values:

`valet ensure -f cli/internal/ensure/test/fixtures/secrets.yaml`

```yaml
applications:
  - name: example-secrets
    resources:
      - secret:
          name: test-secret-1
          namespace: default
          entries:
            aws_access_id:
              envVar: AWS_ACCESS_KEY_ID
            aws_secret_key:
              envVar: AWS_SECRET_ACCESS_KEY
      - secret:
          name: test-secret-2
          namespace: default
          entries:
            test-file:
              file: Makefile
      - secret:
          name: test-secret-3
          namespace: default
          entries:
            test-file:
              gcloudKmsEncryptedFile:
                ciphertextFile: test-cloudbuild/ci/id_rsa.enc
                gcloudProject: solo-public
                keyring: build
                key: build-key
```

#### Ensuring clusters

A cluster definition can be included in a Valet config in order to ensure the state of the cluster. 
Currently, Valet can ensure Minikube and Kubernetes clusters. 

Here is an example of ensuring a default minikube cluster: 
`valet ensure -f cli/internal/ensure/test/fixtures/minikube.yaml`

```yaml
cluster:
  minikube: {}
```

Here is an example for GKE. Note: if deploying to GKE, the global 
`GOOGLE_APPLICATION_CREDENTIALS` environment variable must be set. The referenced credentials must 
grant GKE cluster admin permissions to Valet. 

`valet ensure -f cli/internal/ensure/test/fixtures/cluster.yaml`
```yaml
cluster:
  gke:
    name: valet-test
    location: us-east1
    project: solo-public
```

#### Running Valet in CI 

Valet publishes a container on release to the `quay.io/solo-io/valet` Docker repo, with all the dependencies
pre-loaded. This makes it easy to invoke in a CI system, such as Google Cloud Build (see cloudbuild.yaml for complete example):

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

#### Setting your local Kubernetes context

Run `valet set-context -f ...` to set your local Kubernetes context to the cluster in the config file. This will 
not ensure or teardown any resources. 

### Teardown

When you are done using a config, Valet can clean it up with `valet teardown -f config.yaml`. 
This can be useful if the cluster gets into a bad state, or if it is no longer useful. **If the config
contains a cluster definition, the cluster will be destroyed.**

### Config

In order to get all the value from Valet, run the following commands to set up a few environment variables:

```
valet config set \
  GITHUB_TOKEN=... \
  AWS_ACCESS_KEY_ID=... \
  AWS_SECRET_ACCESS_KEY=... \
  LICENSE_KEY=... \
  GOOGLE_APPLICATION_CREDENTIALS=...
```

The github token is needed for downloading enterprise glooctl. The AWS vars are needed for creating AWS secrets and upstreams, as well as creating automatic DNS mappings in Route53. The license key is used for installing enterprise Gloo, and the google application credentials are used when managing GKE clusters.

### Build

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

## Installing

The easiest way to install Valet is to run: 

`go get -u github.com/solo-io/valet`

This will put the latest valet binary in your `$GOPATH/bin` directory.   
