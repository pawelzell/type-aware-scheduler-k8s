package main

import (
	"fmt"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	client "github.com/influxdata/influxdb1-client/v2"
	"log"
	"os"
	"strings"
	"time"
	scheduler_config "type-aware-scheduler/scheduler-config"
)

type DBClient struct {
	client client.Client
	database string
}

func NewDBClient(addr string, username string, password string, database string) (r DBClient, err error) {
	// TODO get addr, user, password from environment
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: addr,
		Username: username,
		Password: password,
	})
	if err == nil {
		r = DBClient{c, database}
	}
	return
}

func formatMetricName(ss ...string) string {
	args := []string{"metric"}
	for _, s := range ss {
		args = append(args, s)
	}
	return strings.Join(args, "/")
}

func addMetricPoint(metricName string, quantity int64, timestamp time.Time, bp client.BatchPoints) (err error) {
	tags := map[string]string{}
	fields := map[string]interface{}{
		"quantity": quantity,
	}
	pt, err := client.NewPoint(metricName, tags, fields, timestamp)
	if err == nil {
		bp.AddPoint(pt)
	}
	return
}

func addPodMetrics(pod v1beta1.PodMetrics, bp client.BatchPoints) (err error) {
	for _, container := range pod.Containers {
		for resource, quantity := range container.Usage {
			metricName := formatMetricName("pod", pod.Namespace, pod.Name, container.Name, resource.String())
			err = addMetricPoint(metricName, quantity.Value(), pod.Timestamp.UTC(), bp)
			if err != nil {
				return
			}
		}
	}
	return
}

func addNodeMetrics(node v1beta1.NodeMetrics, bp client.BatchPoints) (err error) {
	for resource, quantity := range node.Usage {
		metricName := formatMetricName("node", node.Name, resource.String())
		err = addMetricPoint(metricName, quantity.Value(), node.Timestamp.UTC(), bp)
		if err != nil {
			return
		}
	}
	return
}

func (c *DBClient) SavePodMetrics(pods v1beta1.PodMetricsList) (err error) {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database: c.database,
		Precision: "s",
	})
	for _, pod := range pods.Items {
		err = addPodMetrics(pod, bp)
		if err != nil {
			return
		}
	}
	return c.client.Write(bp)
}

func (c *DBClient) SaveNodeMetrics(nodes v1beta1.NodeMetricsList) (err error) {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database: c.database,
		Precision: "s",
	})
	for _, node := range nodes.Items {
		err = addNodeMetrics(node, bp)
		if err != nil {
			return
		}
	}
	return c.client.Write(bp)
}

func addSampleDatapoint(c client.Client) {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "type_aware_scheduler",
		Precision: "s",
	})

	// Create a point and add to batch
	tags := map[string]string{}
	fields := map[string]interface{}{
		"test_data":   42.42,
	}
	pt, err := client.NewPoint("test_data", tags, fields, time.Now())
	if err != nil {
		log.Println("Error creating data point: ", err.Error())
		panic(err.Error())
	}
	bp.AddPoint(pt)
	err = c.Write(bp)
	if err != nil {
		log.Println("Error inserting data point: ", err.Error())
		panic(err.Error())
	}
	log.Printf("Sample datapoint added\n")
}

func main() {
	log.Printf("Starting db client test\n")
	time.Sleep(5 * time.Second)
	//10.0.150.128
	addr := "http://" + os.Getenv(scheduler_config.InfluxDBHostnameEnvKey) + ":8086"
	username := os.Getenv(scheduler_config.InfluxDBUsernameEnvKey)
	password := os.Getenv(scheduler_config.InfluxDBPasswordEnvKey)
	database := os.Getenv(scheduler_config.InfluxDBDatabaseEnvKey)
	log.Printf("config: %s %s %s %s\n", addr, username, password, database)
	time.Sleep(5 * time.Second)
	c, err := NewDBClient(addr, username, password, database)
	log.Printf("Client created\n")
	time.Sleep(5 * time.Second)
	if err != nil {
		log.Fatalln("Error creating InfluxDB Client: ", err.Error())
	}
	log.Printf("Client created successfully\n")
	time.Sleep(5 * time.Second)
	q := client.Query{Command: "create DATABASE " + database}
	r, err := c.client.Query(q)
	if err == nil && r.Error() == nil {
		log.Println(r.Results)
	}
	time.Sleep(5 * time.Second)
	//defer c.client.Close()
	for {
		fmt.Printf("Trying to add sample\n")
		addSampleDatapoint(c.client)
		time.Sleep(15 * time.Second)
	}
}
