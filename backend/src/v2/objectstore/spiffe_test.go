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
	"errors"
	"net/http"
	"testing"

	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

type fakeX509Source struct {
	closed bool
}

func (*fakeX509Source) GetX509SVID() (*x509svid.SVID, error) {
	return nil, errors.New("not used by this test")
}

func (*fakeX509Source) GetX509BundleForTrustDomain(spiffeid.TrustDomain) (*x509bundle.Bundle, error) {
	return nil, errors.New("not used by this test")
}

func (s *fakeX509Source) Close() error {
	s.closed = true
	return nil
}

func TestSPIFFEWorkloadAPISourceIsLazy(t *testing.T) {
	sourceCreated := false
	originalNewX509Source := newX509Source
	newX509Source = func(context.Context) (x509Source, error) {
		sourceCreated = true
		return &fakeX509Source{}, nil
	}
	t.Cleanup(func() { newX509Source = originalNewX509Source })

	ctx, closeSource := WithSPIFFEWorkloadAPI(context.Background())
	client, err := spiffeS3HTTPClient(ctx, "")
	require.NoError(t, err)
	assert.Nil(t, client)
	assert.False(t, sourceCreated)
	require.NoError(t, closeSource())
}

func TestSPIFFES3HTTPClientRequiresSocket(t *testing.T) {
	t.Setenv(workloadapi.SocketEnv, "")
	ctx, closeSource := WithSPIFFEWorkloadAPI(context.Background())
	t.Cleanup(func() { require.NoError(t, closeSource()) })

	_, err := spiffeS3HTTPClient(ctx, "spiffe://example.org/seaweedfs")
	require.Error(t, err)
	assert.Contains(t, err.Error(), workloadapi.SocketEnv)
}

func TestSPIFFES3HTTPClientRejectsInvalidServerID(t *testing.T) {
	t.Setenv(workloadapi.SocketEnv, "unix:///spiffe-workload-api/spire-agent.sock")
	ctx, closeSource := WithSPIFFEWorkloadAPI(context.Background())
	t.Cleanup(func() { require.NoError(t, closeSource()) })

	_, err := spiffeS3HTTPClient(ctx, "not-a-spiffe-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid S3 SPIFFE server ID")
}

func TestSPIFFES3HTTPClientBuildsMTLSClientAndClosesSource(t *testing.T) {
	t.Setenv(workloadapi.SocketEnv, "unix:///spiffe-workload-api/spire-agent.sock")
	source := &fakeX509Source{}
	originalNewX509Source := newX509Source
	newX509Source = func(context.Context) (x509Source, error) {
		return source, nil
	}
	t.Cleanup(func() { newX509Source = originalNewX509Source })

	ctx, closeSource := WithSPIFFEWorkloadAPI(context.Background())
	client, err := spiffeS3HTTPClient(ctx, "spiffe://example.org/seaweedfs")
	require.NoError(t, err)
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.NotNil(t, transport.TLSClientConfig)
	assert.False(t, source.closed)

	require.NoError(t, closeSource())
	assert.True(t, source.closed)
}

func TestCreateS3BucketSessionUsesProviderSPIFFEConfig(t *testing.T) {
	t.Setenv(workloadapi.SocketEnv, "unix:///spiffe-workload-api/spire-agent.sock")
	source := &fakeX509Source{}
	originalNewX509Source := newX509Source
	newX509Source = func(context.Context) (x509Source, error) {
		return source, nil
	}
	t.Cleanup(func() { newX509Source = originalNewX509Source })
	ctx, closeSource := WithSPIFFEWorkloadAPI(context.Background())
	t.Cleanup(func() { require.NoError(t, closeSource()) })

	clientSet := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "s3-provider-secret", Namespace: "testnamespace"},
		Data: map[string][]byte{
			"access-key": []byte("access"),
			"secret-key": []byte("secret"),
		},
	})
	sessionInfo := &SessionInfo{
		Provider: "s3",
		Params: map[string]string{
			"region":         "us-east-1",
			"endpoint":       "seaweedfs.example.org:9000",
			"disableSSL":     "false",
			"fromEnv":        "false",
			"secretName":     "s3-provider-secret",
			"accessKeyKey":   "access-key",
			"secretKeyKey":   "secret-key",
			"forcePathStyle": "true",
			"spiffeServerID": "spiffe://example.org/seaweedfs",
		},
	}

	client, err := createS3BucketSession(ctx, "testnamespace", sessionInfo, clientSet)
	require.NoError(t, err)
	require.NotNil(t, client)

	httpClient, ok := client.Options().HTTPClient.(*http.Client)
	require.True(t, ok)
	transport, ok := httpClient.Transport.(*http.Transport)
	require.True(t, ok)
	assert.NotNil(t, transport.TLSClientConfig)
}
