package interference

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"strings"
	scheduler_config "type-aware-scheduler/scheduler-config"
)

type PodInfo struct {
	TaskType int
	Load float64
}

type CoefficientsMatrix = [][]float64

type ModelType struct {
	NTaskTypes int
	NodeToCoefficients map[string]CoefficientsMatrix
	TypeToResourceRequests []float64
	NodeToResourceCapacity map[string]float64
}
// Type To resource Requests - currently only one resource type
// Node Capacity

var Model = ModelType{
	4, map[string]CoefficientsMatrix{ // baati / "kind-worker"
		"baati": CoefficientsMatrix{{0.09236406129745652,0.06968222664414211,0.1647130881065321,0.20015221017304366},
			{0.0466460482266046,0.05732161544142513,0.10657494482534081,0.10810948076749577},
			{0.05359758455154095,0.062162787980730924,0.10183621207551628,0.18190055622891266},
			{0.07236663478169568,0.053645734317726616,0.12738164345964662,0.21759115500509985},
		},
		"dosa": CoefficientsMatrix{{0.0437716396715501,0.05309571481979834,0.05778327994778269,0.0478036633107744},
			{0.13532284678468662,0.10337024868620512,0.1644856778556154,0.133996920283637},// dosa / "kind-control-plane"
			{0.030551469559096957,0.025204660199400788,0.2617647088479882,0.04293269861596552},
			{0.06232254815329062,0.060185071267793024,0.05227936747270895,0.10634860864271509},
		},
	},
			[]float64{0., 0., 3., 0},
			map[string]float64{"baati": 25., "dosa": 25.},
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
	t, found := scheduler_config.RoleToType[role]
	if !found {
		err = errors.New(fmt.Sprintf("Interference: ERROR - unknown pod role %s\n", role))
		return
	}
	typeId, found := scheduler_config.TypeStringToId[t]
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
