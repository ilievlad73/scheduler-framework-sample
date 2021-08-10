package sample

import (
	"context"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const (
	// Name is plugin name
	Name = "sample"
)

type Args struct {
	KubeConfig string `json:"kubeconfig,omitempty"`
	Master     string `json:"master,omitempty"`
}

var _ framework.PreFilterPlugin = &Sample{}
var _ framework.FilterPlugin = &Sample{}
var _ framework.PreBindPlugin = &Sample{}
var _ framework.ScorePlugin = &Sample{}
var _ framework.PermitPlugin = &Sample{}
var _ framework.ReservePlugin = &Sample{}
var _ framework.PostBindPlugin = &Sample{}

type Sample struct {
	args    *Args
	handle  framework.FrameworkHandle
	bindMap map[string]bool
}

func (pl *Sample) Name() string {
	return Name
}

/* UTILS */
func getPodScheduleTimeoutLabel(pod *v1.Pod) int {
	scheduleTimeoutSeconds := pod.Labels["scheduleTimeoutSeconds"]
	timeoutSeconds, err := strconv.Atoi(scheduleTimeoutSeconds)
	if err != nil {
		return 30 // default scheduler timeout, todo export this
	}

	return timeoutSeconds
}

func getPodAppName(pod *v1.Pod) string {
	return pod.Labels["app"]
}

func getPodTopology(pod *v1.Pod) string {
	return pod.Labels["topology"]
}

func removeEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func getPodDependencies(pod *v1.Pod) []string {
	labelsString := pod.Labels["depends-on"]
	return removeEmptyStrings(strings.Split(labelsString, "__"))
}

func isPodBind(podName string, bindMap map[string]bool) bool {
	isBind, ok := bindMap[podName]
	if ok != true {
		return false
	}

	return isBind
}

func markPodBind(podName string, bindMap map[string]bool) {
	bindMap[podName] = true
}

func markPodUnBind(podName string, bindMap map[string]bool) {
	delete(bindMap, podName)
}

func checkAllDependenciesBind(podNames []string, bindMap map[string]bool) bool {
	for _, s := range podNames {
		if isPodBind(s, bindMap) == false {

			return false
		}
	}

	return true
}

/* END UTILS */

// TODO: sort pods form queue based on priority, topology key, and creation time

func (pl *Sample) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	klog.V(3).Infof("PREFILTER POD : %v", pod.Name)

	scheduleTimeout := getPodScheduleTimeoutLabel(pod)
	podDependencies := getPodDependencies(pod)
	topology := getPodTopology(pod)
	appName := getPodAppName(pod)

	/* log important labels */
	klog.V(3).Infoln("Schedule timeout seconds %v", scheduleTimeout)
	klog.V(3).Infoln("Pod dependencies %v", podDependencies)
	klog.V(3).Infoln("Pod dependencies len %v", len(podDependencies))
	klog.V(3).Infoln("Pod topolofy: %v", topology)
	klog.V(3).Infoln("Pod app name: %v", appName)

	if len(podDependencies) == 0 {
		return framework.NewStatus(framework.Success, "")
	}

	if checkAllDependenciesBind(podDependencies, pl.bindMap) {
		return framework.NewStatus(framework.Success, "")
	}

	return framework.NewStatus(framework.Unschedulable, "")
}

func (pl *Sample) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (s *Sample) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node *nodeinfo.NodeInfo) *framework.Status {
	klog.V(3).Infof("FILTER POD : %v", pod.Name)
	return framework.NewStatus(framework.Success, "")
}

func (pl *Sample) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	klog.V(3).Infof("SCORING POD : %v", pod.Name)
	return 0, framework.NewStatus(framework.Success, "")
}

func (pl *Sample) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func (pl *Sample) Reserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	klog.V(3).Infof("RESERVE THE POD: %v", pod.Name)
	return nil
}

func (pl *Sample) Permit(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (*framework.Status, time.Duration) {
	klog.V(3).Infof("PERMIT ALLOWS THE POD: %v", pod.Name)
	return framework.NewStatus(framework.Success, ""), 0
}

func (pl *Sample) PreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	// nodeInfo, err := s.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	// if err != nil {
	// 	return framework.NewStatus(framework.Error, err.Error())
	// }

	klog.V(3).Infof("PREBIND POD : %v", pod.Name)
	return framework.NewStatus(framework.Success, "")
}

func (pl *Sample) PostBind(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeName string) {
	klog.V(3).Infof("POSTBIND POD : %v", pod.Name)
	markPodBind(getPodAppName(pod), pl.bindMap)
}

func New(plArgs *runtime.Unknown, handle framework.FrameworkHandle) (framework.Plugin, error) {
	args := &Args{}
	if err := framework.DecodeInto(plArgs, args); err != nil {
		return nil, err
	}
	klog.V(3).Infof("--------> args: %+v", args)
	return &Sample{
		args:    args,
		handle:  handle,
		bindMap: make(map[string]bool),
	}, nil
}
