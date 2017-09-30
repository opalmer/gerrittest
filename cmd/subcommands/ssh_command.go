package cmd

import (
	"errors"
	"fmt"

	"github.com/opalmer/gerrittest"
	"github.com/spf13/cobra"
)

// GetSSHCommand returns a command to run to connect over ssh.
var GetSSHCommand = &cobra.Command{
	Use:   "get-ssh-command",
	Short: "Constructs and returns the SSH command based on the running Gerrit instance.",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := getString(cmd, "json")
		if path == "" {
			return errors.New("--json not provided")
		}

		info, err := gerrittest.LoadJSON(path)
		if err != nil {
			return err
		}

		sshcmd, err := gerrittest.GetSSHCommand(info)
		if err != nil {
			return err
		}
		fmt.Println(sshcmd)
		return nil
	},
}

func init() {
	addCommonFlags(GetSSHCommand)
}
