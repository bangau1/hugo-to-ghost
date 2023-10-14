/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
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
	Short: "A simple CLI that helps migrating contents from Hugo (Markdown) to Ghost",
	Long: `A simple CLI that helps migrating contents from Hugo (Markdown) to Ghost.

This CLI program will output json file that can later be imported to Ghost Import functionality (via Admin dashboard).

For example:
- hugo-to-ghost --contentDir ./content/english/post > ghost-content.json
	to output the JSON file containing all Hugo's posts that conformed with the Ghost Importing Content format:
- hugo-to-ghost --contentDir ./content/english/post --staticContentPrefixChanges "/img/uploads/,/content/images/hugo/"
	to change the image assets in Hugo (that previously located in /img/uploads/ path) to /content/images/hugo folder inside the ghost installation
`,
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
	contentDir                 string
	isRecursive                bool = false // TODO(bangau1): to support recursive mode
	staticContentPrefixChanges string

	deduplicatePosts bool
	ghostAdminAPIKey string
	ghostURL         string
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	rootCmd.Flags().StringVar(&contentDir, "contentDir", "", "The content directory where the frontmatter posts are located. Example: 'content/english/posts/'")
	rootCmd.Flags().StringVar(&staticContentPrefixChanges, "staticContentPrefixChanges", "", "To change the static content prefix. Example: /img/uploads/,/content/images/hugo/ ")
	rootCmd.Flags().BoolVar(&deduplicatePosts, "deduplicate", false, "Set it true if you want to detect the same post in Ghost. Ghost doesn't do deduplication functionality, hence if you don't set it and import the JSON file multiple times, it will create new post everytime (instead of updating it)")
	rootCmd.Flags().StringVar(&ghostAdminAPIKey, "ghostAdminAPIKey", "", "The Ghost's Admin API Key. It's only needed if you set --deduplicate flag")
	rootCmd.Flags().StringVar(&ghostURL, "ghostUrl", "http://localhost:8080", "The Ghost's URL It's only needed if you set --deduplicate flag")
	// TODO(bangau1): add recursive mode

	rootCmd.MarkFlagRequired("contentDir")
}

func cmdRun(cmd *cobra.Command, args []string) {
	staticContentPrefixChangesRules := make([]string, 0)
	if len(staticContentPrefixChanges) > 0 {
		staticContentPrefixChangesRules = strings.Split(staticContentPrefixChanges, ",")
	}
	if len(staticContentPrefixChangesRules)%2 > 0 {
		log.Fatal("staticContentPrefixChanges is invalid.")
	}
	var ghostApi pkg.GhostAdminAPI
	if deduplicatePosts {
		if ghostAdminAPIKey == "" {
			log.Fatal("--ghostAdminAPIKey is required if --deduplicate is set")
		}
		ghostApi = pkg.NewGhostAdminAPI(ghostURL, ghostAdminAPIKey)
	}

	if !isRecursive {
		files, err := os.ReadDir(contentDir)
		if err != nil {
			log.Fatal(err)
		}
		ghostContents := make([]pkg.GhostContent, 0)
		for _, file := range files {
			// if the file is a markdown file (.md)
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				mdFilePath, err := filepath.Abs(contentDir + "/" + file.Name())
				if err != nil {
					log.Fatal(err)
				}

				// read the file and convert it to MarkdownPost
				post, err := pkg.NewPostFromFrontMatterDocFile(mdFilePath)
				if err != nil {
					log.Fatal(err)
				}

				// apply changes to the post's static content prefix path
				for i := 0; i < len(staticContentPrefixChangesRules); i += 2 {
					post.ChangeStaticContentPrefix(staticContentPrefixChangesRules[i], staticContentPrefixChangesRules[i+1])
				}

				// convert MarkdownPost to GhostContent
				ghostContent, err := pkg.NewGhostContentFromMarkdownPost(post)
				if err != nil {
					log.Fatal(err)
				}

				if deduplicatePosts && post.Slug != "" {
					existingPost, err := ghostApi.GetPostBySlug(context.Background(), ghostContent.Slug)
					if err != nil {
						if !errors.Is(err, pkg.ErrNotFound) {
							log.Fatal(err)
						}
					}

					// override the id of ghostContent to existingPost.id
					ghostContent.Id = existingPost.Id
					ghostContent.UUID = existingPost.UUID
				}
				ghostContents = append(ghostContents, ghostContent)
			}
		}

		if len(ghostContents) > 0 {
			// TODO(bangau1): the ghost import tool can't deduplicate the content
			// So we may need to fetch the ghost's content first and deduplicate it here
			// by assigning the same id/uuid if we detect there is already the same title content there.
			importData := pkg.NewGhostImportData(ghostContents...)
			jsonData, err := importData.ToJson()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(jsonData)
		} else {
			log.Println("WARN: no content being processed")
		}
	} else {
		// TODO(bangau1)
	}
}
