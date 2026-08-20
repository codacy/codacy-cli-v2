package domain

// InitFlags represents the flags for the init command
type InitFlags struct {
	ApiToken     string
	ProjectToken string
	Provider     string
	Organization string
	Repository   string
}

// HasRemoteToken reports whether a token allowing remote configuration download was provided.
func (f InitFlags) HasRemoteToken() bool {
	return f.ApiToken != "" || f.ProjectToken != ""
}

// HasRepositoryCoordinates reports whether the repository every remote read is scoped to
// was fully identified.
func (f InitFlags) HasRepositoryCoordinates() bool {
	return f.Provider != "" && f.Organization != "" && f.Repository != ""
}
