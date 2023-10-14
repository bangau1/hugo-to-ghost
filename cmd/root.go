/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bangau1/hugo-to-ghost/pkg"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hugo-to-ghost",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: cmdRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	contentDir  string
	isRecursive bool = false // TODO(bangau1): to support recursive mode
)

func init() {
	rootCmd.Flags().StringVar(&contentDir, "contentDir", "", "The content directory where the frontmatter posts are located. Example: 'content/english/posts/'")
	// TODO(bangau1): add recursive mode

	rootCmd.MarkFlagRequired("contentDir")
}

func cmdRun(cmd *cobra.Command, args []string) {

	if !isRecursive {
		files, err := os.ReadDir(contentDir)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			// if the file is a markdown file (.md)
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				mdFilePath, err := filepath.Abs(contentDir + "/" + file.Name())
				if err != nil {
					log.Fatal(err)
				}

				post, err := pkg.NewPostFromFrontMatterDocFile(mdFilePath)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("%+v\n", post)
			}
		}
	} else {
		// TODO(bangau1)
	}
}
