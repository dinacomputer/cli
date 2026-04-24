package cli

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dinacomputer/cli/internal/api"
	"github.com/dinacomputer/cli/internal/term"
	"github.com/spf13/cobra"
)

const signalsProduct = "dina-cli"

var (
	bugTitle       string
	bugDescription string
	bugSeverity    string
	bugContext     string
	bugContextFile string
)

var bugCmd = &cobra.Command{
	Use:   "bug",
	Short: "Submit a bug report",
	Long:  "Submit a bug report to the Sokkel Signals API. Authentication is optional — if you are logged in, the submitter identity is recorded.",
	Example: `  # interactive — prompts for title, description, severity
  dina feedback bug

  # non-interactive with context from a file
  dina feedback bug \
    --title "deploy --wait hangs on failed builds" \
    --description "Expected exit 1 on build_failed; the process stalled for 10m" \
    --severity high \
    --context-file ./build.log`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := loadContext(bugContext, bugContextFile)
		if err != nil {
			return err
		}

		title := bugTitle
		desc := bugDescription
		severity := strings.ToLower(bugSeverity)

		if title == "" || desc == "" {
			if err := requireInteractive("bug", "--title", "--description"); err != nil {
				return err
			}
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Title").
						Description("Short, human summary of the bug").
						Value(&title).
						Validate(requireNonEmpty("title")),
					huh.NewText().
						Title("Description").
						Description("What happened, what was expected, repro steps").
						Value(&desc).
						Validate(requireNonEmpty("description")),
					huh.NewSelect[string]().
						Title("Severity").
						Options(
							huh.NewOption("(unspecified)", ""),
							huh.NewOption("low", "low"),
							huh.NewOption("medium", "medium"),
							huh.NewOption("high", "high"),
						).
						Value(&severity),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
		}

		if severity != "" && severity != "low" && severity != "medium" && severity != "high" {
			return fmt.Errorf("invalid severity %q: must be low, medium, or high", severity)
		}

		client := api.NewAnonymousClient()
		resp, err := client.SubmitBug(api.BugBody{
			Product:     signalsProduct,
			Title:       title,
			Description: desc,
			Severity:    severity,
			OS:          fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			Version:     Version,
			Context:     ctx,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Bug report submitted (id: %s)\n", resp.ID)
		return nil
	},
}

var (
	featureTitle       string
	featureDescription string
)

var featureCmd = &cobra.Command{
	Use:     "feature",
	Short:   "Submit a feature request",
	Long:    "Submit a feature request to the Sokkel Signals API.",
	Example: `  dina feedback feature --title "Support PATCH on hostnames" --description "..."`,
	RunE: func(cmd *cobra.Command, args []string) error {
		title := featureTitle
		desc := featureDescription

		if title == "" || desc == "" {
			if err := requireInteractive("feature request", "--title", "--description"); err != nil {
				return err
			}
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Title").
						Description("Short summary of the feature request").
						Value(&title).
						Validate(requireNonEmpty("title")),
					huh.NewText().
						Title("Description").
						Description("What would you like to see, and why?").
						Value(&desc).
						Validate(requireNonEmpty("description")),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
		}

		client, err := api.NewClient()
		if err != nil {
			return err
		}
		resp, err := client.SubmitFeatureRequest(api.FeatureRequestBody{
			Product:     signalsProduct,
			Title:       title,
			Description: desc,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Feature request submitted (id: %s)\n", resp.ID)
		return nil
	},
}

var (
	feedbackMessage string
	feedbackRating  int
)

var feedbackCmd = &cobra.Command{
	Use:   "feedback",
	Short: "Submit general feedback",
	Long: `Submit general feedback to the Sokkel Signals API.

For bug reports, use ` + "`dina feedback bug`" + `. For feature requests, use
` + "`dina feedback feature`" + `. Without a subcommand, this command submits an
open-ended message with an optional 1-5 rating.`,
	Example: `  dina feedback --message "The new error messages are much clearer, thanks!" --rating 5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		message := feedbackMessage
		ratingStr := ""
		if cmd.Flags().Changed("rating") {
			ratingStr = strconv.Itoa(feedbackRating)
		}

		if message == "" {
			if err := requireInteractive("feedback", "--message"); err != nil {
				return err
			}
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewText().
						Title("Message").
						Description("What's on your mind?").
						Value(&message).
						Validate(requireNonEmpty("message")),
					huh.NewSelect[string]().
						Title("Rating (optional)").
						Options(
							huh.NewOption("(skip)", ""),
							huh.NewOption("1 — poor", "1"),
							huh.NewOption("2", "2"),
							huh.NewOption("3", "3"),
							huh.NewOption("4", "4"),
							huh.NewOption("5 — excellent", "5"),
						).
						Value(&ratingStr),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
		}

		rating := 0
		if ratingStr != "" {
			r, err := strconv.Atoi(ratingStr)
			if err != nil || r < 1 || r > 5 {
				return fmt.Errorf("invalid rating %q: must be 1-5", ratingStr)
			}
			rating = r
		}

		client, err := api.NewClient()
		if err != nil {
			return err
		}
		resp, err := client.SubmitFeedback(api.FeedbackBody{
			Product: signalsProduct,
			Message: message,
			Rating:  rating,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Feedback submitted (id: %s)\n", resp.ID)
		return nil
	},
}

func loadContext(inline, path string) (string, error) {
	if inline != "" && path != "" {
		return "", fmt.Errorf("--context and --context-file are mutually exclusive")
	}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("reading context file: %w", err)
		}
		return string(data), nil
	}
	return inline, nil
}

func requireNonEmpty(field string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}

// requireInteractive ensures stdin is a TTY and --no-input is not set before
// prompting. Returns a clear error listing the flags the user should pass
// instead if not.
func requireInteractive(what string, flags ...string) error {
	if noInput {
		return fmt.Errorf("--no-input was set but %s input is missing; pass %s", what, strings.Join(flags, " and "))
	}
	if !term.IsStdinTTY() {
		return fmt.Errorf("stdin is not a terminal; pass %s to submit %s non-interactively", strings.Join(flags, " and "), what)
	}
	return nil
}

func init() {
	bugCmd.Flags().StringVarP(&bugTitle, "title", "t", "", "Short summary of the bug")
	bugCmd.Flags().StringVarP(&bugDescription, "description", "d", "", "What happened, what was expected, repro steps")
	bugCmd.Flags().StringVarP(&bugSeverity, "severity", "s", "", "Severity: low, medium, or high")
	bugCmd.Flags().StringVar(&bugContext, "context", "", "Optional context (logs, stack trace, env)")
	bugCmd.Flags().StringVar(&bugContextFile, "context-file", "", "Path to a file whose contents will be submitted as context")

	featureCmd.Flags().StringVarP(&featureTitle, "title", "t", "", "Short summary of the feature request")
	featureCmd.Flags().StringVarP(&featureDescription, "description", "d", "", "Full description of the feature request")

	feedbackCmd.Flags().StringVarP(&feedbackMessage, "message", "m", "", "Feedback message")
	feedbackCmd.Flags().IntVarP(&feedbackRating, "rating", "r", 0, "Optional 1-5 rating")

	feedbackCmd.AddCommand(bugCmd)
	feedbackCmd.AddCommand(featureCmd)
	rootCmd.AddCommand(feedbackCmd)
}
