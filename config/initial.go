package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/MerlinKodo/clash-rev/component/mmdb"
	C "github.com/MerlinKodo/clash-rev/constant"
	"github.com/MerlinKodo/clash-rev/log"
)

const (
	mmdbURL           = "https://cdn.jsdelivr.net/gh/MerlinKodo/maxmind-geoip@release/Country.mmdb"
	initialConfig     = "mixed-port: 7890"
	directoryPerm     = 0o755
	filePerm          = 0o644
	httpClientTimeout = 60 * time.Second
)

var httpClient = &http.Client{Timeout: httpClientTimeout}

func downloadMMDB(path string) (err error) {
	resp, err := httpClient.Get(mmdbURL)
	if err != nil {
		return fmt.Errorf("failed to get MMDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK HTTP status code: %d", resp.StatusCode)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, filePerm)
	if err != nil {
		return fmt.Errorf("error while opening the file: %w", err)
	}

	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("error while closing the file: %w", cerr)
		}
	}()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("failed to write MMDB file: %w", err)
	}

	return nil
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
		log.Infoln("MMDB not found, starting download")
		return attemptDownload()
	} else if !mmdb.Verify() {
		log.Warnln("MMDB is invalid, removing and re-downloading")
		if err := os.Remove(mmdbPath); err != nil {
			return fmt.Errorf("failed to remove invalid MMDB: %w", err)
		}
		return attemptDownload()
	}

	return nil
}

// Init prepare necessary files
func Init(dir string) error {
	// initial homedir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, directoryPerm); err != nil {
			return fmt.Errorf("failed to create config directory %s: %w", dir, err)
		}
	}

	// initial config.yaml
	if _, err := os.Stat(C.Path.Config()); os.IsNotExist(err) {
		log.Infoln("Can't find config, create a initial config file")
		f, err := os.OpenFile(C.Path.Config(), os.O_CREATE|os.O_WRONLY, filePerm)
		if err != nil {
			return fmt.Errorf("failed to create config directory %s: %w", C.Path.Config(), err)
		}
		f.Write([]byte(initialConfig))
		f.Close()
	}

	// initial mmdb
	if err := initMMDB(); err != nil {
		return fmt.Errorf("failed to initialize MMDB: %w", err)
	}
	return nil
}
