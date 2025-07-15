package docs

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	//routes "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Determine the status of an ACME",
		Long:  "Determine the status of an ACME",
		Run:   status,
	}

}

func status(_ *cobra.Command, _ []string) {

	ctx := context.Background()

	fmt.Println("Running status...")

	kubeClient, err := CreateKubeClient()
	if err != nil {
		glog.Fatalf("failed to create kube client: %v", err)
	}

	namespaces, err := kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		glog.Fatalf("failed to list namespaces: %v", err)
	}

	for _, namespace := range namespaces.Items {
		glog.Infof("Checking status of namespace %s", namespace.Name)
	}

	kubeClient.DiscoveryClient.RESTClient().Patch()

	argoApplication := v1alpha1.Application{}

}
