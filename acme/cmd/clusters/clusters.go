package clusters

import (
	"github.com/openshift-online/bootstrap/acme/pkg/api"
)

func GetRegions() []*api.ClusterDeploymentConfig {

	return []*api.ClusterDeploymentConfig{
		{
			Name:            "cluster-01",
			Namespace:       "cluster-01",
			BaseDomain:      "rosa.mturansk-test.csu2.i3.devshift.org",
			AWSCreds:        "aws-creds",
			Region:          "us-east-1",
			ClusterImageSet: "img4.19.0-multi-appsub",
			InstallConfig:   "install-config",
			PullSecret:      "pull-secret",
		},
		//{
		//	Name:            "cluster-02",
		//	BaseDomain:      "rosa.mturansk-test.csu2.i3.devshift.org",
		//	AWSCreds:        "aws-creds",
		//	Region:          "eu-west-1",
		//	ClusterImageSet: "img4.19.0-multi-appsub",
		//	InstallConfig:   "install-config",
		//	PullSecret:      "pull-secret",
		//},
		//{
		//	Name:            "cluster-03",
		//	BaseDomain:      "rosa.mturansk-test.csu2.i3.devshift.org",
		//	AWSCreds:        "aws-creds",
		//	Region:          "ap-southeast-1",
		//	ClusterImageSet: "img4.19.0-multi-appsub",
		//	InstallConfig:   "install-config",
		//	PullSecret:      "pull-secret",
		//},
		//{
		//	Name:            "cluster-04",
		//	BaseDomain:      "rosa.mturansk-test.csu2.i3.devshift.org",
		//	AWSCreds:        "aws-creds",
		//	Region:          "sa-east-1",
		//	ClusterImageSet: "img4.19.0-multi-appsub",
		//	InstallConfig:   "install-config",
		//	PullSecret:      "pull-secret",
		//},
	}
}
