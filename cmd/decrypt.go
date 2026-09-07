package cmd

import (
	"fmt"
	"os"

	envcrypto "github.com/BLemine/envguard/internal/crypto"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	decryptLocal      string
	decryptOut        string
	decryptPassphrase string
	decryptForce      bool
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt a .env.enc file back into a .env file",
	Long:  `Decrypts an AES-256-GCM encrypted .env.enc file using a passphrase. Fails loudly if the passphrase is wrong. Will not overwrite an existing .env unless --force is set.`,
	Example: `  envguard decrypt --passphrase=mysecret
  envguard decrypt --local=.env.staging.enc --out=.env.staging --passphrase=mysecret
  envguard decrypt --passphrase=mysecret --force`,
	Run: func(cmd *cobra.Command, args []string) {
		pass := resolvePassphrase(decryptPassphrase)
		if pass == "" {
			fmt.Fprintln(os.Stderr, "✗ passphrase required: use --passphrase or set ENVGUARD_PASSPHRASE")
			os.Exit(1)
		}

		if !decryptForce {
			if _, err := os.Stat(decryptOut); err == nil {
				fmt.Fprintf(os.Stderr, "✗ %s already exists — use --force to overwrite\n", decryptOut)
				os.Exit(1)
			}
		}

		data, err := os.ReadFile(decryptLocal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Error reading %s: %v\n", decryptLocal, err)
			os.Exit(1)
		}

		plaintext, err := envcrypto.Decrypt(data, pass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

		if err := writeDecryptedFile(decryptOut, plaintext, decryptForce); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Error writing %s: %v\n", decryptOut, err)
			os.Exit(1)
		}

		color.New(color.FgGreen).Printf("✓ %s decrypted → %s\n", decryptLocal, decryptOut)
	},
}

func init() {
	decryptCmd.Flags().StringVar(&decryptLocal, "local", ".env.enc", "Path to the encrypted file to decrypt")
	decryptCmd.Flags().StringVar(&decryptOut, "out", ".env", "Path to write the decrypted output")
	decryptCmd.Flags().StringVar(&decryptPassphrase, "passphrase", "", "Passphrase for decryption (or set ENVGUARD_PASSPHRASE)")
	decryptCmd.Flags().BoolVar(&decryptForce, "force", false, "Overwrite the output file if it already exists")
}

// Restrict permissions before writing secrets, including when replacing a file.
func writeDecryptedFile(path string, plaintext []byte, force bool) error {
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if force {
		flags = os.O_WRONLY | os.O_CREATE
	}
	f, err := os.OpenFile(path, flags, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Chmod(0600); err != nil {
		return err
	}
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Write(plaintext); err != nil {
		return err
	}
	return f.Close()
}
