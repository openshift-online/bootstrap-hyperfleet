package docs

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func CreateKubeClient() (*kubernetes.Clientset, error) {
	var restConfig *rest.Config
	var err error

	restConfig, err = rest.InClusterConfig()

	if err != nil {
		glog.Infof("Error creating in-cluster kube config: %v\n", err)
		glog.Info("attempting out-of-cluster kubeconfig")

		var kubeconfig *string
		if kcfg := os.Getenv("KUBECONFIG"); kcfg != "" {
			kubeconfig = flag.String("kubeconfig", kcfg, "absoluate path to kubeconfig file")
		} else if home := os.Getenv("HOME"); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "ocm-mturansk", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		restConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			glog.Fatalf("Error creating out-of-cluster kube config: %v\n", err)
			return nil, fmt.Errorf("could not created kube client: %s", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
