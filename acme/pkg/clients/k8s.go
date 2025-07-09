package clients

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Client struct {
	clientset kubernetes.Interface
}

type PodInfo struct {
	Name      string
	Namespace string
	Status    string
	Ready     bool
}

type DeploymentInfo struct {
	Name      string
	Namespace string
	Replicas  int32
	Ready     int32
}

func NewClient(clientset kubernetes.Interface) *Client {
	return &Client{clientset: clientset}
}

func (c *Client) GetPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var podInfos []PodInfo
	for _, pod := range pods.Items {
		podInfos = append(podInfos, PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			Ready:     isPodReady(&pod),
		})
	}

	return podInfos, nil
}

func (c *Client) GetDeployments(ctx context.Context, namespace string) ([]DeploymentInfo, error) {
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var deploymentInfos []DeploymentInfo
	for _, deployment := range deployments.Items {
		deploymentInfos = append(deploymentInfos, DeploymentInfo{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			Replicas:  *deployment.Spec.Replicas,
			Ready:     deployment.Status.ReadyReplicas,
		})
	}

	return deploymentInfos, nil
}

func (c *Client) GetNodes(ctx context.Context) ([]corev1.Node, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return nodes.Items, nil
}

func (c *Client) CreateDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	_, err := c.clientset.AppsV1().Deployments(deployment.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	return nil
}

func (c *Client) DeleteDeployment(ctx context.Context, name, namespace string) error {
	err := c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	return nil
}

func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
