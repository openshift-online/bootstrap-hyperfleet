package api

import (
	hivev1 "github.com/openshift/hive/apis/hive/v1"
)

func NewRegionalCluster(config *ClusterDeploymentConfig) RegionalCluster {
	return RegionalCluster{
		ClusterDeploymentConfig: *config,
		ClusterDeployment:       NewClusterDeployment(config),
		InstallConfig:          *NewInstallConfig(config),
		MachinePool:            NewMachinePool(config),
		ManagedCluster:         NewManagedCluster(config),
	}
}

type RegionalCluster struct {
	ClusterDeploymentConfig ClusterDeploymentConfig
	ClusterDeployment       hivev1.ClusterDeployment
	InstallConfig           InstallConfig
	MachinePool             hivev1.MachinePool
	ManagedCluster          ManagedCluster
}