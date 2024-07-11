package utils

import "github.com/spf13/cobra"

func New() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.AddCommand(NewVideoSlice())
	return cmd
}
