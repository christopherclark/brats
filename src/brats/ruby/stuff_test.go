package ruby_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Masterminds/semver"
	"github.com/cloudfoundry/libbuildpack"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Basic example of how to clone a repository using clone options.
func GitGet(directory, language, branch string) (string, error) {
	// directory, err := ioutil.TempDir("", fmt.Sprintf("%s-buildpack", language))
	// if err != nil {
	// 	return "", err
	// }

	r, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:           fmt.Sprintf("https://github.com/cloudfoundry/%s-buildpack", language),
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return "", err
	}

	ref, err := r.Head()
	if err != nil {
		return "", err
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return "", err
	}

	return commit.String(), err
}

func GetSortedDepVersions(dep string) ([]string, error) {
	manifest, _ := libbuildpack.NewManifest(os.Getenv("BP_DIR"), libbuildpack.NewLogger(os.Stdout), time.Now())
	versions := manifest.AllDependencyVersions(dep)

	vs := make([]*semver.Version, len(versions))
	for i, v := range versions {
		var err error
		if vs[i], err = semver.NewVersion(v); err != nil {
			return []string{}, err
		}
	}
	sort.Sort(semver.Collection(vs))

	for i, _ := range vs {
		versions[i] = vs[i].Original()
	}

	return versions, nil
}

func AddDotProfileScriptToApp(templatePath string) error {
	return ioutil.WriteFile(filepath.Join(templatePath, ".profile"), []byte(
		`#!/usr/bin/env bash

echo PROFILE_SCRIPT_IS_PRESENT_AND_RAN

BASHCODE
`), 0755)
}
