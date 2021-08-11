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

func (podState SamplePodState) String() string {
	return fmt.Sprintf("{isPending:%v, isRunning:%v, isCompleted:%v, isTerminated: %v}", podState.isPending, podState.isRunning, podState.isCompleted, podState.isTerminated)
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

func (pod SamplePod) String() string {
	return fmt.Sprintf("{app: %v, topology: %v,scheduleTimeoutSeconds:%v, isPending:%v, isRunning:%v, isCompleted:%v, isTerminated: %v, completeDependsOn: %v}",
		pod.app, pod.topology, pod.scheduleTimeoutSeconds, pod.isPending, pod.isRunning, pod.isCompleted, pod.isTerminated, pod.completeDependsOn)
}

func InitSamplePodsMap() map[string]*SamplePod {
	return make(map[string]*SamplePod)
}

func InitPodState(dependencies []string, podStateMap map[string]*SamplePodState, samplePods map[string]*SamplePod) {
	for _, dependency := range dependencies {
		podState := new(SamplePodState)
		dependencyGlobalValue := samplePods[dependency]
		podState.isPending = dependencyGlobalValue.isPending
		podState.isRunning = dependencyGlobalValue.isRunning
		podState.isCompleted = dependencyGlobalValue.isCompleted
		podState.isTerminated = dependencyGlobalValue.isTerminated
		podStateMap[dependency] = podState
	}
}

func InitSamplePod(app string, topology string, scheduleTimeoutSeconds int, completeDependsOn []string, samplePods map[string]*SamplePod) {
	/* check if somebody else initialized this */
	existingSamplePod, ok := samplePods[app]
	if ok {
		existingSamplePod.completeDependsOn = make(map[string]*SamplePodState)
		InitPodState(completeDependsOn, existingSamplePod.completeDependsOn, samplePods)
		return
	}

	/* init */
	samplePod := new(SamplePod)
	samplePod.app = app
	samplePod.topology = topology
	samplePod.scheduleTimeoutSeconds = scheduleTimeoutSeconds
	samplePod.completeDependsOn = make(map[string]*SamplePodState)
	InitPodState(completeDependsOn, samplePod.completeDependsOn, samplePods)
	samplePods[app] = samplePod
}

func MarkCompleteDependencyOnAsRunning(app string, pod *SamplePod) {
	dependency, ok := pod.completeDependsOn[app]
	if ok == false {
		return
	}

	dependency.isRunning = true
	dependency.isCompleted = false
	dependency.isPending = false
	dependency.isTerminated = false
}

func MarkPodAsRunnning(pod *v1.Pod, samplePods map[string]*SamplePod) {
	appName := AppName(pod)
	samplePod := samplePods[AppName(pod)]
	podTopology := samplePod.topology

	/* mark on yourself */
	samplePod.isRunning = true
	samplePod.isCompleted = false
	samplePod.isPending = false
	samplePod.isTerminated = false

	for _, otherPod := range samplePods {
		if otherPod.topology == podTopology {
			MarkCompleteDependencyOnAsRunning(appName, otherPod)
		}
	}

}

func MarkCompleteDependencyOnAsCompleted(app string, pod *SamplePod) {
	dependency, ok := pod.completeDependsOn[app]
	if ok == false {
		return
	}
	dependency.isRunning = false
	dependency.isCompleted = true
	dependency.isPending = false
	dependency.isTerminated = false
}

func MarkPodAsCompleted(pod *v1.Pod, samplePods map[string]*SamplePod) {
	appName := AppName(pod)
	samplePod := samplePods[AppName(pod)]
	podTopology := samplePod.topology
	/* mark on yourself */
	samplePod.isRunning = false
	samplePod.isCompleted = true
	samplePod.isPending = false
	samplePod.isTerminated = false

	for _, otherPod := range samplePods {
		if otherPod.topology == podTopology {
			MarkCompleteDependencyOnAsCompleted(appName, otherPod)
		}
	}

}

func AreCompleteDependsOnCompletedV2(pod *v1.Pod, samplePods map[string]*SamplePod) bool {
	appName := AppName(pod)
	podSample := samplePods[appName]
	completeDependsOn := podSample.completeDependsOn
	klog.Infof("AreCompleteDependsOnCompletedV2 %v", completeDependsOn)
	for _, dependencyPod := range completeDependsOn {
		if dependencyPod.isCompleted == false {
			return false
		}
	}

	return true
}

func AreCompleteDependsOnRunningV2(pod *v1.Pod, samplePods map[string]*SamplePod) bool {
	appName := AppName(pod)
	podSample := samplePods[appName]
	completeDependsOn := podSample.completeDependsOn
	for _, dependencyPod := range completeDependsOn {
		if dependencyPod.isRunning == false {
			return false
		}
	}

	return true
}
