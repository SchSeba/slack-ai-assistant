package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/spf13/cobra"

	"github.com/SchSeba/slack-ai-assistant/pkg/agent"
	"github.com/SchSeba/slack-ai-assistant/pkg/database"
	"github.com/SchSeba/slack-ai-assistant/pkg/llm"
	slackbot "github.com/SchSeba/slack-ai-assistant/pkg/slack-bot"
)

var (
	slackBotToken string
	slackAppToken string
	debug         bool
	workers       int
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&slackBotToken, "bot-token", "b", "", "Slack Bot Token (required)")
	rootCmd.PersistentFlags().StringVarP(&slackAppToken, "app-token", "a", "", "Slack App Token (required)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "w", 10, "Number of workers for the agent")

	// Mark required flags
	if err := rootCmd.MarkPersistentFlagRequired("bot-token"); err != nil {
		log.Fatalf("Failed to mark bot-token as required: %v", err)
	}
	if err := rootCmd.MarkPersistentFlagRequired("app-token"); err != nil {
		log.Fatalf("Failed to mark app-token as required: %v", err)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "slack-ai-assistant",
	Short: "A Slack AI Assistant Bot",
	Long: `A Slack AI Assistant Bot that can respond to messages and interact with users.
This bot uses socket mode for real-time communication with Slack.`,
	Run: func(cmd *cobra.Command, args []string) {
		startSlackBot()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func startSlackBot() {
	fmt.Printf("🚀 Starting Slack AI Assistant Bot with %d workers...\n", workers)

	if slackBotToken == "" || slackAppToken == "" {
		log.Fatal("❌ Both bot-token and app-token are required")
	}

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle shutdown signals
	go func() {
		sig := <-sigChan
		fmt.Printf("🛑 Received signal %v, shutting down gracefully...\n", sig)
		cancel()
	}()

	db, err := database.NewDatabase("slack-ai-assistant.db")
	if err != nil {
		//nolint:gocritic // this is a critical error, so we should log it and exit
		log.Fatalf("❌ Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Auto migrate the database
	err = db.AutoMigrate()
	if err != nil {
		//nolint:gocritic // this is a critical error, so we should log it and exit
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}

	appMentionChannel := make(chan *slackevents.AppMentionEvent, 100)
	slashCommandChannel := make(chan *slack.SlashCommand, 100)

	slackBot, err := slackbot.NewSlackBot(slackBotToken, slackAppToken, appMentionChannel, slashCommandChannel, debug)
	if err != nil {
		//nolint:gocritic // this is a critical error, so we should log it and exit
		log.Fatalf("❌ Failed to create Slack bot: %v", err)
	}

	llmClient := llm.NewLLMClient()

	agentProcess := agent.NewAgent(db, slackBot, llmClient, appMentionChannel, slashCommandChannel, workers)
	fmt.Println("👋 Starting Slack AI Assistant Bot...")
	agentProcess.Start(ctx)
	fmt.Println("👋 Shutting down Slack AI Assistant Bot...")
}

func main() {
	Execute()
}
