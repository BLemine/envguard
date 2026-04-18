package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	envcrypto "github.com/yourusername/envguard/internal/crypto"
)

var (
	encryptLocal      string
	encryptOut        string
	encryptPassphrase string
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt a .env file into an AES-256-GCM encrypted .env.enc",
	Long:  `Encrypts your .env file using AES-256-GCM with a key derived from your passphrase via Argon2id. The passphrase is never stored.`,
	Example: `  envguard encrypt --passphrase=mysecret
  envguard encrypt --local=.env.staging --out=.env.staging.enc --passphrase=mysecret
  ENVGUARD_PASSPHRASE=mysecret envguard encrypt`,
	Run: func(cmd *cobra.Command, args []string) {
		pass := resolvePassphrase(encryptPassphrase)
		if pass == "" {
			fmt.Fprintln(os.Stderr, "✗ passphrase required: use --passphrase or set ENVGUARD_PASSPHRASE")
			os.Exit(1)
		}

		plaintext, err := os.ReadFile(encryptLocal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Error reading %s: %v\n", encryptLocal, err)
			os.Exit(1)
		}

		ciphertext, err := envcrypto.Encrypt(plaintext, pass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Encryption failed: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(encryptOut, ciphertext, 0600); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Error writing %s: %v\n", encryptOut, err)
			os.Exit(1)
		}

		color.New(color.FgGreen).Printf("✓ %s encrypted → %s (share this with your team)\n", encryptLocal, encryptOut)
	},
}

func init() {
	encryptCmd.Flags().StringVar(&encryptLocal, "local", ".env", "Path to the .env file to encrypt")
	encryptCmd.Flags().StringVar(&encryptOut, "out", ".env.enc", "Path to write the encrypted output")
	encryptCmd.Flags().StringVar(&encryptPassphrase, "passphrase", "", "Passphrase for encryption (or set ENVGUARD_PASSPHRASE)")
}
