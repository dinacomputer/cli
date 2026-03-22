package cli

import (
	"fmt"
	"os"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/dinacomputer/cli/internal/archive"
	"github.com/spf13/cobra"
)

var (
	deployApp      string
	deployTag      string
	deployPort     int
	deployReplicas int
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application",
	Long:  "Deploy an application to the Dina platform. If --tag is not provided, the current directory is zipped and uploaded as source.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		var port *int
		if cmd.Flags().Changed("port") {
			port = &deployPort
		}
		var replicas *int
		if cmd.Flags().Changed("replicas") {
			replicas = &deployReplicas
		}

		if deployTag != "" {
			// Deploy with a pre-built image.
			input := api.DeployInput{Image: deployTag}
			if port != nil {
				input.Port = port
			}
			if replicas != nil {
				input.Replicas = replicas
			}
			dep, err := client.Deploy(deployApp, input)
			if err != nil {
				return err
			}
			printDeployment(dep)
			return nil
		}

		// No tag — zip and upload source code.
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		fmt.Println("Packaging source code...")
		zipData, err := archive.ZipDir(cwd)
		if err != nil {
			return fmt.Errorf("zipping source: %w", err)
		}
		fmt.Printf("Uploading archive (%.1f KB)...\n", float64(len(zipData))/1024)

		dep, err := client.DeploySource(deployApp, zipData)
		if err != nil {
			return err
		}
		printDeployment(dep)
		return nil
	},
}

func printDeployment(dep *api.Deployment) {
	fmt.Printf("Deployment %s created\n", dep.ID)
	fmt.Printf("  Status:  %s\n", dep.Status)
	fmt.Printf("  Source:  %s (%s)\n", dep.SourceRef, dep.SourceType)
	if dep.Image != "" {
		fmt.Printf("  Image:   %s\n", dep.Image)
	}
	fmt.Printf("  Port:    %d\n", dep.Port)
	fmt.Printf("  Replicas: %d\n", dep.Replicas)
}

func init() {
	deployCmd.Flags().StringVarP(&deployApp, "app", "a", "", "Application name")
	deployCmd.Flags().StringVarP(&deployTag, "tag", "t", "", "Pre-built image tag (if omitted, source is uploaded)")
	deployCmd.Flags().IntVarP(&deployPort, "port", "p", 8080, "Container port the app listens on")
	deployCmd.Flags().IntVar(&deployReplicas, "replicas", 1, "Number of replicas")
	deployCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(deployCmd)
}
