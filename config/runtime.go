package config

import "fmt"

type Runtime struct {
	name string
	version string
	tools []ConfigTool
}

func (r *Runtime) Name() string {
	return r.name
}

func (r *Runtime) Version() string {
	return r.version
}

func (r *Runtime) Tools() []ConfigTool {
	return r.tools
}

func (r *Runtime) AddTool(tool *ConfigTool) {
	r.tools = append(r.tools, *tool)
}

func (r *Runtime) FullName() string {
	return fmt.Sprintf("%s-%s", r.name, r.version)
}

func (r *Runtime) GetTool(name string) *ConfigTool  {
	// TODO might become a map
	for _, tool := range r.tools {
		if tool.name == name {
			return &tool
		}
	}
	return nil
}