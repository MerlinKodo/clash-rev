package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/MerlinKodo/clash-rev/config"
	C "github.com/MerlinKodo/clash-rev/constant"
	"github.com/MerlinKodo/clash-rev/constant/features"
	"github.com/MerlinKodo/clash-rev/hub"
	"github.com/MerlinKodo/clash-rev/hub/executor"
	"github.com/MerlinKodo/clash-rev/log"

	"go.uber.org/automaxprocs/maxprocs"
)

var (
	version            bool
	testConfig         bool
	geodataMode        bool
	homeDir            string
	configFile         string
	externalUI         string
	externalController string
	secret             string
)

func init() {
	flag.StringVar(&homeDir, "d", os.Getenv("CLASH_HOME_DIR"), "specify configuration directory, env: CLASH_HOME_DIR")
	flag.StringVar(&configFile, "f", os.Getenv("CLASH_CONFIG_FILE"), "specify configuration file, env: CLASH_CONFIG_FILE")
	flag.StringVar(&externalUI, "ext-ui", os.Getenv("CLASH_OVERRIDE_EXTERNAL_UI_DIR"), "override external ui directory, env: CLASH_OVERRIDE_EXTERNAL_UI_DIR")
	flag.StringVar(&externalController, "ext-ctl", os.Getenv("CLASH_OVERRIDE_EXTERNAL_CONTROLLER"), "override external controller address, env: CLASH_OVERRIDE_EXTERNAL_CONTROLLER")
	flag.StringVar(&secret, "secret", os.Getenv("CLASH_OVERRIDE_SECRET"), "override secret for RESTful API, env: CLASH_OVERRIDE_SECRET")
	flag.BoolVar(&geodataMode, "m", false, "set geodata mode")
	flag.BoolVar(&version, "v", false, "show current version of clash")
	flag.BoolVar(&testConfig, "t", false, "test configuration and exit")
	flag.Parse()
}

func main() {
	setupMaxProcs()
	if version {
		printVersion()
		return
	}

	if homeDir != "" {
		if !filepath.IsAbs(homeDir) {
			currentDir, _ := os.Getwd()
			homeDir = filepath.Join(currentDir, homeDir)
		}
		C.SetHomeDir(homeDir)
	}

	if configFile != "" {
		if !filepath.IsAbs(configFile) {
			currentDir, _ := os.Getwd()
			configFile = filepath.Join(currentDir, configFile)
		}
		C.SetConfig(configFile)
	} else {
		configFile := filepath.Join(C.Path.HomeDir(), C.Path.Config())
		C.SetConfig(configFile)
	}

	if geodataMode {
		C.GeodataMode = true
	}

	if err := config.Init(C.Path.HomeDir()); err != nil {
		log.Fatalln("Initial configuration directory error: %s", err.Error())
	}

	if testConfig {
		testConfiguration()
		return
	}

	options := parseOptions()
	if err := hub.Parse(options...); err != nil {
		log.Fatalln("Parse config error: %s", err.Error())
	}

	defer executor.Shutdown()

	handleSignals()
	fmt.Println("Clash Rev is running now, press Ctrl+C to exit.")
	select {}
}

func setupMaxProcs() {
	_, _ = maxprocs.Set(maxprocs.Logger(func(string, ...any) {}))
}

func printVersion() {
	fmt.Printf(
		"Clash Rev Version: %s\nOS: %s\nArchitecture: %s\nGo Version: %s\nBuild Time: %s\n",
		C.Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), C.BuildTime)
	if len(features.TAGS) != 0 {
		fmt.Printf("Use tags: %s\n", strings.Join(features.TAGS, ", "))
	}
}

func testConfiguration() {
	if _, err := executor.Parse(); err != nil {
		log.Errorln(err.Error())
		fmt.Printf("configuration file %s test failed\n", C.Path.Config())
		os.Exit(1)
	}
	fmt.Printf("configuration file %s test is successful\n", C.Path.Config())
}

func parseOptions() []hub.Option {
	var options []hub.Option
	if externalUI != "" {
		options = append(options, hub.WithExternalUI(externalUI))
	}
	if externalController != "" {
		options = append(options, hub.WithExternalController(externalController))
	}
	if secret != "" {
		options = append(options, hub.WithSecret(secret))
	}
	return options
}

func handleSignals() {
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
