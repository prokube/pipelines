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

func testS3Credentials(secretName string) *S3Credentials {
	return &S3Credentials{SecretRef: &S3SecretRef{
		SecretName:   secretName,
		AccessKeyKey: "access-key",
		SecretKeyKey: "secret-key",
	}}
}
