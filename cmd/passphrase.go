package cmd

import "os"

func resolvePassphrase(flag string) string {
	if flag != "" {
		return flag
	}
	return os.Getenv("ENVGUARD_PASSPHRASE")
}
