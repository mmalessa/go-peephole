package kubetools

import (
	"context"
	"fmt"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

/*
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/labels"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
*/

type kubeTools struct {
	clientset          *kubernetes.Clientset
	kubeConfigFileName *string
}

// NewKubetools Elo, elo
func NewKubetools() *kubeTools {
	k := &kubeTools{}
	return k
}

func (k *kubeTools) setKubeConfigFileName() {
	if k.kubeConfigFileName != nil {
		return
	}
	home := homedir.HomeDir()
	if home == "" {
		panic("HomeDir not found!")
	}
	configFileName := filepath.Join(home, ".kube", "config")
	k.kubeConfigFileName = &configFileName
}

func (k *kubeTools) setClientset() {
	if k.clientset != nil {
		return
	}
	k.setKubeConfigFileName()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *k.kubeConfigFileName)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	k.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

func (k *kubeTools) getService(namespace string, serviceName string) *v1.Service {
	svc, err := k.clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	return svc
}

func (k *kubeTools) getPodPort(svc *v1.Service, servicePort int32) int32 {
	for _, port := range svc.Spec.Ports {
		if port.Port == servicePort {
			return port.TargetPort.IntVal
		}
	}
	panic("Port " + string(servicePort) + " not found in service spec")
}

func (k *kubeTools) getPodsInService(namespace string, svc *v1.Service) *v1.PodList {
	set := labels.Set(svc.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	podList, err := k.clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		panic(err.Error())
	}
	return podList
}

func (k *kubeTools) getPodFromPodList(podList *v1.PodList) *v1.Pod {
	for _, item := range podList.Items {
		return &item
	}
	panic("Empty podList.Items")
}

func (k *kubeTools) forward() {
	fmt.Println("TODO")
}

func (k *kubeTools) RedirectServicePort(namespace string, serviceName string, servicePort int32, localPort int32) {
	k.setClientset()
	svc := k.getService(namespace, serviceName)
	// debug
	// s, _ := json.MarshalIndent(svc.Spec, "", "\t")
	// fmt.Print("Service Spec:\n", string(s), "\n")

	podList := k.getPodsInService(namespace, svc)
	pod := k.getPodFromPodList(podList)
	podPort := k.getPodPort(svc, servicePort)

	fmt.Printf("Forward service: %s:%d (%s:%d) to localhost:%d\n", serviceName, servicePort, pod.Name, podPort, localPort)
	k.forward()
}
