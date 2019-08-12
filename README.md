# Valet

Valet is a tool for automatically ensuring the state of kubernetes clusters, Solo products, and demo applications. 

## Installing

The easiest way to install Valet is to run: 

`go get -u github.com/solo-io/valet`

This will put the latest valet binary in your `$GOPATH/bin` directory. 

## Example Usage

### Example 1: Deploy Open Source Gloo on Minikube

For example, given the following configuration in `demo.yaml`:

```yaml
cluster:
  type: minikube
gloo:
  version: v0.18.15
demos:
  petclinic: {}
```

Run `valet ensure -f demo.yaml` to get a minikube cluster running gloo v0.18.15 with the petclinic demo resources. If minikube is already running, or Gloo is already installed, valet will detect that and not re-create when possible. 

### Example 2: Deploy Enterprise Gloo on GKE

A similar config could be provided for creating a GKE cluster running enterprise Gloo:

```yaml
cluster:
  type: gke
  name: gloo-demo
  project: solo-test-236622
  location: us-central1-a
gloo:
  version: v0.18.6
  enterprise: true
  license-key: LICENSE_KEY
demos:
  petclinic: {}
```

Note that to interact with GKE, you need the `GOOGLE_APPLICATION_CREDENTIALS` environment variable set. Also, currently artifacts are downloaded from Github, so a Github token is required to download glooctl for enterprise Gloo. So you could deploy the demo with the following command:

`GOOGLE_APPLICATION_CREDENTIALS=/path/to/creds.json GITHUB_TOKEN=token valet ensure -f demo.yaml`
