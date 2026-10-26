package grok

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

const CouncilMemberProfile = "[profiles.council-member]"

func EnsureSandboxProfile() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	grokDir := filepath.Join(home, ".grok")
	if err := os.MkdirAll(grokDir, 0755); err != nil {
		return err
	}
	sandboxPath := filepath.Join(grokDir, "sandbox.toml")
	data, err := os.ReadFile(sandboxPath)
	if err == nil {
		if bytes.Contains(data, []byte(CouncilMemberProfile)) {
			return nil
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	optmemDir := filepath.Join(home, ".local", "share", "grok-agy", "optmem")
	entry := fmt.Sprintf("%s\nextends = \"read-only\"\nread_write = [\"%s\"]\n",
		CouncilMemberProfile, optmemDir)

	if len(data) > 0 && !bytes.HasSuffix(data, []byte("\n")) {
		entry = "\n" + entry
	}

	f, err := os.OpenFile(sandboxPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if len(data) == 0 {
		_ = os.Chmod(sandboxPath, 0600)
	}

	if _, err := f.WriteString(entry); err != nil {
		return err
	}
	return nil
}
