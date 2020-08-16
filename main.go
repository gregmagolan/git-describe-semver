package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-git/go-git/v5"
)

// GenerateVersionOptions ...
type GenerateVersionOptions struct {
	FallbackTagName string
	DropTagNamePrefix bool
}

// GenerateVersion ...
func GenerateVersion(tagName string, counter int, headHash string, opts GenerateVersionOptions) (*string, error) {
	devPreRelease := []string{"dev", strconv.Itoa(counter), "g" + (headHash)[0:7]}
	if tagName == "" {
		version := SemVerParse(opts.FallbackTagName)
		if version == nil {
			return nil, fmt.Errorf("unable to parse fallback tag")
		}
		version.PreRelease = devPreRelease
		if opts.DropTagNamePrefix {
			version.Prefix = ""
		}
		result := version.String()
		return &result, nil
	}
	version := SemVerParse(tagName)
	if version == nil {
		return nil, fmt.Errorf("unable to parse tag")
	}
	if counter == 0 {
		result := version.String()
		return &result, nil
	}
	if len(version.PreRelease) > 0 {
		version = &SemVer{
			Prefix:        version.Prefix,
			Major:         version.Major,
			Minor:         version.Minor,
			Patch:         version.Patch,
			PreRelease:    append(version.PreRelease, devPreRelease...),
			BuildMetadata: append([]string{}, version.BuildMetadata...),
		}
	} else {
		version = &SemVer{
			Prefix:        version.Prefix,
			Major:         version.Major,
			Minor:         version.Minor,
			Patch:         version.Patch + 1,
			PreRelease:    devPreRelease,
			BuildMetadata: append([]string{}, version.BuildMetadata...),
		}
	}
	if opts.DropTagNamePrefix {
		version.Prefix = ""
	}
	result := version.String()
	return &result, nil
}

// Run ...
func Run(dir string, opts GenerateVersionOptions) (*string, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to open git repository: %v", err)
	}
	tagName, counter, headHash, err := GitDescribe(*repo)
	if err != nil {
		return nil, fmt.Errorf("unable to describe commit: %v", err)
	}
	result, err := GenerateVersion(*tagName, *counter, *headHash, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to generate version: %v", err)
	}
	return result, nil
}

func main() {
	fallback := flag.String("fallback", "", "The first version to fallback to should there be no tag")
	dropPrefix := flag.Bool("drop-prefix", false, "Drop prefix from output")
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to determine current directory: %v\n", err)
	}
	opts := GenerateVersionOptions{
		FallbackTagName: *fallback,
		DropTagNamePrefix: *dropPrefix,
	}
	result, err := Run(dir, opts)
	if err != nil {
		log.Fatalf("unable to generate version: %v\n", err)
	}
	fmt.Println(*result)
}
