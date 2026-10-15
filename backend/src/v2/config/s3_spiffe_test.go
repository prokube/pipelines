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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestS3ProviderSPIFFEConfigYAML(t *testing.T) {
	providers := BucketProviders{}
	err := yaml.Unmarshal([]byte(`s3:
  default:
    endpoint: seaweedfs.example.org:9000
    credentials:
      fromEnv: false
      secretRef:
        secretName: s3creds
        accessKeyKey: AWS_ACCESS_KEY_ID
        secretKeyKey: AWS_SECRET_ACCESS_KEY
    tls:
      spiffe:
        serverID: spiffe://example.org/seaweedfs
`), &providers)
	require.NoError(t, err)

	session, err := providers.S3.ProvideSessionInfo("s3://pipeline-artifacts/run/output")
	require.NoError(t, err)
	assert.Equal(t, "spiffe://example.org/seaweedfs", session.Params["spiffeServerID"])
}

func TestS3ProviderUnsignedConfigYAML(t *testing.T) {
	providers := BucketProviders{}
	err := yaml.Unmarshal([]byte(`s3:
  default:
    endpoint: seaweedfs.example.org:9000
    unsigned: true
    tls:
      spiffe:
        serverID: spiffe://example.org/seaweedfs
`), &providers)
	require.NoError(t, err)

	session, err := providers.S3.ProvideSessionInfo("s3://pipeline-artifacts/run/output")
	require.NoError(t, err)
	assert.Equal(t, "true", session.Params["unsigned"])
	assert.NotContains(t, session.Params, "fromEnv")
	assert.NotContains(t, session.Params, "secretName")
}

func TestS3ProviderSPIFFEServerID(t *testing.T) {
	provider := S3ProviderConfig{
		Default: &S3ProviderDefault{
			Endpoint:    "seaweedfs.example.org:9000",
			Credentials: testS3Credentials("default-secret"),
			TLS: &S3TLSConfig{SPIFFE: &S3SPIFFEConfig{
				ServerID: "spiffe://example.org/seaweedfs",
			}},
		},
		Overrides: []S3Override{{
			BucketName:  "without-spiffe",
			KeyPrefix:   "artifacts",
			Credentials: testS3Credentials("override-secret"),
			TLS:         &S3TLSConfig{},
		}, {
			BucketName:  "other-server",
			KeyPrefix:   "artifacts",
			Credentials: testS3Credentials("other-secret"),
			TLS: &S3TLSConfig{SPIFFE: &S3SPIFFEConfig{
				ServerID: "spiffe://example.org/other",
			}},
		}},
	}

	inherited, err := provider.ProvideSessionInfo("s3://default/artifacts/file")
	require.NoError(t, err)
	assert.Equal(t, "spiffe://example.org/seaweedfs", inherited.Params["spiffeServerID"])

	disabled, err := provider.ProvideSessionInfo("s3://without-spiffe/artifacts/file")
	require.NoError(t, err)
	assert.NotContains(t, disabled.Params, "spiffeServerID")

	overridden, err := provider.ProvideSessionInfo("s3://other-server/artifacts/file")
	require.NoError(t, err)
	assert.Equal(t, "spiffe://example.org/other", overridden.Params["spiffeServerID"])
}

func TestS3ProviderUnsignedValidation(t *testing.T) {
	spiffeTLS := &S3TLSConfig{SPIFFE: &S3SPIFFEConfig{ServerID: "spiffe://example.org/seaweedfs"}}
	disableSSL := true

	tests := []struct {
		name     string
		provider S3ProviderConfig
		path     string
		wantErr  string
	}{
		{
			name: "default credentials and unsigned are mutually exclusive",
			provider: S3ProviderConfig{Default: &S3ProviderDefault{
				Unsigned: true, Credentials: testS3Credentials("default-secret"), TLS: spiffeTLS,
			}},
			path:    "s3://bucket/artifacts",
			wantErr: "default unsigned and credentials are mutually exclusive",
		},
		{
			name:     "unsigned requires SPIFFE",
			provider: S3ProviderConfig{Default: &S3ProviderDefault{Unsigned: true}},
			path:     "s3://bucket/artifacts",
			wantErr:  "unsigned requires tls.spiffe.serverID",
		},
		{
			name: "unsigned rejects whitespace-only SPIFFE ID",
			provider: S3ProviderConfig{Default: &S3ProviderDefault{
				Unsigned: true, TLS: &S3TLSConfig{SPIFFE: &S3SPIFFEConfig{ServerID: " \t\n "}},
			}},
			path:    "s3://bucket/artifacts",
			wantErr: "unsigned requires tls.spiffe.serverID",
		},
		{
			name: "unsigned rejects disabled TLS",
			provider: S3ProviderConfig{Default: &S3ProviderDefault{
				Unsigned: true, TLS: spiffeTLS, DisableSSL: &disableSSL,
			}},
			path:    "s3://bucket/artifacts",
			wantErr: "unsigned requires disableSSL to be false",
		},
		{
			name: "unsigned rejects explicit HTTP endpoint",
			provider: S3ProviderConfig{Default: &S3ProviderDefault{
				Endpoint: " \tHtTp://seaweedfs.example.org:9000 ", Unsigned: true, TLS: spiffeTLS,
			}},
			path:    "s3://bucket/artifacts",
			wantErr: "unsigned endpoint scheme must be HTTPS",
		},
		{
			name: "unsigned rejects empty endpoint",
			provider: S3ProviderConfig{Default: &S3ProviderDefault{
				Unsigned: true, TLS: spiffeTLS,
			}},
			path:    "s3://bucket/artifacts",
			wantErr: "unsigned requires a non-empty endpoint",
		},
		{
			name: "unsigned rejects non-HTTPS endpoint scheme",
			provider: S3ProviderConfig{Default: &S3ProviderDefault{
				Endpoint: "ftp://seaweedfs.example.org:9000", Unsigned: true, TLS: spiffeTLS,
			}},
			path:    "s3://bucket/artifacts",
			wantErr: "unsigned endpoint scheme must be HTTPS",
		},
		{
			name: "override credentials and unsigned are mutually exclusive",
			provider: S3ProviderConfig{
				Default: &S3ProviderDefault{Credentials: testS3Credentials("default-secret")},
				Overrides: []S3Override{{
					BucketName: "bucket", KeyPrefix: "artifacts", Unsigned: true,
					Credentials: testS3Credentials("override-secret"), TLS: spiffeTLS,
				}},
			},
			path:    "s3://bucket/artifacts/file",
			wantErr: "override unsigned and credentials are mutually exclusive",
		},
		{
			name: "unsigned override rejects removed inherited SPIFFE config",
			provider: S3ProviderConfig{
				Default: &S3ProviderDefault{Credentials: testS3Credentials("default-secret"), TLS: spiffeTLS},
				Overrides: []S3Override{{
					BucketName: "bucket", KeyPrefix: "artifacts", Unsigned: true, TLS: &S3TLSConfig{},
				}},
			},
			path:    "s3://bucket/artifacts/file",
			wantErr: "unsigned requires tls.spiffe.serverID",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.provider.ProvideSessionInfo(test.path)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.wantErr)
		})
	}
}

func TestS3ProviderUnsignedRejectsQueryStringBypass(t *testing.T) {
	spiffeTLS := &S3TLSConfig{SPIFFE: &S3SPIFFEConfig{ServerID: " spiffe://example.org/seaweedfs "}}
	provider := S3ProviderConfig{
		Default: &S3ProviderDefault{Endpoint: "seaweedfs.example.org:9000", Unsigned: true, TLS: spiffeTLS},
		Overrides: []S3Override{
			{
				BucketName: "signed", KeyPrefix: "artifacts", Credentials: testS3Credentials("override-secret"),
			},
			{
				BucketName: "unsigned", KeyPrefix: "artifacts", Unsigned: true,
			},
		},
	}

	_, err := provider.ProvideSessionInfo("s3://default/artifacts?region=us-east-1")
	require.ErrorContains(t, err, "unsigned does not support S3 query parameters")

	matchingUnsignedOverride := S3ProviderConfig{
		Default: &S3ProviderDefault{Endpoint: "seaweedfs.example.org:9000", Credentials: testS3Credentials("default-secret"), TLS: spiffeTLS},
		Overrides: []S3Override{{
			BucketName: "unsigned", KeyPrefix: "artifacts", Unsigned: true,
		}},
	}
	_, err = matchingUnsignedOverride.ProvideSessionInfo("s3://unsigned/artifacts?region=us-east-1")
	require.ErrorContains(t, err, "unsigned does not support S3 query parameters")

	signed, err := provider.ProvideSessionInfo("s3://signed/artifacts?region=us-east-1")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"fromEnv": "true"}, signed.Params)

	withoutQuery, err := provider.ProvideSessionInfo("s3://default/artifacts")
	require.NoError(t, err)
	assert.Equal(t, "spiffe://example.org/seaweedfs", withoutQuery.Params["spiffeServerID"])
}

func TestS3ProviderUnsignedCanonicalizesEndpoint(t *testing.T) {
	provider := S3ProviderConfig{Default: &S3ProviderDefault{
		Endpoint: " \tHTTPS://seaweedfs.example.org:9000/path ",
		Unsigned: true,
		TLS:      &S3TLSConfig{SPIFFE: &S3SPIFFEConfig{ServerID: "spiffe://example.org/seaweedfs"}},
	}}

	session, err := provider.ProvideSessionInfo("s3://bucket/artifacts")
	require.NoError(t, err)
	assert.Equal(t, "https://seaweedfs.example.org:9000/path", session.Params["endpoint"])
}

func TestS3ProviderOverrideReplacesInheritedAuth(t *testing.T) {
	spiffeTLS := &S3TLSConfig{SPIFFE: &S3SPIFFEConfig{ServerID: "spiffe://example.org/seaweedfs"}}
	provider := S3ProviderConfig{
		Default: &S3ProviderDefault{Endpoint: "seaweedfs.example.org:9000", Unsigned: true, TLS: spiffeTLS},
		Overrides: []S3Override{
			{
				BucketName: "credentialed", KeyPrefix: "artifacts",
				Credentials: testS3Credentials("override-secret"),
			},
		},
	}

	credentialed, err := provider.ProvideSessionInfo("s3://credentialed/artifacts/file")
	require.NoError(t, err)
	assert.NotContains(t, credentialed.Params, "unsigned")
	assert.Equal(t, "false", credentialed.Params["fromEnv"])
	assert.Equal(t, "override-secret", credentialed.Params["secretName"])

	provider.Default = &S3ProviderDefault{
		Endpoint: "seaweedfs.example.org:9000", Credentials: testS3Credentials("default-secret"), TLS: spiffeTLS,
	}
	provider.Overrides[0] = S3Override{BucketName: "unsigned", KeyPrefix: "artifacts", Unsigned: true}
	unsigned, err := provider.ProvideSessionInfo("s3://unsigned/artifacts/file")
	require.NoError(t, err)
	assert.Equal(t, "true", unsigned.Params["unsigned"])
	assert.NotContains(t, unsigned.Params, "fromEnv")
	assert.NotContains(t, unsigned.Params, "secretName")
	assert.NotContains(t, unsigned.Params, "accessKeyKey")
	assert.NotContains(t, unsigned.Params, "secretKeyKey")
}

func testS3Credentials(secretName string) *S3Credentials {
	return &S3Credentials{SecretRef: &S3SecretRef{
		SecretName:   secretName,
		AccessKeyKey: "access-key",
		SecretKeyKey: "secret-key",
	}}
}
