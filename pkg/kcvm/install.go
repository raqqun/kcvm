package cmds

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

// InstallCmd declares install command
var InstallCmd = cli.Command{
	Name:      "install",
	Usage:     "Install kubectl version and start using it",
	ArgsUsage: "VERSION",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "force",
			Usage: "download kubectl again even if it is already installed",
		},
	},
	Action: func(c *cli.Context) error {
		kubectlVersion, err := versionArg(c, "kcvm list remote")
		if err != nil {
			return err
		}

		binPath := kubectlBinPath(kubectlVersion)

		_, err = os.Stat(binPath)
		switch {
		case err == nil && !c.Bool("force"):
			log.Info("kubectl " + kubectlVersion + " is already installed")
		case err == nil || os.IsNotExist(err):
			err = os.MkdirAll(binDir(), 0755)
			if err != nil {
				return err
			}

			err = downloadKubectl(kubectlVersion, binPath)
			if err != nil {
				return err
			}

			log.Info("kubectl " + kubectlVersion + " installed")
		default:
			return err
		}

		err = activate(binPath)
		if err != nil {
			return err
		}

		log.Info("Now using kubectl " + kubectlVersion)

		return nil
	},
}

// downloadKubectl downloads a kubectl binary to destPath and verifies its checksum.
// The binary is written to a temporary file next to destPath and renamed into place
// once complete, so an interrupted or corrupt download never leaves a broken kubectl.
func downloadKubectl(kubectlVersion string, destPath string) error {
	url := fmt.Sprintf("%s/%s/bin/%s/%s/kubectl", KubectlDownloadURL, kubectlVersion, GOOS, GOARCH)

	client := resty.New().
		SetRetryCount(3).
		SetTimeout(10 * time.Minute)

	resp, err := client.R().
		SetDoNotParseResponse(true).
		Get(url)
	if err != nil {
		return err
	}
	body := resp.RawBody()
	defer body.Close()

	switch resp.StatusCode() {
	case http.StatusOK:
	case http.StatusNotFound, http.StatusForbidden:
		return fmt.Errorf("kubectl %s does not exist for %s/%s", kubectlVersion, GOOS, GOARCH)
	default:
		return fmt.Errorf("Downloading %s failed: %s", url, resp.Status())
	}

	checksum, err := fetchChecksum(client, url+".sha256")
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(filepath.Dir(destPath), "."+filepath.Base(destPath)+".*.tmp")
	if err != nil {
		return err
	}
	// No-op once the temporary file has been renamed to destPath
	defer os.Remove(tempFile.Name())

	hash := sha256.New()
	_, err = io.Copy(io.MultiWriter(tempFile, hash), body)
	if err != nil {
		tempFile.Close()
		return fmt.Errorf("Downloading %s failed: %w", url, err)
	}

	err = tempFile.Close()
	if err != nil {
		return err
	}

	if checksum != "" {
		sum := hex.EncodeToString(hash.Sum(nil))
		if sum != checksum {
			return fmt.Errorf("Checksum mismatch for kubectl %s: expected %s, got %s", kubectlVersion, checksum, sum)
		}
	}

	err = os.Chmod(tempFile.Name(), 0755)
	if err != nil {
		return err
	}

	return os.Rename(tempFile.Name(), destPath)
}

// fetchChecksum returns the published sha256 of a kubectl binary, or "" when the
// release predates published checksums
func fetchChecksum(client *resty.Client, url string) (string, error) {
	resp, err := client.R().Get(url)
	if err != nil {
		return "", err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
	case http.StatusNotFound, http.StatusForbidden:
		log.Warn("No checksum published for " + url + ", skipping verification")
		return "", nil
	default:
		return "", fmt.Errorf("Downloading %s failed: %s", url, resp.Status())
	}

	// The file contains the hex digest, optionally followed by the file name
	fields := strings.Fields(resp.String())
	if len(fields) == 0 || len(fields[0]) != sha256.Size*2 {
		return "", fmt.Errorf("Invalid checksum file %s", url)
	}

	return strings.ToLower(fields[0]), nil
}
