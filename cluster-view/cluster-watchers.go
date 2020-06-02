package cluster_view

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	"sync"
	"type-aware-scheduler/interference"
	scheduler_config "type-aware-scheduler/scheduler-config"
)

func InitClusterView(config *rest.Config, podsChan chan <- PodData,
		quit chan struct{}, schedulerName_ string) {
	schedulerName = schedulerName_
	log.Println("InitPodNodesWatcher log")
	initClusterViewStruct(podsChan)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	log.Println("InitPodNodesWatcher")
	factory := informers.NewSharedInformerFactory(clientset, 0)

	nodeInformer := factory.Core().V1().Nodes()
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    handleNodeChange(handleNodeAdd),
		DeleteFunc: handleNodeChange(handleNodeDelete),
		UpdateFunc: handleNodeUpdate,
	})

	podInformer := factory.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    handlePodChange(handlePodAdd),
		DeleteFunc: handlePodChange(handlePodDelete),
		UpdateFunc: handlePodUpdate,
	})

	factory.Start(quit)
	//return nodeInformer.Lister()
}

func initClusterViewStruct(podsChan chan <- PodData) {
	log.Println("InitClusterView")
	clusterViewLock = new(sync.RWMutex)
	podLookup = make(map[PodId]PodData)
	nodeLookup = make(map[string]NodeData)
	podToBeScheduled = podsChan
}

func handleNodeChange(fn nodeHanderFn) func(obj interface{}) {
	return func(obj interface{}) {
		node, ok := obj.(*v1.Node)
		if !ok {
			log.Println("this is not a node")
			return
		}
		fn(node)
	}
}

func handlePodChange(fn podHanderFn) func(obj interface{}) {
	return func(obj interface{}) {
		pod, ok := obj.(*v1.Pod)
		if !ok {
			log.Println("this is not a pod")
			return
		}
		fn(pod)
	}
}

func shouldSchedule(pod *v1.Pod) bool {
	return pod.Spec.NodeName == "" && pod.Spec.SchedulerName == schedulerName
}

// TODO handle pods with other schedulers
func handlePodAdd(pod *v1.Pod) {
	clusterViewLock.Lock()
	defer clusterViewLock.Unlock()
	log.Printf("Handle pod add %s scheduler %s\n", PodToString(pod), pod.Spec.SchedulerName)
	podId := PodId {pod.Name, pod.Namespace}
	_, found := podLookup[podId]
	if found {
		log.Fatal("Trying to add pod that is already in the system %s\n", PodIdString(podId))
		return
	}
	podData := PodData {
		Interference: interference.PodInfo{
			TaskType: 0,
			Load:     0,
		},
		Data: pod.DeepCopy(),
	}
	if shouldSchedule(pod) {
		taskInfo, err := interference.PredictPodInfo(pod)
		if err != nil {
			log.Fatalf(err.Error())
		}
		podData.Interference = taskInfo
	}
	podLookup[podId] = podData
	if shouldSchedule(pod) {
		podToBeScheduled <- podData
	}
}

func handlePodDelete(pod *v1.Pod) {
	clusterViewLock.Lock()
	defer clusterViewLock.Unlock()
	log.Printf("Handle pod delete %s scheduler %s\n", PodToString(pod), pod.Spec.SchedulerName)
	podId := PodId {pod.Name, pod.Namespace}
	podData, found := podLookup[podId]
	if !found {
		log.Fatal("Trying to delete pod that is not in the system %s\n", PodIdString(podId))
		return
	}
	unbindPodFromNode(podData)
	delete(podLookup, podId)
}

func handlePodUpdate(oldObj interface{}, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	newPod, ok2 := newObj.(*v1.Pod)
	if !ok || !ok2 {
		return
	}
	clusterViewLock.Lock()
	defer clusterViewLock.Unlock()
	log.Printf("Update pod %s / %s\n", PodToString(oldPod), PodToString(newPod))
	// TODO
}

func handleNodeAdd(node *v1.Node) {
	clusterViewLock.Lock()
	defer clusterViewLock.Unlock()
	log.Printf("Handle node add %s\n", node.Name)
	_, found := nodeLookup[node.Name]
	if found {
		log.Fatal("Trying to add node that is already in the system %s\n", node.Name)
		return
	}
	nodeLookup[node.Name] = NodeData {
		TypeToLoad: new([scheduler_config.NumberOfTaskTypes]float64)[:],
		Data: node.DeepCopy(),
	}
}

func handleNodeDelete(node *v1.Node) {
	clusterViewLock.Lock()
	defer clusterViewLock.Unlock()
	log.Printf("Handle node delete %s\n", node.Name)
	_, found := nodeLookup[node.Name]
	if !found {
		log.Fatal("Trying to delete node that is not in the system %s\n", node.Name)
		return
	}
	delete(nodeLookup, node.Name)
}

func handleNodeUpdate(oldObj interface{}, newObj interface{}) {
	oldNode, ok := oldObj.(*v1.Node)
	newNode, ok2 := newObj.(*v1.Node)
	if !ok || !ok2 {
		log.Print("Node update - oldObj or newObj is not a node")
		return
	}
	clusterViewLock.Lock()
	defer clusterViewLock.Unlock()
	log.Printf("Update node %s / %s\n", oldNode.Name, newNode.Name)
	newNodeData := NodeData {
		Data : newNode.DeepCopy(),
	}
	oldNodeData, found := nodeLookup[oldNode.Name]
	if found {
		newNodeData.TypeToLoad = oldNodeData.TypeToLoad
	} else {
		newNodeData.TypeToLoad = new([scheduler_config.NumberOfTaskTypes]float64)[:]
		log.Fatal("Trying to update node that cannot be found in system %s\n", oldNode.Name)
	}
	nodeLookup[newNode.Name] = newNodeData
	// TODO can we assume that for each pod we will have update event?
	// TODO what if we override data for some old node?
}

