/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"bufio"
	//"flag"
	"fmt"
	"os"
	//"path/filepath"
	"github.com/golang/glog"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiresrc "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	k8stpv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	//"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	//Creating K8s interface
	/*var kubeconfig *string

	//find kubeconfig file in home directory to establish a connection interface
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)*/
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	fmt.Println("Creating pod/node listener...")
	stopCh := make(chan struct{})
	//schedulerName := "kube-scheduler-minikube"
	//podController := NewPodWatcher(schedulerName, clientset)
	//nodeWatcher := NewNodeWatcher(clientset)

	//go RunPodWatcher(stopCh, 10, *podController)
	//go RunNodeWatcher(stopCh, 10, *nodeWatcher)

	prompt()
	//Call kubenetes to create a depolyment object 
	//called "demo-custom-dep" with 1 cpu
	fmt.Println("Creating deployment...")
	glog.Infof("Creating deployment...")
	CreateDeployment("demo-custom-dep", 0,deploymentsClient)
	
	// Update Deployment
	prompt()
	fmt.Println("Updating deployment...")
	glog.Infof("Updating deployment...")
	UpdateDeployment("demo-custom-dep", deploymentsClient)
	
	// List Deployments
	prompt()
	fmt.Printf("Listing deployments in namespace %q:\n", apiv1.NamespaceDefault)
	glog.Infof("info: Listing deployments in namespace %q:\n", apiv1.NamespaceDefault)
	ListDeployment(deploymentsClient)

	// Delete Deployment
	prompt()
	fmt.Println("Deleting deployment...")
	DeleteDeployment("demo-custom-dep", deploymentsClient)

	
	// We block here.
	<-stopCh
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

// Run starts node watcher.
func RunNodeWatcher(stopCh <-chan struct{}, nWorkers int, nodeWatcher cache.Controller) {
	defer utilruntime.HandleCrash()

	// The workers can stop when we are done.
	//defer nw.nodeWorkQueue.ShutDown()
	defer fmt.Println("Shutting down NodeWatcher")
	fmt.Println("Getting node updates...")

	go nodeWatcher.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, nodeWatcher.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	fmt.Println("Starting node watching workers")
	//for i := 0; i < nWorkers; i++ {
	//	go wait.Until(nw.nodeWorker, time.Second, stopCh)
	//}
	<-stopCh
	fmt.Println("Stopping node watcher")
}

func NewNodeWatcher(client kubernetes.Interface) *cache.Controller {
	fmt.Println("Starting NodeWatcher...")

	_, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(alo metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Nodes().List(alo)
			},
			WatchFunc: func(alo metav1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Nodes().Watch(alo)
			},
		},
		&apiv1.Node{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				_, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					fmt.Printf("AddFunc: error getting key %v", err)
				}
				fmt.Println("Node object added.")
			},
			UpdateFunc: func(old, new interface{}) {
				_, err := cache.MetaNamespaceKeyFunc(new)
				if err != nil {
					fmt.Printf("UpdateFunc: error getting key %v", err)
				}
				fmt.Println("Node object updated.")
			},
			DeleteFunc: func(obj interface{}) {
				_, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					fmt.Printf("DeleteFunc: error getting key %v", err)
				}
				fmt.Println("Node object deleted.")
			},
		},
	)
	return &controller;
}


// Run starts a pod watcher.
func RunPodWatcher(stopCh <-chan struct{}, nWorkers int, podCtl cache.Controller) {
	defer utilruntime.HandleCrash()

	// The workers can stop when we are done.
	fmt.Println("Getting pod updates...")

	go podCtl.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, podCtl.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	fmt.Println("Pod watcher cache synced.")
	<-stopCh
	fmt.Println("Stopping pod watcher.")
}

// NewPodWatcher initialize a PodWatcher.
func NewPodWatcher( schedulerName string, client kubernetes.Interface) *cache.Controller{

	schedulerSelector := fields.Everything()
	podSelector := labels.Everything()

	// schedulerName is only available in Kubernetes >= 1.6.
	//schedulerSelector = fields.ParseSelectorOrDie("spec.schedulerName==" + schedulerName)

	_, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(alo metav1.ListOptions) (runtime.Object, error) {
				alo.FieldSelector = schedulerSelector.String()
				alo.LabelSelector = podSelector.String()
				return client.CoreV1().Pods("").List(alo)
			},
			WatchFunc: func(alo metav1.ListOptions) (watch.Interface, error) {
				alo.FieldSelector = schedulerSelector.String()
				alo.LabelSelector = podSelector.String()
				return client.CoreV1().Pods("").Watch(alo)
			},
		},
		&apiv1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				_, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					fmt.Printf("AddFunc: error getting key %v", err)
				}
				fmt.Println("Pod object added.")
			},
			UpdateFunc: func(old, new interface{}) {
				_, err := cache.MetaNamespaceKeyFunc(new)
				if err != nil {
					fmt.Printf("UpdateFunc: error getting key %v", err)
				}
				fmt.Println("Pod object updated.")
				//podWatcher.enqueuePodUpdate(key, old, new)
			},
			DeleteFunc: func(obj interface{}) {
				_, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					fmt.Printf("DeleteFunc: error getting key %v", err)
				}
				fmt.Println("Pod object deleted.")
				//podWatcher.enqueuePodDeletion(key, obj)
			},
		},
	)
	fmt.Println("Pod listener created...")
	return &controller
}


func int32Ptr(i int32) *int32 { return &i }


func DeleteDeployment(podName string, deploymentsClient k8stpv1.DeploymentInterface){
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(podName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted deployment.")
}

/*List all deployments*/
func ListDeployment(deploymentsClient k8stpv1.DeploymentInterface){
	list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
		fmt.Printf(" ")
	}
}

func CreateDeployment(podName string, cpuLimit int64, deploymentsClient k8stpv1.DeploymentInterface){
	cpuQuantity := &apiresrc.Quantity{}
	cpuQuantity.Set(cpuLimit)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU: *cpuQuantity,
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	return
}

func BindPodToNode(clientSet kubernetes.Interface, bindInfo BindInfo) {
	for {
		err := clientSet.CoreV1().Pods(bindInfo.Namespace).Bind(&apiv1.Binding{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: bindInfo.Name,
			},
			Target: apiv1.ObjectReference{
				Namespace: bindInfo.Namespace,
				Name:      bindInfo.Nodename,
			}})
		if err != nil {
			fmt.Printf("Could not bind pod:%s to nodeName:%s, error: %v", bindInfo.Name, bindInfo.Nodename, err)
		}
	}
}

/*func UpdatePodSpec(cpuLimit apiresrc.Quantity, memLimit apiresrc.Quantity, result *appsv1.Deployment){
		//    You have two options to Update() this Deployment:
	//
	//    1. Modify the "deployment" variable and call: Update(deployment).
	//       This works like the "kubectl replace" command and it overwrites/loses changes
	//       made by other clients between you Create() and Update() the object.
	//    2. Modify the "result" returned by Get() and retry Update(result) until
	//       you no longer get a conflict error. This way, you can preserve changes made
	//       by other clients between Create() and Update(). This is implemented below
	//			 using the retry utility package included with client-go. (RECOMMENDED)
	//
	// More Info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency

	resLimit := &apiv1.ResourceList{
		apiv1.ResourceCPU: cpuLimit, 
		apiv1.ResourceMemory: memLimit,
	}
	result.Spec.Template.Spec.Containers[0].Resources.Limits = *resLimit
}*/
func UpdateDeployment(podName string, deploymentsClient k8stpv1.DeploymentInterface){
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := deploymentsClient.Get(podName, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}

		result.Spec.Replicas = int32Ptr(1)                           // reduce replica count
		result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13" // change nginx version
		_, updateErr := deploymentsClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
	fmt.Println("Updated deployment...")
}

// BindInfo
//k8sclient.BindChannel <- k8sclient.BindInfo{Name: podIdentifier.Name, Namespace: podIdentifier.Namespace, Nodename: nodeName}
type BindInfo struct {
	Name      string
	Namespace string
	Nodename  string
}


