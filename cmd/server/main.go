package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SchSeba/slack-ai-assistant/pkg/agent"
	"github.com/SchSeba/slack-ai-assistant/pkg/database"
	"github.com/SchSeba/slack-ai-assistant/pkg/llm"
	slackbot "github.com/SchSeba/slack-ai-assistant/pkg/slack-bot"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/spf13/cobra"
)

var (
	slackBotToken string
	slackAppToken string
	debug         bool
	workers       int
	botUser       *slack.AuthTestResponse // Store bot user info globally
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&slackBotToken, "bot-token", "b", "", "Slack Bot Token (required)")
	rootCmd.PersistentFlags().StringVarP(&slackAppToken, "app-token", "a", "", "Slack App Token (required)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "w", 10, "Number of workers for the agent")

	// Mark required flags
	rootCmd.MarkPersistentFlagRequired("bot-token")
	rootCmd.MarkPersistentFlagRequired("app-token")
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

	// Create a context that can be cancelled
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
		log.Fatalf("❌ Failed to create database: %v", err)
	}
	defer db.Close()

	// Auto migrate the database
	err = db.AutoMigrate()
	if err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}

	appMentionChannel := make(chan *slackevents.AppMentionEvent, 100)
	slashCommandChannel := make(chan *slack.SlashCommand, 100)

	slackBot, err := slackbot.NewSlackBot(slackBotToken, slackAppToken, appMentionChannel, slashCommandChannel, debug)
	if err != nil {
		log.Fatalf("❌ Failed to create Slack bot: %v", err)
	}

	llmClient := llm.NewLLMClient()

	agent := agent.NewAgent(db, slackBot, llmClient, appMentionChannel, slashCommandChannel, workers)
	fmt.Println("👋 Starting Slack AI Assistant Bot...")
	agent.Start(ctx)
	fmt.Println("👋 Shutting down Slack AI Assistant Bot...")
}

func main() {
	Execute()
}
