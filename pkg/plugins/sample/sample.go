package sample

import (
	"context"

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

type Sample struct {
	args   *Args
	handle framework.FrameworkHandle
}

func (s *Sample) Name() string {
	return Name
}

func (pl *Sample) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	klog.V(3).Infof("Prefilter pod: %v", pod.Name)
	return framework.NewStatus(framework.Success, "")
}

func (pl *Sample) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (s *Sample) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node *nodeinfo.NodeInfo) *framework.Status {
	klog.V(3).Infof("Filter pod: %v", pod.Name)
	return framework.NewStatus(framework.Success, "")
}

func (s *Sample) PreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	// nodeInfo, err := s.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	// if err != nil {
	// 	return framework.NewStatus(framework.Error, err.Error())
	// }
	klog.V(3).Infof("Prebind node: %v", pod.Name)
	return framework.NewStatus(framework.Success, "")
}

func New(plArgs *runtime.Unknown, handle framework.FrameworkHandle) (framework.Plugin, error) {
	args := &Args{}
	if err := framework.DecodeInto(plArgs, args); err != nil {
		return nil, err
	}
	klog.V(3).Infof("--------> args: %+v", args)
	return &Sample{
		args:   args,
		handle: handle,
	}, nil
}
