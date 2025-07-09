package manager

import (
	"context"
	"fmt"
	"github.com/openshift-online/bootstrap/acme/pkg/clients"
)

type Manager struct {
	k8sClient  *clients.Client
	argoClient *clients.ArgoClient
}

type ClusterStatus struct {
	NodeCount   int
	PodCount    int
	AppCount    int
	HealthyApps int
}

func NewManager(k8sClient *clients.Client, argoClient *clients.ArgoClient) *Manager {
	return &Manager{
		k8sClient:  k8sClient,
		argoClient: argoClient,
	}
}

func (m *Manager) CreateApplication(ctx context.Context, spec *clients.ApplicationSpec) error {
	return m.argoClient.CreateApplication(ctx, spec)
}

func (m *Manager) ListApplications(ctx context.Context) ([]clients.ApplicationStatus, error) {
	return m.argoClient.ListApplications(ctx)
}

func (m *Manager) SyncApplication(ctx context.Context, name string) error {
	return m.argoClient.SyncApplication(ctx, name)
}

func (m *Manager) DeleteApplication(ctx context.Context, name string) error {
	return m.argoClient.DeleteApplication(ctx, name)
}

func (m *Manager) GetClusterStatus(ctx context.Context) (*ClusterStatus, error) {
	// Get nodes
	nodes, err := m.k8sClient.GetNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	// Get pods across all namespaces
	pods, err := m.k8sClient.GetPods(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	// Get applications
	apps, err := m.argoClient.ListApplications(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	}

	// Count healthy apps
	healthyApps := 0
	for _, app := range apps {
		if app.Health == "Healthy" {
			healthyApps++
		}
	}

	return &ClusterStatus{
		NodeCount:   len(nodes),
		PodCount:    len(pods),
		AppCount:    len(apps),
		HealthyApps: healthyApps,
	}, nil
}

func (m *Manager) GetApplicationHealth(ctx context.Context, name string) (string, error) {
	app, err := m.argoClient.GetApplication(ctx, name)
	if err != nil {
		return "", err
	}

	return string(app.Status.Health.Status), nil
}
