package pod

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ilievlad73/scheduler-framework-sample/pkg/plugins/helpers"
	v1 "k8s.io/api/core/v1"
)

/* LABELS UTILS */

func ScheduleTimeout(pod *v1.Pod) int {
	scheduleTimeoutSeconds := pod.Labels["scheduleTimeoutSeconds"]
	timeoutSeconds, err := strconv.Atoi(scheduleTimeoutSeconds)
	if err != nil {
		return 30 // default scheduler timeout, todo export this
	}

	return timeoutSeconds
}

func AppName(pod *v1.Pod) string {
	return pod.Labels["app"]
}

func TopologyName(pod *v1.Pod) string {
	return pod.Labels["topology"]
}

func CompleteDependsOnList(pod *v1.Pod) []string {
	labelsString := pod.Labels["complete-depends-on"]
	return helpers.RemoveEmptyStrings(strings.Split(labelsString, "__"))
}

func CompleteDependencyOffList(pod *v1.Pod) []string {
	labelsString := pod.Labels["complete-dependency-off"]
	return helpers.RemoveEmptyStrings(strings.Split(labelsString, "__"))
}

func RunningDependsOnList(pod *v1.Pod) []string {
	labelsString := pod.Labels["running-depends-on"]
	return helpers.RemoveEmptyStrings(strings.Split(labelsString, "__"))
}

func RunningDependencyOffList(pod *v1.Pod) []string {
	labelsString := pod.Labels["running-dependency-off"]
	return helpers.RemoveEmptyStrings(strings.Split(labelsString, "__"))
}

/* STATUS UTILS */

func StatusPhase(pod *v1.Pod) string {
	return string(pod.Status.Phase)
}

func IsPending(pod *v1.Pod) bool {
	return StatusPhase(pod) == "Pending"
}

func IsCompleted(pod *v1.Pod) bool {
	return StatusPhase(pod) == "Succeeded"
}

func IsRunning(pod *v1.Pod) bool {
	return StatusPhase(pod) == "Running"
}

func IsTerminating(pod *v1.Pod) bool {
	return StatusPhase(pod) == "Terminating"
}

/* END STATUS UTILS */

/* BIND STRATEGY UTILS */

func IsBind(podName string, bindMap map[string]bool) bool {
	isBind, ok := bindMap[podName]
	if ok != true {
		return false
	}

	return isBind
}

func MarkAsBind(podName string, bindMap map[string]bool) {
	bindMap[podName] = true
}

func MarkPodAsUnbind(podName string, bindMap map[string]bool) {
	delete(bindMap, podName)
}

func CheckAllDependenciesBind(podNames []string, bindMap map[string]bool) bool {
	for _, s := range podNames {
		if IsBind(s, bindMap) == false {

			return false
		}
	}

	return true
}

/* END BIND STRATEGY UTILS */

func displayPodMetadataName(pod *v1.Pod) {
	for key, value := range pod.Labels {
		fmt.Printf("Labels: key[%s] value[%s]\n", key, value)
	}

	for key, value := range pod.ObjectMeta.Labels {
		fmt.Printf("Meta labels: key[%s] value[%s]\n", key, value)
	}
}
