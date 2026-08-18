package cmdutils

import (
	"errors"
	"fmt"
	"log"
	"os"

	"codacy/cli-v2/domain"

	"github.com/spf13/cobra"
)

// ProjectTokenEnvVar is also the codacy-coverage-reporter's variable, so it is often
// exported for unrelated reasons - see ResolveProjectToken.
const ProjectTokenEnvVar = "CODACY_PROJECT_TOKEN"

// AddCloudFlags adds the common cloud-related flags to a cobra command.
// The flags will be bound to the provided flags struct.
func AddCloudFlags(cmd *cobra.Command, flags *domain.InitFlags) {
	cmd.Flags().StringVar(&flags.ApiToken, "api-token", "", "Optional Codacy account API token. If defined, configurations will be fetched from Codacy")
	cmd.Flags().StringVar(&flags.ProjectToken, "project-token", "", "Optional Codacy repository token, an alternative to api-token for fetching configurations from Codacy. Falls back to CODACY_PROJECT_TOKEN when provider, organization and repository are given. See https://docs.codacy.com/codacy-api/api-tokens/#repository-api-tokens")
	cmd.Flags().StringVar(&flags.Provider, "provider", "", "Provider (e.g., gh, bb, gl) to fetch configurations from Codacy. Required when api-token or project-token is provided")
	cmd.Flags().StringVar(&flags.Organization, "organization", "", "Remote organization name to fetch configurations from Codacy. Required when api-token or project-token is provided")
	cmd.Flags().StringVar(&flags.Repository, "repository", "", "Remote repository name to fetch configurations from Codacy. Required when api-token or project-token is provided")
}

// ResolveProjectToken falls back to CODACY_PROJECT_TOKEN, but only once the repository
// coordinates are known. The variable is the coverage reporter's and is routinely exported
// CI-wide, so on its own it must not switch a local run into remote mode.
func ResolveProjectToken(flags *domain.InitFlags) {
	if flags.HasRemoteToken() || !flags.HasRepositoryCoordinates() {
		return
	}
	flags.ProjectToken = os.Getenv(ProjectTokenEnvVar)
}

// ValidateRemoteFlags rejects a token given without the repository coordinates every
// remote read needs, so commands fail with a clear message instead of requesting a URL
// with empty path segments.
func ValidateRemoteFlags(flags domain.InitFlags) error {
	if !flags.HasRemoteToken() || flags.HasRepositoryCoordinates() {
		return nil
	}
	return errors.New("when using --api-token or --project-token, you must also provide --provider, --organization, and --repository flags")
}

// exit is swappable so the failure path stays testable.
var exit = os.Exit

// PrepareRemoteFlags resolves the environment fallback and rejects an incomplete remote
// setup, so every command that takes cloud flags behaves the same way.
func PrepareRemoteFlags(cmd *cobra.Command, flags *domain.InitFlags) {
	ResolveProjectToken(flags)

	err := ValidateRemoteFlags(*flags)
	if err == nil {
		return
	}

	fmt.Printf("Error: %s.\n", err)
	fmt.Println("Please provide all required flags and try again.")
	fmt.Println()
	if errHelp := cmd.Help(); errHelp != nil {
		log.Printf("Warning: Failed to display command help: %v", errHelp)
	}
	exit(1)
}
