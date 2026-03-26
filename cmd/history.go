package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/guilhermezuriel/git-resume/internal/storage"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "List all repositories with saved summaries",
	RunE: func(cmd *cobra.Command, args []string) error {
		repos, err := storage.ListRepos()
		if err != nil {
			return err
		}

		printHeader("All Repositories")

		if len(repos) == 0 {
			fmt.Println("  [INFO] No repositories found")
			fmt.Println()
			fmt.Println("  Generate summaries in a git repository first.")
			fmt.Println()
			return nil
		}

		fmt.Println()
		fmt.Fprintf(os.Stdout, "  %-20s %-34s %s\n", "REPOSITORY", "PATH", "SUMMARIES")
		fmt.Fprintf(os.Stdout, "  %s\n", repeat("─", 65))

		for _, repo := range repos {
			shortPath := repo.Path
			if len(shortPath) > 32 {
				shortPath = "..." + shortPath[len(shortPath)-29:]
			}
			fmt.Fprintf(os.Stdout, "  %-20s %-34s %d\n",
				truncate(repo.Name, 20),
				shortPath,
				repo.Count,
			)
		}
		fmt.Println()
		return nil
	},
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
