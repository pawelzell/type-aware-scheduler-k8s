package cluster_view

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"log"
	"strings"
)

func getApplicationInstance(pod PodData) string {
	var index = strings.LastIndex(pod.Node, "ai_")
	if index < 0 {
		return ""
	}
	return pod.Node[index:len(pod.Node)]
}

func BindPodToNode(id PodId, nodeName string) error {
	clusterViewLock.Lock()
	defer clusterViewLock.Unlock()
	pod, found := podLookup[id]
	if !found {
		return errors.New("Pod did not found")
	}

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
	for _, otherPod := range podLookup {
		if getApplicationInstance(pod) == getApplicationInstance(otherPod) {
			return pod.Node
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

