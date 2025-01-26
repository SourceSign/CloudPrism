package common

import (
	"regexp"
	"strings"
)

type ApplicationEnvironment uint

const (
	AppEnvSandbox ApplicationEnvironment = iota
	AppEnvDevelopment
	AppEnvIntegration
	AppEnvProduction
)

func (s ApplicationEnvironment) Name() string {
	switch s {
	case AppEnvSandbox:
		return "Sandbox"
	case AppEnvDevelopment:
		return "Development"
	case AppEnvIntegration:
		return "Integration"
	case AppEnvProduction:
		return "Production"
	}

	return "Unknown"
}

func (s ApplicationEnvironment) ID() string {
	switch s {
	case AppEnvSandbox:
		return "sbx"
	case AppEnvDevelopment:
		return "dev"
	case AppEnvIntegration:
		return "int"
	case AppEnvProduction:
		return "prd"
	}

	return "etc"
}

type Application string

func (a Application) ID() string {
	id := strings.ToLower(strings.ReplaceAll(string(a), " ", "-"))
	reg := regexp.MustCompile("[^a-zA-Z0-9-]+")

	return reg.ReplaceAllString(id, "")
}

type StateStoreNameType string

func StateStoreName(app Application, env ApplicationEnvironment) StateStoreNameType {
	return StateStoreNameType(app.ID() + "-" + env.ID())
}
