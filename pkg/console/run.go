package console

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"osub/pkg/parser"
	"osub/pkg/resolve"
	"osub/pkg/shared"
	"osub/pkg/shared/types"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the osub service",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := shared.ReadConfig()

		if err != nil {
			fmt.Println("Cannot read osub config json file: ", err)
		}

		for {
			servers, err := setup(conf)

			if err != nil {
				log.Fatalf("%v", err)
			}

			fmt.Println((servers))

			duration, err := resolve.Interval(conf.Interval)

			if err != nil {
				log.Fatalf("Error parsing Interval string: %v", err)
			}

			time.Sleep(duration)
		}
	},
}

func init() {
	RootCmd.AddCommand(RunCmd)
}

func setup(conf *types.OsubConfig) ([]types.OsubServerConfig, error) {
	var servers []types.OsubServerConfig

	for _, sub := range conf.Subscriptions {
		resp, err := http.Get(sub.URL)

		if err != nil {
			log.Fatalf("Error fetching subscription: %v", err)
		}

		body, err := io.ReadAll(resp.Body)

		if err != nil {
			log.Fatalf("Error request subscription link: %v", err)
		}

		links, err := parser.Subscription(string(body))

		if err != nil {
			log.Fatalf("Error parsing Subscription link: %v", err)
		}

		for _, link := range links {
			if strings.HasPrefix(link, shared.VMESS_PREFIX) {
				config, err := parser.Vmess(link)

				if err != nil {
					log.Fatalf("Error parsing Vmess link: %v", err)
				}

				servers = append(servers, *config)
			}

			if strings.HasPrefix(link, shared.SS_PREFIX) {
				config, err := parser.Shadowsocks(link)

				if err != nil {
					log.Fatalf("Error parsing Vmess link: %v", err)
				}

				servers = append(servers, *config)
			}
		}

		err = resp.Body.Close()

		if err != nil {
			log.Fatalf("Response closed error: %v", err)
		}
	}

	return servers, nil
}
