package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/automaxprocs/maxprocs"
)

func setupMaxProcs() {
	_, _ = maxprocs.Set(maxprocs.Logger(func(string, ...any) {}))
}

func resolvePath(path string) string {
	if !filepath.IsAbs(path) {
		currentDir, _ := os.Getwd()
		return filepath.Join(currentDir, path)
	}
	return path
}

func parseHeader(headerString string) map[string]string {
	header := make(map[string]string)
	if headerString != "" {
		headerList := strings.Split(headerString, ",")
		for _, v := range headerList {
			kv := strings.Split(v, ":")
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				header[key] = value
			}
		}
	}
	return header
}

func downloadFile(url string, filepath string, header map[string]string) error {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	for key, value := range header {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
