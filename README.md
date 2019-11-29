# Type-aware-scheduler

An experimental Kubernetes scheduler, created to test the model from the paper [Optimizing egalitarian performance in the side-effects model of colocation for data center resource management](https://arxiv.org/abs/1610.07339) on a real Kubernetes cluster. 

## Testing environment

I use [Kind - Kubernetes in Docker](https://github.com/kubernetes-sigs/kind) to setup testing cluster. Tested on Ubuntu 18.04.3 LTS, with kind v0.5.1. 

## Running cluster and scheduler

Requirments:
- Kind v0.5.1
- kubectl v 1.15.3
- Go 1.13 

To run the test cluster execute:

`cd deploy && ./up.sh`

To run metrics-server, from deploy directory execute:

`./metrics-up.sh`

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

