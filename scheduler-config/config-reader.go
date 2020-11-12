package scheduler_config

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
	"type-aware-scheduler/interference"
)

const refreshConfigInterval = 15 * time.Second

type OfflineSchedulingExperiment struct {
	Id string
	TaskTypesIds []int
}

type OfflineExpConfigReader struct {
	experiment     OfflineSchedulingExperiment
	experimentRead bool
	configPath     string
	//updateIntervalSeconds time.Duration
	experimentQueue       chan<- OfflineSchedulingExperiment
}

func NewConfigReader(configPath string, experimentQueue chan<- OfflineSchedulingExperiment) OfflineExpConfigReader {
	return OfflineExpConfigReader{
		experimentRead:        false,
		configPath:            configPath,
		//updateIntervalSeconds: 30 * time.Second,
		experimentQueue:       experimentQueue,
	}
}

// Expected header format:
// # scheduler <id> <comma separated list of types without spaces>
func TryToParseOfflineExperimentFromLine(line string) (exp OfflineSchedulingExperiment, err error) {
	if !strings.HasPrefix(line, "#") {
		err = errors.New("experiment line has to start with #")
		return
	}
	components := strings.Split(line, " ")
	if len(components) != 4 {
		err = errors.New(fmt.Sprintf("experiment line must have 4 space separated components " +
			"got %d", len(components)))
		return
	}
	if components[1] != "scheduler" {
		err = errors.New("experiment line must have word scheduler as a 2nd component")
		return
	}
	types := strings.Split(components[3], ",")
	taskTypeIds := make([]int, 0)
	for _, t := range types {
		typeId, found := interference.TypeStringToId[t]
		if !found {
			err = errors.New(fmt.Sprintf("unknown task type %s", t))
			return
		}
		taskTypeIds = append(taskTypeIds, typeId)
	}
	exp = OfflineSchedulingExperiment{components[2], taskTypeIds}
	return
}

func (r*OfflineExpConfigReader) ReadConfig() (exp OfflineSchedulingExperiment, err error) {
	f, err := os.Open(r.configPath)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for ; scanner.Scan(); {
		line := scanner.Text()
		exp, err = TryToParseOfflineExperimentFromLine(line)
		if err == nil {
			return
		}
	}
	err = errors.New("unsupported file format - a line with an experiment description not found")
	return
}

func areEqual(experiment OfflineSchedulingExperiment, newExperiment OfflineSchedulingExperiment) bool {
	return reflect.DeepEqual(experiment, newExperiment)
}

func (r*OfflineExpConfigReader) Run() {
	for {
		newExperiment, err := r.ReadConfig()
		if err == nil {
			if !r.experimentRead || !areEqual(r.experiment, newExperiment) {
				r.experiment = newExperiment
				r.experimentRead = true
				r.experimentQueue <- r.experiment
				log.Printf("OfflineExpConfigReader: %s exp config parsed\n", r.experiment.Id)
			} else {
				//log.Printf("Config not updated\n")
			}
		} else {
			log.Printf(err.Error())
		}
		time.Sleep(refreshConfigInterval)
	}
}
