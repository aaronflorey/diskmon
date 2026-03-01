package cli

import (
	"fmt"

	"diskmon/internal/config"

	"github.com/spf13/cobra"
)

func newConfigCmd(cfg *config.Config) *cobra.Command {
	configCmd := &cobra.Command{Use: "config", Short: "Config tools"}
	configCmd.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: "Validate merged config",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}
			fmt.Println("config is valid")
			return nil
		},
	})
	return configCmd
}
