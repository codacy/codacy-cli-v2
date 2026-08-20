package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasRemoteToken(t *testing.T) {
	tests := []struct {
		name     string
		flags    InitFlags
		expected bool
	}{
		{"no token", InitFlags{}, false},
		{"account token", InitFlags{ApiToken: "api"}, true},
		{"repository token", InitFlags{ProjectToken: "project"}, true},
		{"both tokens", InitFlags{ApiToken: "api", ProjectToken: "project"}, true},
		{"coordinates without token", InitFlags{Provider: "gh", Organization: "org", Repository: "repo"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.flags.HasRemoteToken())
		})
	}
}

func TestHasRepositoryCoordinates(t *testing.T) {
	tests := []struct {
		name     string
		flags    InitFlags
		expected bool
	}{
		{"complete", InitFlags{Provider: "gh", Organization: "org", Repository: "repo"}, true},
		{"empty", InitFlags{}, false},
		{"missing provider", InitFlags{Organization: "org", Repository: "repo"}, false},
		{"missing organization", InitFlags{Provider: "gh", Repository: "repo"}, false},
		{"missing repository", InitFlags{Provider: "gh", Organization: "org"}, false},
		{"token alone is not coordinates", InitFlags{ProjectToken: "tok"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.flags.HasRepositoryCoordinates())
		})
	}
}
