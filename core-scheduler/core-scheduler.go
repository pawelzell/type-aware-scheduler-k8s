package core_scheduler

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
	cluster_view "type-aware-scheduler/cluster-view"
	"type-aware-scheduler/interference"
	scheduler_config "type-aware-scheduler/scheduler-config"
)

const loadEps = 0.000001

type Scheduler struct {
	clientset  *kubernetes.Clientset
	podQueue   <-chan cluster_view.PodData
	DecisionMaker SchedulingDecisionMaker
	SchedulerName string
}

type SchedulingDecisionMaker interface {
	MakeSchedulingDecision(cluster_view.PodData, []cluster_view.NodeData) string
}

type GreedySchedulingDecisionMaker struct {
	Model interference.ModelType
}

func (m *GreedySchedulingDecisionMaker) MakeSchedulingDecision(pod cluster_view.PodData, nodes []cluster_view.NodeData) string {
	n := len(nodes)
	if n <= 0 {
		return ""
	}
	// Compute current maximum cost
	currentCost := math.Inf(-1)
	for _, node := range nodes {
		nodeName := node.Data.Name
		currentCost = math.Max(currentCost, computeNodeMaxCost(node, m.Model.NodeToCoefficients[nodeName], m.Model.NTaskTypes))
	}
	bestCost := math.Inf(1)
	bestNode := ""
	for _, node := range nodes {
		node.TypeToLoad[pod.Interference.TaskType] += pod.Interference.Load
		nodeName := node.Data.Name
		newCost := math.Max(currentCost, computeNodeMaxCost(node, m.Model.NodeToCoefficients[nodeName], m.Model.NTaskTypes))
		node.TypeToLoad[pod.Interference.TaskType] -= pod.Interference.Load
		if newCost < bestCost {
			bestCost = 	newCost
			bestNode = nodeName
		}
	}
	return bestNode
}

type OfflineSchedulingDecisionMaker struct {
	Model interference.ModelType
	UpdateChan <-chan scheduler_config.OfflineSchedulingExperiment
	ExperimentLock *sync.RWMutex

	// Map from node name and task type id to number of tasks at the end of scheduling
	SchedulingDesiredState map[string][]float64
	Experiment             scheduler_config.OfflineSchedulingExperiment
}

func NewOfflineSchedulingDecisionMaker(updateChan <-chan scheduler_config.OfflineSchedulingExperiment) OfflineSchedulingDecisionMaker {
	model, err := interference.GetInterferenceModel()
	if err != nil {
		log.Fatal("Failed to obtain the interference model")
	}
	return OfflineSchedulingDecisionMaker{
		Model:                  model,
		UpdateChan:             updateChan,
		ExperimentLock:         new(sync.RWMutex),
		SchedulingDesiredState: nil,
		Experiment:             scheduler_config.OfflineSchedulingExperiment{"", nil},
	}
}

func (m *OfflineSchedulingDecisionMaker) MakeSchedulingDecision(pod cluster_view.PodData, nodes []cluster_view.NodeData) string {
	if m.SchedulingDesiredState == nil{
		return ""
	}
	m.ExperimentLock.Lock()
	defer m.ExperimentLock.Unlock()
	// Schedule on the first free node
	typeId := pod.Interference.TaskType
	choosenNode := ""
	for _, nodeData := range nodes {
		if nodeData.TypeToLoad[typeId] + loadEps < m.SchedulingDesiredState[nodeData.Data.Name][typeId] {
			choosenNode = nodeData.Data.Name
			break
		}
	}
	if choosenNode == "" {
		log.Printf("ERROR: got unexpected pod %s type %d load %f outside of offline schedule plan",
			pod.Data.Name, pod.Interference.TaskType, pod.Interference.Load)
	}
	return choosenNode
}

func computeTypeTaskCount(model interference.ModelType, exp scheduler_config.OfflineSchedulingExperiment) []int {
	result := make([]int, model.NTaskTypes)
	for _, typeId := range exp.TaskTypesIds {
		result[typeId]++
	}
	return result
}

func checkResourceConstraints(model interference.ModelType, nodeLookup []string, schedule map[string][]float64) bool {
	for _, nodeName := range nodeLookup {
		nodeSchedule := schedule[nodeName]
		resourceRequest := 0.
		for i := 0; i < model.NTaskTypes; i++ {
			resourceRequest += model.TypeToResourceRequests[i] * nodeSchedule[i]
		}
		if resourceRequest > model.NodeToResourceCapacity[nodeName] {
			return false
		}
	}
	return true
}

func ComputeScheduleCost(model interference.ModelType, nodeLookup []string, schedule map[string][]float64) float64 {
	maxCost := math.Inf(-1)
	for _, nodeName := range nodeLookup {
		coefficients := model.NodeToCoefficients[nodeName]
		nodeSchedule := schedule[nodeName]
		for i := 0; i < model.NTaskTypes; i++ {
			if nodeSchedule[i] <= 0. {
				continue
			}
			// We count only load of the others
			nodeSchedule[i] -= 1
			typeCost := 0.
			for j := 0; j < model.NTaskTypes; j++ {
				typeCost += coefficients[i][j] * nodeSchedule[j]
			}
			nodeSchedule[i] += 1
			maxCost = math.Max(maxCost, typeCost)
		}
	}
	return maxCost
}

func copySchedule(dest map[string][]float64, src map[string][]float64, nodeLookup []string) {
	for _, nodeName := range nodeLookup {
		for i := 0; i < len(src[nodeName]); i++ {
			dest[nodeName][i] = src[nodeName][i]
		}
	}
}

func (s *OptimalScheduleSolver) solveHelper(curType int, curNode int) {
	if curType >= s.typeCount {
		if !checkResourceConstraints(s.model, s.nodeLookup, s.curSchedule) {
			return // Tasks do not fit in nodes resource limits
		}
		curCost := ComputeScheduleCost(s.model, s.nodeLookup, s.curSchedule)
		//log.Printf("Compute schedule cost %f\n", curCost)
		if curCost < s.bestScheduleCost {
			s.bestScheduleCost = curCost
			copySchedule(s.bestSchedule, s.curSchedule, s.nodeLookup)
		}
		return
	}
	if curNode >= s.nodeCount {
		if s.typeToTaskCount[curType] <= 0 {
			s.solveHelper(curType+1, 0)
		}
		return
	}
	curNodeName := s.nodeLookup[curNode]
	for taskCount := 0 ; taskCount <= s.typeToTaskCount[curType]; taskCount++ {
		s.typeToTaskCount[curType] -= taskCount
		s.curSchedule[curNodeName][curType] += float64(taskCount)
		s.solveHelper(curType, curNode + 1)
		s.typeToTaskCount[curType] += taskCount
		s.curSchedule[curNodeName][curType] -= float64(taskCount)
	}
}

func (s *OptimalScheduleSolver) Solve() (map[string][]float64, float64) {
	s.solveHelper(0, 0)
	if s.bestScheduleCost >= math.Inf(1.) {
		panic("No schedule found")
	}
	return s.bestSchedule, s.bestScheduleCost
}

type OptimalScheduleSolver struct {
	model            interference.ModelType
	typeToTaskCount  []int
	nodeLookup       []string
	nodeCount        int
	typeCount        int
	curSchedule      map[string][]float64
	bestSchedule     map[string][]float64
	bestScheduleCost float64
}

func NewOptimalScheduleSolver(model interference.ModelType, experiment scheduler_config.OfflineSchedulingExperiment) OptimalScheduleSolver {
	curResult := make(map[string][]float64)
	bestResult := make(map[string][]float64)
	nodeLookup := make([]string, 0)
	for k, _ := range model.NodeToCoefficients {
		nodeLookup = append(nodeLookup, k)
		curResult[k] = make([]float64, model.NTaskTypes)
		bestResult[k] = make([]float64, model.NTaskTypes)
	}
	typeToTaskCount := computeTypeTaskCount(model, experiment)

	return OptimalScheduleSolver{
		model,
		typeToTaskCount,
		nodeLookup,
		len(nodeLookup),
		model.NTaskTypes,
		curResult,
		bestResult,
		math.Inf(1),
	}
}

func (m *OfflineSchedulingDecisionMaker) ComputeSchedulingDesiredState(experiment scheduler_config.OfflineSchedulingExperiment) (map[string][]float64, float64) {
	log.Printf("Computing optimal schedule for experiment %s\n", experiment.Id)
	s := NewOptimalScheduleSolver(m.Model, experiment)
	schedule, cost := s.Solve()
	log.Printf("Best schedule has cost %f\n", cost)
	log.Println(schedule)
	return schedule, cost
}

func (m *OfflineSchedulingDecisionMaker) RunExperimentWatcher() {
	for {
		exp := <-m.UpdateChan
		newPlan, cost := m.ComputeSchedulingDesiredState(exp)
		log.Printf("DecisionMaker: Got new offline experiment config with id %s, computed cost %f\n", exp.Id, cost)
		m.ExperimentLock.Lock()
		m.Experiment = exp
		m.SchedulingDesiredState = newPlan
		m.ExperimentLock.Unlock()
	}
}

func NewScheduler(config *rest.Config, podQueue <-chan cluster_view.PodData,
		decisionMaker SchedulingDecisionMaker, schedulerName string) Scheduler {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err) }
	return Scheduler{
		clientset:  clientset,
		podQueue:   podQueue,
		DecisionMaker: decisionMaker,
		SchedulerName: schedulerName,
	}
}

func (s *Scheduler) Run(quit chan struct{}) {
	wait.Until(s.ScheduleOne, 0, quit)
}

func (s *Scheduler) ScheduleOne() {
	podData := <- s.podQueue
	podFullName := cluster_view.PodToString(podData.Data)
	log.Printf("Got pod for scheduling %s\n", podFullName)

	node := cluster_view.GetNodeByApplicationInstance(podData)
	newLoad := 1.
	if node == "" {
		nodes := cluster_view.GetNodesForScheduling()
		node = s.DecisionMaker.MakeSchedulingDecision(podData, nodes)
	} else {
		newLoad = 0.
		log.Printf("Placing pod %s on the same node %s as an other pod from the same application instance\n",
			podData.Data.Name, node)
	}
	err := s.bindPod(podData.Data, node)
	if err != nil {
		log.Println("failed to bind pod", err.Error())
		return
	}
	podId := cluster_view.PodId{podData.Data.Name, podData.Data.Namespace}
	err = cluster_view.BindPodToNode(podId, newLoad, node)
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

func computeNodeMaxCost(node cluster_view.NodeData, coefficients [][]float64, nTaskTypes int) float64 {
	maxCost := math.Inf(-1)
	for i := 0; i < nTaskTypes; i++ {
		if node.TypeToLoad[i] <= 0. {
			continue
		}
		typeCost := 0.
		for j := 0; j < nTaskTypes; j++ {
			typeCost += coefficients[i][j] * node.TypeToLoad[j]
		}
		maxCost = math.Max(maxCost, typeCost)
	}
	return maxCost
}

type RandomSchedulingDecisionMaker struct {
	Model interference.ModelType
	Random *rand.Rand
	Schedule []int
	SchedulePos int
	FirstNodePr float64
	UpdateChan <-chan scheduler_config.OfflineSchedulingExperiment
	ExperimentLock *sync.RWMutex
}

func NewRandomSchedulingDecisionMaker(updateChan <-chan scheduler_config.OfflineSchedulingExperiment) RandomSchedulingDecisionMaker {
	model, err := interference.GetInterferenceModel()
	if err != nil {
		log.Fatal("Failed to obtain the interference model")
	}
	seed := time.Now().UnixNano()
	log.Printf("Initialize random generator with seed %d\n", seed)
	return RandomSchedulingDecisionMaker{
		Model:                  model,
		Random: rand.New(rand.NewSource(seed)),
		Schedule: make([]int, 0),
		SchedulePos: 0,
		FirstNodePr: scheduler_config.RandomSchedulerFirstNodeProbability,
		UpdateChan: updateChan,
		ExperimentLock: new(sync.RWMutex),
	}
}

func (m *RandomSchedulingDecisionMaker) MakeSchedulingDecision(pod cluster_view.PodData, nodes []cluster_view.NodeData) string {
	m.ExperimentLock.Lock()
	defer m.ExperimentLock.Unlock()
	sort.Sort(NodesByName(nodes)) // We do not have a guarantee on node order
	if len(nodes) < 2 || m.SchedulePos >= len(m.Schedule) {
		return ""
	}
	selectedNodeId := m.Schedule[m.SchedulePos]
	selectedNode := nodes[selectedNodeId].Data.Name
	log.Printf("Selecting node no %d which is %s according to schedule pos %d\n", selectedNodeId, selectedNode, m.SchedulePos)
	m.SchedulePos += 1
	return selectedNode
}

func (m *RandomSchedulingDecisionMaker) GetRandomSchedule(exp scheduler_config.OfflineSchedulingExperiment) []int {
	n := len(exp.TaskTypesIds)
	res := make([]int, n)
	for i:=0; i < n; i++ {
		res[i] = 0
		if float64(i) >= m.FirstNodePr * float64(n) {
			res[i] = 1
		}
	}
	m.Random.Shuffle(n, func(i, j int) {
		res[i], res[j] = res[j], res[i]
	})
	return res
}

func (m *RandomSchedulingDecisionMaker) RunExperimentWatcher() {
	for {
		exp := <-m.UpdateChan
		log.Printf("DecisionMaker: Got new offline experiment config with id %s\n", exp.Id)
		m.ExperimentLock.Lock()
		m.Schedule = m.GetRandomSchedule(exp)
		m.SchedulePos = 0
		m.ExperimentLock.Unlock()
	}
}


type RoundRobinSchedulingDecisionMaker struct {
	Model            interference.ModelType
	LastSelectedNode int
}

func NewRoundRobinSchedulingDecisionMaker() RoundRobinSchedulingDecisionMaker {
	model, err := interference.GetInterferenceModel()
	if err != nil {
		log.Fatal("Failed to obtain the interference model")
	}
	return RoundRobinSchedulingDecisionMaker{
		Model:            model,
		LastSelectedNode: -1,
	}
}

// https://golang.org/pkg/sort/
type NodesByName []cluster_view.NodeData
func (a NodesByName) Len() int           { return len(a) }
func (a NodesByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a NodesByName) Less(i, j int) bool { return a[i].Data.Name < a[j].Data.Name }

func (m *RoundRobinSchedulingDecisionMaker) MakeSchedulingDecision(pod cluster_view.PodData, nodes []cluster_view.NodeData) string {
	sort.Sort(NodesByName(nodes)) // We do not have a guarantee on node order
	n := len(nodes)
	if n <= 0 {
		return ""
	}
	m.LastSelectedNode = (m.LastSelectedNode +1) % n
	return nodes[m.LastSelectedNode].Data.Name
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
			Component: s.SchedulerName,
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


