/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"github.com/enmand/quarid/internal/logger"
	"github.com/enmand/quarid/internal/matrix"
	"github.com/samber/mo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: start,
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringP("homeserver", "s", "", "The homeserver to connect to")
	startCmd.MarkFlagRequired("homeserver")

	startCmd.Flags().StringP("user", "u", "", "The user to connect as")
	startCmd.MarkFlagRequired("user")

	startCmd.Flags().String("pickle", "", "The pickle key to use for crypto. DEFAULT: ''")
	startCmd.Flags().String("password", "", "The password to use for login. DEFAULT: ''")
	startCmd.Flags().String("database-path", "crypto.db", "The path to the database file. DEFAULT: 'crypto.db'")
	startCmd.MarkFlagsRequiredTogether("pickle", "password", "database-path")

	viper.BindPFlags(startCmd.Flags())
}

func start(c *cobra.Command, args []string) (err error) {
	ctx := context.Background()
	log := logger.New(mo.Some(logger.GetWriter(debug)))

	homeserver, err := c.Flags().GetString("homeserver")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get homeserver flag")
		return
	}

	client, err := matrix.NewClient(
		homeserver,
		"quarid",
		matrix.WithLogger[string](log),
		matrix.WithE2EE(
			MustGetFlag[string]("pickle"),
			MustGetFlag[string]("password"),
			MustGetFlag[string]("database-path"),
		),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create client")
		return
	}

	log.Info().Msg("Starting client")
	if err = client.Sync(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to sync")
		return
	}

	return nil
}
