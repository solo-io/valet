# Valet

Valet is a tool for automatically ensuring the state of kubernetes clusters, Solo products, and demo applications. 

## Installing

The easiest way to install Valet is to run: 

`go get -u github.com/solo-io/valet`

This will put the latest valet binary in your `$GOPATH/bin` directory. 

## Global config 

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

## Example Usage

### Example 1: Deploy Open Source Gloo on Minikube

This requires no setup. 

Run `valet ensure -f https://raw.githubusercontent.com/solo-io/valet-config/master/latest-gloo-minikube.yaml` to get a minikube cluster running the latest gloo, with the petclinic demo resources. 

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

