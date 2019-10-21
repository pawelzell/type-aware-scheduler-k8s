package main

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sync"
	cluster_view "type-aware-scheduler/cluster-view"
	"type-aware-scheduler/scheduler-config"
	inter "type-aware-scheduler/interference"
	metrics "type-aware-scheduler/metrics-collector"
)

func main() {
	fmt.Println("Starting core scheduler")
	var wg sync.WaitGroup
	podsMetricsChan := make(chan v1beta1.PodMetrics, 100)
	nodesMetricsChan := make(chan v1beta1.NodeMetrics, 100)
	podChan := make(chan *v1.Pod, 100)
	quitChan := make(chan struct{})
	defer close(podsMetricsChan)
	defer close(nodesMetricsChan)
	defer close(podChan)
	defer close(quitChan)
	config, err := scheduler_config.GetConfigOutOfCluster()
	if err != nil {
		panic(err.Error())
	}
	cluster_view.InitPodNodesWatchers(config, podChan, quitChan)

	wg.Add(2)
	go inter.TrainInterferenceModel(&wg, podsMetricsChan, nodesMetricsChan)
	go metrics.CollectMetrics(config, &wg, podsMetricsChan, nodesMetricsChan)
	wg.Wait()
}
