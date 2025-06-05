package cmdutils

import (
	"codacy/cli-v2/domain"

	"github.com/spf13/cobra"
)

// AddCloudFlags adds the common cloud-related flags to a cobra command.
// The flags will be bound to the provided flags struct.
func AddCloudFlags(cmd *cobra.Command, flags *domain.InitFlags) {
	cmd.Flags().StringVar(&flags.ApiToken, "api-token", "", "Optional Codacy API token. If defined, configurations will be fetched from Codacy")
	cmd.Flags().StringVar(&flags.Provider, "provider", "", "Provider (e.g., gh, bb, gl) to fetch configurations from Codacy. Required when api-token is provided")
	cmd.Flags().StringVar(&flags.Organization, "organization", "", "Remote organization name to fetch configurations from Codacy. Required when api-token is provided")
	cmd.Flags().StringVar(&flags.Repository, "repository", "", "Remote repository name to fetch configurations from Codacy. Required when api-token is provided")
}
