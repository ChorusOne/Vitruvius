// This file provides a version command.
//
// ```bash
// $ vitruvius version
// Version 0.0.1
// ```

package commands

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/ChorusOne/Hippias/cmd/hippias/types"
)

// Version creates the cobra struct required to implement a version command.
// TODO: Automatically pull this from git tag/git commit hash
func Version(_ *types.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print Version. (1.0.0)",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Version 1.0.0")
		},
	}
}
