package acme

import (
	"github.com/openshift-online/bootstrap/acme/pkg/api"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewInstallConfig(config *api.ClusterDeploymentConfig) *corev1.Secret {

	ic := InstallConfig{
		APIVersion: "v1",
		BaseDomain: config.BaseDomain,
		Metadata: metav1.ObjectMeta{
			Name: config.Name,
		},
		ControlPlane: MachinePool{
			Name:           "master",
			Replicas:       config.MasterReplicas,
			Hyperthreading: "Enabled",
			Architecture:   "amd64",
			Platform: MachinePoolPlatform{
				AWS: &AWSMachinePoolPlatform{
					Type: config.MasterInstanceType,
					Zones: []string{
						config.Region,
					},
					RootVolume: RootVolume{
						IOPS: 4000,
						Size: 100,
						Type: "io1",
					},
				},
			},
		},
		Compute: []MachinePool{
			{
				Name:           "worker",
				Replicas:       1,
				Hyperthreading: "Enabled",
				Architecture:   "amd64",
				Platform: MachinePoolPlatform{
					AWS: &AWSMachinePoolPlatform{
						Type: config.WorkerInstanceType,
						Zones: []string{
							config.Region,
						},
						RootVolume: RootVolume{
							IOPS: 2000,
							Size: 100,
							Type: "io1",
						},
					},
				},
			},
		},
		Networking: Networking{
			NetworkType: "OVNKubernetes",
			ClusterNetwork: []ClusterNetworkEntry{
				{
					CIDR:       "10.128.0.0/14",
					HostPrefix: 23,
				},
			},
			ServiceNetwork: []string{
				"172.30.0.0/16",
			},
			MachineNetwork: []MachineNetworkEntry{
				{
					CIDR: "10.0.0.0/16",
				},
			},
		},
		Platform: Platform{
			AWS: &AWSPlatform{
				Region: config.Region,
			},
		},
		PullSecret: "", // This will be omitted in the final YAML
	}

	y, err := yaml.Marshal(ic)
	if err != nil {
		return nil
	}

	secret := api.NewSecret("install-config", config.Namespace)
	secret.Data = map[string][]byte{
		"install-config.yaml": []byte(y),
	}

	return secret
}

// InstallConfig is the top-level structure for the install-config.yaml file.
type InstallConfig struct {
	APIVersion   string            `yaml:"apiVersion"`
	Metadata     metav1.ObjectMeta `yaml:"metadata"`
	BaseDomain   string            `yaml:"baseDomain"`
	ControlPlane MachinePool       `yaml:"controlPlane"`
	Compute      []MachinePool     `yaml:"compute"`
	Networking   Networking        `yaml:"networking"`
	Platform     Platform          `yaml:"platform"`
	PullSecret   string            `yaml:"pullSecret,omitempty"`
}

// MachinePool defines the configuration for a group of machines, like the control plane or computes.
type MachinePool struct {
	Architecture   string              `yaml:"architecture"`
	Hyperthreading string              `yaml:"hyperthreading"`
	Name           string              `yaml:"name"`
	Replicas       int64               `yaml:"replicas"`
	Platform       MachinePoolPlatform `yaml:"platform"`
}

// MachinePoolPlatform holds platform-specific configuration for a MachinePool.
type MachinePoolPlatform struct {
	AWS *AWSMachinePoolPlatform `yaml:"aws,omitempty"`
}

// AWSMachinePoolPlatform defines AWS-specific machine pool settings.
type AWSMachinePoolPlatform struct {
	Zones      []string   `yaml:"zones,omitempty"`
	RootVolume RootVolume `yaml:"rootVolume"`
	Type       string     `yaml:"type"`
}

// RootVolume defines the settings for the root EBS volume.
type RootVolume struct {
	IOPS int    `yaml:"iops"`
	Size int    `yaml:"size"`
	Type string `yaml:"type"`
}

// Networking defines the networking configuration for the cluster.
type Networking struct {
	NetworkType    string                `yaml:"networkType"`
	ClusterNetwork []ClusterNetworkEntry `yaml:"clusterNetwork"`
	MachineNetwork []MachineNetworkEntry `yaml:"machineNetwork"`
	ServiceNetwork []string              `yaml:"serviceNetwork"`
}

// ClusterNetworkEntry defines the CIDR and host prefix for the cluster network.
type ClusterNetworkEntry struct {
	CIDR       string `yaml:"cidr"`
	HostPrefix int    `yaml:"hostPrefix"`
}

// MachineNetworkEntry defines the CIDR for the machine network.
type MachineNetworkEntry struct {
	CIDR string `yaml:"cidr"`
}

// Platform defines the cloud provider for the cluster.
type Platform struct {
	AWS *AWSPlatform `yaml:"aws,omitempty"`
}

// AWSPlatform holds AWS-specific configuration.
type AWSPlatform struct {
	Region string `yaml:"region"`
}
