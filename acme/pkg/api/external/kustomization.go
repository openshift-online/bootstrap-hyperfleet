// Code generated structs for kustomize.config.k8s.io/v1beta1 API.
// This file contains Go structs for the Kustomize Configuration API.
// Generated from official Kustomize API definitions.

package external

import (
	"github.com/openshift-online/bootstrap/acme/pkg/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewKustomization(config *api.ClusterDeploymentConfig) Kustomization {
	return Kustomization{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kustomize.config.k8s.io/v1beta1",
			Kind:       "Kustomization",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
		},
		Resources: []string{},
		CommonLabels: map[string]string{
			"cluster": config.Name,
		},
	}
}

// Kustomization defines a kustomization configuration
type Kustomization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Resources specifies relative paths to files holding YAML representations of kubernetes API objects
	Resources []string `json:"resources,omitempty"`

	// Bases are relative paths or git repository URLs to directories containing kustomization.yaml files
	Bases []string `json:"bases,omitempty"`

	// CommonLabels to add to all objects and selectors
	CommonLabels map[string]string `json:"commonLabels,omitempty"`

	// CommonAnnotations to add to all objects
	CommonAnnotations map[string]string `json:"commonAnnotations,omitempty"`

	// NamePrefix will prefix the names of all resources mentioned in the kustomization file
	NamePrefix string `json:"namePrefix,omitempty"`

	// NameSuffix will suffix the names of all resources mentioned in the kustomization file
	NameSuffix string `json:"nameSuffix,omitempty"`

	// Namespace to add to all objects
	Namespace string `json:"namespace,omitempty"`

	// Images is a list of (image name, new name, new tag or digest) for changing image names, tags or digests
	Images []Image `json:"images,omitempty"`

	// Replicas is a list of (resource name, count) for changing the number of replicas
	Replicas []Replica `json:"replicas,omitempty"`

	// PatchesStrategicMerge specifies the relative path to a file containing a strategic merge patch
	PatchesStrategicMerge []string `json:"patchesStrategicMerge,omitempty"`

	// PatchesJson6902 is a list of json 6902 patches and their targets
	PatchesJson6902 []Patch `json:"patchesJson6902,omitempty"`

	// Patches is a list of patches and their targets
	Patches []Patch `json:"patches,omitempty"`

	// ConfigMapGenerator is a list of configmaps to generate from local or literal sources
	ConfigMapGenerator []ConfigMapArgs `json:"configMapGenerator,omitempty"`

	// SecretGenerator is a list of secrets to generate from local or literal sources
	SecretGenerator []SecretArgs `json:"secretGenerator,omitempty"`

	// GeneratorOptions modify behavior of all ConfigMap and Secret generators
	GeneratorOptions *GeneratorOptions `json:"generatorOptions,omitempty"`

	// Vars allow things modified by kustomize to be injected into a kubernetes object specification
	Vars []Var `json:"vars,omitempty"`

	// Configurations is a list of transformer configuration files
	Configurations []string `json:"configurations,omitempty"`

	// Generators is a list of files containing custom generators
	Generators []string `json:"generators,omitempty"`

	// Transformers is a list of files containing transformers
	Transformers []string `json:"transformers,omitempty"`

	// CRDs is a list of CRD files for configuring CRDs
	CRDs []string `json:"crds,omitempty"`
}

// Image contains an image name, a new name, a new tag or digest, which will replace the original name and tag
type Image struct {
	// Name is a tag-less image name
	Name string `json:"name,omitempty"`

	// NewName is the value used to replace the original name
	NewName string `json:"newName,omitempty"`

	// NewTag is the value used to replace the original tag
	NewTag string `json:"newTag,omitempty"`

	// Digest is the value used to replace the original image tag
	Digest string `json:"digest,omitempty"`
}

// Replica contains resource name and replica count
type Replica struct {
	// Name of the resource
	Name string `json:"name"`

	// Count is the number of replicas
	Count int64 `json:"count"`
}

// Patch contains patch and its targets
type Patch struct {
	// Path is a relative file path to the patch file
	Path string `json:"path,omitempty"`

	// Patch is the content of a patch
	Patch string `json:"patch,omitempty"`

	// Target points to the resources that the patch should be applied to
	Target *Selector `json:"target,omitempty"`

	// Options is a list of options for the patch
	Options map[string]bool `json:"options,omitempty"`
}

// Selector specifies a set of resources
type Selector struct {
	Group              string `json:"group,omitempty"`
	Version            string `json:"version,omitempty"`
	Kind               string `json:"kind,omitempty"`
	Namespace          string `json:"namespace,omitempty"`
	Name               string `json:"name,omitempty"`
	AnnotationSelector string `json:"annotationSelector,omitempty"`
	LabelSelector      string `json:"labelSelector,omitempty"`
	FieldSelector      string `json:"fieldSelector,omitempty"`
}

// ConfigMapArgs contains the metadata of how to generate a configmap
type ConfigMapArgs struct {
	// GeneratorArgs for the configmap
	GeneratorArgs `json:",inline"`
}

// SecretArgs contains the metadata of how to generate a secret
type SecretArgs struct {
	// GeneratorArgs for the secret
	GeneratorArgs `json:",inline"`

	// Type of the secret
	Type string `json:"type,omitempty"`
}

// GeneratorArgs contains arguments common to generators
type GeneratorArgs struct {
	// Name of the generated resource
	Name string `json:"name,omitempty"`

	// Namespace of the generated resource
	Namespace string `json:"namespace,omitempty"`

	// Behavior of the generator
	Behavior string `json:"behavior,omitempty"`

	// DataSources for the generator
	DataSources `json:",inline"`

	// Options for the generator
	Options *GeneratorOptions `json:"options,omitempty"`
}

// DataSources contains sources of data for generators
type DataSources struct {
	// LiteralSources is a list of literal sources
	LiteralSources []string `json:"literals,omitempty"`

	// FileSources is a list of file sources
	FileSources []string `json:"files,omitempty"`

	// EnvSources is a list of env sources
	EnvSources []string `json:"envs,omitempty"`
}

// GeneratorOptions modify behavior of all generators
type GeneratorOptions struct {
	// Labels to add to all generated resources
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations to add to all generated resources
	Annotations map[string]string `json:"annotations,omitempty"`

	// DisableNameSuffixHash if true disables the default behavior of adding a suffix to the names
	DisableNameSuffixHash bool `json:"disableNameSuffixHash,omitempty"`

	// Immutable if true add to all generated resources
	Immutable bool `json:"immutable,omitempty"`
}

// Var represents a variable whose value will be sourced from a field in a Kubernetes object
type Var struct {
	// Name of the variable
	Name string `json:"name"`

	// ObjRef must refer to a Kubernetes resource under the same kustomization
	ObjRef VarReference `json:"objref"`

	// FieldRef refers to the field of the object referred to by ObjRef
	FieldRef FieldSelector `json:"fieldref,omitempty"`
}

// VarReference refers to a field in a Kubernetes object
type VarReference struct {
	// APIVersion of the referent
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the referent
	Kind string `json:"kind"`

	// Name of the referent
	Name string `json:"name"`

	// Namespace of the referent
	Namespace string `json:"namespace,omitempty"`
}

// FieldSelector contains the fieldPath to the object field
type FieldSelector struct {
	FieldPath string `json:"fieldPath,omitempty"`
}
