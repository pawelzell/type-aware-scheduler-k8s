package interference

import (
	"bufio"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"os"
	"strconv"
	"strings"
)

type PodInfo struct {
	TaskType int
	Load float64
}

type CoefficientsMatrix = [][]float64

type ModelType struct {
	Description string
	NTaskTypes int
	NodeToCoefficients map[string]CoefficientsMatrix
	TypeToResourceRequests []float64
	NodeToResourceCapacity map[string]float64
}

const interferenceConfigPath = "scheduler_config"
const sectionSeparator = "-"
const descriptionFieldPrefix = "description: "
const defaultNodeCapacity = 60.
const defaultHadoopSize = 3.

var NumberOfTaskTypes = 0
var TypeStringToId = map[string]int{}
var Model = ModelType{}
var RoleToType = map[string]string{"ycsb": "redis_ycsb", "redis": "redis_ycsb",
	"wrk": "wrk", "apache": "wrk",
	"hadoopmaster": "hadoop", "hadoopslave": "hadoop",
	"sysbench": "sysbench", "mysql": "sysbench",
	"linpack": "linpack"}

func InitializeFromConfigFile() (err error) {
	f, err := os.Open(interferenceConfigPath)
	if err != nil{
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	// Read header section
	for ; scanner.Scan(); {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, sectionSeparator) {
			break
		}
		return errors.New("Unexpected start of line in header section")
	}
	err = readTypeListSection(scanner)
	if err != nil {
		return
	}
	err = readModelSection(scanner)
	if err != nil {
		return
	}
	return nil
}

func readTypeListSection(scanner *bufio.Scanner) (err error) {
	var parsed = false
	for ; scanner.Scan(); {
		line := scanner.Text()
		if strings.HasPrefix(line, sectionSeparator) {
			break
		}
		if parsed {
			return errors.New("TypeList section has multiple lines - expected single line")
		}
		line = strings.TrimSpace(line)
		types := strings.Split(line, " ")
		for i, t := range types {
			TypeStringToId[t] = i
		}
		NumberOfTaskTypes = len(types)
		parsed = true
	}
	if !parsed {
		return errors.New("TypeList section empty")
	}
	return nil
}

func readHeader(scanner* bufio.Scanner) (err error) {
	if !scanner.Scan() {
		return errors.New("Error when trying to read model desciption")
	}
	line := strings.TrimSpace(scanner.Text())
	if !strings.HasPrefix(line, descriptionFieldPrefix) {
		return errors.New("Error model descripiton incorrect field")
	}
	line = strings.TrimPrefix(line, descriptionFieldPrefix)
	Model.Description = line
	return nil
}

func readModelSection(scanner *bufio.Scanner) (err error) {
	Model.NTaskTypes = NumberOfTaskTypes
	Model.NodeToCoefficients = make(map[string]CoefficientsMatrix)
	err = readHeader(scanner)
	if err != nil {
		return err
	}
	for ; scanner.Scan(); {
		line := scanner.Text()
		if strings.HasPrefix(line, sectionSeparator) {
			break
		}
		node := strings.TrimSpace(line)
		matrix := make([][]float64, NumberOfTaskTypes)
		for i := 0 ; i < NumberOfTaskTypes; i++ {
			matrix[i] = make([]float64, NumberOfTaskTypes)
			if !scanner.Scan() {
				return errors.New("Error while reading model matrix")

			}
			line = strings.TrimSpace(scanner.Text())
			cells := strings.Split(line, " ")
			if len(cells) != NumberOfTaskTypes {
				return errors.New("Unexpected number of cells in a matrix row")
			}
			for j:=0; j < NumberOfTaskTypes; j++ {
				val, err := strconv.ParseFloat(cells[j], 64)
				if err != nil {
					return errors.New("Failed to parse matrix cell as float")
				}
				matrix[i][j] = val
			}
		}
		Model.NodeToCoefficients[node] = matrix
	}
	Model.TypeToResourceRequests = make([]float64, NumberOfTaskTypes)
	Model.TypeToResourceRequests[TypeStringToId["hadoop"]] = defaultHadoopSize
	Model.NodeToResourceCapacity = make(map[string]float64)
	for node, _ := range Model.NodeToCoefficients {
		Model.NodeToResourceCapacity[node] = defaultNodeCapacity
	}
	return nil
}

func PredictPodInfo(pod *v1.Pod) (result PodInfo, err error){
	result = PodInfo{0, 1.}
	components := strings.Split(pod.Name, "-")
	if len(components) < 5 {
		err = errors.New(fmt.Sprintf("Interference: ERROR - pod name has unexpected number of components %d\n",
			len(components)))
		return
	}
	role := components[4]
	t, found := RoleToType[role]
	if !found {
		err = errors.New(fmt.Sprintf("Interference: ERROR - unknown pod role %s\n", role))
		return
	}
	typeId, found := TypeStringToId[t]
	if !found {
		err = errors.New(fmt.Sprintf("Interference: ERROR - unknown pod type %s\n", t))
		return
	}
	result.TaskType = typeId
	return
}

func GetInterferenceModel() (result ModelType, err error){
	result = Model
	return
}

func PrintConfig() {
	fmt.Println("Print interference config")
	fmt.Println("Model:")
	fmt.Println(Model.Description)
	for node, matrix := range Model.NodeToCoefficients {
		fmt.Println(node)
		for i := 0; i < NumberOfTaskTypes; i++ {
			for j := 0; j < NumberOfTaskTypes; j++ {
				fmt.Printf("%f ", matrix[i][j])
			}
			fmt.Println("")
		}
	}
	fmt.Printf("NumberOfTasks: %d\n", Model.NTaskTypes)
	fmt.Println("NodeToResourceCapacity:")
	for n, c := range Model.NodeToResourceCapacity {
		fmt.Printf("%s %f; ", n, c)
	}
	fmt.Println("\nTypeToResourceRequests:")
	for _, r := range Model.TypeToResourceRequests {
		fmt.Printf("%f; ", r)
	}
	fmt.Println("\nTypeStringToId:")
	for t, i := range TypeStringToId {
		fmt.Printf("%s %d; ", t, i)
	}
	fmt.Println("")
}
