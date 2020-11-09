package cmd

import (
	"os"
	"strings"

	"github.com/edv1n/metagit/internal/metagit"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var mappingFlag *string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the metagit server",
	Long:  `Serve will start the interface to listen to HTTP traffic. It will redirect git related traffic according to the mapping rules.`,
	Run: func(cmd *cobra.Command, args []string) {
		s := metagit.NewServer()

		mapping := *mappingFlag

		if mapping == "" {
			mapping = os.Getenv("MAPPING")
		}

		mappings := strings.Split(strings.Trim(mapping, " "), ",")
		for _, v := range mappings {
			ss := strings.Split(v, "=")

			if len(ss) == 2 {
				if err := s.AddMapping(ss[0], ss[1]); err != nil {
					logrus.Fatal(err.Error())
					os.Exit(1)
				}
			}
		}

		addr := ":8000"

		if port := os.Getenv("PORT"); port != "" {
			addr = ":" + port
		}

		if len(args) > 0 {
			addr = args[0]
		}

		logrus.Info("Serving at " + addr)
		if err := s.Serve(addr); err != nil {
			logrus.Fatal(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	mappingFlag = serveCmd.Flags().StringP("mapping", "m", "", "the mapping of how an git host shall map with another git prefix, add multiple mappping by comma(,) as the seperator. e.g. a.com=github.com/a,b.com=github.com/b")
}
