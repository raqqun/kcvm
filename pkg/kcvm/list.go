package cmds

import (
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-version"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

type GithubTag struct {
	Name       string `json:"name"`
	ZipballURL string `json:"zipball_url"`
	TarballURL string `json:"tarball_url"`
	Commit     struct {
		Sha string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	NodeID string `json:"node_id"`
}

type GithubTags []*GithubTag

// ListCmd lists remote or locally installed kubectl versions
var ListCmd = cli.Command{
	Name:  "list",
	Usage: "List kubectl version",
	Subcommands: []*cli.Command{
		{
			Name:  "remote",
			Usage: "List upstream kubectl versions",
			Action: func(c *cli.Context) error {
				client := resty.New()
				var respTags GithubTags

				page := 1
				for {
					resp, err := client.R().
						SetHeader("Accept", "application/json").
						SetResult(GithubTags{}).
						Get(fmt.Sprintf("https://api.github.com/repos/kubernetes/kubernetes/tags?per_page=100&page=%d", page))

					if err != nil {
						return err
					}

					tags := resp.Result().(*GithubTags)
					if len(*tags) == 0 {
						break // No more pages
					}

					for _, tag := range *tags {
						respTags = append(respTags, tag)
					}
					page++
				}

				for i := range respTags {
					switch {
					case strings.Contains(respTags[i].Name, "beta"):
						fallthrough
					case strings.Contains(respTags[i].Name, "alpha"):
						fallthrough
					case strings.Contains(respTags[i].Name, "rc"):
					default:
						fmt.Println(respTags[i].Name)
					}
				}

				return nil
			},
		},
		{
			Name:  "local",
			Usage: "List locally installed kubectl version",
			Action: func(c *cli.Context) error {
				files, err := ioutil.ReadDir(path.Join(KcvmPath, "bin"))
				if err != nil {
					return err
				}

				sort.SliceStable(files, func(i, j int) bool {
					vi, _ := version.NewVersion(strings.TrimPrefix(files[i].Name(), "kubectl-"))
					vj, _ := version.NewVersion(strings.TrimPrefix(files[j].Name(), "kubectl-"))
					return vi.GreaterThan(vj)
				})

				for _, file := range files {
					switch file.Name() {
					case "kubectl":
						break
					default:
						fmt.Println(strings.TrimPrefix(file.Name(), "kubectl-"))
					}
				}

				return nil
			},
		},
	},
}
