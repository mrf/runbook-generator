package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mrf/runbook-generator/internal/ai"
	"github.com/mrf/runbook-generator/internal/generator"
	"github.com/mrf/runbook-generator/internal/history"
	"github.com/mrf/runbook-generator/internal/processor"
)

var (
	version = "dev"

	fromFlag   int
	toFlag     int
	outputFlag string
	titleFlag  string
)

var rootCmd = &cobra.Command{
	Use:   "runbook-gen",
	Short: "Generate runbooks from zsh history",
	Long: `Runbook Generator analyzes zsh command history between specified command
numbers, then produces a structured markdown runbook that others can follow
to reproduce the same workflow.

Command numbers match what you see when running 'history' in zsh.

The tool intelligently removes duplicates, infers intent from command
sequences, and scrubs sensitive data like passwords and API keys.`,
	RunE: run,
}

func Execute() error {
	return rootCmd.Execute()
}

func SetVersion(v string) {
	version = v
}

func init() {
	rootCmd.Flags().IntVarP(&fromFlag, "from", "f", 0, "start command number (required)")
	rootCmd.Flags().IntVarP(&toFlag, "to", "t", 0, "end command number (required)")
	rootCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "output file path (default: stdout)")
	rootCmd.Flags().StringVar(&titleFlag, "title", "Runbook", "runbook title")

	rootCmd.MarkFlagRequired("from")
	rootCmd.MarkFlagRequired("to")
}

func run(cmd *cobra.Command, args []string) error {
	// Create extractor (uses ~/.zsh_history)
	extractor, err := history.NewExtractor()
	if err != nil {
		return err
	}

	// Extract history
	entries, err := extractor.Extract(fromFlag, toFlag)
	if err != nil {
		return fmt.Errorf("failed to extract history: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Extracted %d commands from history\n", len(entries))

	// Check if AI features are available
	var aiClient *ai.Client
	if ai.Available() {
		aiClient = ai.NewClient()
		fmt.Fprintf(os.Stderr, "AI features enabled (ANTHROPIC_API_KEY detected)\n")
	}

	// Process: deduplicate
	ctx := context.Background()
	if aiClient != nil {
		fmt.Fprintf(os.Stderr, "Running AI-powered deduplication...\n")
		dedupResult, err := aiClient.DeduplicateCommands(ctx, entries)
		if err != nil {
			fmt.Fprintf(os.Stderr, "AI deduplication failed, falling back to standard: %v\n", err)
			dedup := processor.NewDedup()
			entries = dedup.Process(entries)
		} else {
			var summaries []string
			entries, summaries = ai.ApplyDedup(entries, dedupResult)
			if len(summaries) > 0 {
				fmt.Fprintf(os.Stderr, "AI deduplication applied: %d groups merged\n", len(summaries))
			}
		}
	} else {
		dedup := processor.NewDedup()
		entries = dedup.Process(entries)
	}
	fmt.Fprintf(os.Stderr, "After deduplication: %d commands\n", len(entries))

	// Process: sanitize
	sanitizer := processor.NewSanitizer()
	entries, redactions := sanitizer.Process(entries)
	if len(redactions) > 0 {
		fmt.Fprintf(os.Stderr, "Sanitized %d sensitive values\n", len(redactions))
	}

	// Process: analyze intent
	analyzer := processor.NewIntentAnalyzer()
	groups := analyzer.Analyze(entries)
	fmt.Fprintf(os.Stderr, "Organized into %d steps\n", len(groups))

	// Enhance with AI explanations if available
	var aiOverview string
	var aiPrerequisites []string
	if aiClient != nil {
		fmt.Fprintf(os.Stderr, "Generating AI explanations...\n")
		explanations, err := aiClient.GenerateExplanations(ctx, groups)
		if err != nil {
			fmt.Fprintf(os.Stderr, "AI explanation generation failed: %v\n", err)
		} else {
			groups = ai.EnhanceGroups(groups, explanations)
			aiOverview = explanations.Overview
			aiPrerequisites = explanations.Prerequisites
			fmt.Fprintf(os.Stderr, "AI explanations added\n")
		}
	}

	// Generate runbook
	gen := generator.NewMarkdownGenerator()
	timeRange := fmt.Sprintf("commands #%d to #%d", fromFlag, toFlag)

	data := generator.RunbookData{
		Title:           titleFlag,
		Generated:       time.Now(),
		TimeRange:       timeRange,
		Groups:          groups,
		RedactedCount:   len(redactions),
		AIOverview:      aiOverview,
		AIPrerequisites: aiPrerequisites,
	}

	output := gen.Generate(data)

	// Write output
	if outputFlag != "" {
		if err := os.WriteFile(outputFlag, []byte(output), 0600); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Runbook written to %s\n", outputFlag)
	} else {
		fmt.Print(output)
	}

	return nil
}
