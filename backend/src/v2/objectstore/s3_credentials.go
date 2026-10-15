// Copyright 2026 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package objectstore

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type spiffeJWTFetcher func(context.Context, string, string) ([]byte, error)

const (
	spiffeJWTFetchAttempts = 20
	spiffeJWTFetchInterval = 500 * time.Millisecond
)

type spiffeWebIdentityProvider struct {
	client          stscreds.AssumeRoleWithWebIdentityAPIClient
	fetchToken      spiffeJWTFetcher
	socket          string
	audience        string
	roleARN         string
	roleSessionName string
}

func getS3CredentialsProvider(ctx context.Context, client kubernetes.Interface, namespace string, params *S3Params) (aws.CredentialsProvider, error) {
	if params.Spiffe != nil {
		return newSpiffeWebIdentityCredentials(params, fetchSpiffeJWTSVID)
	}
	return getS3BucketCredential(ctx, client, namespace, params.SecretName, params.SecretKeyKey, params.AccessKeyKey)
}

func (p *spiffeWebIdentityProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	token, err := fetchSpiffeJWTWithRetry(ctx, p.fetchToken, p.socket, p.audience, spiffeJWTFetchAttempts, spiffeJWTFetchInterval)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("failed to fetch SPIFFE JWT-SVID: %w", err)
	}
	provider := stscreds.NewWebIdentityRoleProvider(p.client, p.roleARN, identityToken(token), func(options *stscreds.WebIdentityRoleOptions) {
		options.RoleSessionName = p.roleSessionName
	})
	return provider.Retrieve(ctx)
}

func fetchSpiffeJWTWithRetry(ctx context.Context, fetchToken spiffeJWTFetcher, socket, audience string, attempts int, interval time.Duration) ([]byte, error) {
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		var token []byte
		token, err = fetchToken(ctx, socket, audience)
		if err == nil {
			return token, nil
		}
		if attempt == attempts {
			break
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, fmt.Errorf("failed after %d attempts: %w", attempts, err)
}

type identityToken []byte

func (t identityToken) GetIdentityToken() ([]byte, error) {
	return t, nil
}

func fetchSpiffeJWTSVID(ctx context.Context, socket, audience string) ([]byte, error) {
	svid, err := workloadapi.FetchJWTSVID(ctx, jwtsvid.Params{Audience: audience}, workloadapi.WithAddr(socket))
	if err != nil {
		return nil, err
	}
	return []byte(svid.Marshal()), nil
}

func newSpiffeWebIdentityCredentials(params *S3Params, fetchToken spiffeJWTFetcher) (aws.CredentialsProvider, error) {
	if params.Spiffe.Audience == "" {
		return nil, fmt.Errorf("missing SPIFFE audience")
	}
	if params.Spiffe.RoleARN == "" {
		return nil, fmt.Errorf("missing SPIFFE role ARN")
	}
	socket, ok := os.LookupEnv(workloadapi.SocketEnv)
	if !ok || socket == "" {
		return nil, fmt.Errorf("%s must be set for SPIFFE authentication", workloadapi.SocketEnv)
	}
	if err := workloadapi.ValidateAddress(socket); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", workloadapi.SocketEnv, err)
	}

	stsConfig := aws.Config{
		Region:      params.Region,
		Credentials: aws.AnonymousCredentials{},
	}
	if params.Spiffe.STSEndpoint != "" {
		endpoint, err := url.Parse(params.Spiffe.STSEndpoint)
		if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Host == "" ||
			endpoint.User != nil || endpoint.Path != "" || endpoint.RawQuery != "" || endpoint.Fragment != "" {
			return nil, fmt.Errorf("invalid SPIFFE STS endpoint %q", params.Spiffe.STSEndpoint)
		}
		stsConfig.BaseEndpoint = aws.String(params.Spiffe.STSEndpoint)
	}

	return aws.NewCredentialsCache(&spiffeWebIdentityProvider{
		client:          sts.NewFromConfig(stsConfig),
		fetchToken:      fetchToken,
		socket:          socket,
		audience:        params.Spiffe.Audience,
		roleARN:         params.Spiffe.RoleARN,
		roleSessionName: params.Spiffe.RoleSessionName,
	}), nil
}

func getS3BucketCredential(
	ctx context.Context,
	clientSet kubernetes.Interface,
	namespace string,
	secretName string,
	bucketSecretKeyKey string,
	bucketAccessKeyKey string,
) (cred *credentials.StaticCredentialsProvider, err error) {
	defer func() {
		if err != nil {
			// wrap error before returning
			err = fmt.Errorf("Failed to get Bucket credentials from secret name=%q namespace=%q: %w", secretName, namespace, err)
		}
	}()
	secret, err := clientSet.CoreV1().Secrets(namespace).Get(
		ctx,
		secretName,
		metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// The k8s secret "Key" for "SecretKey" and "AccessKey"
	accessKey := string(secret.Data[bucketAccessKeyKey])
	secretKey := string(secret.Data[bucketSecretKeyKey])

	if accessKey != "" && secretKey != "" {
		s3Creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
		return &s3Creds, err
	}
	return nil, fmt.Errorf("could not find specified keys '%s' or '%s'", bucketAccessKeyKey, bucketSecretKeyKey)
}
