package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/thelonelyghost/vault-aws-credential-protocol/internal/aws"
	"github.com/thelonelyghost/vault-aws-credential-protocol/internal/cache"
)

func run(ctx context.Context, path string, sessionName string, shouldCache bool) {
	client, err := vault.New(
		vault.WithEnvironment(),
	)
	if err != nil {
		log.Fatal(err)
	}

	cred, err := getAwsCred(ctx, client, path, sessionName, shouldCache)
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(cred)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(os.Stdout, string(data[:]))
}

func getAwsCred(ctx context.Context, client *vault.Client, path string, sessionName string, shouldCache bool) (cred aws.AwsCredentialProtocolOutput, err error) {
	cacheKey := fmt.Sprintf("%s-%s", path, sessionName)

	if shouldCache {
		cred, err = cache.GetCache(cacheKey)
		if err == nil {
			log.Println(err)
			return
		}
		err = nil
	}

	// We're being conservative here by setting the starting time to _before_
	// we make the AWS STS request, that way we err on the side of thinking it
	// expires a second or two (max) before it actually does.
	current := time.Now()

	payload := map[string]any{
		"role_session_name": sessionName,
	}

	resp, err := client.Write(ctx, path, payload)
	if err != nil {
		log.Fatal(err)
	}
	cred.Version = 1
	cred.AccessKeyId = resp.Data["access_key"].(string)
	cred.SecretAccessKey = resp.Data["secret_key"].(string)
	if val, ok := resp.Data["session_token"]; ok {
		cred.SessionToken = val.(string)
	} else if val, ok := resp.Data["security_token"]; ok {
		cred.SessionToken = val.(string)
	}
	expiry, err := time.ParseDuration(fmt.Sprintf("%ds", resp.LeaseDuration))
	if err != nil {
		return
	}
	cred.Expiration = current.Add(expiry).Format(time.RFC3339)

	if shouldCache {
		err = cache.SetCache(cacheKey, cred)
		if err != nil {
			log.Fatal(err)
		}
	}
	return
}
