# Introduction to Containers / K8s / OpenShift

The labs in this repo were tested in a Fedora 35 system with the following hardware and software versions:

Hardware:

 - 4 vCPUs
 - 4 GiB RAM
 - 30 GB HD

Software:
    
  - Podman v3.0
  - Podman Compose v0.1
  - Kubectl v1.21.1
  - Kind v0.11.1

## Preparing the system for the labs

1. Connect to the Fedora 35 system
2. Install Podman and podman-compose

    ~~~sh
    sudo dnf install -y podman podman-compose
    ~~~
3. Install Kubectl 

    ~~~sh
    sudo curl -L https://dl.k8s.io/release/v1.21.1/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl
    sudo chmod +x /usr/local/bin/kubectl
    ~~~
4. Install Kind

    ~~~sh
    sudo curl -L https://github.com/kubernetes-sigs/kind/releases/download/v0.11.1/kind-linux-amd64 -o /usr/local/bin/kind
    sudo chmod +x /usr/local/bin/kind
    ~~~