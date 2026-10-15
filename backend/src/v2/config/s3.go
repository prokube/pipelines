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
	// if FromEnv is false then exactly one of SecretRef or Spiffe is required
	SecretRef *S3SecretRef `json:"secretRef"`
	Spiffe    *S3Spiffe    `json:"spiffe"`
}

type S3Spiffe struct {
	Audience        string `json:"audience"`
	RoleARN         string `json:"roleArn"`
	RoleSessionName string `json:"roleSessionName"`
	STSEndpoint     string `json:"stsEndpoint"`
}
type S3Override struct {
	Endpoint string `json:"endpoint"`
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

func (p S3ProviderConfig) ProvideSessionInfo(path string) (objectstore.SessionInfo, error) {
	bucketConfig, err := objectstore.ParseBucketPathToConfig(path)
	if err != nil {
		return objectstore.SessionInfo{}, err
	}
	bucketName := bucketConfig.BucketName
	bucketPrefix := bucketConfig.Prefix
	queryString := bucketConfig.QueryString

	invalidConfigErr := func(err error) error {
		return fmt.Errorf("invalid provider config: %w", err)
	}

	params := map[string]string{}

	// If provider config did not have a matching configuration for the provider inferred from pipelineroot,
	// use blob.OpenBucket(ctx, config.bucketURL()) by setting "FromEnv = True".
	if p.Default == nil && p.Overrides == nil {
		params["fromEnv"] = strconv.FormatBool(true)
		return objectstore.SessionInfo{
			Provider: "s3",
			Params:   params,
		}, nil
	}
	if queryString != "" {
		var credentials *S3Credentials
		if p.Default != nil {
			credentials = p.Default.Credentials
		}
		if override := p.getOverrideByPrefix(bucketName, bucketPrefix); override != nil {
			credentials = override.Credentials
		}
		if credentials != nil && credentials.Spiffe != nil {
			return objectstore.SessionInfo{}, invalidConfigErr(fmt.Errorf("query parameters are not supported with spiffe credentials"))
		}
		params["fromEnv"] = strconv.FormatBool(true)
		return objectstore.SessionInfo{
			Provider: "s3",
			Params:   params,
		}, nil
	}

	if p.Default == nil || p.Default.Credentials == nil {
		return objectstore.SessionInfo{}, invalidConfigErr(fmt.Errorf("missing default credentials"))
	}
	if !p.Default.Credentials.FromEnv && p.Default.Credentials.SecretRef == nil && p.Default.Credentials.Spiffe == nil {
		return objectstore.SessionInfo{}, invalidConfigErr(fmt.Errorf("missing default secretref or spiffe"))
	}

	params["endpoint"] = p.Default.Endpoint
	params["region"] = p.Default.Region

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

	if err := applyS3Credentials(params, p.Default.Credentials); err != nil {
		return objectstore.SessionInfo{}, invalidConfigErr(fmt.Errorf("invalid default credentials: %w", err))
	}

	// Set defaults
	sessionInfo := objectstore.SessionInfo{
		Provider: "s3",
		Params:   params,
	}

	// If there's a matching override, then override defaults with provided configs
	override := p.getOverrideByPrefix(bucketName, bucketPrefix)
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
		if override.Credentials == nil {
			return objectstore.SessionInfo{}, invalidConfigErr(fmt.Errorf("missing override credentials"))
		}
		if !override.Credentials.FromEnv && override.Credentials.SecretRef == nil && override.Credentials.Spiffe == nil {
			return objectstore.SessionInfo{}, invalidConfigErr(fmt.Errorf("missing override secretref or spiffe"))
		}
		if err := applyS3Credentials(params, override.Credentials); err != nil {
			return objectstore.SessionInfo{}, invalidConfigErr(fmt.Errorf("invalid override credentials: %w", err))
		}
	}
	return sessionInfo, nil
}

func applyS3Credentials(params map[string]string, credentials *S3Credentials) error {
	for _, key := range []string{
		"secretName", "accessKeyKey", "secretKeyKey", "spiffeAudience", "spiffeRoleArn",
		"spiffeRoleSessionName", "spiffeStsEndpoint",
	} {
		delete(params, key)
	}

	params["fromEnv"] = strconv.FormatBool(credentials.FromEnv)
	if credentials.FromEnv {
		if credentials.Spiffe != nil {
			return fmt.Errorf("fromEnv cannot be combined with spiffe")
		}
		return nil
	}
	if credentials.SecretRef != nil && credentials.Spiffe != nil {
		return fmt.Errorf("secretRef and spiffe are mutually exclusive")
	}
	if credentials.Spiffe != nil {
		if credentials.Spiffe.Audience == "" {
			return fmt.Errorf("missing spiffe audience")
		}
		if credentials.Spiffe.RoleARN == "" {
			return fmt.Errorf("missing spiffe roleArn")
		}
		params["spiffeAudience"] = credentials.Spiffe.Audience
		params["spiffeRoleArn"] = credentials.Spiffe.RoleARN
		params["spiffeRoleSessionName"] = credentials.Spiffe.RoleSessionName
		params["spiffeStsEndpoint"] = credentials.Spiffe.STSEndpoint
		return nil
	}
	if credentials.SecretRef == nil {
		return fmt.Errorf("missing secretRef or spiffe")
	}
	params["secretName"] = credentials.SecretRef.SecretName
	params["accessKeyKey"] = credentials.SecretRef.AccessKeyKey
	params["secretKeyKey"] = credentials.SecretRef.SecretKeyKey
	return nil
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
