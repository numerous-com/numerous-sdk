package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var (
	listDir         string
	listOrgSlug     string
	listCollection  string
	listRemote      bool
	listAPIURL      string
	listAccessToken string
)

var listCmd = &cobra.Command{
	Use:   "list [directory]",
	Short: "List tasks in a task collection or collections in an organization",
	Long: `List all available tasks in a task collection or all collections in an organization.

By default, lists remote collections and tasks from the Numerous platform.
Use --local to list tasks from a local task collection directory.

Organization is required to identify which workspace to use.

Examples:
  numerous task list --org my-org                           # List all collections in organization
  numerous task list --org my-org --collection my-tasks    # List tasks in specific collection
  numerous task list --org my-org --local ./my-tasks       # List tasks in local directory`,
	RunE: listTasksCmd,
}

func init() {
	// Organization is mandatory
	listCmd.Flags().StringVarP(&listOrgSlug, "org", "", "", "Organization slug (required)")
	listCmd.MarkFlagRequired("org")

	// Collection is optional
	listCmd.Flags().StringVarP(&listCollection, "collection", "c", "", "Task collection name (optional)")

	// Local vs remote execution
	listCmd.Flags().BoolVar(&listRemote, "remote", true, "List remote collections/tasks (default)")
	listCmd.Flags().BoolVar(&localExecution, "local", false, "List local task collection")
	listCmd.Flags().StringVarP(&listDir, "dir", "d", ".", "Directory containing the task collection (local mode only)")

	// API configuration
	listCmd.Flags().StringVar(&listAPIURL, "api-url", "", "API endpoint URL (defaults to environment variable)")
	listCmd.Flags().StringVar(&listAccessToken, "token", "", "Access token (defaults to environment variable)")

	listCmd.MarkFlagDirname("dir")
}

func listTasksCmd(cmd *cobra.Command, args []string) error {
	// Determine directory for local mode
	if len(args) > 0 && localExecution {
		listDir = args[0]
	}

	if localExecution {
		return listLocalTasks()
	} else {
		return listRemoteTasks()
	}
}

func listLocalTasks() error {
	// Load task manifest from local directory
	manifestPath := filepath.Join(listDir, "numerous-task.toml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("no task manifest found at %s", manifestPath)
	}

	var manifest TaskManifest
	if _, err := toml.DecodeFile(manifestPath, &manifest); err != nil {
		return fmt.Errorf("failed to parse task manifest: %w", err)
	}

	return listTasks(&manifest)
}

func listRemoteTasks() error {
	// Get API configuration
	if listAPIURL == "" {
		listAPIURL = os.Getenv("NUMEROUS_API_URL")
		if listAPIURL == "" {
			listAPIURL = "http://localhost:8080/graphql" // Default for local development
		}
	}

	if listAccessToken == "" {
		listAccessToken = os.Getenv("NUMEROUS_ACCESS_TOKEN")
		if listAccessToken == "" {
			fmt.Println("Warning: No access token found. Set NUMEROUS_ACCESS_TOKEN or use --token flag.")
		}
	}

	if listCollection == "" {
		// List all collections in the organization
		return listCollectionsInOrganization(listOrgSlug, listAPIURL, listAccessToken)
	} else {
		// List tasks in a specific collection
		return listTasksInCollection(listOrgSlug, listCollection, listAPIURL, listAccessToken)
	}
}

func listCollectionsInOrganization(orgSlug, apiURL, accessToken string) error {
	fmt.Printf("🏢 Task Collections in Organization: %s\n\n", orgSlug)

	// TODO: Implement GraphQL query to list collections
	// For now, show what the command would do
	fmt.Printf("Would query API at: %s\n", apiURL)
	fmt.Printf("Organization: %s\n", orgSlug)
	fmt.Printf("\n📋 Available Collections:\n")
	fmt.Printf("  (This will be implemented with GraphQL query)\n")
	fmt.Printf("\nTo list tasks in a specific collection:\n")
	fmt.Printf("  numerous task list --org %s --collection <collection-name>\n", orgSlug)

	return nil
}

func listTasksInCollection(orgSlug, collectionName, apiURL, accessToken string) error {
	fmt.Printf("📦 Tasks in Collection: %s (Organization: %s)\n\n", collectionName, orgSlug)

	// TODO: Implement GraphQL query to list tasks in collection
	// For now, show what the command would do
	fmt.Printf("Would query API at: %s\n", apiURL)
	fmt.Printf("Organization: %s\n", orgSlug)
	fmt.Printf("Collection: %s\n", collectionName)
	fmt.Printf("\n📋 Available Tasks:\n")
	fmt.Printf("  (This will be implemented with GraphQL query)\n")
	fmt.Printf("\nTo run a task:\n")
	fmt.Printf("  numerous task run <task-name> --org %s --collection %s\n", orgSlug, collectionName)

	return nil
}
