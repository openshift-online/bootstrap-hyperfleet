package clusters

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewManagedCluster(config *ClusterDeploymentConfig) ManagedCluster {
	return ManagedCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.open-cluster-management.io/v1",
			Kind:       "ManagedCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Name,
			Labels: map[string]string{
				"name":   config.Name,
				"cloud":  "Amazon",
				"vendor": "OpenShift",
				"cluster.open-cluster-management.io/clusterset": "default",
			},
		},
		Spec: ManagedClusterSpec{
			HubAcceptsClient: true,
		},
	}
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,path=managedclusters

// ManagedCluster is the representation of a cluster that will be managed by the hub.
// When a ManagedCluster is created, a corresponding ManagedCluster object is created in the hub cluster in the cluster's namespace.
// A hub-side controller will approve the join of the cluster and then a managed-side agent will be deployed in the cluster
// to establish the connection between the managed cluster and the hub.
// This is an alpha API, so we reserve the right to change this API in the future.
type ManagedCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the ManagedCluster.
	Spec ManagedClusterSpec `json:"spec,omitempty"`

	// Status represents the current status of the ManagedCluster.
	Status ManagedClusterStatus `json:"status,omitempty"`
}

// ManagedClusterSpec defines the desired state of the ManagedCluster.
type ManagedClusterSpec struct {
	// HubAcceptsClient specifies whether the hub agent on the managed cluster can be accepted.
	// The managed cluster agent generates a CSR on the hub cluster to get a client certificate to connect to the hub.
	// If this field is set to true, the hub agent will be automatically approved; otherwise, the CSR will be required to be approved manually.
	HubAcceptsClient bool `json:"hubAcceptsClient"`

	// LeaseDurationSeconds is the lease duration for the managed cluster,
	// the managed cluster has to update its lease every leaseDurationSeconds.
	// +optional
	LeaseDurationSeconds int32 `json:"leaseDurationSeconds,omitempty"`

	// ManagedClusterClientConfigs represents a list of the apiserver address of the managed cluster.
	// If it is empty, the managed cluster has no accessible address to be visited from hub.
	// +optional
	ManagedClusterClientConfigs []ClientConfig `json:"managedClusterClientConfigs,omitempty"`
}

// ClientConfig represents the apiserver address of the managed cluster.
type ClientConfig struct {
	// URL is the URL of apiserver endpoint of the managed cluster.
	URL string `json:"url"`

	// CABundle is the ca bundle of apiserver endpoint of the managed cluster.
	// If it is not specified, the apiserver of the managed cluster is visiting without ca bundle.
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`
}

// ManagedClusterStatus represents the current status of the ManagedCluster.
type ManagedClusterStatus struct {
	// Conditions contains the different condition statuses for this managed cluster.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Capacity represents the total resource capacity from all node-agents on the managed cluster.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Capacity map[string]runtime.RawExtension `json:"capacity,omitempty"`

	// Allocatable represents the total allocatable resources on the managed cluster.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Allocatable map[string]runtime.RawExtension `json:"allocatable,omitempty"`

	// Version represents the kubernetes version of the managed cluster.
	// +optional
	Version ManagedClusterVersion `json:"version,omitempty"`

	// ClusterClaims represents cluster information that is useful for the hub to schedule work to the managed cluster.
	// The claims are received from the managed cluster agent.
	// +optional
	ClusterClaims []ClusterClaim `json:"clusterClaims,omitempty"`
}

// ManagedClusterVersion represents the kubernetes version of the managed cluster.
type ManagedClusterVersion struct {
	// Kubernetes is the kubernetes version of the managed cluster.
	// +optional
	Kubernetes string `json:"kubernetes,omitempty"`
}

// ClusterClaim represents a Cluster-wide resource that is not specific to any namespace.
// It is used to hold information about cluster status.
type ClusterClaim struct {
	// Name is the name of the cluster claim.
	Name string `json:"name"`
	// Value is the value of the cluster claim.
	Value string `json:"value"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManagedClusterList is a list of ManagedCluster objects
type ManagedClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// A list of ManagedCluster objects.
	Items []ManagedCluster `json:"items"`
}
