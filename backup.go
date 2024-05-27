package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
)

type VaultPolicyBackup struct {
	Policies VaultPolicies `json:"policies"`
}

type VaultPolicies []VaultPolicy

type VaultPolicy struct {
	Name   string `json:"name"`
	Policy string `json:"policy"`
}

func convertVaultPolicyBackupToJSON(vaultPolicyBackup VaultPolicyBackup) ([]byte, error) {
	vaultPolicyBackupJSON, err := toJSON(vaultPolicyBackup)
	if err != nil {
		return nil, err
	}
	return vaultPolicyBackupJSON, nil
}

func backupVaultPolicies(client *api.Client, policyNames []string, quietProgress bool) VaultPolicyBackup {
	vaultPolicies := make(VaultPolicies, 0)

	fmt.Fprintf(os.Stdout, "\nbacking up the following vault policies in vault: %+v\n", policyNames)

	// Note: Ignore root policy as there's nothing to backup in it

	// Backup all Vault policies
	for _, policyName := range policyNames {
		if policyName == "root" {
			// Ignore root policy as it has no content
			continue
		}
		policy := getVaultPolicy(client, policyName)

		if quietProgress {
			fmt.Fprintf(os.Stdout, ".")
		} else {
			fmt.Fprintf(os.Stdout, "\nbacking up `%s` vault policy\n", policyName)
			fmt.Fprintf(os.Stdout, "\n`%s` vault policy rules: %+v\n", policyName, policy)
		}

		vaultPolicies = append(vaultPolicies, VaultPolicy{Name: policyName, Policy: policy})
	}

	fmt.Fprintf(os.Stdout, "\n")

	return VaultPolicyBackup{Policies: vaultPolicies}
}

func getVaultPolicy(client *api.Client, policyName string) string {
	vaultPolicy, err := client.Sys().GetPolicy(policyName)
	// TODO: Think on this.
	// Note: We are abruptly stopping here, if we cannot read the policy,
	// instead of letting the caller handle the error.
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading '%s' vault policy from vault: %s\n", policyName, err)
		os.Exit(1)
	}
	return vaultPolicy
}
