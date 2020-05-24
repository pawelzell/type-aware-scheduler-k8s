package main

import (
	"fmt"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sync"
	"time"
	cluster_view "type-aware-scheduler/cluster-view"
	inter "type-aware-scheduler/interference"
	metrics "type-aware-scheduler/metrics-collector"
	core "type-aware-scheduler/core-scheduler"
	scheduler_config "type-aware-scheduler/scheduler-config"
)

type NodeTest struct {
	Data float32
}

type NodeDataTest struct {
	Vector []float32
	Data *NodeTest
}

func test() {
	var lookup  = make(map[string]NodeDataTest)
	lookup["foo"] = NodeDataTest{
		Data: new(NodeTest),
		Vector: new([2]float32)[:],
	}
	fmt.Printf("Vector data %f\n", lookup["foo"].Vector[0])
	data, _ := lookup["foo"]
	data.Vector[0] += 1
	fmt.Printf("Vector data %f\n", data.Vector[0])
	fmt.Printf("Vector data %f\n", lookup["foo"].Vector[0])
}

func main() {
	fmt.Println("Starting core scheduler v2")
	var wg sync.WaitGroup
	podsMetricsChan := make(chan v1beta1.PodMetrics, 100)
	nodesMetricsChan := make(chan v1beta1.NodeMetrics, 100)
	podsChan := make(chan cluster_view.PodData, 100)
	quitChan := make(chan struct{})
	defer close(podsMetricsChan)
	defer close(nodesMetricsChan)
	defer close(podsChan)
	defer close(quitChan)

	offlineExpConfigChan := make(chan scheduler_config.OfflineSchedulingExperiment)
	defer close(offlineExpConfigChan)
	offlineExpConfigReader := scheduler_config.NewConfigReader(scheduler_config.OfflineExpConfigPath,
		offlineExpConfigChan)
	offlineSchedulerDecisionMaker := core.NewOfflineSchedulingDecisionMaker(offlineExpConfigChan)
	go offlineSchedulerDecisionMaker.RunExperimentWatcher()
	go offlineExpConfigReader.Run()


	config, err := scheduler_config.GetConfigInCluster()
	if err != nil {
		panic(err.Error())
	}
	cluster_view.InitClusterView(config, podsChan, quitChan)

	wg.Add(3) // TODO change to 3 when collect metrics enabled
	go inter.TrainInterferenceModel(&wg, podsMetricsChan, nodesMetricsChan)
	go metrics.CollectMetrics(config, &wg, podsMetricsChan, nodesMetricsChan)
	go ClusterViewPrinter()
	scheduler := core.NewScheduler(config, podsChan, &offlineSchedulerDecisionMaker)
	scheduler.Run(quitChan)
	wg.Wait()
}

func ClusterViewPrinter() {
	for {
		cluster_view.PrintClusterView()
		time.Sleep(15 * time.Second)
	}
}


