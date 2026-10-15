// Copyright 2024 The Kubeflow Authors
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
	"fmt"
	"strconv"
	"strings"

	"github.com/kubeflow/pipelines/backend/src/v2/objectstore"
)

type S3ProviderConfig struct {
	Default *S3ProviderDefault `json:"default"`
	// optional, ordered, the auth config for the first matching prefix is used
	Overrides []S3Override `json:"Overrides"`
}

type S3ProviderDefault struct {
	Endpoint    string         `json:"endpoint"`
	Credentials *S3Credentials `json:"credentials"`
	Unsigned    bool           `json:"unsigned"`
	TLS         *S3TLSConfig   `json:"tls"`
	// optional for any non aws s3 provider
	Region string `json:"region"`
	// optional
	DisableSSL *bool `json:"disableSSL"`
	// optional
	ForcePathStyle *bool `json:"forcePathStyle"`
	// optional
	MaxRetries *int `json:"maxRetries"`
}

type S3Credentials struct {
	// optional
	FromEnv bool `json:"fromEnv"`
	// if FromEnv is False then SecretRef is required
	SecretRef *S3SecretRef `json:"secretRef"`
}
type S3Override struct {
	Endpoint string       `json:"endpoint"`
	TLS      *S3TLSConfig `json:"tls"`
	Unsigned bool         `json:"unsigned"`
	// optional for any non aws s3 provider
	Region string `json:"region"`
	// optional
	DisableSSL *bool  `json:"disableSSL"`
	BucketName string `json:"bucketName"`
	KeyPrefix  string `json:"keyPrefix"`
	// required
	Credentials *S3Credentials `json:"credentials"`
	// optional
	ForcePathStyle *bool `json:"forcePathStyle"`
	// optional
	MaxRetries *int `json:"maxRetries"`
}
type S3SecretRef struct {
	SecretName string `json:"secretName"`
	// The k8s secret "Key" for "Artifact SecretKey" and "Artifact AccessKey"
	AccessKeyKey string `json:"accessKeyKey"`
	SecretKeyKey string `json:"secretKeyKey"`
}

type S3TLSConfig struct {
	SPIFFE *S3SPIFFEConfig `json:"spiffe"`
}

type S3SPIFFEConfig struct {
	ServerID string `json:"serverID"`
}

func (p S3ProviderConfig) ProvideSessionInfo(path string) (objectstore.SessionInfo, error) {
	bucketConfig, err := objectstore.ParseBucketPathToConfig(path)
	if err != nil {
		return objectstore.SessionInfo{}, err
	}
	bucketName := bucketConfig.BucketName
	bucketPrefix := bucketConfig.Prefix
	queryString := bucketConfig.QueryString
	override := p.getOverrideByPrefix(bucketName, bucketPrefix)

	invalidConfigErr := func(err error) error {
		return fmt.Errorf("invalid provider config: %w", err)
	}

	params := map[string]string{}

	// Without provider config, or with query parameters in signed mode, use
	// blob.OpenBucket(ctx, config.bucketURL()) by setting "FromEnv = True".
	if p.Default == nil && p.Overrides == nil {
		params["fromEnv"] = strconv.FormatBool(true)
		return objectstore.SessionInfo{
			Provider: "s3",
			Params:   params,
		}, nil
	}
	if queryString != "" && !effectiveUnsigned(p.Default, override) {
		params["fromEnv"] = strconv.FormatBool(true)
		return objectstore.SessionInfo{
			Provider: "s3",
			Params:   params,
		}, nil
	}

	if p.Default == nil {
		return objectstore.SessionInfo{}, invalidConfigErr(fmt.Errorf("missing default credentials"))
	}

	params["endpoint"] = p.Default.Endpoint
	params["region"] = p.Default.Region
	setSPIFFEServerID(params, p.Default.TLS)

	if p.Default.DisableSSL == nil {
		params["disableSSL"] = strconv.FormatBool(false)
	} else {
		params["disableSSL"] = strconv.FormatBool(*p.Default.DisableSSL)
	}

	if p.Default.ForcePathStyle == nil {
		params["forcePathStyle"] = strconv.FormatBool(true)
	} else {
		params["forcePathStyle"] = strconv.FormatBool(*p.Default.ForcePathStyle)
	}
	if p.Default.MaxRetries == nil {
		params["maxRetries"] = strconv.FormatInt(5, 10)
	} else {
		params["maxRetries"] = strconv.FormatInt(int64(*p.Default.MaxRetries), 10)
	}

	if err := setS3AuthParams(params, p.Default.Unsigned, p.Default.Credentials, "default"); err != nil {
		return objectstore.SessionInfo{}, invalidConfigErr(err)
	}

	// Set defaults
	sessionInfo := objectstore.SessionInfo{
		Provider: "s3",
		Params:   params,
	}

	// If there's a matching override, then override defaults with provided configs
	if override != nil {
		if override.Endpoint != "" {
			sessionInfo.Params["endpoint"] = override.Endpoint
		}
		if override.Region != "" {
			sessionInfo.Params["region"] = override.Region
		}
		if override.DisableSSL != nil {
			sessionInfo.Params["disableSSL"] = strconv.FormatBool(*override.DisableSSL)
		}
		if override.ForcePathStyle != nil {
			sessionInfo.Params["forcePathStyle"] = strconv.FormatBool(*override.ForcePathStyle)
		}
		if override.MaxRetries != nil {
			sessionInfo.Params["maxRetries"] = strconv.FormatInt(int64(*p.Default.MaxRetries), 10)
		}
		if override.TLS != nil {
			delete(params, "spiffeServerID")
			setSPIFFEServerID(params, override.TLS)
		}
		if err := setS3AuthParams(params, override.Unsigned, override.Credentials, "override"); err != nil {
			return objectstore.SessionInfo{}, invalidConfigErr(err)
		}
	}
	if err := validateUnsignedS3Params(params); err != nil {
		return objectstore.SessionInfo{}, invalidConfigErr(err)
	}
	if queryString != "" {
		return objectstore.SessionInfo{}, invalidConfigErr(fmt.Errorf("unsigned does not support S3 query parameters"))
	}
	return sessionInfo, nil
}

func effectiveUnsigned(defaultConfig *S3ProviderDefault, override *S3Override) bool {
	if override != nil {
		return override.Unsigned
	}
	return defaultConfig != nil && defaultConfig.Unsigned
}

func setS3AuthParams(params map[string]string, unsigned bool, credentials *S3Credentials, configName string) error {
	delete(params, "unsigned")
	delete(params, "fromEnv")
	delete(params, "secretName")
	delete(params, "accessKeyKey")
	delete(params, "secretKeyKey")

	if unsigned && credentials != nil {
		return fmt.Errorf("%s unsigned and credentials are mutually exclusive", configName)
	}
	if unsigned {
		params["unsigned"] = strconv.FormatBool(true)
		return nil
	}
	if credentials == nil {
		return fmt.Errorf("missing %s credentials", configName)
	}

	params["fromEnv"] = strconv.FormatBool(credentials.FromEnv)
	if credentials.FromEnv {
		return nil
	}
	if credentials.SecretRef == nil {
		return fmt.Errorf("missing %s secretref", configName)
	}
	params["secretName"] = credentials.SecretRef.SecretName
	params["accessKeyKey"] = credentials.SecretRef.AccessKeyKey
	params["secretKeyKey"] = credentials.SecretRef.SecretKeyKey
	return nil
}

func validateUnsignedS3Params(params map[string]string) error {
	if params["unsigned"] != strconv.FormatBool(true) {
		return nil
	}
	if params["spiffeServerID"] == "" {
		return fmt.Errorf("unsigned requires tls.spiffe.serverID")
	}
	if params["disableSSL"] == strconv.FormatBool(true) {
		return fmt.Errorf("unsigned requires disableSSL to be false")
	}
	endpoint, err := objectstore.NormalizeUnsignedS3Endpoint(params["endpoint"])
	if err != nil {
		return err
	}
	params["endpoint"] = endpoint
	return nil
}

func setSPIFFEServerID(params map[string]string, tls *S3TLSConfig) {
	if tls != nil && tls.SPIFFE != nil {
		serverID := strings.TrimSpace(tls.SPIFFE.ServerID)
		if serverID != "" {
			params["spiffeServerID"] = serverID
		}
	}
}

// getOverrideByPrefix returns first matching bucketname and prefix in overrides
func (p S3ProviderConfig) getOverrideByPrefix(bucketName, prefix string) *S3Override {
	for _, override := range p.Overrides {
		if override.BucketName == bucketName && strings.HasPrefix(prefix, override.KeyPrefix) {
			return &override
		}
	}
	return nil
}
