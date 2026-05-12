package cmd

import (
	"fmt"
	"os"

	"github.com/MescoCzubinski/skill-cli/core"
)

func Remote(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli remote <url>")
		os.Exit(1)
	}
	remoteURL := args[0]

	err := core.ValidateRemoteURL(remoteURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.GitAvailable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	isRepo := core.IsGitRepo()
	if isRepo {
		err = updateExistingRemote(remoteURL)
	} else {
		err = initFreshRepo(remoteURL)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.SyncSkillFiles()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("set remote: %s\n", remoteURL)
}

func initFreshRepo(remoteURL string) error {
	err := core.GitInitRepo()
	if err != nil {
		return err
	}

	err = core.GitAddOrigin(remoteURL)
	if err != nil {
		return err
	}

	remoteEmpty, err := core.IsRemoteEmpty()
	if err != nil {
		return err
	}

	hasLocal, err := core.HasLocalChanges()
	if err != nil {
		return err
	}

	if remoteEmpty {
		return seedEmptyRemote()
	}

	err = core.GitFetchOrigin()
	if err != nil {
		return err
	}

	if hasLocal {
		return mergeLocalIntoRemote()
	}

	return core.GitCheckoutTrackMain()
}

func seedEmptyRemote() error {
	err := core.GitAddAll()
	if err != nil {
		return err
	}

	err = core.GitCommitAllowEmpty("init")
	if err != nil {
		return err
	}

	err = core.GitBranchMain()
	if err != nil {
		return err
	}

	return core.GitPushMain()
}

func mergeLocalIntoRemote() error {
	err := core.GitAddAll()
	if err != nil {
		return err
	}

	err = core.GitCommit("local")
	if err != nil {
		return err
	}

	err = core.GitMergeTheirsMain()
	if err != nil {
		return err
	}

	err = core.GitBranchMain()
	if err != nil {
		return err
	}

	err = core.GitSetUpstreamMain()
	if err != nil {
		return err
	}

	return core.GitPushMain()
}

func updateExistingRemote(remoteURL string) error {
	err := core.GitSetOrigin(remoteURL)
	if err != nil {
		return err
	}

	empty, err := core.IsRemoteEmpty()
	if err != nil {
		return err
	}

	if empty {
		return core.GitPushMain()
	}
	return core.GitPull()
}
