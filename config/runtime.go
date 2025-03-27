package config

import (
	"fmt"
)

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

func (r *Runtime) populateInfo() {
	switch r.Name() {
	case "node":
		r.info = genInfoNode(r)
	case "eslint":
		r.info = genInfoEslint(r)
	case "dart":
		r.info = genInfoDart(r)
	case "flutter":
		r.info = genInfoFlutter(r)
		/*case "dart":
		r.info = genInfoDart(r)*/
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
