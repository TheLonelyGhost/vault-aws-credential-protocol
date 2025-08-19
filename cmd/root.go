package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

var (
	VersionSuffix string

	vaultMount     string
	vaultRole      string
	skipCache      bool
	stsSessionName string

	rootCmd = &cobra.Command{
		Use:     "vault-aws-credential-protocol --mount VAULT_MOUNT_PATH --role VAULT_ROLE_NAME [--no-cache] [--session-name SESSION_NAME]",
		Version: Version,
		Short:   "Grab AWS creds from Vault and expose them in the AWS Credential Protocol format",

		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()
			path := fmt.Sprintf("%s/sts/%s", vaultMount, vaultRole)
			run(ctx, path, stsSessionName, !skipCache)
			return
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&vaultMount, "mount", "m", "", "Path to where the AWS secrets engine is mounted. If located outside of the root namespace, prefix the path with that namespace.")
	rootCmd.MarkFlagRequired("mount")
	rootCmd.Flags().StringVarP(&vaultRole, "role", "r", "", "Name of the desired role on the given AWS secrets engine")
	rootCmd.MarkFlagRequired("role")
	rootCmd.Flags().BoolVar(&skipCache, "no-cache", false, "Skip reading creds from, and writing creds to cache")
	rootCmd.Flags().StringVar(&stsSessionName, "session-name", "vault-aws-credential-protocol", "session-name")
}

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
