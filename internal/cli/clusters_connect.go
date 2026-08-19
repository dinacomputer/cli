package cli

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var clustersConnectSetCurrent bool

var clustersConnectCmd = &cobra.Command{
	Use:   "connect <cluster>",
	Short: "Add a kubectl context for a cluster",
	Long:  "Fetches a cluster's connection details and writes a kubectl context to your kubeconfig. The context authenticates through the Dina CLI: kubectl invokes `dina clusters credentials` to mint a short-lived token on each call, the same way the gcloud GKE auth plugin works.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		client, err := api.NewClient()
		if err != nil {
			return err
		}
		Infof("Fetching %s...\n", name)
		cluster, err := client.GetCluster(name)
		if err != nil {
			return err
		}
		if cluster.Endpoint == "" {
			return fmt.Errorf("cluster %q is missing an endpoint — it may not be ready yet", name)
		}

		// Dina's proxy endpoint is served with a public TLS cert, so no CA is
		// pinned in the normal case. If the server ever returns one (base64 PEM),
		// decode it; clientcmd re-encodes CertificateAuthorityData on write. If
		// it isn't valid base64, assume the server already sent raw PEM.
		var caPEM []byte
		if cluster.CACertificate != "" {
			decoded, err := base64.StdEncoding.DecodeString(cluster.CACertificate)
			if err != nil {
				decoded = []byte(cluster.CACertificate)
			}
			caPEM = decoded
		}

		// Point kubectl at this same binary so the exec plugin resolves even if
		// `dina` isn't on kubectl's PATH.
		self, err := os.Executable()
		if err != nil || self == "" {
			self = "dina"
		}

		key := "dina-" + name

		po := clientcmd.NewDefaultPathOptions()
		cfg, err := po.GetStartingConfig()
		if err != nil {
			return fmt.Errorf("reading kubeconfig: %w", err)
		}

		cfg.Clusters[key] = &clientcmdapi.Cluster{
			Server:                   cluster.Endpoint,
			CertificateAuthorityData: caPEM,
		}
		cfg.AuthInfos[key] = &clientcmdapi.AuthInfo{
			Exec: &clientcmdapi.ExecConfig{
				APIVersion:         "client.authentication.k8s.io/v1",
				Command:            self,
				Args:               []string{"clusters", "credentials", name},
				InstallHint:        "The Dina CLI is required to authenticate to this cluster. Install it from https://dina.sh.",
				InteractiveMode:    clientcmdapi.IfAvailableExecInteractiveMode,
				ProvideClusterInfo: false,
			},
		}
		cfg.Contexts[key] = &clientcmdapi.Context{
			Cluster:  key,
			AuthInfo: key,
		}
		if clustersConnectSetCurrent {
			cfg.CurrentContext = key
		}

		if err := clientcmd.ModifyConfig(po, *cfg, true); err != nil {
			return fmt.Errorf("writing kubeconfig: %w", err)
		}

		fmt.Printf("Added kubectl context %q.\n", key)
		if clustersConnectSetCurrent {
			fmt.Printf("Switched current context to %q.\n", key)
		} else {
			fmt.Printf("Select it with: kubectl config use-context %s\n", key)
		}
		return nil
	},
}

func init() {
	clustersConnectCmd.Flags().BoolVar(&clustersConnectSetCurrent, "set-current", true, "Set the new context as kubectl's current context")
	clustersCmd.AddCommand(clustersConnectCmd)
}
