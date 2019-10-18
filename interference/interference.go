package interference

import (
	"fmt"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sync"
)

func TrainInterferenceModel(wg *sync.WaitGroup, podsChan chan v1beta1.PodMetrics, nodesChan chan v1beta1.NodeMetrics) {
	for {
		select {
		case podMetrics := <-podsChan:
			fmt.Printf("Interference got pod metrics %s:%s\n", podMetrics.ObjectMeta.Namespace,
				podMetrics.ObjectMeta.Name)
		case nodeMetrics := <-nodesChan:
			fmt.Printf("Interference got node metrics %s\n", nodeMetrics.ObjectMeta.Name)
		}
	}
}

func getTaskInterferenceInfo() {

}
