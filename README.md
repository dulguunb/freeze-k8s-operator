# k8s-builder: Deploymentfreezer

I have added a test manifest in the directory `test-manifests` where it contains nginx pod deployment and deploymentfreezer CRD at the same time. So that you can test it easily.

## Getting Started

To do this, make sure you have all the prerequisites such as the CRD is already installed.

1) **Install the CRDs into the cluster:**

```sh
make install
```

2) **Build your own image and deploy**

```sh
make docker-build IMG=deployment-freezer:latest
```

```sh
make deploy IMG=deployment-freezer:latest
```

2) **Or you can use already built image:**

```sh
make deploy IMG=ghcr.io/dulguunb/freeze-k8s-operator:cdbf6e979fec689e3405440088b76c16821ff080
```

### Prerequisites
- go version v1.24.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

