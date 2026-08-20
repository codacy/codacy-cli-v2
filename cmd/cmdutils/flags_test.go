package cmdutils

import (
	"io"
	"testing"

	"codacy/cli-v2/domain"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func completeFlags() domain.InitFlags {
	return domain.InitFlags{Provider: "gh", Organization: "org", Repository: "repo"}
}

func TestAddCloudFlags(t *testing.T) {
	t.Setenv(ProjectTokenEnvVar, "token-from-env")

	cmd := &cobra.Command{Use: "test"}
	flags := domain.InitFlags{}
	AddCloudFlags(cmd, &flags)

	for _, name := range []string{"api-token", "project-token", "provider", "organization", "repository"} {
		assert.NotNil(t, cmd.Flags().Lookup(name), "flag %s should be registered", name)
	}

	// the environment must not leak in at registration time - see ResolveProjectToken
	assert.Equal(t, "", cmd.Flags().Lookup("project-token").DefValue)
	assert.Equal(t, domain.InitFlags{}, flags)
}

func TestAddCloudFlagsBindsValues(t *testing.T) {
	cmd := &cobra.Command{Use: "test", Run: func(*cobra.Command, []string) {}}
	flags := domain.InitFlags{}
	AddCloudFlags(cmd, &flags)

	cmd.SetArgs([]string{"--project-token", "tok", "--provider", "gh", "--organization", "org", "--repository", "repo"})
	assert.NoError(t, cmd.Execute())

	assert.Equal(t, domain.InitFlags{ProjectToken: "tok", Provider: "gh", Organization: "org", Repository: "repo"}, flags)
	assert.NoError(t, ValidateRemoteFlags(flags))
}

func TestResolveProjectToken(t *testing.T) {
	tests := []struct {
		name          string
		env           string
		flags         domain.InitFlags
		expectedToken string
	}{
		{"environment applies once coordinates are known", "env-token", completeFlags(), "env-token"},
		{"environment ignored without coordinates", "env-token", domain.InitFlags{}, ""},
		{"environment ignored with partial coordinates", "env-token", domain.InitFlags{Provider: "gh", Organization: "org"}, ""},
		{"explicit flag wins over environment", "env-token", domain.InitFlags{ProjectToken: "flag-token", Provider: "gh", Organization: "org", Repository: "repo"}, "flag-token"},
		{"account token is not overwritten", "env-token", domain.InitFlags{ApiToken: "api", Provider: "gh", Organization: "org", Repository: "repo"}, ""},
		{"unset environment leaves flags alone", "", completeFlags(), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(ProjectTokenEnvVar, tt.env)

			flags := tt.flags
			ResolveProjectToken(&flags)

			assert.Equal(t, tt.expectedToken, flags.ProjectToken)
		})
	}
}

func TestValidateRemoteFlags(t *testing.T) {
	tests := []struct {
		name        string
		flags       domain.InitFlags
		expectError bool
	}{
		{"no token needs no coordinates", domain.InitFlags{}, false},
		{"no token with partial coordinates", domain.InitFlags{Provider: "gh"}, false},
		{"account token with coordinates", domain.InitFlags{ApiToken: "api", Provider: "gh", Organization: "org", Repository: "repo"}, false},
		{"repository token with coordinates", domain.InitFlags{ProjectToken: "project", Provider: "gh", Organization: "org", Repository: "repo"}, false},
		{"repository token without coordinates", domain.InitFlags{ProjectToken: "project"}, true},
		{"account token without coordinates", domain.InitFlags{ApiToken: "api"}, true},
		{"token missing provider", domain.InitFlags{ProjectToken: "project", Organization: "org", Repository: "repo"}, true},
		{"token missing organization", domain.InitFlags{ProjectToken: "project", Provider: "gh", Repository: "repo"}, true},
		{"token missing repository", domain.InitFlags{ProjectToken: "project", Provider: "gh", Organization: "org"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRemoteFlags(tt.flags)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// stopped reports whether PrepareRemoteFlags terminated the command. The exit stub panics
// so a helper that kept running past it would fail the assertion instead of passing.
func stopped(t *testing.T, flags *domain.InitFlags) bool {
	t.Helper()

	original := exit
	defer func() { exit = original }()

	type exitCall struct{ code int }
	exit = func(code int) { panic(exitCall{code}) }

	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(io.Discard)

	didExit := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				call, ok := r.(exitCall)
				assert.True(t, ok, "unexpected panic: %v", r)
				assert.Equal(t, 1, call.code)
				didExit = true
			}
		}()
		PrepareRemoteFlags(cmd, flags)
	}()

	return didExit
}

func TestPrepareRemoteFlags(t *testing.T) {
	t.Run("stops on a token without coordinates", func(t *testing.T) {
		flags := domain.InitFlags{ProjectToken: "tok"}
		assert.True(t, stopped(t, &flags))
	})

	t.Run("continues without any token", func(t *testing.T) {
		flags := domain.InitFlags{}
		assert.False(t, stopped(t, &flags))
	})

	t.Run("continues when the environment token is ignored", func(t *testing.T) {
		t.Setenv(ProjectTokenEnvVar, "env-token")

		flags := domain.InitFlags{}
		assert.False(t, stopped(t, &flags))
		assert.Equal(t, "", flags.ProjectToken)
	})

	t.Run("adopts the environment token with coordinates", func(t *testing.T) {
		t.Setenv(ProjectTokenEnvVar, "env-token")

		flags := completeFlags()
		assert.False(t, stopped(t, &flags))
		assert.Equal(t, "env-token", flags.ProjectToken)
	})
}
