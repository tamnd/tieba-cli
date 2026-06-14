package cli

import (
	"github.com/spf13/cobra"
)

// hotCmd returns the hot command that lists Baidu Tieba hot topics.
func (a *App) hotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hot",
		Short: "List Baidu Tieba hot topics (百度贴吧热门话题)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			a.progressf("fetching hot topics...")
			topics, err := a.client.Hot(cmd.Context(), n)
			if err != nil {
				return codeError(exitError, err)
			}
			return a.renderOrEmpty(topics, len(topics))
		},
	}
}
