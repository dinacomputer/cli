package cli

import (
	"github.com/dinacomputer/cli/internal/skills"
	"github.com/spf13/cobra"
)

var installSkills bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		if installSkills {
			return skills.Install()
		}
		return cmd.Help()
	},
}

func init() {
	installCmd.Flags().BoolVar(&installSkills, "skills", false, "Install agent skills for the Dina CLI")
	rootCmd.AddCommand(installCmd)
}
