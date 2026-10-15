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
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

type x509Source interface {
	x509svid.Source
	x509bundle.Source
	Close() error
}

type spiffeSourceManager struct {
	ctx    context.Context
	mu     sync.Mutex
	source x509Source
}

type spiffeSourceManagerKey struct{}

var newX509Source = func(ctx context.Context) (x509Source, error) {
	return workloadapi.NewX509Source(ctx)
}

// WithSPIFFEWorkloadAPI adds a lazy Workload API source to the context. The
// source is opened only when an S3 provider config contains a SPIFFE server ID.
func WithSPIFFEWorkloadAPI(ctx context.Context) (context.Context, func() error) {
	manager := &spiffeSourceManager{ctx: ctx}
	return context.WithValue(ctx, spiffeSourceManagerKey{}, manager), manager.Close
}

func (m *spiffeSourceManager) Source() (x509Source, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.source != nil {
		return m.source, nil
	}
	source, err := newX509Source(m.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SPIFFE Workload API source: %w", err)
	}
	m.source = source
	return source, nil
}

func (m *spiffeSourceManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.source == nil {
		return nil
	}
	return m.source.Close()
}

func spiffeS3HTTPClient(ctx context.Context, serverIDValue string) (*http.Client, error) {
	serverIDValue = strings.TrimSpace(serverIDValue)
	if serverIDValue == "" {
		return nil, nil
	}
	if strings.TrimSpace(os.Getenv(workloadapi.SocketEnv)) == "" {
		return nil, fmt.Errorf("S3 SPIFFE TLS requires %s to be set", workloadapi.SocketEnv)
	}
	serverID, err := spiffeid.FromString(serverIDValue)
	if err != nil {
		return nil, fmt.Errorf("invalid S3 SPIFFE server ID: %w", err)
	}
	manager, ok := ctx.Value(spiffeSourceManagerKey{}).(*spiffeSourceManager)
	if !ok {
		return nil, fmt.Errorf("S3 SPIFFE TLS requires an initialized Workload API context")
	}
	source, err := manager.Source()
	if err != nil {
		return nil, err
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsconfig.MTLSClientConfig(
		source,
		source,
		tlsconfig.AuthorizeID(serverID),
	)
	return &http.Client{Transport: transport}, nil
}
