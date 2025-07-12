// Code generated structs for agent.open-cluster-management.io/v1 KlusterletAddonConfig API.
// This file contains Go structs for the ACM KlusterletAddonConfig CRD.
// Generated from official ACM API definitions.

package external

import (
	"github.com/openshift-online/bootstrap/acme/pkg/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewKlusterletAddonConfig(config *api.ClusterDeploymentConfig) KlusterletAddonConfig {
	return KlusterletAddonConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "agent.open-cluster-management.io/v1",
			Kind:       "KlusterletAddonConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Name,
		},
		Spec: KlusterletAddonConfigSpec{
			ClusterName:      config.Name,
			ClusterNamespace: config.Name,
			ClusterLabels: map[string]string{
				"cloud":  "Amazon",
				"name":   config.Name,
				"vendor": "OpenShift",
			},
			ApplicationManager: ApplicationManagerConfig{
				Enabled: true,
			},
			PolicyController: PolicyController{
				Enabled: true,
			},
			SearchCollector: SearchCollector{
				Enabled: true,
			},
			CertPolicyController: CertPolicyController{
				Enabled: true,
			},
		},
	}
}

// KlusterletAddonConfig is the configuration for KlusterletAddon
type KlusterletAddonConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KlusterletAddonConfigSpec   `json:"spec,omitempty"`
	Status KlusterletAddonConfigStatus `json:"status,omitempty"`
}

// KlusterletAddonConfigSpec defines the desired state of KlusterletAddonConfig
type KlusterletAddonConfigSpec struct {
	// ClusterName is the name of the managed cluster
	ClusterName string `json:"clusterName"`

	// ClusterNamespace is the namespace of the managed cluster
	ClusterNamespace string `json:"clusterNamespace"`

	// ClusterLabels are the labels on the managed cluster
	ClusterLabels map[string]string `json:"clusterLabels,omitempty"`

	// ApplicationManager defines the configurations of ApplicationManager
	ApplicationManager ApplicationManagerConfig `json:"applicationManager,omitempty"`

	// PolicyController defines the configurations of PolicyController
	PolicyController PolicyController `json:"policyController,omitempty"`

	// SearchCollector defines the configurations of SearchCollector
	SearchCollector SearchCollector `json:"searchCollector,omitempty"`

	// CertPolicyController defines the configurations of CertPolicyController
	CertPolicyController CertPolicyController `json:"certPolicyController,omitempty"`

	// IAMPolicyController defines the configurations of IAMPolicyController
	IAMPolicyController IAMPolicyController `json:"iamPolicyController,omitempty"`

	// Version is the version of the KlusterletAddon
	Version string `json:"version,omitempty"`
}

// ApplicationManagerConfig defines the configurations of ApplicationManager
type ApplicationManagerConfig struct {
	// Enabled indicates whether ApplicationManager is enabled
	Enabled bool `json:"enabled"`

	// ArgocdCluster indicates whether to connect to Argo CD cluster
	ArgocdCluster bool `json:"argocdCluster,omitempty"`

	// ProxyConfig defines the proxy configurations
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
}

// PolicyController defines the configurations of PolicyController
type PolicyController struct {
	// Enabled indicates whether PolicyController is enabled
	Enabled bool `json:"enabled"`

	// ProxyConfig defines the proxy configurations
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
}

// SearchCollector defines the configurations of SearchCollector
type SearchCollector struct {
	// Enabled indicates whether SearchCollector is enabled
	Enabled bool `json:"enabled"`

	// ProxyConfig defines the proxy configurations
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
}

// CertPolicyController defines the configurations of CertPolicyController
type CertPolicyController struct {
	// Enabled indicates whether CertPolicyController is enabled
	Enabled bool `json:"enabled"`

	// ProxyConfig defines the proxy configurations
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
}

// IAMPolicyController defines the configurations of IAMPolicyController
type IAMPolicyController struct {
	// Enabled indicates whether IAMPolicyController is enabled
	Enabled bool `json:"enabled"`

	// ProxyConfig defines the proxy configurations
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
}

// ProxyConfig defines the proxy configurations
type ProxyConfig struct {
	// HTTPProxy is the HTTP proxy URL
	HTTPProxy string `json:"httpProxy,omitempty"`

	// HTTPSProxy is the HTTPS proxy URL
	HTTPSProxy string `json:"httpsProxy,omitempty"`

	// NoProxy is the list of hostnames/domains that should not use proxy
	NoProxy string `json:"noProxy,omitempty"`

	// CABundle is the custom CA bundle for proxy
	CABundle []byte `json:"caBundle,omitempty"`
}

// KlusterletAddonConfigStatus defines the observed state of KlusterletAddonConfig
type KlusterletAddonConfigStatus struct {
	// Conditions is the list of conditions and their status
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ApplicationManagerStatus defines the status of ApplicationManager
	ApplicationManagerStatus AddonStatus `json:"applicationManagerStatus,omitempty"`

	// PolicyControllerStatus defines the status of PolicyController
	PolicyControllerStatus AddonStatus `json:"policyControllerStatus,omitempty"`

	// SearchCollectorStatus defines the status of SearchCollector
	SearchCollectorStatus AddonStatus `json:"searchCollectorStatus,omitempty"`

	// CertPolicyControllerStatus defines the status of CertPolicyController
	CertPolicyControllerStatus AddonStatus `json:"certPolicyControllerStatus,omitempty"`

	// IAMPolicyControllerStatus defines the status of IAMPolicyController
	IAMPolicyControllerStatus AddonStatus `json:"iamPolicyControllerStatus,omitempty"`
}

// AddonStatus defines the status of an addon
type AddonStatus struct {
	// Enabled indicates whether the addon is enabled
	Enabled bool `json:"enabled"`

	// Available indicates whether the addon is available
	Available metav1.ConditionStatus `json:"available,omitempty"`

	// Message is the status message
	Message string `json:"message,omitempty"`
}

// KlusterletAddonConfigList contains a list of KlusterletAddonConfig
type KlusterletAddonConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KlusterletAddonConfig `json:"items"`
}
