package config

import (
	"fmt"
	"os"

	C "github.com/MerlinKodo/clash-rev/constant"
	"github.com/MerlinKodo/clash-rev/log"
)

// Init prepare necessary files
func Init(dir string) error {
	// initial homedir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return fmt.Errorf("can't create config directory %s: %s", dir, err.Error())
		}
	}

	// initial config.yaml
	configPath := C.Path.Config()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Infoln("Can't find config, create an initial config file")
		f, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("can't create file %s: %s", configPath, err.Error())
		}
		f.Write([]byte(`mixed-port: 7890`))
		f.Close()
	}

	return nil
}
