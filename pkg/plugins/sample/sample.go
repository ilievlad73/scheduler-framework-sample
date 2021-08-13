package sample

import (
	"context"
	"time"

	"github.com/ilievlad73/scheduler-framework-sample/pkg/plugins/client"
	informerUtils "github.com/ilievlad73/scheduler-framework-sample/pkg/plugins/informer"
	podUtils "github.com/ilievlad73/scheduler-framework-sample/pkg/plugins/pod"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const (
	// Name is plugin name
	Name                    = "sample"
	schedulerTimeoutSeconds = 30
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
	args            *Args
	handle          framework.FrameworkHandle
	clientset       *kubernetes.Clientset
	clientsetConfig *rest.Config
	samplePods      map[string]*podUtils.SamplePod
}

func (pl *Sample) Name() string {
	return Name
}

/* END UTILS */

// TODO: sort pods form queue based on priority, topology key, and creation time

func (pl *Sample) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	klog.V(3).Infof("Prefilter pod : %v, app : %v, running deps %v, complete deps %v",
		pod.Name, podUtils.AppName(pod), podUtils.RunningDependsOnList(pod), podUtils.CompleteDependsOnList(pod))

	podUtils.InitSamplePod(podUtils.AppName(pod), podUtils.TopologyName(pod), podUtils.ScheduleTimeout(pod),
		podUtils.CompleteDependsOnList(pod), podUtils.RunningDependsOnList(pod), podUtils.SkipScheduleTimes(pod), pl.samplePods)
	klog.V(3).Infof("Sample pods from prefilter", pl.samplePods[podUtils.AppName(pod)])

	if podUtils.ShouldSkipScheduler(pod, pl.samplePods) {
		klog.V(3).Infof("Reject due to ShouldSkipScheduler")
		return framework.NewStatus(framework.Unschedulable, "")
	}

	if !podUtils.AreRunningDependsOnPendingOrRunning(pod, pl.samplePods) {
		klog.V(3).Infof("Reject due to AreRunningDependsOnRunningOrPending")
		return framework.NewStatus(framework.Unschedulable, "")
	}

	if !podUtils.AreRunninDependsOnPendingOrRunningLessThanTwoLayers(pod, pl.samplePods) {
		klog.V(3).Infof("Reject due to AreRunninDependsOnPendingOrRunningLessThanTwoLayers")
		return framework.NewStatus(framework.Unschedulable, "")
	}

	if !podUtils.AreCompleteDependsOnRunningOrComplete(pod, pl.samplePods) {
		klog.V(3).Infof("Reject due to AreCompleteDependsOnRunningOrComplete")
		return framework.NewStatus(framework.Unschedulable, "")
	}

	return framework.NewStatus(framework.Success, "")
}

func (pl *Sample) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (pl *Sample) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node *nodeinfo.NodeInfo) *framework.Status {
	klog.V(3).Infof("Filter pod : %v", pod.Name)

	return framework.NewStatus(framework.Success, "")
}

func (pl *Sample) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	klog.V(3).Infof("Scoring pod : %v", pod.Name)
	return 0, framework.NewStatus(framework.Success, "")
}

func (pl *Sample) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func (pl *Sample) Reserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	klog.V(3).Infof("Reserve the pod: %v", pod.Name)
	return nil
}

func (pl *Sample) Permit(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (*framework.Status, time.Duration) {
	klog.V(3).Infof("Permit the pod: %v", pod.Name)

	if !podUtils.AreRunningDependsOnRunningSince(pod, pl.samplePods, podUtils.POD_RUNNING_HEALTY_TIMEOUT) {
		time.AfterFunc(time.Duration(podUtils.POD_RUNNING_HEALTY_TIMEOUT)*time.Second, func() {
			podUtils.AllowWaitingPods(Name, pl.handle, pl.samplePods)
		})

		klog.Infof("Pod: %v is waiting to be scheduled to node due to running deps since: %v", pod.Name, nodeName)
		return framework.NewStatus(framework.Wait, ""), time.Duration(podUtils.RUNNING_DEPENDS_ON_WAIT_TIMEOUT) * time.Second
	}

	if !podUtils.AreCompleteDependsOnCompleted(pod, pl.samplePods) {
		klog.Infof("Pod: %v is waiting to be scheduled to node due to complete deps: %v", pod.Name, nodeName)
		return framework.NewStatus(framework.Wait, ""), time.Duration(podUtils.COMPLETE_DEPENDS_ON_WAIT_TIMEOUT) * time.Second
	}

	klog.V(3).Infof("Permit allows the pod: %v, app %v", pod.Name, podUtils.AppName(pod))
	return framework.NewStatus(framework.Success, ""), 0
}

func (pl *Sample) PreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	// nodeInfo, err := s.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	// if err != nil {
	// 	return framework.NewStatus(framework.Error, err.Error())
	// }

	klog.V(3).Infof("Prebind pod : %v", pod.Name)
	return framework.NewStatus(framework.Success, "")
}

func (pl *Sample) PostBind(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeName string) {
	klog.V(3).Infof("Postbind pod : %v", pod.Name)
}

// rejectPod rejects pod in cache
func (pl *Sample) rejectPod(uid types.UID) {
	waitingPod := pl.handle.GetWaitingPod(uid)
	if waitingPod == nil {
		return
	}
	waitingPod.Reject(Name)
}

func New(plArgs *runtime.Unknown, handle framework.FrameworkHandle) (framework.Plugin, error) {
	samplePods := podUtils.InitSamplePodsMap()

	args := &Args{}
	klog.V(3).Infof("--------> args: %+v", args)

	if err := framework.DecodeInto(plArgs, args); err != nil {
		return nil, err
	}

	clientset, clientsetConfig, err := client.Connect()
	if err != nil {
		return nil, err
	}

	factory := informers.NewSharedInformerFactory(clientset, time.Hour*24)
	controller := informerUtils.NewPodLoggingController(factory, handle, clientset, samplePods, Name)

	stop := make(chan struct{})
	err = controller.Run(stop)
	if err != nil {
		return nil, err
	}

	return &Sample{
		args:            args,
		handle:          handle,
		clientset:       clientset,
		clientsetConfig: clientsetConfig,
		samplePods:      samplePods,
	}, nil
}
