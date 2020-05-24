package cluster_view

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"log"
	"strings"
)

func getApplicationInstance(pod PodData) (result string, err error) {
	podName := pod.Data.Name
	var index = strings.LastIndex(podName, "ai-")
	if index < 0 {
		err = errors.New(fmt.Sprintf("ai- substring not found in pod name %s", podName))
		return
	}
	result = podName[index:]
	return
}

func BindPodToNode(id PodId, load float64, nodeName string) error {
	clusterViewLock.Lock()
	defer clusterViewLock.Unlock()
	pod, found := podLookup[id]
	if !found {
		return errors.New("Pod did not found")
	}
	pod.Interference.Load = load
	podLookup[id] = pod

	node, found := nodeLookup[nodeName]
	if !found {
		return errors.New("Node not found")
	}
	node.TypeToLoad[pod.Interference.TaskType] += pod.Interference.Load
	podLookup[id] = PodData{
		Interference: pod.Interference,
		Node:         nodeName,
		Data:         pod.Data,
	}
	return nil
}

// We might want to schedule all pods from one application instance (AI) into one node.
// If an other pod from given AI has been scheduled, returns name of its node.
func GetNodeByApplicationInstance(pod PodData) string {
	clusterViewLock.Lock()
	defer clusterViewLock.Unlock()
	podId := PodId {pod.Data.Name, pod.Data.Namespace}
	podAI, err := getApplicationInstance(pod)
	if err != nil {
		return ""
	}
	for otherPodId, otherPod := range podLookup {
		if podId == otherPodId {
			continue
		}
		otherPodAI, err := getApplicationInstance(otherPod)
		if err != nil {
			continue
		}
		if (podAI == otherPodAI) && (otherPod.Node != "" ) {
			return otherPod.Node
		}
	}
	return ""
}

func GetNodesForScheduling() []NodeData {
	clusterViewLock.Lock()
	defer clusterViewLock.Unlock()
	result := []NodeData{}
	for name, node := range nodeLookup {
		// TODO better way to filter
		if strings.HasSuffix(name, "control-plane") {
			continue
		}
		result = append(result, node)
	}
	return result
}

func PrintClusterView() {
	clusterViewLock.Lock()
	defer clusterViewLock.Unlock()
	for k, node := range nodeLookup {
		valuesList := []string{}
		for _, v := range node.TypeToLoad {
			valuesList = append(valuesList, fmt.Sprintf("%f", v))
		}
		log.Printf("ClusterView: %s: %s\n", k, strings.Join(valuesList, " "))
	}
}

func unbindPodFromNode(pod PodData) {
	if pod.Node == "" {
		return
	}
	node, found := nodeLookup[pod.Node]
	if !found {
		return
	}
	node.TypeToLoad[pod.Interference.TaskType] -= pod.Interference.Load
}

func PodToString(pod *v1.Pod) string {
	return pod.Namespace + ":" + pod.Name
}

func PodIdString(id PodId) string {
	// TODO implement interface
	return id.Namespace + ":" + id.Name
}

