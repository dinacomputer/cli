package cli

import (
	"github.com/dinacomputer/cli/internal/skills"
	"github.com/spf13/cobra"
)

var installSkills bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install resources",
	Long:  "Install resources for the Dina CLI. Use --skills to install agent skill files that teach AI tools (Claude Code, Cursor, VS Code Copilot, Gemini, OpenAI Codex) how to use the Dina CLI.",
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
