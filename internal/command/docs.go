package command

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	docsFormat       string = "markdown"
	docsValidFormats        = []string{"markdown", "yaml", "manpage"}
	docsOutputDir    string = "./docs"
)

func init() {
	if flags := docsCmd.Flags(); flags != nil {
		flags.StringVarP(&docsFormat, "format", "f", docsFormat, fmt.Sprintf("desired format for generated documents {%s}", strings.Join(docsValidFormats, ", ")))
		flags.StringVarP(&docsOutputDir, "output-dir", "o", docsOutputDir, "output directory for generated documents")
	}

	rootCmd.AddCommand(docsCmd)
}

var (
	docsCmd = &cobra.Command{
		Use:   "docs",
		Short: "Generate documentation for Gotracker",
		Long:  `Generate the documentation for Gotracker in one of many different formats.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch docsFormat {
			case "markdown", "md":
				if err := os.MkdirAll(docsOutputDir, 0755); err != nil {
					return err
				}
				return doc.GenMarkdownTree(rootCmd, docsOutputDir)
			case "yaml", "yml":
				if err := os.MkdirAll(docsOutputDir, 0755); err != nil {
					return err
				}
				return doc.GenYamlTree(rootCmd, docsOutputDir)
			case "manpage", "man":
				if err := os.MkdirAll(docsOutputDir, 0755); err != nil {
					return err
				}
				manHead := &doc.GenManHeader{
					Title:   "GOTRACKER",
					Section: "3",
				}
				return doc.GenManTree(rootCmd, manHead, docsOutputDir)
			default:
				return errors.New("unsupported document format")
			}
		},
	}
)
