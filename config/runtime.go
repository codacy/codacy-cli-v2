package config

import (
	"fmt"
)

// Note that this is only used by tools
type Runtime struct {
	name    string
	version string
	info    map[string]string
}

func (r *Runtime) Name() string {
	return r.name
}

func (r *Runtime) Version() string {
	return r.version
}

func (r *Runtime) FullName() string {
	return fmt.Sprintf("%s-%s", r.name, r.version)
}

func (r *Runtime) Info() map[string]string {
	return r.info
}

// populateInfo populates the runtime info
func (r *Runtime) populateInfo() {
	switch r.Name() {
	case "eslint":
		r.info = genInfoEslint(r)
	case "trivy":
		r.info = genInfoTrivy(r)
	}
}

// genInfoTrivy generates the info map for Trivy
func genInfoTrivy(r *Runtime) map[string]string {
	return map[string]string{
		"name":        r.name,
		"version":     r.version,
		"description": "Container and Filesystem Vulnerability Scanner",
		"binary":      "trivy",
	}
}

func NewRuntime(name string, version string) *Runtime {
	r := Runtime{
		name:    name,
		version: version,
	}
	r.populateInfo()
	return &r
}
