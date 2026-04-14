package cmd

import "github.com/spf13/cobra"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Lucy on current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// subcmdInit is an alias for initCmd for backward compatibility.
// TODO: Remove after cmd/cmd.go is migrated to Cobra.
var subcmdInit = initCmd

func init() {
	rootCmd.AddCommand(initCmd)
}
