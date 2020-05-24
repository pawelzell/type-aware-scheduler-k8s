/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package metrics_collector

import (
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"sync"
	"time"
	db_client "type-aware-scheduler/db-client"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const fetchMetricsInterval = 30 * time.Second

func CollectMetrics(config *rest.Config, wg *sync.WaitGroup, podsMetricsChan chan v1beta1.PodMetrics,
	nodesMetricsChan chan v1beta1.NodeMetrics) {
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
		// TODO: handle pods and nodes with the same generic code
		// PODS
		podMetricsList, err := clientset.MetricsV1beta1().PodMetricses("").List(metav1.ListOptions{})
		if err == nil {
			//log.Printf("Collector: Got metrics for %d pods\n", len(podMetricsList.Items))
			err = dbClient.SavePodMetrics(podMetricsList)
			if err != nil {
				log.Println("metrics-collector: ERROR - failed to save pod metrics to database")
			}
			for _, podMetrics := range podMetricsList.Items {
				podsMetricsChan <- podMetrics
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
			}
			for _, nodeMetrics := range nodeMetricsList.Items {
				nodesMetricsChan <- nodeMetrics
			}
		} else {
			log.Println("metrics-collector: ERROR - failed to get nodes metrics from kubernetes api")
		}
		time.Sleep(fetchMetricsInterval)
	}
}

// Examples for error handling:
// - Use helper functions e.g. errors.IsNotFound()
// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
