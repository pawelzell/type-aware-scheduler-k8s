package scheduler_config

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

const SchedulerName = "type-aware-scheduler"
const InfluxDBUsernameEnvKey = "INFLUXDB_USERNAME"
const InfluxDBPasswordEnvKey = "INFLUXDB_PASSWORD"
const InfluxDBHostnameEnvKey = "INFLUXDB_HOST"
const InfluxDBDatabaseEnvKey = "INFLUXDB_DATABASE"
const OfflineExpConfigPath = "exp"
// TODO load number of task types from yaml configuration
const NumberOfTaskTypes = 4
var TypeStringToId = map[string]int{"redis_ycsb": 0, "wrk": 1, "hadoop": 2, "linpack": 3}

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

