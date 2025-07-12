package acme

import (
	"github.com/openshift-online/bootstrap/acme/pkg/api"
	"github.com/openshift-online/bootstrap/acme/pkg/api/external"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	v1 "k8s.io/api/core/v1"
)

func NewRegionalCluster(config *api.ClusterDeploymentConfig) RegionalCluster {
	return RegionalCluster{
		ClusterDeploymentConfig: *config,
		Namespace:               api.NewNamespace(config.Namespace),
		ClusterDeployment:       external.NewClusterDeployment(config),
		InstallConfig:           NewInstallConfig(config),
		MachinePool:             external.NewMachinePool(config),
		ManagedCluster:          external.NewManagedCluster(config),
		KlusterletAddonConfig:   external.NewKlusterletAddonConfig(config),
	}
}

type RegionalCluster struct {
	ClusterDeploymentConfig api.ClusterDeploymentConfig
	Namespace               *v1.Namespace
	ClusterDeployment       hivev1.ClusterDeployment
	InstallConfig           *v1.Secret
	MachinePool             hivev1.MachinePool
	ManagedCluster          external.ManagedCluster
	KlusterletAddonConfig   external.KlusterletAddonConfig
	Kustomization           external.Kustomization
}
