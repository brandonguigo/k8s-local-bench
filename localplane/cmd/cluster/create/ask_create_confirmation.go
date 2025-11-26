package create

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// askCreateConfirmation prompts the user to confirm cluster creation unless
// the --yes flag is provided. Returns true to proceed, false to abort.
func askCreateConfirmation(cmd *cobra.Command, clusterName string) bool {
	yes, _ := cmd.Flags().GetBool("yes")
	if yes {
		return true
	}

	prompt := promptui.Select{
		Label: fmt.Sprintf("Proceed to create kind cluster '%s'?", clusterName),
		Items: []string{"Yes", "No"},
	}

	i, _, err := prompt.Run()
	if err != nil {
		log.Info().Err(err).Msg("aborting cluster creation")
		return false
	}

	if i == 0 {
		return true
	}

	log.Info().Msg("aborting cluster creation")
	return false
}
