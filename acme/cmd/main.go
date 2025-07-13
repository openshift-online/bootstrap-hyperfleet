package main

import (
	"encoding/json"
	"fmt"
	"github.com/openshift-online/bootstrap/acme/pkg/api/external"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/openshift-online/bootstrap/acme/cmd/clusters"
	"github.com/openshift-online/bootstrap/acme/pkg/api/acme"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"
)

var (
	kubeconfig string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "acme",
		Short: "ACME cluster configuration generator",
		Long:  "Generate OpenShift cluster manifests for GitOps deployment",
	}

	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	rootCmd.AddCommand(NewClustersCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func NewClustersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clusters",
		Short: "Generate cluster configuration manifests",
		Run:   runClusters,
	}
}

func runClusters(cmd *cobra.Command, args []string) {
	desiredClusters := clusters.GetRegions()

	for _, spec := range desiredClusters {
		rc := acme.NewRegionalCluster(spec)
		kustomization := external.NewKustomization(&rc.ClusterDeploymentConfig)

		files := map[string]interface{}{
			"namespace.json":             rc.Namespace,
			"clusterdeployment.json":     rc.ClusterDeployment,
			"installconfig.json":         rc.InstallConfig,
			"machinepool.json":           rc.MachinePool,
			"managedcluster.json":        rc.ManagedCluster,
			"klusterletaddonconfig.json": rc.KlusterletAddonConfig,
		}

		for filename, obj := range files {
			path := "./clusters/overlay/" + spec.Name + "/" + filename
			if err := WriteFile(path, toJSON(obj)); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", filename, err)
			}

			kustomization.Resources = append(kustomization.Resources, filename)
		}

		if err := WriteFile("./clusters/overlay/"+spec.Name+"/kustomization.yaml", toYAML(kustomization)); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing kustomization.yaml: %v\n", err)
		}
	}
}

func toJSON(obj interface{}) string {
	cleanObj := removeRuntimeFields(obj)
	y, err := json.MarshalIndent(cleanObj, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(y)
}

func removeRuntimeFields(obj interface{}) interface{} {
	objBytes, err := json.Marshal(obj)
	if err != nil {
		return obj
	}
	
	var objMap map[string]interface{}
	if err := json.Unmarshal(objBytes, &objMap); err != nil {
		return obj
	}
	
	// Remove status field
	delete(objMap, "status")
	
	// Remove creationTimestamp from metadata
	if metadata, ok := objMap["metadata"].(map[string]interface{}); ok {
		delete(metadata, "creationTimestamp")
		if len(metadata) == 0 {
			delete(objMap, "metadata")
		}
	}
	
	return objMap
}

func toYAML(obj interface{}) string {
	cleanObj := removeRuntimeFields(obj)
	y, err := yaml.Marshal(cleanObj)
	if err != nil {
		return err.Error()
	}
	return string(y)
}

func GetProjectRootDir() string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(filepath.Join(b, "..", ".."))
	return basepath
}

func ReadFile(file string) (string, error) {
	// If the value is in quotes, unquote it
	unquotedFile, err := strconv.Unquote(file)
	if err != nil {
		// values without quotes will raise an error, ignore it.
		unquotedFile = file
	}

	// If no file is provided, leave val unchanged.
	if unquotedFile == "" {
		return "", nil
	}

	// Ensure the absolute file path is used
	absFilePath := unquotedFile
	if !filepath.IsAbs(unquotedFile) {
		absFilePath = filepath.Join(GetProjectRootDir(), unquotedFile)
	}

	// Read the file
	buf, err := os.ReadFile(absFilePath)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func WriteFile(path, contents string) error {
	// If the value is in quotes, unquote it
	unquotedFile, err := strconv.Unquote(path)
	if err != nil {
		// values without quotes will raise an error, ignore it.
		unquotedFile = path
	}

	// If no path is provided, leave val unchanged.
	if unquotedFile == "" {
		return nil
	}

	// Ensure the absolute path path is used
	absFilePath := unquotedFile
	if !filepath.IsAbs(unquotedFile) {
		absFilePath = filepath.Join(GetProjectRootDir(), unquotedFile)
	}

	destDir := filepath.Dir(absFilePath)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}
	}

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		log.Fatalf("No destination directory %s", destDir)
	}

	err = os.WriteFile(absFilePath, []byte(contents), 0644)
	if err != nil {
		log.Fatalf("error: failed to write file: %v", err)
	}

	fmt.Printf("Successfully wrote object to %s\n", path)

	return nil
}

func getKubeConfig() (*rest.Config, error) {
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// Try in-cluster config first
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
