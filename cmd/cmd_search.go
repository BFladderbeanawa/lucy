package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/mclucy/lucy/logger"
	"github.com/mclucy/lucy/syntax"
	"github.com/mclucy/lucy/tools"
	"github.com/mclucy/lucy/tui"
	"github.com/mclucy/lucy/types"
	"github.com/mclucy/lucy/upstream/routing"
	"github.com/spf13/cobra"
)

const (
	flagIndexName  = "index"
	flagClientName = "client"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for mods and plugins",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return CompletePackageIDSuggestions(context.Background(), "search", toComplete)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		index, _ := cmd.Flags().GetString(flagIndexName)
		if !types.SearchSort(index).Valid() {
			return errors.New("--index must be one of \"relevance\", \"downloads\", \"newest\"")
		}
		return validateSourceFlag(cmd)
	},
	RunE: runWithErrorLogging(actionSearch),
}

// subcmdSearch is an alias for searchCmd for backward compatibility.
// TODO: Remove after cmd/cmd.go is migrated to Cobra.
var subcmdSearch = searchCmd

func init() {
	searchCmd.Flags().StringP(flagIndexName, "i", "relevance", "Index search results by INDEX")
	searchCmd.Flags().BoolP(flagClientName, "c", false, "Also show client-only mods in results")
	addJsonFlag(searchCmd)
	addLongFlag(searchCmd)
	addNoStyleFlag(searchCmd)
	addSourceFlag(searchCmd)
	_ = searchCmd.RegisterFlagCompletionFunc(flagSourceName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var vals []string
		for _, c := range StaticSourceCandidates() {
			vals = append(vals, c.Value)
		}
		return FilterByPrefixStrings(vals, toComplete), cobra.ShellCompDirectiveNoFileComp
	})
	_ = searchCmd.RegisterFlagCompletionFunc(flagIndexName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var vals []string
		for _, c := range StaticSortCandidates() {
			vals = append(vals, c.Value)
		}
		return FilterByPrefixStrings(vals, toComplete), cobra.ShellCompDirectiveNoFileComp
	})
	rootCmd.AddCommand(searchCmd)
}

func actionSearch(cmd *cobra.Command, args []string) error {
	p := syntax.Parse(args[0])
	index, _ := cmd.Flags().GetString(flagIndexName)
	client, _ := cmd.Flags().GetBool(flagClientName)
	long, _ := cmd.Flags().GetBool(flagLongName)
	sourceArg, _ := cmd.Flags().GetString(flagSourceName)
	specifiedSource := types.ParseSource(sourceArg)

	options := types.SearchOptions{
		IncludeClient: client,
		SortBy:        types.SearchSort(index),
	}

	out := &tui.Data{}
	providers, err := routing.ResolveProviders(p.Platform, specifiedSource)
	if err != nil {
		errArg := sourceArg
		if specifiedSource == types.SourceAuto {
			errArg = p.Platform.String()
		}
		logger.Fatal(fmt.Errorf("%w: %s", err, errArg))
	}

	results, errs := routing.SearchMany(providers, p.Name, options)
	for _, err := range errs {
		providerErr := fmt.Errorf(
			"search on %s failed: %w",
			err.Source.Title(),
			err.Err,
		)
		if specifiedSource == types.SourceAuto && len(providers) > 1 {
			logger.ReportWarn(providerErr)
			continue
		}
		logger.ReportWarn(providerErr)
	}

	if err := searchResultError(results, errs); err != nil {
		return err
	}

	for _, res := range results {
		appendToSearchOutput(out, long, res)
	}

	tui.Flush(out)
	return nil
}

func searchResultError(
	results []types.SearchResults,
	providerErrors []routing.ProviderError,
) error {
	if len(results) > 0 || len(providerErrors) == 0 {
		return nil
	}
	joined := make([]error, 0, len(providerErrors))
	for _, providerErr := range providerErrors {
		joined = append(joined, providerErr)
	}
	return errors.Join(joined...)
}

func appendToSearchOutput(
	out *tui.Data,
	showAll bool,
	res types.SearchResults,
) {
	var results []string
	for _, r := range res.Projects {
		results = append(results, r.String())
	}

	if len(out.Fields) != 0 {
		out.Fields = append(
			out.Fields, &tui.FieldSeparator{
				Length: 0,
				Dim:    false,
			},
		)
	}

	out.Fields = append(
		out.Fields,
		&tui.FieldAnnotation{
			Annotation: "Results from " + res.Source.Title(),
		},
	)

	if res.Source == types.SourceModrinth && len(res.Projects) == 100 {
		out.Fields = append(
			out.Fields,
			&tui.FieldAnnotation{
				Annotation: "* only showing the top 100",
			},
		)
	}

	out.Fields = append(
		out.Fields,
		&tui.FieldShortText{
			Title: "#  ",
			Text:  strconv.Itoa(len(res.Projects)),
		},
		&tui.FieldDynamicColumnLabels{
			Title:  ">>>",
			Labels: results,
			MaxLines: tools.Ternary(
				showAll,
				0,
				tools.TermHeight()-6,
			),
		},
	)
}
