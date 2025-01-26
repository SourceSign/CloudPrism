package common

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/apex/log"
)

type DefaultStateStore struct {
	// The name of the name to use for the state store.
	name string
	// The base path to the directory to use for the state store folder.
	path string
}

func GetDefaultStateStore(path, name string) (StateStore, error) {
	log.WithFields(log.Fields{
		"path": path,
		"name": name,
	}).Debug("DefaultStateStore.GetDefaultStateStore()")

	if path == "" {
		path = "."
	}

	if name == "" {
		name = ".statestore"
	}

	return &DefaultStateStore{
		name: name,
		path: path,
	}, nil
}

// Creates and/or logs in to a state store and returns its URL string.
func (ss *DefaultStateStore) StoreOpen() (string, error) {
	statePath := path.Join(ss.path, ss.name)
	absPath, _ := filepath.Abs(statePath)
	stateURI := "file://" + absPath

	log.WithField("absPath", absPath).Debug("DefaultStateStore.StoreOpen()")

	log.WithField("stateUri", stateURI).Debug("DefaultStateStore.StoreOpen()")

	err := os.MkdirAll(statePath, os.ModePerm)
	if err != nil {
		log.WithField("statePath", statePath).WithError(err).Error("DefaultStateStore create directory failed")

		return "", err
	}

	// TODO: get this from the config
	err = os.Setenv("PULUMI_CONFIG_PASSPHRASE", "we need a persistent token for the project here")
	if err != nil {
		log.WithField("statePath", statePath).WithError(err).Error("DefaultStateStore set env failed")

		return "", err
	}

	cmd := exec.Command("pulumi", "login", path.Clean(stateURI)) // #nosec G204
	pwd, _ := os.Getwd()
	cmd.Dir = pwd
	_, err = cmd.Output()

	if err != nil {
		log.WithField("stateUri", stateURI).WithError(err).Error("DefaultStateStore login failed")

		return "", err
	}

	return stateURI, nil
}

// Closes or logs out of a state store, without deleting any data.
func (ss *DefaultStateStore) StoreClose() error {
	statePath := path.Join(ss.path, ss.name)
	absPath, _ := filepath.Abs(statePath)
	stateURI := "file://" + absPath

	log.WithField("stateUri", stateURI).Debug("DefaultStateStore.StoreClose()")

	cmd := exec.Command("pulumi", "logout", path.Clean(stateURI)) // #nosec G204
	cmd.Dir = path.Dir(statePath)

	if _, err := cmd.Output(); err != nil {
		log.WithField("stateUri", stateURI).WithError(err).Error("DefaultStateStore logout failed")

		return err
	}

	return nil
}

// Deletes the state store, including all data when the force parameter is true.
func (ss *DefaultStateStore) StoreDelete(force bool) error {
	statePath := path.Join(ss.path, ss.name)

	log.WithField("statePath", statePath).Debug("DefaultStateStore.StoreDelete()")

	if force {
		err := os.RemoveAll(statePath)
		if err != nil {
			log.WithField("statePath", statePath).WithError(err).Error("DefaultStateStore remove all failed")

			return err
		}
	} else {
		err := os.Remove(statePath)
		if err != nil {
			log.WithField("statePath", statePath).WithError(err).Error("DefaultStateStore remove failed")

			return err
		}
	}

	return nil
}
