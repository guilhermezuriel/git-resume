package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	gitpkg "github.com/guilhermezuriel/git-resume/internal/git"
	"github.com/guilhermezuriel/git-resume/internal/storage"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all summaries for the current repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !gitpkg.IsGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}

		info, err := gitpkg.GetRepoInfo()
		if err != nil {
			return err
		}

		repoDir := storage.RepoDirFor(info.ID)
		files, err := storage.ListResumes(repoDir)
		if err != nil {
			return err
		}

		printHeader(fmt.Sprintf("Summaries: %s", info.Name))

		if len(files) == 0 {
			fmt.Println("  [INFO] No summaries found for this repository")
			fmt.Println()
			fmt.Println("  Generate your first summary with:")
			fmt.Println("    git-resume")
			fmt.Println("    git-resume --enrich")
			fmt.Println()
			return nil
		}

		fmt.Println()
		fmt.Fprintf(os.Stdout, "  %-44s %s\n", "FILE", "MODIFIED")
		fmt.Fprintf(os.Stdout, "  %s\n", repeat("─", 65))

		for _, f := range files {
			fmt.Fprintf(os.Stdout, "  %-44s %s\n",
				f.Name,
				f.ModTime.Format("2006-01-02 15:04"),
			)
		}
		fmt.Println()
		return nil
	},
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func printHeader(title string) {
	width := 60
	fmt.Printf("\n┌%s┐\n", repeat("─", width))
	padding := (width - len(title) - 2) / 2
	line := fmt.Sprintf("│%s%s%s│",
		spaces(padding),
		title,
		spaces(width-padding-len(title)),
	)
	fmt.Println(line)
	fmt.Printf("└%s┘\n", repeat("─", width))
}

func spaces(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += " "
	}
	return s
}
