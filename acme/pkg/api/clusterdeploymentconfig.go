package api

type ClusterDeploymentConfig struct {
	Name            string
	BaseDomain      string
	AWSCreds        string
	Region          string
	ClusterImageSet string
	InstallConfig   string
	PullSecret      string

	MasterReplicas     int64
	MasterInstanceType string

	WorkerReplicas     int64
	WorkerInstanceType string
}