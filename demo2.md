# Introduction to Kubernetes

In this demo we are going to see how to deploy a Kubernetes test cluster using [Kind](https://github.com/kubernetes-sigs/kind). We will deploy an application and expose it so we can access it from our browser.

## Running a test cluster with Kind

1. Create a test-cluster with Kind:

    > **NOTE**: Below cluster definition do extra stuff so we can deploy an ingress controller afterwards.
    ~~~sh
    cat <<EOF | sudo kind create cluster --config=-
    kind: Cluster
    name: test-cluster
    apiVersion: kind.x-k8s.io/v1alpha4
    nodes:
    - role: control-plane
      image: docker.io/kindest/node@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6
      kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: ingress-ready=true
      extraPortMappings:
      - containerPort: 80
        hostPort: 80
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
    - role: worker
      image: docker.io/kindest/node@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6
    EOF
    ~~~
2. Copy the kubeconfig from root to your regular user

    ~~~sh
    mkdir -p ~/.kube/
    sudo cp /root/.kube/config ~/.kube/config
    sudo chmod 644 ~/.kube/config
    ~~~
3. Deploy the NGINX Ingress Controller and wait for it rollout

    ~~~sh
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
    kubectl -n ingress-nginx rollout status deployment ingress-nginx-controller
    ~~~

## First contanct with a Kubernetes

Now we have our Kubernetes test-cluster, but it's the first time that we interact with one, so let's get to know the basics.

1. Kubernetes clusters have nodes, those nodes can be control-plane nodes or compute nodes:

    ~~~sh
    kubectl get nodes

    NAME                         STATUS   ROLES                  AGE   VERSION
    test-cluster-control-plane   Ready    control-plane,master   17m   v1.21.1
    test-cluster-worker          Ready    <none>                 16m   v1.21.1
    ~~~
2. We have namespaces. Namespaces provides a mechanism for isolating groups of resources within a single cluster. Names of resources need to be unique within a namespace, but not across namespaces. Namespace-based scoping is applicable only for namespaced objects (e.g. Deployments, Services, etc) and not for cluster-wide objects (e.g. StorageClass, Nodes, PersistentVolumes, etc).

    ~~~sh
    kubectl get namespaces

    NAME                 STATUS   AGE
    default              Active   22m
    kube-node-lease      Active   22m
    kube-public          Active   22m
    kube-system          Active   22m
    local-path-storage   Active   22m
    ~~~
3. We also have pods. Pods are the smallest deployable units of computing that you can create and manage in Kubernetes. A pod can contain from 1 to many containers inside.


    ~~~sh
    kubectl get pods -A

    NAMESPACE            NAME                                                 READY   STATUS    RESTARTS   AGE
    kube-system          coredns-558bd4d5db-85zhb                             1/1     Running   0          23m
    kube-system          coredns-558bd4d5db-b54xr                             1/1     Running   0          23m
    kube-system          etcd-test-cluster-control-plane                      1/1     Running   0          23m
    kube-system          kindnet-nhjtp                                        1/1     Running   0          23m
    kube-system          kindnet-r7dgx                                        1/1     Running   0          23m
    kube-system          kube-apiserver-test-cluster-control-plane            1/1     Running   0          23m
    kube-system          kube-controller-manager-test-cluster-control-plane   1/1     Running   0          23m
    kube-system          kube-proxy-r9grx                                     1/1     Running   0          23m
    kube-system          kube-proxy-tvhj8                                     1/1     Running   0          23m
    kube-system          kube-scheduler-test-cluster-control-plane            1/1     Running   0          23m
    local-path-storage   local-path-provisioner-547f784dff-jrh72              1/1     Running   0          23m
    ~~~  
4. Services are also part of Kubernetes. A service is an abstract way to expose an application running on a set of Pods as a network service.

    ~~~sh
    kubectl get services -A

    NAMESPACE     NAME         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                  AGE
    default       kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP                  24m
    kube-system   kube-dns     ClusterIP   10.96.0.10   <none>        53/UDP,53/TCP,9153/TCP   24m
    ~~~
5. There are a lot more objects that we're not going to cover, these are the basics and you will get introduced to new objects as you go.

## Deploying test application

In the previous lab we deployed a Pacman application, let's get it deployed on Kubernetes!

We have the required manifests present in these folders:

- [Pacman game deployment files](./demo2-assets/pacman/)
- [MongoDB deployment files](./demo2-assets/mongo/)

1. Deploy MongoDB

2. Deploy Pacman
3. Access the game
