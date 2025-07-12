package acme

import "github.com/openshift-online/bootstrap/acme/pkg/api"

type CentralControlPlane struct {
	RegionalClusters        []RegionalCluster
	ClusterDeploymentConfig api.ClusterDeploymentConfig
}
