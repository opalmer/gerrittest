package cmd

import (
	//"github.com/andygrunwald/go-gerrit"
	"github.com/spf13/cobra"
	//"fmt"
)

var CreateAccount = &cobra.Command{
	Use: "create-account",
	RunE: func(cmd *cobra.Command, args []string) error {
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return err
		}
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return err
		}
		_ = username
		_ = password

		// FIXME This needs to come from the json file
		//url := fmt.Sprintf("http://localhost")
		//gerrit.NewClient("http://lo")

		return nil
	},
}

func init() {
	CreateAccount.Flags().StringP(
		"username", "u", "admin",
		"The name of the user to create")
	CreateAccount.Flags().StringP(
		"password", "p", "secret",
		"The password to associate with the user.")
	CreateAccount.Flags().String(
		"json", "",
		"The json file containing information about the container.")
}
