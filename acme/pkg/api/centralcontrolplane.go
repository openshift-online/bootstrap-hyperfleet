package api

type CentralControlPlane struct {
	RegionalClusters []RegionalCluster
	ClusterDeploymentConfig ClusterDeploymentConfig
}