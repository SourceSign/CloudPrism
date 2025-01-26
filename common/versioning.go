package common

import (
	"strings"
	"time"

	"github.com/apex/log"
	"gopkg.in/src-d/go-git.v4"
)

func LatestGitHash(path string) (string, bool) {
	res := ""
	clean := false

	log.WithFields(log.Fields{
		"path": path,
	}).Debug("Checking if path is under Git control")

	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return res, clean
	}

	ref, err := repo.Head()
	if err != nil {
		return res, clean
	}

	wt, err := repo.Worktree()
	if err != nil {
		return res, clean
	}

	status, err := wt.Status()
	if err != nil {
		return res, clean
	}

	clean = status.IsClean()

	// get the hash shortened
	res = ref.Hash().String()[:8]

	return res, clean
}

// Returns a timestamp in a filename-friendly version of the RFC3339Nano format.
func TimeStamp(t time.Time) string {
	return strings.ReplaceAll(strings.ReplaceAll(t.UTC().Format(time.RFC3339), ":", ""), "-", "")
}

func GetDeploymentTargetName(path string) string {
	timestamp := TimeStamp(time.Now())

	hash, clean := LatestGitHash(path)

	// if we have a hash and it is clean, just use the hash
	if clean {
		log.WithFields(log.Fields{
			"path":    path,
			"target":  hash,
			"isClean": clean,
		}).Debug("DeploymentTarget: clean Git hash")

		return hash
	}

	// append a timestamp to dirty hashes
	if hash != "" {
		log.WithFields(log.Fields{
			"path":    path,
			"target":  hash + "-" + timestamp,
			"isClean": clean,
		}).Debug("DeploymentTarget: Git hash plus timestamp")

		return hash + "-" + timestamp
	}

	// use a random name + timestamp otherwise
	randomName := GetRandomName() + "-" + timestamp

	log.WithFields(log.Fields{
		"path":   path,
		"target": randomName,
	}).Debug("DeploymentTarget: random name")

	return randomName
}
