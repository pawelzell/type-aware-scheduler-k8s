package main

import (
	"bufio"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"log"
	db_client "type-aware-scheduler/db-client"
)

const collectOSMetricsInterval = 30 * time.Second
const collectOSMetricsEndpointURL = "http://localhost:9100/metrics"
const osMetricsMeasurementName = "metric/os"
var metricWhitelistPrefixes = [...]string{""}

func shouldCollectMetric(text string) bool {
	for _, prefix := range metricWhitelistPrefixes {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}

func parseOSMetricDatapoint(text string) (result db_client.Datapoint, err error)  {
	ind := strings.LastIndex(text, " ")
	if ind == -1 {
		err = errors.New("Unexpected format of line of metric endpoint response (a space character not found): " + text)
		return
	}
	result.Key = text[:ind]
	result.Value, err = strconv.ParseFloat(text[ind+1:len(text)], 64)
	if err != nil {
		return
	}
	return
}

func handleCollectOSMetricsResponse(db db_client.DBClient, resp *http.Response) {
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("ERROR: Response status code %d != 200. Will try again later.", resp.StatusCode)
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	datapoints := make([]db_client.Datapoint, 0)
	log.Println("Read response")
	for ; scanner.Scan(); {
		text := scanner.Text()
		if strings.HasPrefix(text, "#") { // Skip comment
			continue
		}
		if !shouldCollectMetric(text) {
			continue
		}
		datapoint, err := parseOSMetricDatapoint(text)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		datapoints = append(datapoints, datapoint)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error while scanning response body.\n")
	}
	//for _, datapoint := range datapoints {
	//	log.Printf("Datapoint %s: %f", datapoint.Key, datapoint.Value)
	//}
	err := db.InsertDatapoints(osMetricsMeasurementName, datapoints)
	if err != nil {
		log.Printf("os-metrics-collector: ERROR - failed to insert metrics datapoints to database %s\n",
			err.Error())
	}
}

func main() {
	db, err := db_client.NewDBClient("http://127.0.0.1:8086", "root", "root", "type_aware_scheduler")
	if err != nil {
		panic(err)
	}
	for {
		resp, err := http.Get(collectOSMetricsEndpointURL)
		if err == nil {
			handleCollectOSMetricsResponse(db, resp)
		} else {
			log.Printf("Failed to collect OS metrics from url %s will try again later\n", collectOSMetricsEndpointURL)
		}
		time.Sleep(collectOSMetricsInterval)
	}

}
