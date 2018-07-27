package cmd

import (
	"github.com/spf13/cobra"
	"github.com/voormedia/kd/pkg/proxy"
)

var cmdProxy = &cobra.Command{
	Use:   "proxy",
	Short: "Connect to Google Cloud SQL instance through a proxy.",
	DisableFlagsInUseLine: true,

	Run: func(_ *cobra.Command, args []string) {
		if err := proxy.Run(log); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cmdRoot.AddCommand(cmdProxy)
}
