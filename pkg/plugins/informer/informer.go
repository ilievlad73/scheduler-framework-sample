// https://github.com/feiskyer/kubernetes-handbook/blob/master/examples/client/informer/informer.go

package informer

import (
	// "flag"

	"fmt"
	"time"

	// "time"

	podUtils "github.com/ilievlad73/scheduler-framework-sample/pkg/plugins/pod"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"

	// "k8s.io/client-go/kubernetes"
	// "k8s.io/component-base/logs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	klog "k8s.io/klog/v2"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	// "k8s.io/client-go/tools/clientcmd"
)

// PodLoggingController logs the name and namespace of pods that are added,
// deleted, or updated
type PodLoggingController struct {
	informerFactory  informers.SharedInformerFactory
	podInformer      coreinformers.PodInformer
	frameworkHandler framework.FrameworkHandle
	clientset        *kubernetes.Clientset
	samplePods       map[string]*podUtils.SamplePod
	sampleName       string
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *PodLoggingController) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so
	// far.
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.podInformer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return nil
}

func (c *PodLoggingController) podAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	klog.Infof("[Informer] pod created: %s/%s", pod.Namespace, pod.Name)
	klog.Infof("[Informer] Add sample pods", c.samplePods)
}

func (c *PodLoggingController) podUpdate(old, new interface{}) {
	oldPod := old.(*v1.Pod)
	newPod := new.(*v1.Pod)
	klog.Infof(
		"[Informer] pod %s/%s updated to pod %s/%s : phase %s", oldPod.Namespace, oldPod.Name, newPod.Namespace, newPod.Name, newPod.Status.Phase,
	)

	isError := false
	/* check for error states, running phase and container not ready => error from container */
	if podUtils.IsRunning(newPod) {
		for _, condition := range newPod.Status.Conditions {
			klog.Infof("Condition from error check %v", condition)
			if condition.Type == "Ready" && condition.Status == "False" {
				isError = true
				break
			}
		}
	}

	if isError {
		klog.Infof("[Informer] mark pod as error")
		podUtils.MarkPodAsError(newPod, c.samplePods)
	} else if podUtils.IsPending(newPod) {
		klog.Infof("[Informer] mark pod as pending")
		podUtils.MarkPodAsPending(newPod, c.samplePods)
	} else if podUtils.IsRunning(newPod) {
		klog.Infof("[Informer] mark pod as running")
		podUtils.MarkPodAsRunnning(newPod, c.samplePods)
		time.AfterFunc(time.Duration(podUtils.POD_RUNNING_HEALTY_TIMEOUT)*time.Second, func() {
			podUtils.AllowWaitingPods(c.sampleName, c.frameworkHandler, c.samplePods)
		})
	} else if podUtils.IsCompleted(newPod) {
		klog.Infof("[Informer] mark pod as completed")
		podUtils.MarkPodAsCompleted(newPod, c.samplePods)
		podUtils.AllowWaitingPods(c.sampleName, c.frameworkHandler, c.samplePods)
	} else {
		klog.Infof("[Informer] mark pod as undefined")
		podUtils.MarkPodAsUndefined(newPod, c.samplePods)
	}

	klog.Infof("[Informer] Update sample pods", c.samplePods)
}

func (c *PodLoggingController) podDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	klog.Infof("[Infomer] pod deleted: %s/%s", pod.Namespace, pod.Name)
	klog.Infof("[Infomer] remove key: %s", podUtils.AppName(pod))
	podUtils.RemoveSamplePod(podUtils.AppName(pod), c.samplePods)
	klog.Infof("[Informer] Delete sample pods", c.samplePods)
}

// NewPodLoggingController creates a PodLoggingController
func NewPodLoggingController(informerFactory informers.SharedInformerFactory, handle framework.FrameworkHandle,
	clientset *kubernetes.Clientset, samplePods map[string]*podUtils.SamplePod, sampleName string) *PodLoggingController {
	podInformer := informerFactory.Core().V1().Pods()

	c := &PodLoggingController{
		informerFactory:  informerFactory,
		podInformer:      podInformer,
		frameworkHandler: handle,
		clientset:        clientset,
		samplePods:       samplePods,
		sampleName:       sampleName,
	}
	podInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.podAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.podUpdate,
			// Called on resource deletion.
			DeleteFunc: c.podDelete,
		},
	)
	return c
}

// var kubeconfig string

// func init() {
// 	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
// }

// func main() {
// 	flag.Parse()
// 	logs.InitLogs()
// 	defer logs.FlushLogs()

// 	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		klog.Fatal(err)
// 	}

// 	factory := informers.NewSharedInformerFactory(clientset, time.Hour*24)
// 	controller := NewPodLoggingController(factory)
// 	stop := make(chan struct{})
// 	defer close(stop)
// 	err = controller.Run(stop)
// 	if err != nil {
// 		klog.Fatal(err)
// 	}
// 	select {}
// }
