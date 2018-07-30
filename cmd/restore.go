package cmd

import (
	"github.com/spf13/cobra"
	"github.com/voormedia/kd/pkg/restore"
)

var cmdRestore = &cobra.Command{
	Use:   "restore",
	Short: "Download and restore a copy of a Google Cloud SQL database.",
	DisableFlagsInUseLine: true,

	Run: func(_ *cobra.Command, args []string) {
		if err := restore.Run(log); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cmdRoot.AddCommand(cmdRestore)
}
