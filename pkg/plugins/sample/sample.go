package sample

import (
	"context"
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
	args   *Args
	handle framework.FrameworkHandle
}

func (pl *Sample) Name() string {
	return Name
}

func (pl *Sample) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	klog.Infof("PREFILTER POD : %v", pod.Name)
	return framework.NewStatus(framework.Success, "")
}

func (pl *Sample) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (s *Sample) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node *nodeinfo.NodeInfo) *framework.Status {
	klog.Infof("FILTER POD : %v", pod.Name)
	return framework.NewStatus(framework.Success, "")
}

func (pl *Sample) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	klog.Infof("SCORING POD : %v", pod.Name)
	return 0, framework.NewStatus(framework.Success, "")
}

func (pl *Sample) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func (pl *Sample) Reserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	klog.Infof("RESERVE THE POD: %v", pod.Name)
	return nil
}

func (pl *Sample) Permit(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (*framework.Status, time.Duration) {
	klog.Infof("PERMIT ALLOWS THE POD: %v", pod.Name)
	return framework.NewStatus(framework.Success, ""), 0
}

func (pl *Sample) PreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	// nodeInfo, err := s.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	// if err != nil {
	// 	return framework.NewStatus(framework.Error, err.Error())
	// }
	klog.Infof("PREBIND NODE : %v", pod.Name)
	return framework.NewStatus(framework.Success, "")
}

func (pl *Sample) PostBind(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeName string) {
	klog.Infof("POSTBIND NODE : %v", pod.Name)
}

func New(plArgs *runtime.Unknown, handle framework.FrameworkHandle) (framework.Plugin, error) {
	args := &Args{}
	if err := framework.DecodeInto(plArgs, args); err != nil {
		return nil, err
	}
	klog.Infof("--------> args: %+v", args)
	return &Sample{
		args:   args,
		handle: handle,
	}, nil
}
