package cmds

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-version"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

// stableVersionPattern matches kubectl release versions, excluding alpha, beta and rc
var stableVersionPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// ListCmd lists remote or locally installed kubectl versions
var ListCmd = cli.Command{
	Name:  "list",
	Usage: "List kubectl version",
	Subcommands: []*cli.Command{
		{
			Name:  "remote",
			Usage: "List upstream kubectl versions",
			Action: func(c *cli.Context) error {
				tags, err := listKubernetesTags()
				if err != nil {
					return err
				}

				var versions []string
				for _, tag := range tags {
					if stableVersionPattern.MatchString(tag) {
						versions = append(versions, tag)
					}
				}

				for _, v := range sortVersions(versions) {
					fmt.Fprintln(c.App.Writer, v)
				}

				return nil
			},
		},
		{
			Name:  "local",
			Usage: "List locally installed kubectl version, the active one is marked with *",
			Action: func(c *cli.Context) error {
				files, err := os.ReadDir(binDir())
				if os.IsNotExist(err) {
					return errors.New("kcvm is not initialized, run: kcvm init")
				}
				if err != nil {
					return err
				}

				var versions []string
				for _, file := range files {
					v := strings.TrimPrefix(file.Name(), "kubectl-")
					if strings.HasPrefix(file.Name(), "kubectl-") && versionPattern.MatchString(v) {
						versions = append(versions, v)
					}
				}

				active := activeVersion()
				for _, v := range sortVersions(versions) {
					if v == active {
						fmt.Fprintln(c.App.Writer, "* "+v)
					} else {
						fmt.Fprintln(c.App.Writer, "  "+v)
					}
				}

				return nil
			},
		},
	},
}

// sortVersions sorts versions from newest to oldest. Versions must match versionPattern.
func sortVersions(versions []string) []string {
	sort.SliceStable(versions, func(i, j int) bool {
		vi := version.Must(version.NewVersion(versions[i]))
		vj := version.Must(version.NewVersion(versions[j]))
		return vi.GreaterThan(vj)
	})
	return versions
}

// listKubernetesTags returns the tag names of the Kubernetes repository.
//
// Tags are listed with a single git protocol v2 "ls-refs" request, the same one
// `git ls-remote --tags` makes. The GitHub REST API would need a dozen paginated
// requests per listing and only allows 60 unauthenticated requests per hour.
func listKubernetesTags() ([]string, error) {
	request := pktLine("command=ls-refs\n") + "0001" + pktLine("ref-prefix refs/tags/\n") + "0000"

	resp, err := resty.New().
		SetRetryCount(3).
		SetTimeout(time.Minute).
		R().
		SetHeader("Git-Protocol", "version=2").
		SetHeader("Content-Type", "application/x-git-upload-pack-request").
		SetBody(request).
		Post(KubernetesRepoURL + "/git-upload-pack")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("Listing kubectl versions from %s failed: %s", KubernetesRepoURL, resp.Status())
	}

	return parseLsRefs(resp.Body())
}

// pktLine encodes a git pkt-line: the line length in 4 hex digits followed by the line
func pktLine(line string) string {
	return fmt.Sprintf("%04x%s", len(line)+4, line)
}

// parseLsRefs extracts tag names from an ls-refs response, a list of pkt-lines of
// the form "<object id> refs/tags/<name>" terminated by a flush packet "0000"
func parseLsRefs(data []byte) ([]string, error) {
	errInvalid := errors.New("Unexpected response while listing kubectl versions")

	var tags []string
	for len(data) >= 4 {
		length, err := strconv.ParseUint(string(data[:4]), 16, 16)
		if err != nil {
			return nil, errInvalid
		}
		if length == 0 {
			return tags, nil
		}
		if length < 4 || int(length) > len(data) {
			return nil, errInvalid
		}

		fields := strings.Fields(string(data[4:length]))
		if len(fields) >= 2 {
			tags = append(tags, strings.TrimPrefix(fields[1], "refs/tags/"))
		}

		data = data[length:]
	}

	return nil, errInvalid
}
