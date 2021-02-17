package kubetools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
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
	restConfig         *rest.Config //?
	stopChan           chan struct{}
	readyChan          chan struct{}
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

func (k *kubeTools) setRestConfig() {
	var err error
	k.restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		panic("Could not load kubernetes configuration file")
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

func (k *kubeTools) dialer(namespace string, podName string) httpstream.Dialer {
	url := k.clientset.CoreV1().RESTClient().Post().
		Namespace(namespace).
		Resource("pods").
		Name(podName).
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(k.restConfig)
	if err != nil {
		panic("Could not create round tripper")
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)
	return dialer
}

func (k *kubeTools) forward(namespace string, podName string, localPort int32, podPort int32) {
	k.stopChan = make(chan struct{}, 1)
	readyChan := make(chan struct{}, 1)
	errChan := make(chan error, 1)

	dialer := k.dialer(namespace, podName)
	ports := []string{
		fmt.Sprintf("%d:%d", localPort, podPort),
	}

	discard := ioutil.Discard
	pf, err := portforward.New(dialer, ports, k.stopChan, readyChan, discard, discard)
	if err != nil {
		panic("Could not port forward into pod")
	}

	go func() {
		errChan <- pf.ForwardPorts()
	}()

	select {
	case err = <-errChan:
		panic("Could not create port forward")
	case <-readyChan:
		// return
		fmt.Println("Forward READY")
	}

}

func (k *kubeTools) RedirectServicePort(namespace string, serviceName string, servicePort int32, localPort int32) {
	k.setClientset()
	k.setRestConfig()

	svc := k.getService(namespace, serviceName)
	// debug
	// s, _ := json.MarshalIndent(svc.Spec, "", "\t")
	// fmt.Print("Service Spec:\n", string(s), "\n")

	podList := k.getPodsInService(namespace, svc)
	pod := k.getPodFromPodList(podList)
	podPort := k.getPodPort(svc, servicePort)

	fmt.Printf("Forward service: %s:%d (%s:%d) to localhost:%d\n", serviceName, servicePort, pod.Name, podPort, localPort)
	k.forward(namespace, pod.Name, localPort, podPort)
}
