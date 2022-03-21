package main

// Code examples drawn from:
//    https://pkg.go.dev/text/tabwriter#Constants
//    https://github.com/kubernetes/client-go/tree/master/examples
//    https://hub.docker.com/_/hello-world
//    https://github.com/feiskyer/kubernetes-handbook/blob/master/examples/client/informer/informer.go
//    https://medium.com/codex/explore-client-go-informer-patterns-4415bb5f1fbd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	// 1. Setup client and connec to cluster
	fmt.Println("Connecting to local cluster ...")

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err)
	}

	// Setup variables
	var (
		namespaceName = "andromeda"
		label         = "app=sample-app"
		podName       = "hello-world-pod"
	)
	// 2. List namespaces
	fmt.Println("\n** Namespaces Before Creation **")
	listNamespaces(clientset)

	// 3. Create new namespace
	fmt.Println("\nCreating namespace ...")
	createNamespace(clientset, namespaceName)
	createNamespace(clientset, "milky-way")

	// List namespaces after namespace creation
	fmt.Println("\n** Namespaces After Creation **")
	listNamespaces(clientset)

	// Launch informer
	time.Sleep(15 * time.Second)
	go launchInformer(clientset, namespaceName)

	// 4. Create static pod
	fmt.Println("\nCreating pod ...")
	createPod(clientset, namespaceName)
	createPod(clientset, "milky-way")
	time.Sleep(15 * time.Second)

	// 5. List pods with label
	fmt.Println("\n** List of Pods **")
	listPods(clientset, &metav1.ListOptions{LabelSelector: label})

	// 6. Delete completed pod
	fmt.Println("\n** Deleting pod ... **")
	deletePod(clientset, namespaceName, podName)
	time.Sleep(15 * time.Second)

	// List pods with label after deletion
	fmt.Println("\n** List of Pods after deletion **")
	listPods(clientset, &metav1.ListOptions{LabelSelector: label})

	fmt.Println()
}

func listNamespaces(clientset *kubernetes.Clientset) {
	namespacesClient := clientset.CoreV1().Namespaces() // TODO decide if creation of clients should be moved out
	namespaceList, err := namespacesClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, namespace := range namespaceList.Items {
		fmt.Printf("\t%s\n", namespace.Name)
	}
}

func createNamespace(clientset *kubernetes.Clientset, namespaceName string) {
	namespaceObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	namespacesClient := clientset.CoreV1().Namespaces()
	namespacesClient.Create(context.TODO(), namespaceObj, metav1.CreateOptions{})
}

func createPod(clientset *kubernetes.Clientset, namespaceName string) {

	// TODO pod definition could be split from creation
	var (
		labelName     = "app"
		labelValue    = "sample-app"
		appName       = "hello-world-pod"
		containerName = "planet-earth"
		dockerImage   = "hello-world:latest"
	)

	// Declare static pod definition
	podObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				labelName: labelValue,
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:  containerName,
					Image: dockerImage,
				},
			},
		},
	}

	// Create static pod from definition
	podClient := clientset.CoreV1().Pods(namespaceName)
	_, err := podClient.Create(context.TODO(), podObj, metav1.CreateOptions{})
	if err != nil {
		log.Warn(err)
	}
}

func listPods(clientset *kubernetes.Clientset, options *metav1.ListOptions) {

	// Setup a writer to make columns pretty
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 2, '\t', 0)

	defer w.Flush()

	fmt.Fprintf(w, "%s\t%s\t", "Pod", "Namespace")
	fmt.Fprintf(w, "\n %s\t%s\t", "----", "----")

	// Go through each namespace looking for the label
	namespacesClient := clientset.CoreV1().Namespaces()
	namespaceList, err := namespacesClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, namespace := range namespaceList.Items {
		podClient := clientset.CoreV1().Pods(namespace.Name)
		podlist, err := podClient.List(context.TODO(), *options)

		if err != nil {
			log.Panic(err)
		}
		for _, pod := range podlist.Items {
			fmt.Fprintf(w, "\n%s\t%s\t", pod.Name, pod.Namespace)
		}
	}
}

func deletePod(clientset *kubernetes.Clientset, namespaceName string, podName string) {

	podClient := clientset.CoreV1().Pods(namespaceName)
	deletePolicy := metav1.DeletePropagationForeground
	err := podClient.Delete(context.TODO(), podName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		log.Warn(err)
	}
}

// Informer Stuff

type Controller struct {
	informerFactory informers.SharedInformerFactory
	podInformer     coreinformers.PodInformer
}

func launchInformer(clientset *kubernetes.Clientset, namespaceName string) {
	factory := informers.NewSharedInformerFactory(clientset, 30*time.Minute)

	podInformer := factory.Core().V1().Pods()

	controller := &Controller{
		informerFactory: factory,
		podInformer:     podInformer,
	}

	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    controller.podAdded,
			UpdateFunc: func(interface{}, interface{}) { log.Info("not implemented") },
			DeleteFunc: controller.podDeleted,
		},
	)

	stop := make(chan struct{})
	defer close(stop)

	controller.informerFactory.Start(stop)
	if !cache.WaitForCacheSync(stop, controller.podInformer.Informer().HasSynced) {
		log.Panic("Unable to sync informer")
		return
	}
}

func (controller *Controller) podAdded(modifiedResource interface{}) {
	pod := modifiedResource.(*corev1.Pod)
	fmt.Printf("\n -> pod: %s created in namespace: %s <-\n", pod.Name, pod.Namespace)
}

func (controller *Controller) podDeleted(modifiedResource interface{}) {
	pod := modifiedResource.(*corev1.Pod)
	fmt.Printf("\n ->pod: %s deleted in namespace: %s <-\n", pod.Name, pod.Namespace)
}
