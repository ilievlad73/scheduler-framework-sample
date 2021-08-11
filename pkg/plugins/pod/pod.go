package pod

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ilievlad73/scheduler-framework-sample/pkg/plugins/helpers"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
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

/* PODS UTILS */

func OtherPods(clientset *kubernetes.Clientset, pod *v1.Pod) ([]v1.Pod, error) {
	podsInfo, err := clientset.CoreV1().Pods(pod.GetNamespace()).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	/* remove current pod from the list */
	filteredPods := []v1.Pod{}
	for i := range podsInfo.Items {
		if podsInfo.Items[i].GetName() != pod.GetName() {
			filteredPods = append(filteredPods, podsInfo.Items[i])
		}
	}

	for _, otherPod := range filteredPods {
		klog.V(3).Infof("Range: Pod app: %v, Pod name: %v, phase %v", AppName(&otherPod), otherPod.Name, otherPod.Status.Phase)
	}

	return filteredPods, nil
}

func AreCompleteDependsOnCompleted(otherPods []v1.Pod, pod *v1.Pod) bool {
	podCompleteDependsOn := CompleteDependsOnList(pod)
	if len(podCompleteDependsOn) == 0 {
		return true
	}

	if len(otherPods) == 0 {
		return false
	}

	for _, otherPod := range otherPods {
		if helpers.StringInSlice(AppName(&otherPod), podCompleteDependsOn) && IsCompleted((&otherPod)) {
			podCompleteDependsOn = helpers.RemoveStringInSlice(AppName(&otherPod), podCompleteDependsOn)
		}
	}

	return len(podCompleteDependsOn) == 0
}

func AreCompleteDependsOnRunning(otherPods []v1.Pod, pod *v1.Pod) bool {
	podCompleteDependsOn := CompleteDependsOnList(pod)
	if len(podCompleteDependsOn) == 0 {
		return true
	}

	if len(otherPods) == 0 {
		return false
	}

	for _, otherPod := range otherPods {
		if helpers.StringInSlice(AppName(&otherPod), podCompleteDependsOn) && IsRunning((&otherPod)) {
			podCompleteDependsOn = helpers.RemoveStringInSlice(AppName(&otherPod), podCompleteDependsOn)
		}
	}

	return len(podCompleteDependsOn) == 0
}

/* PODS END UTILS */

/**
*
*
podAppName -> { appName,
				scheduleTimeoutSeconds,
				completeDependsOnStatus: podAppName -> {
															isRunning: false,
															isCompleted: false,
															isPending: false
														}

				}

*/

/* POD MANAGEMENT DATA STRUCTURE */

type SamplePodState struct {
	isPending    bool
	isRunning    bool
	isCompleted  bool
	isTerminated bool
}

type SamplePod struct {
	app                    string
	topology               string
	scheduleTimeoutSeconds int
	isPending              bool
	isRunning              bool
	isCompleted            bool
	isTerminated           bool
	completeDependsOn      map[string]*SamplePodState
}

func InitSamplePodsMap() map[string]*SamplePod {
	return make(map[string]*SamplePod)
}

func InitSamplePod(app string, topology string, scheduleTimeoutSeconds int, completeDependsOn []string, samplePods map[string]*SamplePod) {
	/* check if somebody else initialized this */
	_, ok := samplePods[app]
	if ok {
		return
	}

	/* init */
	samplePod := new(SamplePod)
	samplePod.app = app
	samplePod.topology = topology
	samplePod.scheduleTimeoutSeconds = scheduleTimeoutSeconds
	samplePod.completeDependsOn = make(map[string]*SamplePodState)
	samplePods[app] = samplePod
}

func InitPodState(dependencies []string, podStateMap map[string]*SamplePodState) {
	for _, dependency := range dependencies {
		podState := new(SamplePodState)
		podStateMap[dependency] = podState
	}
}

func MarkCompleteDependencyOnAsCompleted(app string, pod *SamplePod) {
	dependency, ok := pod.completeDependsOn[app]
	if ok == false {
		return
	}
	dependency.isCompleted = true
}

func MarkCompleteDependencyOnAsRunning(app string, pod *SamplePod) {
	dependency, ok := pod.completeDependsOn[app]
	if ok == false {
		return
	}
	dependency.isRunning = true
}

func MarkPodAsCompleted(app string, samplePods map[string]*SamplePod) {
	pod := samplePods[app]
	podTopology := pod.topology
	/* mark on yourself */
	pod.isRunning = true

	for _, otherPod := range samplePods {
		if otherPod.topology == podTopology {
			MarkCompleteDependencyOnAsCompleted(app, otherPod)
		}
	}

}

func MarkPodAsRunnning(app string, samplePods map[string]*SamplePod) {
	pod := samplePods[app]
	podTopology := pod.topology
	/* mark on yourself */
	pod.isRunning = true

	for _, otherPod := range samplePods {
		if otherPod.topology == podTopology {
			MarkCompleteDependencyOnAsRunning(app, otherPod)
		}
	}

}
