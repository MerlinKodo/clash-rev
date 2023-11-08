package config

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/MerlinKodo/clash-rev/component/mmdb"
	C "github.com/MerlinKodo/clash-rev/constant"
	"github.com/MerlinKodo/clash-rev/log"
)

func downloadMMDB(path string) (err error) {
	resp, err := http.Get("https://cdn.jsdelivr.net/gh/MerlinKodo/maxmind-geoip@release/Country.mmdb")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)

	return err
}

func initMMDB() error {
	mmdbPath := C.Path.MMDB()

	attemptDownload := func() error {
		if err := downloadMMDB(mmdbPath); err != nil {
			return fmt.Errorf("can't download MMDB: %w", err)
		}
		return nil
	}

	if _, err := os.Stat(mmdbPath); os.IsNotExist(err) {
		log.Infoln("Can't find MMDB, start download")
		return attemptDownload()
	} else if !mmdb.Verify() {
		log.Warnln("MMDB invalid, remove and download")
		if err := os.Remove(mmdbPath); err != nil {
			return fmt.Errorf("can't remove invalid MMDB: %w", err)
		}
		return attemptDownload()
	}

	return nil
}

// Init prepare necessary files
func Init(dir string) error {
	// initial homedir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return fmt.Errorf("can't create config directory %s: %s", dir, err.Error())
		}
	}

	// initial config.yaml
	if _, err := os.Stat(C.Path.Config()); os.IsNotExist(err) {
		log.Infoln("Can't find config, create a initial config file")
		f, err := os.OpenFile(C.Path.Config(), os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("can't create file %s: %s", C.Path.Config(), err.Error())
		}
		f.Write([]byte(`mixed-port: 7890`))
		f.Close()
	}

	// initial mmdb
	if err := initMMDB(); err != nil {
		return fmt.Errorf("can't initial MMDB: %w", err)
	}
	return nil
}
