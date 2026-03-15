package skills

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
)

const skillDirName = "dina-cli"

// Install runs the interactive skill installation flow.
func Install() error {
	// 1. Pick an agent
	var agentSlug string
	options := make([]huh.Option[string], len(SupportedAgents))
	for i, a := range SupportedAgents {
		options[i] = huh.NewOption(a.Name, a.Slug)
	}

	err := huh.NewSelect[string]().
		Title("Which AI agent do you want to install the Dina skill for?").
		Options(options...).
		Value(&agentSlug).
		Run()
	if err != nil {
		return err
	}

	var agent Agent
	for _, a := range SupportedAgents {
		if a.Slug == agentSlug {
			agent = a
			break
		}
	}

	// 2. Pick scope
	var scope string
	err = huh.NewSelect[string]().
		Title("Install scope?").
		Options(
			huh.NewOption("Project (current directory)", "project"),
			huh.NewOption("User (global, all projects)", "user"),
		).
		Value(&scope).
		Run()
	if err != nil {
		return err
	}

	// 3. Resolve target directory
	var baseDir string
	switch scope {
	case "project":
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		baseDir = filepath.Join(cwd, agent.ProjectDir)
	case "user":
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}
		baseDir = filepath.Join(home, agent.UserDir)
	}

	skillDir := filepath.Join(baseDir, skillDirName)
	skillFile := filepath.Join(skillDir, "SKILL.md")

	// 4. Check if already installed
	if _, err := os.Stat(skillFile); err == nil {
		var overwrite bool
		err = huh.NewConfirm().
			Title("Skill already installed at " + skillFile + ". Overwrite?").
			Value(&overwrite).
			Run()
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Println("Installation cancelled.")
			return nil
		}
	}

	// 5. Write the skill
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return fmt.Errorf("creating skill directory: %w", err)
	}

	if err := os.WriteFile(skillFile, []byte(SkillMD()), 0o644); err != nil {
		return fmt.Errorf("writing SKILL.md: %w", err)
	}

	fmt.Printf("\nInstalled Dina CLI skill for %s at:\n  %s\n", agent.Name, skillFile)
	return nil
}
