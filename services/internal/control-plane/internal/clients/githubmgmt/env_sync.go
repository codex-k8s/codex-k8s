package githubmgmt

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v82/github"
	"golang.org/x/crypto/nacl/box"
)

type githubEnvPublicKey struct {
	KeyID string
	Key   [32]byte
}

func (c *Client) EnsureEnvironment(ctx context.Context, token string, owner string, repo string, envName string) error {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	envName = strings.TrimSpace(envName)
	if owner == "" || repo == "" || envName == "" {
		return fmt.Errorf("owner, repo and env name are required")
	}

	client := c.clientWithToken(token)
	if _, _, err := client.Repositories.CreateUpdateEnvironment(ctx, owner, repo, envName, &gh.CreateUpdateEnvironment{}); err != nil {
		return fmt.Errorf("ensure github environment %s/%s:%s: %w", owner, repo, envName, err)
	}
	return nil
}

func (c *Client) ListEnvSecretNames(ctx context.Context, token string, owner string, repo string, envName string) (map[string]struct{}, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	envName = strings.TrimSpace(envName)
	if owner == "" || repo == "" || envName == "" {
		return nil, fmt.Errorf("owner, repo and env name are required")
	}

	client := c.clientWithToken(token)
	repoID, err := getGitHubRepoID(ctx, client, owner, repo)
	if err != nil {
		return nil, err
	}

	out := make(map[string]struct{})
	page := 1
	for {
		secrets, resp, err := client.Actions.ListEnvSecrets(ctx, int(repoID), envName, &gh.ListOptions{PerPage: 100, Page: page})
		if err != nil {
			return nil, fmt.Errorf("list github env secrets %s/%s:%s: %w", owner, repo, envName, err)
		}
		if secrets != nil {
			for _, secret := range secrets.Secrets {
				if secret == nil {
					continue
				}
				name := strings.TrimSpace(secret.Name)
				if name == "" {
					continue
				}
				out[name] = struct{}{}
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return out, nil
}

func (c *Client) ListEnvVariableValues(ctx context.Context, token string, owner string, repo string, envName string) (map[string]string, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	envName = strings.TrimSpace(envName)
	if owner == "" || repo == "" || envName == "" {
		return nil, fmt.Errorf("owner, repo and env name are required")
	}

	client := c.clientWithToken(token)
	out := make(map[string]string)
	page := 1
	for {
		vars, resp, err := client.Actions.ListEnvVariables(ctx, owner, repo, envName, &gh.ListOptions{PerPage: 100, Page: page})
		if err != nil {
			return nil, fmt.Errorf("list github env variables %s/%s:%s: %w", owner, repo, envName, err)
		}
		if vars != nil {
			for _, v := range vars.Variables {
				if v == nil {
					continue
				}
				name := strings.TrimSpace(v.Name)
				if name == "" {
					continue
				}
				out[name] = strings.TrimSpace(v.Value)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return out, nil
}

func (c *Client) UpsertEnvVariable(ctx context.Context, token string, owner string, repo string, envName string, key string, value string) error {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	envName = strings.TrimSpace(envName)
	key = strings.TrimSpace(key)
	if owner == "" || repo == "" || envName == "" || key == "" {
		return fmt.Errorf("owner, repo, env name and key are required")
	}

	client := c.clientWithToken(token)
	payload := &gh.ActionsVariable{Name: key, Value: value}

	if _, err := client.Actions.UpdateEnvVariable(ctx, owner, repo, envName, payload); err == nil {
		return nil
	} else if !isGitHubNotFound(err) && !isGitHubConflict(err) && !isGitHubUnprocessable(err) {
		return fmt.Errorf("update github env variable %s/%s:%s:%s: %w", owner, repo, envName, key, err)
	}

	if _, err := client.Actions.CreateEnvVariable(ctx, owner, repo, envName, payload); err != nil {
		if isGitHubConflict(err) || isGitHubUnprocessable(err) {
			if _, updateErr := client.Actions.UpdateEnvVariable(ctx, owner, repo, envName, payload); updateErr != nil {
				return fmt.Errorf("update github env variable %s/%s:%s:%s after conflict: %w", owner, repo, envName, key, updateErr)
			}
			return nil
		}
		return fmt.Errorf("create github env variable %s/%s:%s:%s: %w", owner, repo, envName, key, err)
	}
	return nil
}

func (c *Client) UpsertEnvSecret(ctx context.Context, token string, owner string, repo string, envName string, key string, value string) error {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	envName = strings.TrimSpace(envName)
	key = strings.TrimSpace(key)
	if owner == "" || repo == "" || envName == "" || key == "" {
		return fmt.Errorf("owner, repo, env name and key are required")
	}

	client := c.clientWithToken(token)
	repoID, err := getGitHubRepoID(ctx, client, owner, repo)
	if err != nil {
		return err
	}

	publicKey, err := getGitHubEnvPublicKey(ctx, client, repoID, envName)
	if err != nil {
		return err
	}
	encrypted, err := encryptGitHubSecretValue(strings.TrimSpace(value), publicKey.Key)
	if err != nil {
		return fmt.Errorf("encrypt github env secret %s/%s:%s:%s: %w", owner, repo, envName, key, err)
	}

	if _, err := client.Actions.CreateOrUpdateEnvSecret(ctx, int(repoID), envName, &gh.EncryptedSecret{
		Name:           key,
		KeyID:          publicKey.KeyID,
		EncryptedValue: encrypted,
	}); err != nil {
		return fmt.Errorf("upsert github env secret %s/%s:%s:%s: %w", owner, repo, envName, key, err)
	}
	return nil
}

func getGitHubRepoID(ctx context.Context, client *gh.Client, owner string, repo string) (int64, error) {
	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return 0, fmt.Errorf("get github repository %s/%s: %w", owner, repo, err)
	}
	id := repository.GetID()
	if id <= 0 {
		return 0, fmt.Errorf("github repository %s/%s has invalid id", owner, repo)
	}
	return id, nil
}

func getGitHubEnvPublicKey(ctx context.Context, client *gh.Client, repoID int64, envName string) (githubEnvPublicKey, error) {
	publicKey, _, err := client.Actions.GetEnvPublicKey(ctx, int(repoID), envName)
	if err != nil {
		return githubEnvPublicKey{}, fmt.Errorf("get github env public key repo_id=%d env=%s: %w", repoID, envName, err)
	}

	keyID := strings.TrimSpace(publicKey.GetKeyID())
	keyEncoded := strings.TrimSpace(publicKey.GetKey())
	if keyID == "" || keyEncoded == "" {
		return githubEnvPublicKey{}, fmt.Errorf("github env public key repo_id=%d env=%s is invalid", repoID, envName)
	}
	decoded, err := base64.StdEncoding.DecodeString(keyEncoded)
	if err != nil {
		return githubEnvPublicKey{}, fmt.Errorf("decode github env public key repo_id=%d env=%s: %w", repoID, envName, err)
	}
	if len(decoded) != 32 {
		return githubEnvPublicKey{}, fmt.Errorf("github env public key repo_id=%d env=%s has invalid length", repoID, envName)
	}
	var key [32]byte
	copy(key[:], decoded)
	return githubEnvPublicKey{KeyID: keyID, Key: key}, nil
}

func encryptGitHubSecretValue(value string, publicKey [32]byte) (string, error) {
	encryptedRaw, err := box.SealAnonymous(nil, []byte(value), &publicKey, rand.Reader)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptedRaw), nil
}
