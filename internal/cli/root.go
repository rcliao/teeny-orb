package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/rcliao/teeny-orb/internal/cli/commands"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "teeny-orb",
	Short: "AI-powered coding assistant with container isolation",
	Long: `teeny-orb is an AI-powered coding assistant that executes all operations 
within containerized environments for security and isolation. It bridges LLM 
capabilities with local development through the Model Context Protocol (MCP).`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting interactive coding session...")
		fmt.Println("Type 'help' for available commands or 'exit' to quit.")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.teeny-orb.yaml)")
	rootCmd.PersistentFlags().String("project", "", "project directory to work with")
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")

	viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Add subcommands
	rootCmd.AddCommand(commands.NewGenerateCmd())
	rootCmd.AddCommand(commands.NewReviewCmd())
	rootCmd.AddCommand(commands.NewSessionCmd())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".teeny-orb")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}