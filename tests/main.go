package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"testing"
	core_scheduler "type-aware-scheduler/core-scheduler"
	"type-aware-scheduler/interference"
	scheduler_config "type-aware-scheduler/scheduler-config"
)

func TestAbs(t *testing.T) {
	got := math.Abs(-1)
	if got != 1 {
		t.Errorf("Abs(-1) = %f; want 1", got)
	}
}

func TestChooseNode(t *testing.T) {
	//interference.Coefficients := {{}}
	got := math.Abs(-1)
	if got != 1 {
		t.Errorf("Abs(-1) = %f; want 1", got)
	}
}

func interferenceTest() {
	model := interference.ModelType{
		2, map[string]interference.CoefficientsMatrix{
			"baati": interference.CoefficientsMatrix{{1., 2}, {2, 1.}},
			"naan": interference.CoefficientsMatrix{{1., 2}, {2, 1.}}},
	}
	experiment := scheduler_config.OfflineSchedulingExperiment{"twoClashing", []int{0, 1, 0, 1}}
	solver := core_scheduler.NewOptimalScheduleSolver(model, experiment)
	schedule, cost := solver.Solve()
	fmt.Println(cost)
	fmt.Println(schedule)

	model = interference.ModelType{
		2, map[string]interference.CoefficientsMatrix{
			"baati": interference.CoefficientsMatrix{{1., 0.5}, {0.5, 1.}},
			"naan": interference.CoefficientsMatrix{{1., 0.5}, {0.5, 1.}}},
	}
	experiment = scheduler_config.OfflineSchedulingExperiment{"better2Mix", []int{0, 1, 0, 1}}
	solver = core_scheduler.NewOptimalScheduleSolver(model, experiment)
	schedule, cost = solver.Solve()
	fmt.Println(cost)
	fmt.Println(schedule)

	model = interference.ModelType{
		2, map[string]interference.CoefficientsMatrix{
			"baati": interference.CoefficientsMatrix{{1., 0.5}, {0.5, 1.}},
			"naan": interference.CoefficientsMatrix{{0.25, 0.5}, {0.5, 0.25}}},
	}
	experiment = scheduler_config.OfflineSchedulingExperiment{"nodeDifference", []int{0, 1, 0, 1}}
	solver = core_scheduler.NewOptimalScheduleSolver(model, experiment)
	schedule, cost = solver.Solve()
	fmt.Println(cost)
	fmt.Println(schedule)

	model = interference.ModelType{
		2, map[string]interference.CoefficientsMatrix{
			"baati": interference.CoefficientsMatrix{{1., 2.}, {0.5, 1.}}},
	}
	experiment = scheduler_config.OfflineSchedulingExperiment{"nonSymetricCoeff", []int{0, 1, 0}}
	solver = core_scheduler.NewOptimalScheduleSolver(model, experiment)
	schedule, cost = solver.Solve()
	fmt.Println(cost)
	fmt.Println(schedule)

}

func testParse(path string) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		line := scanner.Text()
		exp, err := scheduler_config.TryToParseOfflineExperimentFromLine(line)
		if err != nil {
			panic(err)
		}
		fmt.Println(exp)
	} else {
		fmt.Printf("Scanner empty")
	}
}

func testReader(path string) {
	expChan := make(chan scheduler_config.OfflineSchedulingExperiment)
	configReader := scheduler_config.NewConfigReader(path, expChan)
	go configReader.Run()
	for {
		exp := <-expChan
		fmt.Println(exp.Id)
		fmt.Println(exp.TaskTypesIds)
	}
}

func main() {
	//path := "exp"
	//testParse(path)
	//testReader(path)
	interferenceTest()

}
