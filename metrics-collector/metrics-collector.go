package main

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"time"
	db_client "type-aware-scheduler/db-client"
	scheduler_config "type-aware-scheduler/scheduler-config"
)

const fetchMetricsInterval = 5 * time.Second

func CollectMetrics(config *rest.Config) {
	log.Println("Starting metrics-collector")
	clientset, err := metricsv.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	dbClient, err := db_client.NewDBClientFromLocalConfig()
	if err != nil {
		panic(err.Error())
	}
	for {
		// PODS
		podMetricsList, err := clientset.MetricsV1beta1().PodMetricses("").List(metav1.ListOptions{})
		if err == nil {
			//log.Printf("Collector: Got metrics for %d pods\n", len(podMetricsList.Items))
			err = dbClient.SavePodMetrics(podMetricsList)
			if err != nil {
				log.Println("metrics-collector: ERROR - failed to save pod metrics to database")
			} else {
				log.Println("Read pods metrics")
			}
		} else {
			log.Println("metrics-collector: ERROR - failed to get pods metrics from kubernetes api")
		}
		// NODES
		nodeMetricsList, err := clientset.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
		if err == nil {
			//log.Printf("Collector: Got metrics for %d nodes\n", len(nodeMetricsList.Items))
			err = dbClient.SaveNodeMetrics(nodeMetricsList)
			if err != nil {
				log.Println("metrics-collector: ERROR - failed to save nodes metrics to database")
			} else {
				log.Println("Read nodes metrics")
			}
	} else {
			log.Println("metrics-collector: ERROR - failed to get nodes metrics from kubernetes api")
		}
		time.Sleep(fetchMetricsInterval)
	}
}

func main() {
	fmt.Println("Starting metrics collector")
	config, err := scheduler_config.GetConfigInCluster()
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}
	CollectMetrics(config)
}
