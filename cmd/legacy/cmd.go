package legacy

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/args"
	deleteapp "numerous.com/cli/cmd/legacy/delete"
	"numerous.com/cli/cmd/legacy/initialize"
	"numerous.com/cli/cmd/legacy/list"
	"numerous.com/cli/cmd/legacy/log"
	"numerous.com/cli/cmd/legacy/publish"
	"numerous.com/cli/cmd/legacy/push"
	"numerous.com/cli/cmd/legacy/unpublish"
)

var LegacyRootCmd = &cobra.Command{
	Use:   "legacy",
	Short: "Commands for managing legacy apps on Numerous",
	Args:  args.SubCommandRequired,
}

func init() {
	LegacyRootCmd.AddCommand(push.PushCmd)
	LegacyRootCmd.AddCommand(log.LogCmd)
	LegacyRootCmd.AddCommand(deleteapp.DeleteCmd)
	LegacyRootCmd.AddCommand(publish.PublishCmd)
	LegacyRootCmd.AddCommand(unpublish.UnpublishCmd)
	LegacyRootCmd.AddCommand(list.ListCmd)
	LegacyRootCmd.AddCommand(initialize.InitCmd)
}
