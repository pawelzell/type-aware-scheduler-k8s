package cluster_view

import (
	v1 "k8s.io/api/core/v1"
	"sync"
	"type-aware-scheduler/interference"
)

type podHanderFn func(*v1.Pod)
type nodeHanderFn func(*v1.Node)

type PodId struct {
	Name string
	Namespace string
}

type PodData struct {
	Interference interference.PodInfo
	Node string // Empty if node is not bound
	Data *v1.Pod
}

type NodeData struct {
	TypeToLoad []float64
	Data *v1.Node
}

type PodIdAndInterference struct {
	Id PodId
	Interference interference.PodInfo
}

var clusterViewLock *sync.RWMutex

var podLookup map[PodId]PodData
var nodeLookup map[string]NodeData
var podToBeScheduled chan <- PodData


