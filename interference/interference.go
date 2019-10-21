package interference

import (
	"fmt"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sync"
)

const testLog = false

func testLogPods(podMetrics * v1beta1.PodMetrics) {
	if !testLog {
		return
	}
	fmt.Printf("Interference got pod metrics %s:%s\n", podMetrics.ObjectMeta.Namespace,
		podMetrics.ObjectMeta.Name)
}

func testLogNodes(nodeMetrics *v1beta1.NodeMetrics) {
	if !testLog {
		return
	}
	fmt.Printf("Interference got node metrics %s\n", nodeMetrics.ObjectMeta.Name)
}

func TrainInterferenceModel(wg *sync.WaitGroup, podsChan chan v1beta1.PodMetrics, nodesChan chan v1beta1.NodeMetrics) {
	for {
		select {
		case podMetrics := <-podsChan:
			testLogPods(&podMetrics)
		case nodeMetrics := <-nodesChan:
			testLogNodes(&nodeMetrics)
		}
	}
}

// TODO
func getTaskInterferenceInfo() {

}
