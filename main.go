package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
)

var usage = `
usage: vault-policy-backup [-quiet|--quiet] [-file|--file <vault-k8s-auth-backup-json-file-path>] [<vault-policy-name>]

Usage of vault-policy-backup:

  -file / --file string
      vault policy backup json file path (default "vault_policy_backup.json")

  -quiet / --quiet
      quiet progress (default false).
      By default vault-policy-backup CLI will show a lot of details
      about the backup process and detailed progress during the
      backup process

  -h / -help / --help
      show help

examples:

# show help
vault-policy-backup -h

# show help
vault-policy-backup --help

# backs up all vault policies from vault
# except the root policy
vault-policy-backup

# backs up allow_read policy from vault.
# it will throw an error if it does not exist
vault-policy-backup allow_read

# quietly backup all vault policies.
# this will just show dots (.) for progress
vault-policy-backup -quiet

# OR you can use --quiet too instead of -quiet

vault-policy-backup --quiet
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s", usage)
		os.Exit(0)
	}
	quietProgress := flag.Bool("quiet", false, "quiet progress")
	vaultPolicyBackupJsonFileName := flag.String("file", "vault_policy_backup.json", "vault policy backup json file path")
	flag.Parse()

	if !(flag.NArg() == 1 || flag.NArg() == 0) {
		fmt.Fprintf(os.Stderr, "invalid number of arguments: %d. expected 0 or 1 arguments.\n\n", flag.NArg())
		flag.Usage()
	}

	config := api.DefaultConfig()
	client, err := api.NewClient(config)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating vault client: %s\n", err)
		os.Exit(1)
	}

	policies := []string{}

	if flag.NArg() == 0 {
		allVaultPolicies, err := client.Sys().ListPolicies()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error listing vault policies: %s\n", err)
			os.Exit(1)
		}
		policies = append(policies, allVaultPolicies...)
	} else {
		vaultPolicyName := flag.Args()[0]
		policies = append(policies, vaultPolicyName)
	}

	vaultPolicyBackup := backupVaultPolicies(client, policies, *quietProgress)

	vaultPolicyBackupJSON, err := convertVaultPolicyBackupToJSON(vaultPolicyBackup)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error converting vault policy backup to json: %s\n", err)
		os.Exit(1)
	}
	err = writeToFile(vaultPolicyBackupJSON, *vaultPolicyBackupJsonFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing vault policy backup to json file: %s\n", err)
		os.Exit(1)
	}
}
