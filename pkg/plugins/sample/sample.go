package sample

import (
	"context"
	"time"

	"github.com/ilievlad73/scheduler-framework-sample/pkg/plugins/client"
	informerUtils "github.com/ilievlad73/scheduler-framework-sample/pkg/plugins/informer"
	podUtils "github.com/ilievlad73/scheduler-framework-sample/pkg/plugins/pod"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	args            *Args
	handle          framework.FrameworkHandle
	bindMap         map[string]bool
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
	klog.V(3).Infof("Prefilter pod : %v, app : %v, dependencies %v", pod.Name, podUtils.AppName(pod), podUtils.CompleteDependsOnList(pod))

	podUtils.InitSamplePod(podUtils.AppName(pod), podUtils.TopologyName(pod), podUtils.ScheduleTimeout(pod), podUtils.CompleteDependsOnList(pod), pl.samplePods)
	klog.V(3).Infof("Sample pods from prefilter", pl.samplePods)

	if podUtils.AreCompleteDependsOnRunningV2(pod, pl.samplePods) {
		return framework.NewStatus(framework.Success, "")
	}

	return framework.NewStatus(framework.Unschedulable, "")
}

func (pl *Sample) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (s *Sample) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node *nodeinfo.NodeInfo) *framework.Status {
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
	klog.V(3).Infof("Permit allows the pod: %v to be scheduled on the node", pod.Name, nodeName)

	if !podUtils.AreCompleteDependsOnCompletedV2(pod, pl.samplePods) {
		klog.Infof("Pod: %v is waiting to be scheduled to node: %v", pod.Name, nodeName)
		return framework.NewStatus(framework.Success, ""), time.Duration(podUtils.ScheduleTimeout(pod)) * time.Second
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
	podUtils.MarkAsBind(podUtils.AppName(pod), pl.bindMap)
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
	controller := informerUtils.NewPodLoggingController(factory, handle, clientset, samplePods)

	stop := make(chan struct{})
	err = controller.Run(stop)
	if err != nil {
		return nil, err
	}

	return &Sample{
		args:            args,
		handle:          handle,
		bindMap:         make(map[string]bool),
		clientset:       clientset,
		clientsetConfig: clientsetConfig,
		samplePods:      samplePods,
	}, nil
}
