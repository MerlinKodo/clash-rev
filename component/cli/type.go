package cli

import (
	"github.com/spf13/cobra"
)

type AppConfig struct {
	homeDir            string
	configFile         string
	configUrl          string
	configUrlHeader    string
	externalUI         string
	externalController string
	secret             string
	geodataMode        bool
	version            bool
	testConfig         bool
}

type App struct {
	RootCmd *cobra.Command
	Config  *AppConfig
}
