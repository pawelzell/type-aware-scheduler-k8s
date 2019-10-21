package cluster_view

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	"type-aware-scheduler/scheduler-config"
)

func InitPodNodesWatchers(config *rest.Config, podQueue chan *v1.Pod,
		quit chan struct{}) {
	log.Println("InitPodNodesWatcher log")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("InitPodNodesWatcher")
	factory := informers.NewSharedInformerFactory(clientset, 0)

	nodeInformer := factory.Core().V1().Nodes()
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node, ok := obj.(*v1.Node)
			if !ok {
				log.Println("this is not a node")
				return
			}
			fmt.Printf("New Node Added to Store: %s\n", node.GetName())
		},
	})

	podInformer := factory.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				log.Println("this is not a pod")
				return
			}
			fmt.Printf("Got new pod %s:%s with scheduler name: %s\n", pod.Namespace,
				pod.Namespace, pod.Spec.SchedulerName)
			if pod.Spec.NodeName == "" && pod.Spec.SchedulerName == scheduler_config.SchedulerName {
				podQueue <- pod
			}
		},
	})

	factory.Start(quit)
	//return nodeInformer.Lister()
}
