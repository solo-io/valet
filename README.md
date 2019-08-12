# Valet

Valet is a tool for automatically ensuring the state of kubernetes clusters, Solo products, and demo applications. For example, given the following configuration in `demo.yaml`:

```yaml
cluster:
  type: minikube
gloo:
  version: v0.18.15
demos:
  petclinic: {}
```

Run `valet ensure -f demo.yaml` to get a minikube cluster running gloo v0.18.15 with the petclinic demo resources.  
