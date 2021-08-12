package pod

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ilievlad73/scheduler-framework-sample/pkg/plugins/helpers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

/* LABELS UTILS */

const (
	PENDING_STATUS            = "Pending"
	RUNNING_STATUS            = "Running"
	COMPLETED_STATUS          = "Succeeded"
	TERMINATED_STATUS         = "Terminating"
	ERROR_STATUS              = "Error"
	CONTAINER_CREATING_STATUS = "ContainerCreating"
	UNDEFINED_STATUS          = ""
)

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

func RunningDependsOnList(pod *v1.Pod) []string {
	labelsString := pod.Labels["running-depends-on"]
	return helpers.RemoveEmptyStrings(strings.Split(labelsString, "__"))
}

/* STATUS UTILS */

func StatusPhase(pod *v1.Pod) string {
	return string(pod.Status.Phase)
}

func IsPending(pod *v1.Pod) bool {
	return StatusPhase(pod) == PENDING_STATUS
}

func IsCompleted(pod *v1.Pod) bool {
	return StatusPhase(pod) == COMPLETED_STATUS
}

func IsRunning(pod *v1.Pod) bool {
	return StatusPhase(pod) == RUNNING_STATUS
}

func IsError(pod *v1.Pod) bool {
	return StatusPhase(pod) == ERROR_STATUS
}

func IsTerminating(pod *v1.Pod) bool {
	return StatusPhase(pod) == TERMINATED_STATUS
}

/* END STATUS UTILS */

/* BIND STRATEGY UTILS */

// func IsBind(podName string, bindMap map[string]bool) bool {
// 	isBind, ok := bindMap[podName]
// 	if ok != true {
// 		return false
// 	}

// 	return isBind
// }

// func MarkAsBind(podName string, bindMap map[string]bool) {
// 	bindMap[podName] = true
// }

// func MarkPodAsUnbind(podName string, bindMap map[string]bool) {
// 	delete(bindMap, podName)
// }

// func CheckAllDependenciesBind(podNames []string, bindMap map[string]bool) bool {
// 	for _, s := range podNames {
// 		if IsBind(s, bindMap) == false {

// 			return false
// 		}
// 	}

// 	return true
// }

/* END BIND STRATEGY UTILS */

// func displayPodMetadataName(pod *v1.Pod) {
// 	for key, value := range pod.Labels {
// 		fmt.Printf("Labels: key[%s] value[%s]\n", key, value)
// 	}

// 	for key, value := range pod.ObjectMeta.Labels {
// 		fmt.Printf("Meta labels: key[%s] value[%s]\n", key, value)
// 	}
// }

// /* PODS UTILS */

// func OtherPods(clientset *kubernetes.Clientset, pod *v1.Pod) ([]v1.Pod, error) {
// 	podsInfo, err := clientset.CoreV1().Pods(pod.GetNamespace()).List(metav1.ListOptions{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	/* remove current pod from the list */
// 	filteredPods := []v1.Pod{}
// 	for i := range podsInfo.Items {
// 		if podsInfo.Items[i].GetName() != pod.GetName() {
// 			filteredPods = append(filteredPods, podsInfo.Items[i])
// 		}
// 	}

// 	for _, otherPod := range filteredPods {
// 		klog.V(3).Infof("Range: Pod app: %v, Pod name: %v, phase %v", AppName(&otherPod), otherPod.Name, otherPod.Status.Phase)
// 	}

// 	return filteredPods, nil
// }

// func AreCompleteDependsOnCompleted(otherPods []v1.Pod, pod *v1.Pod) bool {
// 	podCompleteDependsOn := CompleteDependsOnList(pod)
// 	if len(podCompleteDependsOn) == 0 {
// 		return true
// 	}

// 	if len(otherPods) == 0 {
// 		return false
// 	}

// 	for _, otherPod := range otherPods {
// 		if helpers.StringInSlice(AppName(&otherPod), podCompleteDependsOn) && IsCompleted((&otherPod)) {
// 			podCompleteDependsOn = helpers.RemoveStringInSlice(AppName(&otherPod), podCompleteDependsOn)
// 		}
// 	}

// 	return len(podCompleteDependsOn) == 0
// }

// func AreCompleteDependsOnRunning(otherPods []v1.Pod, pod *v1.Pod) bool {
// 	podCompleteDependsOn := CompleteDependsOnList(pod)
// 	if len(podCompleteDependsOn) == 0 {
// 		return true
// 	}

// 	if len(otherPods) == 0 {
// 		return false
// 	}

// 	for _, otherPod := range otherPods {
// 		if helpers.StringInSlice(AppName(&otherPod), podCompleteDependsOn) && IsRunning((&otherPod)) {
// 			podCompleteDependsOn = helpers.RemoveStringInSlice(AppName(&otherPod), podCompleteDependsOn)
// 		}
// 	}

// 	return len(podCompleteDependsOn) == 0
// }

/* PODS END UTILS */

/* POD MANAGEMENT DATA STRUCTURE */

type SamplePodState struct {
	status string
}

func (podState SamplePodState) String() string {
	return fmt.Sprintf("{status:%v}", podState.status)
}

type SamplePod struct {
	app                    string
	topology               string
	scheduleTimeoutSeconds int
	status                 string
	completeDependsOn      map[string]*SamplePodState
	runningDependsOn       map[string]*SamplePodState
}

func (pod SamplePod) String() string {
	return fmt.Sprintf("{app: %v, topology: %v,scheduleTimeoutSeconds:%v, status:%v, runningDependsOn: %v, completeDependsOn: %v}",
		pod.app, pod.topology, pod.scheduleTimeoutSeconds, pod.status, pod.runningDependsOn, pod.completeDependsOn)
}

func InitSamplePodsMap() map[string]*SamplePod {
	return make(map[string]*SamplePod)
}

func InitDependenciesPodState(dependencies []string, podStateMap map[string]*SamplePodState, samplePods map[string]*SamplePod) {
	for _, dependency := range dependencies {
		podState := new(SamplePodState)
		dependencyGlobalValue, ok := samplePods[dependency]
		podStateMap[dependency] = podState
		if ok {
			podState.status = dependencyGlobalValue.status
		}
	}
}

func InitSamplePod(app string, topology string, scheduleTimeoutSeconds int, completeDependsOn []string, runningDependsOn []string, samplePods map[string]*SamplePod) {
	/* check if somebody else initialized this */
	existingSamplePod, ok := samplePods[app]
	if ok {
		existingSamplePod.runningDependsOn = make(map[string]*SamplePodState)
		InitDependenciesPodState(runningDependsOn, existingSamplePod.runningDependsOn, samplePods)
		existingSamplePod.completeDependsOn = make(map[string]*SamplePodState)
		InitDependenciesPodState(completeDependsOn, existingSamplePod.completeDependsOn, samplePods)
		return
	}

	/* init */
	samplePod := new(SamplePod)
	samplePod.app = app
	samplePod.topology = topology
	samplePod.scheduleTimeoutSeconds = scheduleTimeoutSeconds
	samplePod.status = PENDING_STATUS
	samplePod.runningDependsOn = make(map[string]*SamplePodState)
	InitDependenciesPodState(runningDependsOn, samplePod.runningDependsOn, samplePods)
	samplePod.completeDependsOn = make(map[string]*SamplePodState)
	InitDependenciesPodState(completeDependsOn, samplePod.completeDependsOn, samplePods)
	samplePods[app] = samplePod
}

func MarkDependencyOnAsError(app string, pod *SamplePod) {
	runningDeppency, ok := pod.runningDependsOn[app]
	if ok != false {
		runningDeppency.status = UNDEFINED_STATUS
	}

	completeDependency, ok := pod.completeDependsOn[app]
	if ok != false {
		completeDependency.status = UNDEFINED_STATUS
	}
}

func MarkPodAsError(pod *v1.Pod, samplePods map[string]*SamplePod) {
	appName := AppName(pod)
	samplePod, ok := samplePods[AppName(pod)]
	if ok == false {
		klog.Infof("Mark pod as pending failed, pod %v not exists in data structure", pod.Name)
		return
	}

	podTopology := samplePod.topology
	/* mark on yourself */
	samplePod.status = UNDEFINED_STATUS

	for _, otherPod := range samplePods {
		if otherPod.topology == podTopology {
			MarkDependencyOnAsError(appName, otherPod)
		}
	}
}

func MarkDependencyOnAsUndefined(app string, pod *SamplePod) {
	runningDeppency, ok := pod.runningDependsOn[app]
	if ok != false {
		runningDeppency.status = UNDEFINED_STATUS
	}

	completeDependency, ok := pod.completeDependsOn[app]
	if ok != false {
		completeDependency.status = UNDEFINED_STATUS
	}
}

func MarkPodAsUndefined(pod *v1.Pod, samplePods map[string]*SamplePod) {
	appName := AppName(pod)
	samplePod, ok := samplePods[AppName(pod)]
	if ok == false {
		klog.Infof("Mark pod as pending failed, pod %v not exists in data structure", pod.Name)
		return
	}

	podTopology := samplePod.topology
	/* mark on yourself */
	samplePod.status = UNDEFINED_STATUS

	for _, otherPod := range samplePods {
		if otherPod.topology == podTopology {
			MarkDependencyOnAsUndefined(appName, otherPod)
		}
	}
}

func MarkDependencyOnAsPending(app string, pod *SamplePod) {
	runningDeppency, ok := pod.runningDependsOn[app]
	if ok != false {
		runningDeppency.status = PENDING_STATUS
	}

	completeDependency, ok := pod.completeDependsOn[app]
	if ok != false {
		completeDependency.status = PENDING_STATUS
	}
}

func MarkPodAsPending(pod *v1.Pod, samplePods map[string]*SamplePod) {
	appName := AppName(pod)
	samplePod, ok := samplePods[AppName(pod)]
	if ok == false {
		klog.Infof("Mark pod as pending failed, pod %v not exists in data structure", pod.Name)
		return
	}

	podTopology := samplePod.topology
	/* mark on yourself */
	samplePod.status = PENDING_STATUS

	for _, otherPod := range samplePods {
		if otherPod.topology == podTopology {
			MarkDependencyOnAsPending(appName, otherPod)
		}
	}
}

func MarkDependencyOnAsRunning(app string, pod *SamplePod) {
	runningDeppency, ok := pod.runningDependsOn[app]
	if ok != false {
		runningDeppency.status = RUNNING_STATUS
	}

	completeDependency, ok := pod.completeDependsOn[app]
	if ok != false {
		completeDependency.status = RUNNING_STATUS
	}
}

func MarkPodAsRunnning(pod *v1.Pod, samplePods map[string]*SamplePod) {
	appName := AppName(pod)
	samplePod, ok := samplePods[AppName(pod)]
	if ok == false {
		klog.Infof("Mark pod as running failed, pod %v not exists in data structure", pod.Name)
		return
	}

	podTopology := samplePod.topology
	/* mark on yourself */
	samplePod.status = RUNNING_STATUS

	for _, otherPod := range samplePods {
		if otherPod.topology == podTopology {
			MarkDependencyOnAsRunning(appName, otherPod)
		}
	}
}

func MarkDependencyOnAsCompleted(app string, pod *SamplePod) {
	runningDeppency, ok := pod.runningDependsOn[app]
	if ok != false {
		runningDeppency.status = COMPLETED_STATUS
	}

	completeDependency, ok := pod.completeDependsOn[app]
	if ok != false {
		completeDependency.status = COMPLETED_STATUS
	}
}

func MarkPodAsCompleted(pod *v1.Pod, samplePods map[string]*SamplePod) {
	appName := AppName(pod)
	samplePod, ok := samplePods[AppName(pod)]
	if ok == false {
		klog.Infof("Mark pod as complete failed, pod %v not exists in data structure", pod.Name)
		return
	}

	podTopology := samplePod.topology
	/* mark on yourself */
	samplePod.status = COMPLETED_STATUS

	for _, otherPod := range samplePods {
		if otherPod.topology == podTopology {
			MarkDependencyOnAsCompleted(appName, otherPod)
		}
	}

}

func AreCompleteDependsOnRunning(pod *v1.Pod, samplePods map[string]*SamplePod) bool {
	appName := AppName(pod)
	podSample := samplePods[appName]
	completeDependsOn := podSample.completeDependsOn
	for _, dependencyPod := range completeDependsOn {
		if dependencyPod.status != RUNNING_STATUS {
			return false
		}
	}

	return true
}

func AreCompleteDependsOnCompleted(pod *v1.Pod, samplePods map[string]*SamplePod) bool {
	appName := AppName(pod)
	podSample := samplePods[appName]
	completeDependsOn := podSample.completeDependsOn
	for _, dependencyPod := range completeDependsOn {
		if dependencyPod.status != COMPLETED_STATUS {
			return false
		}
	}

	return true
}

func AreCompleteDependsOnRunningOrComplete(pod *v1.Pod, samplePods map[string]*SamplePod) bool {
	appName := AppName(pod)
	podSample := samplePods[appName]
	completeDependsOn := podSample.completeDependsOn
	for _, dependencyPod := range completeDependsOn {
		if dependencyPod.status != RUNNING_STATUS && dependencyPod.status != COMPLETED_STATUS {
			return false
		}
	}

	return true
}

func AreRunningDependsOnRunning(pod *v1.Pod, samplePods map[string]*SamplePod) bool {
	appName := AppName(pod)
	podSample := samplePods[appName]
	runningDependsOn := podSample.runningDependsOn
	for _, dependencyPod := range runningDependsOn {
		if dependencyPod.status != RUNNING_STATUS {
			return false
		}
	}

	return true
}

func AreRunningDependsOnRunningOrPending(pod *v1.Pod, samplePods map[string]*SamplePod) bool {
	appName := AppName(pod)
	podSample := samplePods[appName]
	runningDependsOn := podSample.runningDependsOn
	for _, dependencyPod := range runningDependsOn {
		if dependencyPod.status != RUNNING_STATUS && dependencyPod.status != PENDING_STATUS {
			return false
		}
	}

	return true
}

/* END POD MANAGEMENT DATA STRUCTURE */
