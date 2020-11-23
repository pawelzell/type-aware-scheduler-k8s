package main

import (
	"fmt"
	"log"
	"os"
	"time"
	cluster_view "type-aware-scheduler/cluster-view"
	core "type-aware-scheduler/core-scheduler"
	"type-aware-scheduler/interference"
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
	log.Println("Starting core scheduler v3")
	err := interference.InitializeFromConfigFile()
	if err != nil {
		panic(err.Error())
	}
	interference.PrintConfig()
	podsChan := make(chan cluster_view.PodData, 100)
	quitChan := make(chan struct{})
	defer close(podsChan)
	defer close(quitChan)
	config, err := scheduler_config.GetConfigInCluster()
	if err != nil {
		panic(err.Error())
	}

	schedulerType := os.Getenv(scheduler_config.SchedulerTypeEnvKey)
	var scheduler core.Scheduler
	log.Printf("Scheduler type %s\n", schedulerType)
	if schedulerType == scheduler_config.RandomSchedulerType {
		log.Println("Initializing random scheduler")
		offlineExpConfigChan := make(chan scheduler_config.OfflineSchedulingExperiment)
		defer close(offlineExpConfigChan)
		offlineExpConfigReader := scheduler_config.NewConfigReader(scheduler_config.OfflineExpConfigPath,
			offlineExpConfigChan)

		schedulerDecisionMaker := core.NewRandomSchedulingDecisionMaker(offlineExpConfigChan)
		go schedulerDecisionMaker.RunExperimentWatcher()
		go offlineExpConfigReader.Run()
		scheduler = core.NewScheduler(config, podsChan, &schedulerDecisionMaker, schedulerType)
	} else if schedulerType == scheduler_config.RoundRobinSchedulerType {
		log.Println("Initializing round robin scheduler")
		schedulerDecisionMaker := core.NewRoundRobinSchedulingDecisionMaker()
		scheduler = core.NewScheduler(config, podsChan, &schedulerDecisionMaker, schedulerType)
	} else if schedulerType == scheduler_config.OfflineSchedulerType {
		log.Println("Initializing type aware offline scheduler")
		offlineExpConfigChan := make(chan scheduler_config.OfflineSchedulingExperiment)
		defer close(offlineExpConfigChan)
		offlineExpConfigReader := scheduler_config.NewConfigReader(scheduler_config.OfflineExpConfigPath,
			offlineExpConfigChan)

		schedulerDecisionMaker := core.NewOfflineSchedulingDecisionMaker(offlineExpConfigChan)
		go schedulerDecisionMaker.RunExperimentWatcher()
		go offlineExpConfigReader.Run()
		scheduler = core.NewScheduler(config, podsChan, &schedulerDecisionMaker, schedulerType)
	} else {
		panic(fmt.Sprintf("Unknown scheduler type read from environment: %s", schedulerType))
	}

	cluster_view.InitClusterView(config, podsChan, quitChan, schedulerType)
	go ClusterViewPrinter()
	scheduler.Run(quitChan)
}

func ClusterViewPrinter() {
	for {
		cluster_view.PrintClusterView()
		time.Sleep(15 * time.Second)
	}
}


