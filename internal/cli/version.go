package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is injected at build time by goreleaser; defaults to dev for local builds.
var version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
}
