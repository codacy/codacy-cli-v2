package domain

// InitFlags represents the flags for the init command
type InitFlags struct {
	ApiToken     string
	Provider     string
	Organization string
	Repository   string
}
