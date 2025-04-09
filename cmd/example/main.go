package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/scagogogo/sonatype-central-sdk/pkg/api"
	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

func main() {
	// Command line flags
	groupID := flag.String("g", "", "GroupID to search for")
	artifactID := flag.String("a", "", "ArtifactID to search for")
	className := flag.String("c", "", "Class name to search for")
	sha1 := flag.String("sha1", "", "SHA1 checksum to search for")
	limit := flag.Int("limit", 10, "Maximum number of results to return")
	flag.Parse()

	// Create client
	client := api.NewClient()

	// Context for API calls
	ctx := context.Background()

	// Perform search based on provided parameters
	if *groupID != "" && *artifactID == "" {
		// Search by group ID
		fmt.Printf("Searching for artifacts with GroupID: %s\n", *groupID)
		results, err := client.SearchByGroupId(ctx, *groupID, *limit)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printArtifacts(results)
	} else if *artifactID != "" && *groupID == "" {
		// Search by artifact ID
		fmt.Printf("Searching for artifacts with ArtifactID: %s\n", *artifactID)
		results, err := client.SearchByArtifactId(ctx, *artifactID, *limit)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printArtifacts(results)
	} else if *groupID != "" && *artifactID != "" {
		// Advanced search for both group and artifact
		fmt.Printf("Searching for artifacts with GroupID: %s and ArtifactID: %s\n", *groupID, *artifactID)
		searchReq := request.NewAdvancedSearchOptions().
			SetGroupId(*groupID).
			SetArtifactId(*artifactID)

		results, err := client.AdvancedSearch(ctx, searchReq, *limit)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printArtifacts(results)
	} else if *className != "" {
		// Search by class name
		fmt.Printf("Searching for artifacts containing class: %s\n", *className)
		results, err := client.SearchByClassName(ctx, *className, *limit)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d artifacts\n", len(results))
		for _, v := range results {
			fmt.Printf("%s:%s:%s (%s)\n", v.GroupId, v.ArtifactId, v.Version, v.Packaging)
		}
	} else if *sha1 != "" {
		// Search by SHA1
		fmt.Printf("Searching for artifact with SHA1: %s\n", *sha1)
		results, err := client.SearchBySha1(ctx, *sha1, *limit)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d artifacts\n", len(results))
		for _, v := range results {
			fmt.Printf("%s:%s:%s (%s)\n", v.GroupId, v.ArtifactId, v.Version, v.Packaging)
		}
	} else {
		fmt.Println("Please provide search parameters. Use -h for help.")
		flag.Usage()
		os.Exit(1)
	}
}

func printArtifacts(artifacts []*response.Artifact) {
	fmt.Printf("Found %d artifacts\n", len(artifacts))
	for _, a := range artifacts {
		fmt.Printf("%s:%s:%s (%s)\n", a.GroupId, a.ArtifactId, a.LatestVersion, a.Packaging)
	}
}
