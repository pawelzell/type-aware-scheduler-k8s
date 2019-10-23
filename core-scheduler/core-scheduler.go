package main

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"math/rand"
	"time"
	cluster_view "type-aware-scheduler/cluster-view"
	scheduler_config "type-aware-scheduler/scheduler-config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Scheduler struct {
	clientset  *kubernetes.Clientset
	podQueue   <-chan cluster_view.PodData
}

func NewScheduler(config *rest.Config, podQueue <-chan cluster_view.PodData) Scheduler {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return Scheduler{
		clientset:  clientset,
		podQueue:   podQueue,
	}
}

func (s *Scheduler) Run(quit chan struct{}) {
	wait.Until(s.ScheduleOne, 0, quit)
}

func (s *Scheduler) ScheduleOne() {
	podData := <- s.podQueue
	podFullName := cluster_view.PodToString(podData.Data)
	log.Printf("Got pod for scheduling %s\n", podFullName)
	node := MakeSchedulingDecision(podData)

	err := s.bindPod(podData.Data, node)
	if err != nil {
		log.Println("failed to bind pod", err.Error())
		return
	}
	podId := cluster_view.PodId{podData.Data.Name, podData.Data.Namespace}
	err = cluster_view.BindPodToNode(podId, node)
	if err != nil {
		log.Println("failed to bind pod in cluster view", err.Error())
		return
	}
	message := fmt.Sprintf("Placed pod %s on %s\n", podFullName, node)
	err = s.emitEvent(podData.Data, message)
	if err != nil {
		log.Println("failed to emit scheduled event", err.Error())
		return
	}
	log.Println(message)
}

func MakeSchedulingDecision(podData cluster_view.PodData) string {
	nodes := cluster_view.GetNodesForScheduling()
	n := len(nodes)
	if n <= 0 {
		return ""
	}
	return nodes[rand.Intn(n)]
}

func (s *Scheduler) emitEvent(p *v1.Pod, message string) error {
	timestamp := time.Now().UTC()
	_, err := s.clientset.CoreV1().Events(p.Namespace).Create(&v1.Event{
		Count:          1,
		Message:        message,
		Reason:         "Scheduled",
		LastTimestamp:  metav1.NewTime(timestamp),
		FirstTimestamp: metav1.NewTime(timestamp),
		Type:           "Normal",
		Source: v1.EventSource{
			Component: scheduler_config.SchedulerName,
		},
		InvolvedObject: v1.ObjectReference{
			Kind:      "Pod",
			Name:      p.Name,
			Namespace: p.Namespace,
			UID:       p.UID,
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: p.Name + "-",
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) bindPod(p *v1.Pod, node string) error {
	return s.clientset.CoreV1().Pods(p.Namespace).Bind(&v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
		},
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       node,
		},
	})
}


