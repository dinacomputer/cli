package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/dinacomputer/cli/internal/archive"
	"github.com/dinacomputer/cli/internal/color"
	"github.com/dinacomputer/cli/internal/term"
	"github.com/spf13/cobra"
)

var (
	deployApp       string
	deployTag       string
	deployPort      int
	deployReplicas  int
	deployBuildArgs []string
	deployWait      bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application",
	Long:  "Deploy an application to the Dina platform. If --tag is not provided, the current directory is zipped and uploaded as source.",
	Example: `  # deploy from the current directory's source
  dina deploy -a my-app

  # deploy a pre-built image
  dina deploy -a my-app --tag registry.example.com/my-app:v1.2

  # deploy with more replicas and block until the deployment is running
  dina deploy -a my-app --replicas 3 --wait

  # deploy with build-time arguments
  dina deploy -a my-app --build-arg NODE_ENV=production --build-arg COMMIT_SHA=$(git rev-parse HEAD)`,
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

		buildArgs, err := parseBuildArgs(deployBuildArgs)
		if err != nil {
			return err
		}

		var dep *api.Deployment

		if deployTag != "" {
			// Deploy with a pre-built image.
			input := api.DeployInput{Image: deployTag}
			if port != nil {
				input.Port = port
			}
			if replicas != nil {
				input.Replicas = replicas
			}
			if len(buildArgs) > 0 {
				input.BuildArgs = buildArgs
			}
			dep, err = client.Deploy(deployApp, input)
			if err != nil {
				return err
			}
		} else {
			// No tag — zip and upload source code.
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			Infoln("Packaging source code...")
			zipData, err := archive.ZipDir(cwd)
			if err != nil {
				return fmt.Errorf("zipping source: %w", err)
			}
			Infof("Uploading archive (%.1f KB)...\n", float64(len(zipData))/1024)

			dep, err = client.DeploySource(deployApp, zipData, buildArgs)
			if err != nil {
				return err
			}
		}

		printDeployment(dep)

		if deployWait {
			return waitForDeployment(client, deployApp, dep.ID)
		}
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

func waitForDeployment(client *api.Client, appName, deploymentID string) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	timeout := time.After(10 * time.Minute)

	start := time.Now()
	lastStatus := ""
	animated := term.IsStderrTTY() && !Quiet()

	for {
		select {
		case <-timeout:
			if animated {
				Infoln()
			}
			return fmt.Errorf("timed out waiting for deployment %s", deploymentID)
		case <-ticker.C:
			dep, err := client.GetDeployment(appName, deploymentID)
			if err != nil {
				if animated {
					Infoln()
				}
				return err
			}
			elapsed := time.Since(start).Truncate(time.Second)
			switch dep.Status {
			case "running":
				runningTag := color.Green("running")
				if animated {
					Infof("\rDeployment is %s (%s)          \n", runningTag, elapsed)
				} else {
					Infof("Deployment is %s (%s)\n", runningTag, elapsed)
				}
				return nil
			case "failed", "build_failed", "deploy_failed":
				failedTag := color.Red(dep.Status)
				if animated {
					Infof("\rDeployment failed: %s (%s)          \n", failedTag, elapsed)
				} else {
					Infof("Deployment failed: %s (%s)\n", failedTag, elapsed)
				}
				return fmt.Errorf("deployment %s failed with status: %s", deploymentID, dep.Status)
			default:
				if dep.Status != lastStatus {
					if animated && lastStatus != "" {
						Infoln()
					}
					if !animated {
						Infof("  %s... %s\n", dep.Status, elapsed)
					}
					lastStatus = dep.Status
				}
				if animated {
					Infof("\r  %s... %s", dep.Status, elapsed)
				}
			}
		}
	}
}

func parseBuildArgs(raw []string) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	m := make(map[string]string, len(raw))
	for _, arg := range raw {
		k, v, ok := strings.Cut(arg, "=")
		if !ok {
			return nil, fmt.Errorf("invalid build arg %q: must be KEY=VALUE", arg)
		}
		m[k] = v
	}
	return m, nil
}

func init() {
	deployCmd.Flags().StringVarP(&deployApp, "app", "a", "", "Application name")
	deployCmd.Flags().StringVarP(&deployTag, "tag", "t", "", "Pre-built image tag (if omitted, source is uploaded)")
	deployCmd.Flags().IntVarP(&deployPort, "port", "p", 8080, "Container port the app listens on")
	deployCmd.Flags().IntVar(&deployReplicas, "replicas", 1, "Number of replicas")
	deployCmd.Flags().StringArrayVar(&deployBuildArgs, "build-arg", nil, "Build argument in KEY=VALUE format (can be repeated)")
	deployCmd.Flags().BoolVarP(&deployWait, "wait", "w", false, "Wait for deployment to complete")
	deployCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(deployCmd)
}
