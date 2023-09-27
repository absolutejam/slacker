package cli

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	RootFlagVerbose = "verbose"
)

func init() {
	viper.AutomaticEnv()

	RootCmd.PersistentFlags().Bool(RootFlagVerbose, false, "Show Debug log output")
	viper.BindPFlag(RootFlagVerbose, RootCmd.PersistentFlags().Lookup(RootFlagVerbose))

	RootCmd.AddCommand(SlackCmd)
}

var RootCmd = &cobra.Command{
	Use:   "slacker",
	Short: "Sends messages to Slack",

	Args: cobra.ExactArgs(1),

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		verbose := viper.GetBool(RootFlagVerbose)

		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}
