package api

import (
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"github.com/openshift/hive/apis/hive/v1/aws"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewClusterDeployment(config *ClusterDeploymentConfig) hivev1.ClusterDeployment {
	return hivev1.ClusterDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "hive.openshift.io/v1",
			Kind:       "ClusterDeployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Name,
		},
		Spec: hivev1.ClusterDeploymentSpec{
			BaseDomain:  config.BaseDomain,
			ClusterName: config.Name,
			ControlPlaneConfig: hivev1.ControlPlaneConfigSpec{
				ServingCertificates: hivev1.ControlPlaneServingCertificateSpec{},
			},
			Platform: hivev1.Platform{
				AWS: &aws.Platform{
					CredentialsSecretRef: corev1.LocalObjectReference{
						Name: config.AWSCreds,
					},
					Region: config.Region,
				},
			},
			Provisioning: &hivev1.Provisioning{
				ImageSetRef: &hivev1.ClusterImageSetReference{
					Name: config.ClusterImageSet,
				},
				InstallConfigSecretRef: &corev1.LocalObjectReference{
					Name: config.InstallConfig,
				},
			},
			PullSecretRef: &corev1.LocalObjectReference{
				Name: config.PullSecret,
			},
		},
	}
}
