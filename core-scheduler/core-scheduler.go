package main

import (
	"fmt"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sync"
	inter "type-aware-scheduler/interference"
	metrics "type-aware-scheduler/metrics-collector"
)

func main() {
	fmt.Println("Starting core scheduler")
	var wg sync.WaitGroup
	podsMetricsChan := make(chan v1beta1.PodMetrics, 100)
	nodesMetricsChan := make(chan v1beta1.NodeMetrics, 100)
	wg.Add(1)
	go inter.TrainInterferenceModel(&wg, podsMetricsChan, nodesMetricsChan)
	wg.Add(1)
	go metrics.CollectMetricsOutOfCluster(&wg, podsMetricsChan, nodesMetricsChan)
	wg.Wait()
}
