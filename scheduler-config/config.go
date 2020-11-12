package scheduler_config

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

const SchedulerTypeEnvKey = "SCHEDULER_TYPE"
const RandomSchedulerType = "random-scheduler"
const RoundRobinSchedulerType = "round-robin-scheduler"
const OfflineSchedulerType = "type-aware-scheduler"
const InfluxDBUsernameEnvKey = "INFLUXDB_USERNAME"
const InfluxDBPasswordEnvKey = "INFLUXDB_PASSWORD"
const InfluxDBHostnameEnvKey = "INFLUXDB_HOST"
const InfluxDBDatabaseEnvKey = "INFLUXDB_DATABASE"
const RandomSchedulerFirstNodeProbability = 1. / 3.
const OfflineExpConfigPath = "exp"
// TODO load number of task types from yaml configuration

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

