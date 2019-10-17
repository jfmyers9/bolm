package commands

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Long:          "bolm yo",
	Short:         "bolm",
	SilenceErrors: true,
	Use:           "bolm",
}

var valuesFile string

func init() {
	deployCommand.Flags().StringVarP(&valuesFile, "values", "f", "", "values file")
	RootCmd.AddCommand(deployCommand)
}

var deployCommand = &cobra.Command{
	RunE:  deploy,
	Short: "bolm",
	Use:   "deploy <helm-chart> -f <values-file>",
}

func deploy(cmd *cobra.Command, args []string) error {
	return nil
}
