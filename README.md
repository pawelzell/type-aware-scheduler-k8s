# Type-aware-scheduler

An experimental Kubernetes scheduler, created to test the model from the paper [Optimizing egalitarian performance in the side-effects model of colocation for data center resource management](https://arxiv.org/abs/1610.07339) on a real Kubernetes cluster. 

## Testing environment

I use [Kind - Kubernetes in Docker](https://github.com/kubernetes-sigs/kind) to setup testing cluster. Tested on Ubuntu 18.04.3 LTS, with kind v0.5.1. 


## Cluster in Kind

Requirments:
- Kind v0.5.1
- kubectl v 1.15.3
- [helm 2.16](https://github.com/helm/helm/releases) - to install metrics server
- Go 1.13 

To run the test cluster execute:

`cd deploy && ./kind-up.sh`

To run metrics-server, from deploy directory execute:

`./kind-metrics-up.sh`

To create a go executable of the scheduler and build its docker image execute:

`cd core-scheduler && ./build.sh`

To deploy a pod from your local image you need to load it to the kind kubernetes cluster with command:

`kind load docker-image type-aware-scheduler`

Once cluster has been set up, metrics-server is up and running, a scheduler docker image is load into cluster, you can deploy the scheduler with:

`kubectl apply -f core-scheduler/deployment.yaml`

To verify if the scheduler has been deployed succesfully and is running, execute:

`kubeclt get pods`

The output might be:

```
NAME                                    READY   STATUS    RESTARTS   AGE
grafana-dbdfbc8b7-9xfxv                 1/1     Running   0          111m
influxdb-7c6995f8fd-k2xmg               1/1     Running   0          111m
type-aware-scheduler-6bf7f4c766-56sjf   1/1     Running   0          4m5s
```

Then you can check logs from the scheduler by executing the following command. Replace the name of the pod with your value obtained from the previous command.

`kubectl logs type-aware-scheduler-6bf7f4c766-56sjf`

To stop kind cluster execute:

`kind delete cluster`

## Cluster on bare metal

I configured the cluster of two machines. I configured baati machine as a master and naan as a worker. For experiments we may want to setup a cluster additionally with dosa machine (was unavailable last time for testing).

Two machines have to be in the same network. Baati and naan should be already connected via p2p openvpn, and have ip addresses 10.9.99.1 and 10.9.99.2. Swap needs to be turn off (done on baati and naan). 

If a machine has less than 15% disc space left, pods deployed on it will be immediatelly evicted caused by disc pressure condition. This threshold can be changed by providing flags on kubelet creation. It should be possible to use `kubeadm init` with --config flag to do that.

[CBTOOL](https://github.com/ibmcb/cbtool) benchmark tool (as well as my setup scripts) works with kubernetes 1.15.x, but not with 1.16. I assume kubeadm and kubelet version 1.15, kubectl version 1.15 or 1.16 are installed on master and worker nodes. If you want to bring more machines into kubernetes cluster you may need to follow [the guide](https://bigstep.com/blog/kubernetes-on-bare-metal-cloud).

### NFS setup
To use influxdb we should configured persistent storage accessible from any machine of our cluster. 

To setup nfs server on baati run
`./nfs-server-init.sh`
Then to setup nfs client on naan run
`./nfs-client-init.sh`

### Cluster setup
On master node (baati) execute:
`./cluster-master-init.sh`

The output of this script will show you a command you need to execute on each node worker machine. The command will be simmilar to:

`sudo kubeadm join 10.9.99.2:6443 --token <token> --discovery-token-ca-cert-hash sha256:<hash>` 


To deploy influx and grafana execute:

`./up.sh`.

To reset the cluster on each master and worker node execute:

`./cluster-reset.sh`

NOTE: metrics-server does not work yet on our cluster.

