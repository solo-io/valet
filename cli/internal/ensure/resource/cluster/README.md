# Clusters

Valet enables managing configurations for entire **Clusters** in a single config. A cluster config is a workflow, 
with an additional `cluster` section.  

Valet currently supports managing Minikube, EKS, and GKE clusters. 
If no cluster is defined in the valet config, valet will ensure the applications against the current kube config. 

## GKE

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

## Minikube

To ensure a Minikube cluster instead, use:
```yaml
cluster:
  minikube: {}
```

By default, this will start a new Minikube cluster with 8192MB RAM, 4 CPUs, and Kube version v1.13.0. 

## EKS 

To ensure an EKS cluster, use:
```yaml
cluster:
  eks: 
    name: ...
    region: ...
```

Note: this requires `eksctl` installed locally.

## Setting your local Kubernetes context

Run `valet set-context -f path/to/config.yaml` to set your local Kubernetes context to the cluster in the config file. This will 
not ensure or teardown any resources.  

## Cluster Values

The cluster information is key to the configuration and workflow that valet offers. In order to facilitate easier use of 
this cluster information, valet will inject certain values from the cluster setup into the `Values` Object for use in 
steps which occur after cluster ensure.

As of right now the 3 different cluster types inject slightly different values:

* GKE: 
    * ClusterName: name of the cluster
    * Project: Gcloud project ID
    * Location: Gcloud location
* EKS:
    * ClusterName: name of the cluster
    * Region: AWS region
* Minikube: 
    * No values as of right now

More values are planned for the future, such as kube context, and kube config.