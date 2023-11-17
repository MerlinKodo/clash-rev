package cli

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"

	"github.com/MerlinKodo/clash-rev/config"
	C "github.com/MerlinKodo/clash-rev/constant"
	"github.com/MerlinKodo/clash-rev/hub"
	"github.com/MerlinKodo/clash-rev/hub/executor"
	"github.com/MerlinKodo/clash-rev/log"

	"github.com/spf13/cobra"
)

func newAppConfig() *AppConfig {
	return &AppConfig{
		homeDir:            os.Getenv("CLASH_HOME_DIR"),
		configFile:         os.Getenv("CLASH_CONFIG_FILE"),
		configUrl:          os.Getenv("CLASH_CONFIG_URL"),
		configUrlHeader:    os.Getenv("CLASH_CONFIG_URL_HEADER"),
		externalUI:         os.Getenv("CLASH_OVERRIDE_EXTERNAL_UI_DIR"),
		externalController: os.Getenv("CLASH_OVERRIDE_EXTERNAL_CONTROLLER"),
		secret:             os.Getenv("CLASH_OVERRIDE_SECRET"),
	}
}

func NewApp() *App {
	app := &App{
		Config: newAppConfig(),
	}
	app.setupRootCmd()
	return app
}

func (a *App) Run() error {
	return a.RootCmd.Execute()
}

func (a *App) setupRootCmd() {
	a.RootCmd = &cobra.Command{
		Use:   "clash",
		Short: "A rule-based tunnel in Go.",
		Long:  `Clash Rev is a rule-based tunnel in Go. Check out the project home page for more information: https://merlinkodo.github.io/Clash-Rev-Doc/`,
		Run:   a.execute,
	}
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.homeDir, "dir", "d", a.Config.homeDir, "Specify configuration directory, env: CLASH_HOME_DIR")
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.configFile, "config", "f", a.Config.configFile, "Specify configuration file, env: CLASH_CONFIG_FILE")
	a.RootCmd.PersistentFlags().StringVar(&a.Config.configUrl, "cfg-url", a.Config.configUrl, "Specify configuration file url, env: CLASH_CONFIG_URL")
	a.RootCmd.PersistentFlags().StringVar(&a.Config.configUrlHeader, "cfg-header", a.Config.configUrlHeader, "Specify configuration file url header, env: CLASH_CONFIG_URL_HEADER")
	a.RootCmd.PersistentFlags().StringVar(&a.Config.externalUI, "ext-ui", a.Config.externalUI, "Override external ui directory, env: CLASH_OVERRIDE_EXTERNAL_UI_DIR")
	a.RootCmd.PersistentFlags().StringVar(&a.Config.externalController, "ext-ctl", a.Config.externalController, "Override external controller address, env: CLASH_OVERRIDE_EXTERNAL_CONTROLLER")
	a.RootCmd.PersistentFlags().StringVar(&a.Config.secret, "secret", a.Config.secret, "override secret, env: CLASH_OVERRIDE_SECRET")
	a.RootCmd.PersistentFlags().BoolVarP(&a.Config.geodataMode, "geodata", "m", false, "Set geodata mode")
	a.RootCmd.PersistentFlags().BoolVarP(&a.Config.version, "version", "v", false, "Print current version of clash")
	a.RootCmd.PersistentFlags().BoolVarP(&a.Config.testConfig, "test", "t", false, "Test configuration and exit")
}

func (a *App) execute(cmd *cobra.Command, args []string) {
	setupMaxProcs()

	if a.Config.version {
		a.printVersion()
		return
	}
	if a.Config.homeDir != "" {
		a.Config.homeDir = resolvePath(a.Config.homeDir)
		C.SetHomeDir(a.Config.homeDir)
	}

	a.Config.configFile = a.resolveConfigFile()
	C.SetConfig(a.Config.configFile)

	if a.Config.geodataMode {
		C.GeodataMode = true
	}

	if err := config.Init(C.Path.HomeDir()); err != nil {
		log.Fatalln("Initial configuration directory error: %s", err.Error())
	}

	if a.Config.testConfig {
		a.testConfiguration()
		return
	}

	options := a.parseOptions()
	if err := hub.Parse(options...); err != nil {
		log.Fatalln("Parse config error: %s", err.Error())
	}

	defer executor.Shutdown()

	a.handleSignals()
	fmt.Println("Clash Rev is running now, press Ctrl+C to exit.")
	select {}
}

func (a *App) printVersion() {
	versionString := "Clash Rev Version: " + C.Version + "\n\n"
	versionString += "OS: " + runtime.GOOS + "\n" + "Architecture: " + runtime.GOARCH + "\n" + "Go Version: " + runtime.Version() + "\n" + "Build Time: " + C.BuildTime + "\n"

	var tags string
	var revision string

	debugInfo, loaded := debug.ReadBuildInfo()
	if loaded {
		for _, setting := range debugInfo.Settings {
			switch setting.Key {
			case "-tags":
				tags = setting.Value
			case "vcs.revision":
				revision = setting.Value
			}
		}
	}
	if tags != "" {
		versionString += "Tags: " + tags + "\n"
	}
	if revision != "" {
		versionString += "Revision: " + revision + "\n"
	}

	if C.CGO_ENABLED {
		versionString += "CGO Enabled: Yes\n"
	} else {
		versionString += "CGO Enabled: No\n"
	}

	fmt.Println(versionString)
}

func (a *App) resolveConfigFile() string {
	if a.Config.configFile != "" {
		return resolvePath(a.Config.configFile)
	} else if a.Config.configUrl != "" {
		log.Infoln("Downloading configuration file from %s", a.Config.configUrl)
		header := parseHeader(a.Config.configUrlHeader)
		configPath := filepath.Join(a.Config.homeDir, "config.yaml")
		if err := downloadFile(a.Config.configUrl, configPath, header); err != nil {
			log.Fatalln("Download configuration file error: %s", err.Error())
		}
		return configPath
	} else {
		return filepath.Join(C.Path.HomeDir(), C.Path.Config())
	}
}

func (a *App) testConfiguration() {
	if _, err := executor.Parse(); err != nil {
		log.Errorln(err.Error())
		fmt.Printf("configuration file %s test failed\n", C.Path.Config())
		os.Exit(1)
	}
	fmt.Printf("configuration file %s test is successful\n", C.Path.Config())
}

func (a *App) parseOptions() []hub.Option {
	var options []hub.Option
	if a.Config.externalUI != "" {
		options = append(options, hub.WithExternalUI(a.Config.externalUI))
	}
	if a.Config.externalController != "" {
		options = append(options, hub.WithExternalController(a.Config.externalController))
	}
	if a.Config.secret != "" {
		options = append(options, hub.WithSecret(a.Config.secret))
	}
	return options
}

func (a *App) handleSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for sig := range sigs {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Infoln("Received SIGINT or SIGTERM. Exiting gracefully...")
				os.Exit(0)
			case syscall.SIGHUP:
				log.Infoln("Received SIGHUP. Reloading...")
				if cfg, err := executor.ParseWithPath(C.Path.Config()); err == nil {
					executor.ApplyConfig(cfg, true)
				} else {
					log.Errorln("Parse config error: %s", err.Error())
				}
			}
		}
	}()
}
