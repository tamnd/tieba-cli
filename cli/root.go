// Package cli builds the tieba command tree on top of the tieba-cli library.
package cli

import (
	"github.com/spf13/cobra"
)

// Build metadata, set via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Root builds the root command and its subtree.
func Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "tieba",
		Short: "Browse Baidu Tieba hot topics (百度贴吧热议)",
		Long: `Browse Baidu Tieba hot topics (百度贴吧热议)

This is a fresh scaffold. Add your commands here on top of the tieba-cli
library package, then wire them into Root with root.AddCommand.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newVersionCmd())
	// TODO: root.AddCommand(newGetCmd()), etc.
	return root
}
