package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/util/io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	hivev1 "github.com/openshift/hive/apis/hive/v1"
)

type ArgoClient struct {
	apiClient apiclient.Client
	appClient application.ApplicationServiceClient
}

type ApplicationSpec struct {
	Name           string
	Namespace      string
	RepoURL        string
	Path           string
	TargetRevision string
	DestServer     string
	DestNamespace  string
}

type ApplicationStatus struct {
	Name   string
	Status string
	Health string
}

func NewArgoClient(server, token string) (*ArgoClient, error) {
	clientOpts := &apiclient.ClientOptions{
		ServerAddr: server,
		AuthToken:  token,
		Insecure:   true, // For development only
	}

	apiClient, err := apiclient.NewClient(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create ArgoCD client: %w", err)
	}

	closer, appClient, err := apiClient.NewApplicationClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create application client: %w", err)
	}
	defer io.Close(closer)

	_ = hivev1.ClusterDeployment{}

	return &ArgoClient{
		apiClient: apiClient,
		appClient: appClient,
	}, nil
}

func (c *ArgoClient) CreateApplication(ctx context.Context, spec *ApplicationSpec) error {
	app := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: spec.Namespace,
		},
		Spec: v1alpha1.ApplicationSpec{
			Project: "default",
			Source: &v1alpha1.ApplicationSource{
				RepoURL:        spec.RepoURL,
				Path:           spec.Path,
				TargetRevision: spec.TargetRevision,
			},
			Destination: v1alpha1.ApplicationDestination{
				Server:    spec.DestServer,
				Namespace: spec.DestNamespace,
			},
			SyncPolicy: &v1alpha1.SyncPolicy{
				Automated: &v1alpha1.SyncPolicyAutomated{
					Prune:    true,
					SelfHeal: true,
				},
				SyncOptions: []string{
					"CreateNamespace=true",
				},
			},
		},
	}

	req := &application.ApplicationCreateRequest{
		Application: app,
	}

	_, err := c.appClient.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	return nil
}

func (c *ArgoClient) ListApplications(ctx context.Context) ([]ApplicationStatus, error) {
	req := &application.ApplicationQuery{}

	appList, err := c.appClient.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}

	var apps []ApplicationStatus
	for _, app := range appList.Items {
		apps = append(apps, ApplicationStatus{
			Name:   app.Name,
			Status: string(app.Status.Sync.Status),
			Health: string(app.Status.Health.Status),
		})
	}

	return apps, nil
}

func (c *ArgoClient) SyncApplication(ctx context.Context, name string) error {
	req := &application.ApplicationSyncRequest{
		Name: &name,
		SyncOptions: &application.SyncOptions{
			Items: []string{
				"CreateNamespace=true",
			},
		},
	}

	_, err := c.appClient.Sync(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to sync application: %w", err)
	}

	return nil
}

func (c *ArgoClient) GetApplication(ctx context.Context, name string) (*v1alpha1.Application, error) {
	req := &application.ApplicationQuery{
		Name: &name,
	}

	app, err := c.appClient.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	return app, nil
}

func (c *ArgoClient) DeleteApplication(ctx context.Context, name string) error {
	req := &application.ApplicationDeleteRequest{
		Name: &name,
	}

	_, err := c.appClient.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	return nil
}

func (c *ArgoClient) WaitForSync(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		app, err := c.GetApplication(ctx, name)
		if err != nil {
			return err
		}

		if app.Status.Sync.Status == v1alpha1.SyncStatusCodeSynced {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for application %s to sync", name)
}
