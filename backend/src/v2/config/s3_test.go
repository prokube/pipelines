// Copyright 2026 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3SpiffeCredentialsPropagation(t *testing.T) {
	provider := S3ProviderConfig{
		Default: &S3ProviderDefault{
			Endpoint: "s3.example.test",
			Region:   "us-east-1",
			Credentials: &S3Credentials{Spiffe: &S3Spiffe{
				Audience:        "default-audience",
				RoleARN:         "arn:example:default",
				RoleSessionName: "default-session",
				STSEndpoint:     "http://sts.default.test",
			}},
		},
		Overrides: []S3Override{{
			BucketName: "bucket",
			KeyPrefix:  "override/",
			Credentials: &S3Credentials{Spiffe: &S3Spiffe{
				Audience:    "override-audience",
				RoleARN:     "arn:example:override",
				STSEndpoint: "http://sts.override.test",
			}},
		}},
	}

	info, err := provider.ProvideSessionInfo("s3://bucket/override/artifact")
	require.NoError(t, err)
	assert.Equal(t, "false", info.Params["fromEnv"])
	assert.Equal(t, "override-audience", info.Params["spiffeAudience"])
	assert.Equal(t, "arn:example:override", info.Params["spiffeRoleArn"])
	assert.Equal(t, "", info.Params["spiffeRoleSessionName"])
	assert.Equal(t, "http://sts.override.test", info.Params["spiffeStsEndpoint"])
	assert.NotContains(t, info.Params, "secretName")

	info, err = provider.ProvideSessionInfo("s3://bucket/default/artifact")
	require.NoError(t, err)
	assert.Equal(t, "default-audience", info.Params["spiffeAudience"])
	assert.Equal(t, "arn:example:default", info.Params["spiffeRoleArn"])
	assert.Equal(t, "default-session", info.Params["spiffeRoleSessionName"])
}

func TestS3SpiffeCredentialsValidation(t *testing.T) {
	tests := []struct {
		name        string
		credentials *S3Credentials
		wantError   string
	}{
		{
			name:        "missing audience",
			credentials: &S3Credentials{Spiffe: &S3Spiffe{RoleARN: "arn:example:role"}},
			wantError:   "missing spiffe audience",
		},
		{
			name:        "missing role ARN",
			credentials: &S3Credentials{Spiffe: &S3Spiffe{Audience: "audience"}},
			wantError:   "missing spiffe roleArn",
		},
		{
			name: "multiple credential sources",
			credentials: &S3Credentials{
				SecretRef: &S3SecretRef{},
				Spiffe:    &S3Spiffe{Audience: "audience", RoleARN: "arn:example:role"},
			},
			wantError: "mutually exclusive",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := S3ProviderConfig{Default: &S3ProviderDefault{Credentials: test.credentials}}
			_, err := provider.ProvideSessionInfo("s3://bucket/artifact")
			require.ErrorContains(t, err, test.wantError)
		})
	}
}

func TestS3FromEnvStillIgnoresSecretRef(t *testing.T) {
	provider := S3ProviderConfig{Default: &S3ProviderDefault{
		Credentials: &S3Credentials{FromEnv: true, SecretRef: &S3SecretRef{SecretName: "ignored"}},
	}}

	info, err := provider.ProvideSessionInfo("s3://bucket/artifact")
	require.NoError(t, err)
	assert.Equal(t, "true", info.Params["fromEnv"])
	assert.NotContains(t, info.Params, "secretName")
}

func TestS3SpiffeCredentialsRejectQueryParameters(t *testing.T) {
	provider := S3ProviderConfig{Default: &S3ProviderDefault{
		Credentials: &S3Credentials{Spiffe: &S3Spiffe{
			Audience: "audience",
			RoleARN:  "arn:example:role",
		}},
	}}

	_, err := provider.ProvideSessionInfo("s3://bucket/artifact?region=us-west-2")
	require.ErrorContains(t, err, "query parameters are not supported with spiffe credentials")
}
