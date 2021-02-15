package kubetools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

type KubeTools struct {
	clientset          *kubernetes.Clientset
	kubeConfigFileName *string
}

func NewKubetools() *KubeTools {
	k := &KubeTools{}
	return k
}

func (k *KubeTools) setKubeConfigFileName() {
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

func (k *KubeTools) setClientset() {
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

func (k *KubeTools) ShowPodsInService(namespace string, serviceName string) {
	k.setClientset()

	var err error
	// get service
	svc, err := k.clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(svc.GetCreationTimestamp())

	s, _ := json.MarshalIndent(svc.Spec, "", "\t")
	fmt.Print("Service Spec:\n", string(s), "\n")

	// set pods for service
	set := labels.Set(svc.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := k.clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		panic(err.Error())
	}
	for _, pod := range pods.Items {
		fmt.Fprintf(os.Stdout, "pod name: %v\n", pod.Name)
	}

}
