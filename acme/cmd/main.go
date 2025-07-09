package main

import (
	"encoding/json"
	"fmt"
	"github.com/openshift-online/bootstrap/acme/cmd/clusters"
	clusterTypes "github.com/openshift-online/bootstrap/acme/pkg/clusters"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfig   string
	argocdServer string
	argocdToken  string
	namespace    string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "k8s-argocd-manager",
		Short: "A tool to manage Kubernetes and ArgoCD resources",
		Long:  "A comprehensive tool for managing Kubernetes deployments and ArgoCD applications",
	}

	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&argocdServer, "argocd-server", "localhost:8080", "ArgoCD server address")
	rootCmd.PersistentFlags().StringVar(&argocdToken, "argocd-token", "", "ArgoCD authentication token")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace")

	// Add subcommands
	//rootCmd.AddCommand(createAppCmd())
	//rootCmd.AddCommand(listAppsCmd())
	//rootCmd.AddCommand(syncAppCmd())
	//rootCmd.AddCommand(statusCmd())
	//rootCmd.AddCommand(deleteAppCmd())

	rootCmd.AddCommand(NewClustersCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func NewClustersCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "clusters",
		Short: "clusters",
		Run:   runClusters,
	}
	return cmd
}

func runClusters(cmd *cobra.Command, args []string) {

	desiredClusters := clusters.GetRegions()

	for _, spec := range desiredClusters {

		cd := clusterTypes.NewClusterDeployment(spec)
		ic := clusterTypes.NewInstallConfig(spec)
		mp := clusterTypes.NewMachinePool(spec)
		mc := clusterTypes.NewManagedCluster(spec)

		//d := "/home/mturansk/projects/src/github.com/openshift-online/bootstrap/clusters/overlay/cluster-04/clusterdeployment.json"

		if err := WriteFile("./clusters/overlay/"+spec.Name+"/clusterdeployment.json", toJSON(cd)); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}

		if err := WriteFile("./clusters/overlay/"+spec.Name+"/installconfig.json", toJSON(ic)); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}

		if err := WriteFile("./clusters/overlay/"+spec.Name+"/machinepool.json", toJSON(mp)); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}

		if err := WriteFile("./clusters/overlay/"+spec.Name+"/managedcluster.json", toJSON(mc)); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}

		//fmt.Println(toJSON(cd))
		//fmt.Println(toJSON(ic))
		//fmt.Println(toJSON(mp))
		//fmt.Println(toJSON(mc))

	}
}

func toJSON(obj interface{}) string {
	j, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(j)
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

// Read the contents of file into string value
func ReadFileValueString(file string, val *string) error {
	fileContents, err := ReadFile(file)
	if err != nil {
		return err
	}

	*val = strings.TrimSuffix(fileContents, "\n")
	return err
}

// Return project root path based on the relative path of this file
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
