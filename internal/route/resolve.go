package route

import (
	"errors"
)

// Resolve determines the worker ID based on override, relay configuration, and origin.
func Resolve(override, relayWorker string, originGitHub bool) (string, error) {
	if override != "" && override != "jules" && override != "local" && override != "auto" {
		return "", errors.New("invalid override")
	}

	var resolved string

	if override == "jules" || override == "local" {
		resolved = override
	} else if override == "auto" {
		if originGitHub {
			resolved = "jules"
		} else {
			resolved = "local"
		}
	} else if relayWorker == "jules" || relayWorker == "local" {
		resolved = relayWorker
	} else {
		if originGitHub {
			resolved = "jules"
		} else {
			resolved = "local"
		}
	}

	if resolved != "jules" && resolved != "local" {
		return "", errors.New("invalid resolved worker")
	}

	return resolved, nil
}
