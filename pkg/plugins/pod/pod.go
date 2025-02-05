package pod

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ilievlad73/scheduler-framework-sample/pkg/plugins/helpers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
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

	STAGE_1_ALLOW_RUNNING_DEPENDS_ON_WAITING_PODS = 15
	STAGE_2_ALLOW_RUNNING_DEPENDS_ON_WAITING_PODS = 25
	STAGE_3_ALLOW_RUNNING_DEPENDS_ON_WAITING_PODS = 35
	STAGE_4_ALLOW_RUNNING_DEPENDS_ON_WAITING_PODS = 45
	STAGE_5_ALLOW_RUNNING_DEPENDS_ON_WAITING_PODS = 55

	POD_RUNNING_HEALTY_TIMEOUT_MS    = 20 * 1000
	RUNNING_DEPENDS_ON_WAIT_TIMEOUT  = 60
	COMPLETE_DEPENDS_ON_WAIT_TIMEOUT = 30
)

func ScheduleTimeout(pod *v1.Pod) int {
	scheduleTimeoutSeconds := pod.Labels["scheduleTimeoutSeconds"]
	timeoutSeconds, err := strconv.Atoi(scheduleTimeoutSeconds)
	if err != nil {
		return 30 // default scheduler timeout, todo export this
	}

	return timeoutSeconds
}

func SkipScheduleTimes(pod *v1.Pod) int {
	skipSchedulerTimes := pod.Labels["skipSchedulerTimes"]
	skipScheduler, err := strconv.Atoi(skipSchedulerTimes)
	if err != nil {
		return 0 // default skip schedule times
	}

	return skipScheduler
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

/* POD MANAGEMENT DATA STRUCTURE */

type SamplePodState struct {
	status         string
	statusUpdateAt int64
}

func (podState SamplePodState) String() string {
	return fmt.Sprintf("{status:%v, statusUpdateAt: %v}", podState.status, podState.statusUpdateAt)
}

type SamplePod struct {
	app               string
	topology          string
	status            string
	statusUpdateAt    int64
	skipScheduleTimes int
	completeDependsOn map[string]*SamplePodState
	runningDependsOn  map[string]*SamplePodState
}

func (pod SamplePod) String() string {
	return fmt.Sprintf("{app: %v, topology: %v, status:%v, statusUpdateAt %v, runningDependsOn: %v, completeDependsOn: %v}",
		pod.app, pod.topology, pod.status, pod.statusUpdateAt, pod.runningDependsOn, pod.completeDependsOn)
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
			podState.statusUpdateAt = dependencyGlobalValue.statusUpdateAt
		}
	}
}

func RemoveSamplePod(app string, samplePods map[string]*SamplePod) {
	delete(samplePods, app)
}

func InitSamplePod(app string, topology string,
	completeDependsOn []string, runningDependsOn []string, skipScheduleTimes int, samplePods map[string]*SamplePod) {
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
	samplePod.skipScheduleTimes = skipScheduleTimes
	samplePod.status = PENDING_STATUS
	samplePod.statusUpdateAt = helpers.GetCurrentTimestamp()
	samplePod.runningDependsOn = make(map[string]*SamplePodState)
	InitDependenciesPodState(runningDependsOn, samplePod.runningDependsOn, samplePods)
	samplePod.completeDependsOn = make(map[string]*SamplePodState)
	InitDependenciesPodState(completeDependsOn, samplePod.completeDependsOn, samplePods)
	samplePods[app] = samplePod
}

func MarkDependencyOnAsError(app string, pod *SamplePod) {
	runningDeppency, ok := pod.runningDependsOn[app]
	if ok != false {
		runningDeppency.status = ERROR_STATUS
		runningDeppency.statusUpdateAt = helpers.GetCurrentTimestamp()
	}

	completeDependency, ok := pod.completeDependsOn[app]
	if ok != false {
		completeDependency.status = ERROR_STATUS
		completeDependency.statusUpdateAt = helpers.GetCurrentTimestamp()
	}
}

func MarkPodAsError(pod *v1.Pod, samplePods map[string]*SamplePod) {
	appName := AppName(pod)
	samplePod, ok := samplePods[AppName(pod)]
	if ok == false {
		klog.Infof("Mark pod as error failed, pod %v not exists in data structure", pod.Name)
		return
	}
	podTopology := samplePod.topology

	if samplePod.status == ERROR_STATUS {
		klog.Infof("Mark pod as error unnecessary", pod.Name)
		return
	}
	/* mark on yourself */
	samplePod.status = ERROR_STATUS
	samplePod.statusUpdateAt = helpers.GetCurrentTimestamp()

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
		runningDeppency.statusUpdateAt = helpers.GetCurrentTimestamp()
	}

	completeDependency, ok := pod.completeDependsOn[app]
	if ok != false {
		completeDependency.status = UNDEFINED_STATUS
		completeDependency.statusUpdateAt = helpers.GetCurrentTimestamp()
	}
}

func MarkPodAsUndefined(pod *v1.Pod, samplePods map[string]*SamplePod) {
	appName := AppName(pod)
	samplePod, ok := samplePods[AppName(pod)]
	if ok == false {
		klog.Infof("Mark pod as undefined failed, pod %v not exists in data structure", pod.Name)
		return
	}
	podTopology := samplePod.topology

	if samplePod.status == UNDEFINED_STATUS {
		klog.Infof("Mark pod as undefined unnecessary", pod.Name)
		return
	}
	/* mark on yourself */
	samplePod.status = UNDEFINED_STATUS
	samplePod.statusUpdateAt = helpers.GetCurrentTimestamp()

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
		runningDeppency.statusUpdateAt = helpers.GetCurrentTimestamp()
	}

	completeDependency, ok := pod.completeDependsOn[app]
	if ok != false {
		completeDependency.status = PENDING_STATUS
		completeDependency.statusUpdateAt = helpers.GetCurrentTimestamp()
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

	if samplePod.status == PENDING_STATUS {
		klog.Infof("Mark pod as pending unnecessary", pod.Name)
		return
	}
	/* mark on yourself */
	samplePod.status = PENDING_STATUS
	samplePod.statusUpdateAt = helpers.GetCurrentTimestamp()

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
		runningDeppency.statusUpdateAt = helpers.GetCurrentTimestamp()
	}

	completeDependency, ok := pod.completeDependsOn[app]
	if ok != false {
		completeDependency.status = RUNNING_STATUS
		completeDependency.statusUpdateAt = helpers.GetCurrentTimestamp()
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

	if samplePod.status == RUNNING_STATUS {
		klog.Infof("Mark pod as running unnecessary", pod.Name)
		return
	}
	/* mark on yourself */
	samplePod.status = RUNNING_STATUS
	samplePod.statusUpdateAt = helpers.GetCurrentTimestamp()

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
		runningDeppency.statusUpdateAt = helpers.GetCurrentTimestamp()

	}

	completeDependency, ok := pod.completeDependsOn[app]
	if ok != false {
		completeDependency.status = COMPLETED_STATUS
		completeDependency.statusUpdateAt = helpers.GetCurrentTimestamp()
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

	if samplePod.status == COMPLETED_STATUS {
		klog.Infof("Mark pod as complete unnecessary", pod.Name)
		return
	}
	/* mark on yourself */
	samplePod.status = COMPLETED_STATUS
	samplePod.statusUpdateAt = helpers.GetCurrentTimestamp()

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

func AreRunningDependsOnRunningSince(pod *v1.Pod, samplePods map[string]*SamplePod, timeoutMilliseconds int64) bool {
	appName := AppName(pod)
	podSample := samplePods[appName]
	runningDependsOn := podSample.runningDependsOn
	for _, dependencyPod := range runningDependsOn {
		if dependencyPod.status != RUNNING_STATUS || (helpers.GetCurrentTimestamp()-dependencyPod.statusUpdateAt) < timeoutMilliseconds {
			return false
		}
	}

	return true
}

func AreRunningDependsOnPendingOrRunning(pod *v1.Pod, samplePods map[string]*SamplePod) bool {
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

func AreRunninDependsOnPendingOrRunningLessThanTwoLayers(pod *v1.Pod, samplePods map[string]*SamplePod) bool {
	appName := AppName(pod)
	podSample := samplePods[appName]
	runningDependsOn := podSample.runningDependsOn
	for dependencyPodName, dependencyPod := range runningDependsOn {
		/* first depth layer */
		if dependencyPod.status != RUNNING_STATUS && dependencyPod.status != PENDING_STATUS {
			return false
		}

		/* second depth layer */
		dependencySamplePod, ok := samplePods[dependencyPodName]
		if !ok {
			return false
		}
		for secondDependencyPodName, secondDependencyPod := range dependencySamplePod.runningDependsOn {
			if secondDependencyPod.status != RUNNING_STATUS && dependencyPod.status != PENDING_STATUS {
				return false
			}

			/* third depth layer */
			secondDependencySamplePod, ok := samplePods[secondDependencyPodName]
			if !ok {
				return false
			}
			for _, thirdDependencyPod := range secondDependencySamplePod.runningDependsOn {
				if thirdDependencyPod.status != RUNNING_STATUS {
					return false
				}
			}
		}
	}

	return true
}

func ShouldSkipScheduler(pod *v1.Pod, samplePods map[string]*SamplePod) bool {
	appName := AppName(pod)
	podSample := samplePods[appName]

	if podSample.skipScheduleTimes == 0 {
		return false
	}

	klog.Infof("Skipping scheduler cycle due to skipSchedulerTimes %v", podSample.skipScheduleTimes)

	podSample.skipScheduleTimes--
	return true
}

/* END POD MANAGEMENT DATA STRUCTURE */

/* PERMIT UTILS */

/* TODO, reject waiting post on running if error */
func AllowWaitingPods(sampleName string, handle framework.FrameworkHandle, samplePods map[string]*SamplePod) {
	klog.Infof("Allowing waiting pods")
	handle.IterateOverWaitingPods(func(waitingPod framework.WaitingPod) {
		pod := waitingPod.GetPod()
		if AreRunningDependsOnRunningSince(pod, samplePods, POD_RUNNING_HEALTY_TIMEOUT_MS) {
			klog.Infof("[Informer] Pod %v passed running deps since check", pod.Name)
			if AreCompleteDependsOnCompleted(pod, samplePods) {
				klog.Infof("[Informer] Pod %v to passed complete deps check", pod.Name)
				waitingPod.Allow(sampleName)
			}
		}
	})
}

func AllowWaitingPodAfterTime(sampleName string, handle framework.FrameworkHandle, samplePods map[string]*SamplePod) {
	time.AfterFunc(time.Duration(STAGE_1_ALLOW_RUNNING_DEPENDS_ON_WAITING_PODS)*time.Second, func() {
		AllowWaitingPods(sampleName, handle, samplePods)
	})

	time.AfterFunc(time.Duration(STAGE_2_ALLOW_RUNNING_DEPENDS_ON_WAITING_PODS)*time.Second, func() {
		AllowWaitingPods(sampleName, handle, samplePods)
	})

	time.AfterFunc(time.Duration(STAGE_3_ALLOW_RUNNING_DEPENDS_ON_WAITING_PODS)*time.Second, func() {
		AllowWaitingPods(sampleName, handle, samplePods)
	})

	time.AfterFunc(time.Duration(STAGE_4_ALLOW_RUNNING_DEPENDS_ON_WAITING_PODS)*time.Second, func() {
		AllowWaitingPods(sampleName, handle, samplePods)
	})

	time.AfterFunc(time.Duration(STAGE_5_ALLOW_RUNNING_DEPENDS_ON_WAITING_PODS)*time.Second, func() {
		AllowWaitingPods(sampleName, handle, samplePods)
	})
}
