package legacy

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/group"
	deleteapp "numerous.com/cli/cmd/legacy/delete"
	"numerous.com/cli/cmd/legacy/initialize"
	"numerous.com/cli/cmd/legacy/list"
	"numerous.com/cli/cmd/legacy/log"
	"numerous.com/cli/cmd/legacy/publish"
	"numerous.com/cli/cmd/legacy/push"
	"numerous.com/cli/cmd/legacy/unpublish"
)

var Cmd = &cobra.Command{
	Use:     "legacy",
	Short:   "Commands for managing legacy apps on Numerous",
	Args:    args.SubCommandRequired,
	GroupID: group.AdditionalCommandsGroupID,
}

func init() {
	Cmd.AddCommand(push.PushCmd)
	Cmd.AddCommand(log.LogCmd)
	Cmd.AddCommand(deleteapp.DeleteCmd)
	Cmd.AddCommand(publish.PublishCmd)
	Cmd.AddCommand(unpublish.UnpublishCmd)
	Cmd.AddCommand(list.ListCmd)
	Cmd.AddCommand(initialize.InitCmd)
}
