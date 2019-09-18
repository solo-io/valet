# Valet   

Valet is a tool that automates building and deploying solo products and demos. It can be used in the dev environment, in CI, or as part of a GitOps deployment workflow.  

## Usage

There are three primary commands for interacting with Valet: **build**, **ensure**, and **teardown**. There is also a **config** command for setting global configuration. 

### Build

Valet offers the (optional) ability to build artifacts from any repo. It currently supports building Go binaries, tagging and pushing docker containers, and producing Helm charts and manifests. Valet typically determines the artifacts to build by reading a file called `artifacts.yaml` in the root of the repo. An example is included in [this repo](artifacts.yaml) for building Valet itself. 

Valet artifacts are always stored locally in the `_artifacts` directory at the root of the repo. 

Valet artifacts are versioned, with the version provided at runtime. This version may be a semver version (i.e. 1.2.3), a SHA, or any other version. Typically, when running locally, you would use git to determine a meaningful, unique, and stable version:

`valet build -v $(git describe --tags --dirty | cut -c 2-)`

If your repo does not have any tags, then you may use this instead:

`valet build -v $(git describe --always --dirty)`

Artifacts can be tagged for upload to the google storage valet bucket by adding `upload: true` to the binary or helm chart. (NOTE: this requires google storage writer permissions on the valet bucket.) 

### Ensure 

Valet offers the ability to set up Kubernetes clusters, download and deploy Solo products (currently only glooctl/Gloo is supported), manage other Kubernetes resources (i.e. demo applications), and handle basic administrative tasks (i.e. setting up DNS mappings). These resources are defined in a yaml config file and passed to valet using `valet ensure -f config.yaml`. 

Currently Valet can manage deploying to Minikube or GKE. If deploying to GKE, the global `GOOGLE_APPLICATION_CREDENTIALS` environment variable must be set. 

Valet first ensures that the Kubernetes cluster is running, and create it if not. Once the cluster is running, if the config specifies installing Gloo, Valet will check to see if `glooctl` is already installed locally at the desired version, and install it if necessary. Then, valet will check the cluster to see if gloo has been deployed at the desired version, and install if necessary. If a different version of Gloo is deploy, valet will first run `glooctl uninstall --all`. If running with the Gloo UI and a DNS entry is requested, then Valet will create it in route53. 

Once Gloo is installed, Valet will deploy petclinic and other resources, and set up an additional DNS entry if desired. 

If no version is specified for installation of Gloo, Valet will install the latest version it can find on github. If a version is provided, it will find and install that specific version. Valet can install using artifacts it produced with `valet build` if two flags are provided:

`valet ensure -f config.yaml --valet-artifacts --gloo-version 0.18.42-8-g6910eec45` 

The first option tells Valet to use google storage to locate artifacts instead of Github. The second tells Valet the specific version of Gloo to use during installation. This version overrides the version specified in the config file. 

### Teardown

When you are done using a cluster, Valet can clean it up with `valet teardown -f config.yaml`. This can be useful if the cluster gets into a bad state, or if it is no longer useful. 

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

## Installing

The easiest way to install Valet is to run: 

`go get -u github.com/solo-io/valet`

This will put the latest valet binary in your `$GOPATH/bin` directory.   

## Example Usage

### Example 1: Deploy Open Source Gloo on Minikube

This requires no setup. 

Run `valet ensure -f https://raw.githubusercontent.com/solo-io/valet-config/master/latest-gloo-minikube.yaml` ([link](https://raw.githubusercontent.com/solo-io/valet-config/master/latest-gloo-minikube.yaml)) to get a minikube cluster running the latest gloo, with the petclinic demo resources. 

```
[11:09:54] rick@Ricks-MBP-2:~/code2/valet$ valet ensure -f https://raw.githubusercontent.com/solo-io/valet-config/master/latest-gloo-minikube.yaml
{"level":"info","ts":"2019-08-15T11:09:57.186-0400","caller":"minikube/cluster.go:50","msg":"Setting kube context to minikube"}
{"level":"warn","ts":"2019-08-15T11:09:57.272-0400","caller":"minikube/cluster.go:43","msg":"Error checking minikube status","out":"host: \nkubelet: \napiserver: \nkubectl: \n"}
{"level":"warn","ts":"2019-08-15T11:09:57.272-0400","caller":"minikube/provision.go:31","msg":"Error checking if cluster is running, destroying","error":"host: \nkubelet: \napiserver: \nkubectl: \n (exit status 1)"}
{"level":"info","ts":"2019-08-15T11:09:57.272-0400","caller":"minikube/cluster.go:73","msg":"Destroying minikube"}
{"level":"info","ts":"2019-08-15T11:09:57.291-0400","caller":"minikube/cluster.go:61","msg":"Creating minikube"}
{"level":"info","ts":"2019-08-15T11:12:09.110-0400","caller":"minikube/cluster.go:50","msg":"Setting kube context to minikube"}
{"level":"info","ts":"2019-08-15T11:12:09.176-0400","caller":"minikube/cmd.go:30","msg":"Minikube is ready"}
{"level":"info","ts":"2019-08-15T11:12:09.423-0400","caller":"gloo/root.go:147","msg":"Setting version to latest release","tag":"v0.18.16"}
{"level":"info","ts":"2019-08-15T11:12:09.465-0400","caller":"gloo/glooctl.go:67","msg":"updated glooctl on path to be this version"}
{"level":"info","ts":"2019-08-15T11:12:09.484-0400","caller":"gloo/gloo.go:143","msg":"Running glooctl install"}
{"level":"info","ts":"2019-08-15T11:12:18.507-0400","caller":"gloo/gloo.go:50","msg":"Waiting for pods"}
{"level":"info","ts":"2019-08-15T11:12:38.854-0400","caller":"gloo/gloo.go:90","msg":"gloo is ready"}
{"level":"info","ts":"2019-08-15T11:12:39.242-0400","caller":"petclinic/root.go:89","msg":"Successfully applied file","file":"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic.yaml"}
{"level":"info","ts":"2019-08-15T11:12:39.642-0400","caller":"petclinic/root.go:89","msg":"Successfully applied file","file":"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-vets.yaml"}
{"level":"info","ts":"2019-08-15T11:12:40.050-0400","caller":"petclinic/root.go:89","msg":"Successfully applied file","file":"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-db.yaml"}
{"level":"info","ts":"2019-08-15T11:12:40.373-0400","caller":"petclinic/root.go:89","msg":"Successfully applied file","file":"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-virtual-service.yaml"}
```

If minikube is already running, or Gloo is already installed, valet will detect that and not re-create when possible. 

### Example 2: Deploy Enterprise Gloo on GKE with DNS domains

This requires setting up the global environment variables as described above. 

Create the following config:

```yaml
cluster:
  type: gke
  gke:
    name: glooe-demo-idit2
    location: us-central1-a
    project: solo-test-236622
gloo:
  enterprise: true
  aws:
    secret: true
    upstream: true
  ui_virtual_service:
    dns:
      hosted_zone: corp.solo.io.
demos:
  petclinic:
    dns:
      hosted_zone: corp.solo.io.
```

Valet will ensure the cluster has the latest GlooE with the demo set up and URLs to access petclinic and the Gloo UI. 

```shell
[11:22:48] rick@Ricks-MacBook-Pro-2:~$ code2/valet/_output/valet ensure -f ~/.valet/config/glooe-demo-idit2.yaml
{"level":"warn","ts":"2019-08-15T11:24:36.727-0400","caller":"file/config.go:47","msg":"Could not read url, trying to read file","error":"Get /Users/rick/.valet/config/glooe-demo-idit2.yaml: unsupported protocol scheme \"\"","path":"/Users/rick/.valet/config/glooe-demo-idit2.yaml"}
{"level":"info","ts":"2019-08-15T11:24:37.308-0400","caller":"gke/provision.go:31","msg":"GKE cluster is running, setting context"}
{"level":"info","ts":"2019-08-15T11:24:37.308-0400","caller":"gke/cluster.go:102","msg":"Setting kube context to GKE"}
{"level":"info","ts":"2019-08-15T11:24:38.073-0400","caller":"gke/cmd.go:52","msg":"gke is ready"}
{"level":"info","ts":"2019-08-15T11:24:38.340-0400","caller":"gloo/root.go:147","msg":"Setting version to latest release","tag":"v0.18.9"}
{"level":"info","ts":"2019-08-15T11:24:38.382-0400","caller":"gloo/glooctl.go:67","msg":"updated glooctl on path to be this version"}
{"level":"info","ts":"2019-08-15T11:24:39.081-0400","caller":"gloo/gloo.go:39","msg":"Gloo is installed at the desired version"}
{"level":"info","ts":"2019-08-15T11:24:39.679-0400","caller":"gloo/root.go:115","msg":"aws secret exists"}
{"level":"info","ts":"2019-08-15T11:24:40.509-0400","caller":"gloo/root.go:94","msg":"aws upstream exists"}
{"level":"info","ts":"2019-08-15T11:24:40.509-0400","caller":"gloo/vs.go:48","msg":"Creating ui virtual service"}
{"level":"info","ts":"2019-08-15T11:24:41.023-0400","caller":"gloo/vs.go:99","msg":"Getting Gloo proxy ip"}
{"level":"info","ts":"2019-08-15T11:24:41.333-0400","caller":"internal/utils.go:13","msg":"Getting current context name"}
{"level":"info","ts":"2019-08-15T11:24:41.395-0400","caller":"gloo/dns.go:36","msg":"Getting hosted zone id"}
{"level":"info","ts":"2019-08-15T11:24:41.622-0400","caller":"gloo/dns.go:83","msg":"Creating dns mapping","hostedZone":"corp.solo.io.","hostedZoneId":"/hostedzone/Z3K5Q8T22D8CRP","domain":"valet-glooui-9479a49b9c.corp.solo.io","ip":"35.222.92.145"}
{"level":"info","ts":"2019-08-15T11:24:41.704-0400","caller":"gloo/vs.go:87","msg":"Patching glooui domain"}
{"level":"info","ts":"2019-08-15T11:24:42.819-0400","caller":"petclinic/root.go:89","msg":"Successfully applied file","file":"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic.yaml"}
{"level":"info","ts":"2019-08-15T11:24:43.665-0400","caller":"petclinic/root.go:89","msg":"Successfully applied file","file":"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-vets.yaml"}
{"level":"info","ts":"2019-08-15T11:24:44.456-0400","caller":"petclinic/root.go:89","msg":"Successfully applied file","file":"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-db.yaml"}
{"level":"info","ts":"2019-08-15T11:24:45.114-0400","caller":"petclinic/root.go:89","msg":"Successfully applied file","file":"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-virtual-service.yaml"}
{"level":"info","ts":"2019-08-15T11:24:45.114-0400","caller":"gloo/vs.go:99","msg":"Getting Gloo proxy ip"}
{"level":"info","ts":"2019-08-15T11:24:45.396-0400","caller":"internal/utils.go:13","msg":"Getting current context name"}
{"level":"info","ts":"2019-08-15T11:24:45.457-0400","caller":"gloo/dns.go:36","msg":"Getting hosted zone id"}
{"level":"info","ts":"2019-08-15T11:24:45.494-0400","caller":"gloo/dns.go:83","msg":"Creating dns mapping","hostedZone":"corp.solo.io.","hostedZoneId":"/hostedzone/Z3K5Q8T22D8CRP","domain":"valet-petclinic-9479a49b9c.corp.solo.io","ip":"35.222.92.145"}
{"level":"info","ts":"2019-08-15T11:24:45.572-0400","caller":"petclinic/root.go:94","msg":"Patching petclinic domain"}
```

