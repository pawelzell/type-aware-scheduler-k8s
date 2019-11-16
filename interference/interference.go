package interference

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sync"
)

type PodInfo struct {
	TaskType uint32
	Load float64
}

type CoefficientsMatrix [][]float64

var Coefficients = CoefficientsMatrix{{1., 2.}, {2., 1.}}

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

// TODO to implement
func PredictPodInfo(pod *v1.Pod) (result PodInfo, err error){
	result = PodInfo{0, 100.}
	return
}

func GetInterferenceCoefficients() (result CoefficientsMatrix, err error){
	result = Coefficients
	return
}
