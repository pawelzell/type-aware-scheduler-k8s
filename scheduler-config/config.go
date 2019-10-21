package scheduler_config

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

const SchedulerName = "type-aware-scheduler"

func GetConfigInCluster() (*rest.Config, error) {
	return rest.InClusterConfig()
}

func GetConfigOutOfCluster() (config *rest.Config, err error) {
	home := os.Getenv("HOME")
	kubeconfig := filepath.Join(home, ".kube", "kind-config-kind")

	// use the current context in kubeconfig
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	return
}

