// Copyright 2023 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package component

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kubeflow/pipelines/backend/src/v2/cacheutils"
	"github.com/kubeflow/pipelines/backend/src/v2/client_manager"
	launcherconfig "github.com/kubeflow/pipelines/backend/src/v2/config"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/kubeflow/pipelines/api/v2alpha1/go/pipelinespec"
	"github.com/kubeflow/pipelines/backend/src/v2/metadata"
	"github.com/kubeflow/pipelines/backend/src/v2/objectstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
	"google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

var addNumbersComponent = &pipelinespec.ComponentSpec{
	Implementation: &pipelinespec.ComponentSpec_ExecutorLabel{ExecutorLabel: "add"},
	InputDefinitions: &pipelinespec.ComponentInputsSpec{
		Parameters: map[string]*pipelinespec.ComponentInputsSpec_ParameterSpec{
			"a": {ParameterType: pipelinespec.ParameterType_NUMBER_INTEGER, DefaultValue: structpb.NewNumberValue(5)},
			"b": {ParameterType: pipelinespec.ParameterType_NUMBER_INTEGER},
		},
	},
	OutputDefinitions: &pipelinespec.ComponentOutputsSpec{
		Parameters: map[string]*pipelinespec.ComponentOutputsSpec_ParameterSpec{
			"Output": {ParameterType: pipelinespec.ParameterType_NUMBER_INTEGER},
		},
	},
}

// Tests that launcher correctly executes the user component and successfully writes output parameters to file.
func Test_executeV2_Parameters(t *testing.T) {
	tests := []struct {
		name          string
		executorInput *pipelinespec.ExecutorInput
		executorArgs  []string
		wantErr       bool
	}{
		{
			"happy pass",
			&pipelinespec.ExecutorInput{
				Inputs: &pipelinespec.ExecutorInput_Inputs{
					ParameterValues: map[string]*structpb.Value{"a": structpb.NewNumberValue(1), "b": structpb.NewNumberValue(2)},
				},
			},
			[]string{"-c", "test {{$.inputs.parameters['a']}} -eq 1 || exit 1\ntest {{$.inputs.parameters['b']}} -eq 2 || exit 1"},
			false,
		},
		{
			"use default value",
			&pipelinespec.ExecutorInput{
				Inputs: &pipelinespec.ExecutorInput_Inputs{
					ParameterValues: map[string]*structpb.Value{"b": structpb.NewNumberValue(2)},
				},
			},
			[]string{"-c", "test {{$.inputs.parameters['a']}} -eq 5 || exit 1\ntest {{$.inputs.parameters['b']}} -eq 2 || exit 1"},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeKubernetesClientset := &fake.Clientset{}
			fakeMetadataClient := metadata.NewFakeClient()
			bucket, err := blob.OpenBucket(context.Background(), "mem://test-bucket")
			assert.Nil(t, err)
			bucketConfig, err := objectstore.ParseBucketConfig("mem://test-bucket/pipeline-root/", nil)
			assert.Nil(t, err)
			_, _, err = executeV2(
				context.Background(),
				test.executorInput,
				addNumbersComponent,
				"sh",
				test.executorArgs,
				bucket,
				bucketConfig,
				fakeMetadataClient,
				"namespace",
				fakeKubernetesClientset,
				"false",
				"",
			)

			if test.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)

			}
		})
	}
}

func Test_executeV2_publishLogs(t *testing.T) {
	tests := []struct {
		name          string
		executorInput *pipelinespec.ExecutorInput
		executorArgs  []string
		wantErr       bool
	}{
		{
			"happy pass",
			&pipelinespec.ExecutorInput{
				Inputs: &pipelinespec.ExecutorInput_Inputs{
					ParameterValues: map[string]*structpb.Value{"a": structpb.NewNumberValue(1), "b": structpb.NewNumberValue(2)},
				},
			},
			[]string{"-c", "test {{$.inputs.parameters['a']}} -eq 1 || exit 1\ntest {{$.inputs.parameters['b']}} -eq 2 || exit 1"},
			false,
		},
		{
			"use default value",
			&pipelinespec.ExecutorInput{
				Inputs: &pipelinespec.ExecutorInput_Inputs{
					ParameterValues: map[string]*structpb.Value{"b": structpb.NewNumberValue(2)},
				},
			},
			[]string{"-c", "test {{$.inputs.parameters['a']}} -eq 5 || exit 1\ntest {{$.inputs.parameters['b']}} -eq 2 || exit 1"},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeKubernetesClientset := &fake.Clientset{}
			fakeMetadataClient := metadata.NewFakeClient()
			bucket, err := blob.OpenBucket(context.Background(), "mem://test-bucket")
			assert.Nil(t, err)
			bucketConfig, err := objectstore.ParseBucketConfig("mem://test-bucket/pipeline-root/", nil)
			assert.Nil(t, err)
			_, _, err = executeV2(
				context.Background(),
				test.executorInput,
				addNumbersComponent,
				"sh",
				test.executorArgs,
				bucket,
				bucketConfig,
				fakeMetadataClient,
				"namespace",
				fakeKubernetesClientset,
				"false",
				"",
			)

			if test.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)

			}
		})
	}
}

func Test_getPlaceholders_WorkspaceArtifactPath(t *testing.T) {
	execIn := &pipelinespec.ExecutorInput{
		Inputs: &pipelinespec.ExecutorInput_Inputs{
			Artifacts: map[string]*pipelinespec.ArtifactList{
				"data": {
					Artifacts: []*pipelinespec.RuntimeArtifact{
						{Uri: "minio://mlpipeline/sample/sample.txt", Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{"_kfp_workspace": structpb.NewBoolValue(true)}}},
					},
				},
			},
		},
	}
	ph, err := getPlaceholders(execIn)
	if err != nil {
		t.Fatalf("getPlaceholders error: %v", err)
	}
	actual := ph["{{$.inputs.artifacts['data'].path}}"]
	expected := filepath.Join(WorkspaceMountPath, ".artifacts", "minio", "mlpipeline", "sample", "sample.txt")
	if actual != expected {
		t.Fatalf("placeholder path mismatch: actual=%q expected=%q", actual, expected)
	}
}

func TestSetNonDefaultBucketSession(t *testing.T) {
	providerClient := fake.NewSimpleClientset(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "kfp-launcher", Namespace: "testnamespace"},
		Data: map[string]string{"providers": `s3:
  default:
    endpoint: https://s3.example.org
    unsigned: true
    tls:
      spiffe:
        serverID: spiffe://example.org/s3
  overrides:
    - bucketName: other
      keyPrefix: imported
      endpoint: HTTPS://s3-override.example.org
      region: override-region
      unsigned: true
      tls:
        spiffe:
          serverID: spiffe://example.org/s3-override
minio:
  default:
    endpoint: https://minio.example.org
    region: minio-region
    unsigned: true
    tls:
      spiffe:
        serverID: spiffe://example.org/minio
`},
	})
	providerConfig, err := launcherconfig.FromConfigMap(context.Background(), providerClient, "testnamespace")
	require.NoError(t, err)

	rootSession := &objectstore.SessionInfo{Provider: "s3", Params: map[string]string{"unsigned": "true"}}
	defaultConfig, err := objectstore.ParseBucketConfig("s3://default/pipeline-root", rootSession)
	require.NoError(t, err)
	s3Config, err := objectstore.ParseBucketConfigForArtifactURI("s3://other/imported/artifact")
	require.NoError(t, err)

	require.NoError(t, setNonDefaultBucketSession("s3://other/imported/artifact", s3Config, defaultConfig, providerConfig))
	require.NotNil(t, s3Config.SessionInfo)
	assert.Equal(t, "https://s3-override.example.org", s3Config.SessionInfo.Params["endpoint"])
	assert.Equal(t, "override-region", s3Config.SessionInfo.Params["region"])
	assert.Equal(t, "spiffe://example.org/s3-override", s3Config.SessionInfo.Params["spiffeServerID"])

	minioConfig, err := objectstore.ParseBucketConfigForArtifactURI("minio://other/imported/artifact")
	require.NoError(t, err)
	require.NoError(t, setNonDefaultBucketSession("minio://other/imported/artifact", minioConfig, defaultConfig, providerConfig))
	require.NotNil(t, minioConfig.SessionInfo)
	assert.Equal(t, "minio", minioConfig.SessionInfo.Provider)
	assert.Equal(t, "https://minio.example.org", minioConfig.SessionInfo.Params["endpoint"])
	assert.Equal(t, "minio-region", minioConfig.SessionInfo.Params["region"])
	assert.Equal(t, "spiffe://example.org/minio", minioConfig.SessionInfo.Params["spiffeServerID"])

	signedSession := &objectstore.SessionInfo{Provider: "s3", Params: map[string]string{"fromEnv": "true"}}
	defaultConfig.SessionInfo = signedSession
	s3Config.SessionInfo = nil
	require.NoError(t, setNonDefaultBucketSession("s3://other/imported/artifact", s3Config, defaultConfig, nil))
	assert.Nil(t, s3Config.SessionInfo, "signed cross-bucket behavior must continue using environment credentials")

	sameBucketConfig, err := objectstore.ParseBucketConfigForArtifactURI("s3://default/other/artifact")
	require.NoError(t, err)
	require.NoError(t, setNonDefaultBucketSession("s3://default/other/artifact", sameBucketConfig, defaultConfig, nil))
	assert.Same(t, signedSession, sameBucketConfig.SessionInfo, "same-bucket signed session reuse must remain unchanged")

	defaultConfig.SessionInfo = rootSession
	err = setNonDefaultBucketSession("s3://other/imported/artifact", s3Config, defaultConfig, nil)
	require.ErrorContains(t, err, "launcher config is required")
}

func TestImporterWorkspaceUsesUnsignedSPIFFEConfig(t *testing.T) {
	launcher := &ImportLauncher{
		launcherV2Options: LauncherV2Options{Namespace: "testnamespace"},
		k8sClient: fake.NewSimpleClientset(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "kfp-launcher", Namespace: "testnamespace"},
			Data: map[string]string{"providers": `s3:
  default:
    endpoint: https://seaweedfs.example.org:9000
    unsigned: true
    tls:
      spiffe:
        serverID: spiffe://example.org/seaweedfs
`},
		}),
	}

	bucketConfig, err := launcher.resolveBucketConfigForURI(context.Background(), "s3://other-bucket/path/artifact")
	require.NoError(t, err)
	require.NotNil(t, bucketConfig.SessionInfo)
	assert.Equal(t, "true", bucketConfig.SessionInfo.Params["unsigned"])
}

func TestImporterRejectsInvalidConfiguredUnsignedSession(t *testing.T) {
	launcher := &ImportLauncher{
		launcherV2Options: LauncherV2Options{Namespace: "testnamespace"},
		k8sClient: fake.NewSimpleClientset(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "kfp-launcher", Namespace: "testnamespace"},
			Data: map[string]string{"providers": `s3:
  default:
    endpoint: HTTP://seaweedfs.example.org:9000
    unsigned: true
    tls:
      spiffe:
        serverID: spiffe://example.org/seaweedfs
`},
		}),
	}

	_, err := launcher.resolveBucketConfigForURI(context.Background(), "s3://imported-bucket/path/artifact")
	require.ErrorContains(t, err, "unsigned endpoint scheme must be HTTPS")
}

func TestImporterWorkspaceFallsBackWhenConfigMapLoadFails(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("get", "configmaps", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("configmap API unavailable")
	})
	launcher := &ImportLauncher{
		launcherV2Options: LauncherV2Options{Namespace: "testnamespace"},
		k8sClient:         client,
	}

	bucketConfig, err := launcher.resolveBucketConfigForURI(context.Background(), "s3://imported-bucket/path/artifact")
	require.NoError(t, err)
	assert.Nil(t, bucketConfig.SessionInfo)
}

func Test_executorInput_compileCmdAndArgs(t *testing.T) {
	executorInputJSON := `{
		"inputs": {
			"parameterValues": {
				"config": {
					"category_ids": "{{$.inputs.parameters['pipelinechannel--category_ids']}}",
					"dump_filename": "{{$.inputs.parameters['pipelinechannel--dump_filename']}}",
					"sphinx_host": "{{$.inputs.parameters['pipelinechannel--sphinx_host']}}",
					"sphinx_port": "{{$.inputs.parameters['pipelinechannel--sphinx_port']}}"
				},
				"pipelinechannel--category_ids": "116",
				"pipelinechannel--dump_filename": "dump_filename_test.txt",
				"pipelinechannel--sphinx_host": "sphinx-default-host.ru",
				"pipelinechannel--sphinx_port": 9312
			}
		},
		"outputs": {
			"artifacts": {
				"dataset": {
					"artifacts": [{
						"type": {
							"schemaTitle": "system.Dataset",
							"schemaVersion": "0.0.1"
						},
						"uri": "s3://aviflow-stage-kfp-artifacts/debug-component-pipeline/ae02034e-bd96-4b8a-a06b-55c99fe9eccb/sayhello/c98ac032-2448-4637-bf37-3ad1e13a112c/dataset"
					}]
				}
			},
			"outputFile": "/tmp/kfp_outputs/output_metadata.json"
		}
	}`

	executorInput := &pipelinespec.ExecutorInput{}
	err := protojson.Unmarshal([]byte(executorInputJSON), executorInput)

	assert.NoError(t, err)

	cmd := "sh"
	args := []string{
		"--executor_input", "{{$}}",
		"--function_to_execute", "sayHello",
	}
	cmd, args, err = compileCmdAndArgs(executorInput, cmd, args)

	assert.NoError(t, err)

	var actualExecutorInput string
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--executor_input" {
			actualExecutorInput = args[i+1]
			break
		}
	}
	assert.NotEmpty(t, actualExecutorInput, "--executor_input not found")

	var parsed map[string]any
	err = json.Unmarshal([]byte(actualExecutorInput), &parsed)
	assert.NoError(t, err)

	inputs := parsed["inputs"].(map[string]any)
	paramValues := inputs["parameterValues"].(map[string]any)
	config := paramValues["config"].(map[string]any)

	assert.Equal(t, "116", config["category_ids"])
	assert.Equal(t, "dump_filename_test.txt", config["dump_filename"])
	assert.Equal(t, "sphinx-default-host.ru", config["sphinx_host"])
	assert.Equal(t, "9312", config["sphinx_port"])
}

func Test_get_log_Writer(t *testing.T) {
	old := osCreateFunc
	defer func() { osCreateFunc = old }()

	osCreateFunc = func(name string) (*os.File, error) {
		tmpdir := t.TempDir()
		file, _ := os.CreateTemp(tmpdir, "*")
		return file, nil
	}

	tests := []struct {
		name        string
		artifacts   map[string]*pipelinespec.ArtifactList
		multiWriter bool
	}{
		{
			"single writer - no key logs",
			map[string]*pipelinespec.ArtifactList{
				"notLog": {},
			},
			false,
		},
		{
			"single writer - key log has empty list",
			map[string]*pipelinespec.ArtifactList{
				"logs": {
					Artifacts: []*pipelinespec.RuntimeArtifact{},
				},
			},
			false,
		},
		{
			"single writer - malformed uri",
			map[string]*pipelinespec.ArtifactList{
				"logs": {
					Artifacts: []*pipelinespec.RuntimeArtifact{
						{
							Uri: "",
						},
					},
				},
			},
			false,
		},
		{
			"multiwriter",
			map[string]*pipelinespec.ArtifactList{
				"executor-logs": {
					Artifacts: []*pipelinespec.RuntimeArtifact{
						{
							Uri: "minio://testinguri",
						},
					},
				},
			},
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := getLogWriter(test.artifacts)
			if test.multiWriter == false {
				assert.Equal(t, os.Stdout, writer)
			} else {
				assert.IsType(t, io.MultiWriter(), writer)
			}
		})
	}
}

// Tests happy and unhappy paths for constructing a new LauncherV2
func Test_NewLauncherV2(t *testing.T) {
	var testCmdArgs = []string{"sh", "-c", "echo \"hello world\""}

	disabledCacheClient, _ := cacheutils.NewClient("ml-pipeline.kubeflow", "8887", true, &tls.Config{})
	var testLauncherV2Deps = client_manager.NewFakeClientManager(
		fake.NewSimpleClientset(),
		metadata.NewFakeClient(),
		disabledCacheClient,
	)

	var testValidLauncherV2Opts = LauncherV2Options{
		Namespace:         "my-namespace",
		PodName:           "my-pod",
		PodUID:            "abcd",
		MLMDServerAddress: "example.com",
		MLMDServerPort:    "1234",
	}

	type args struct {
		executionID       int64
		executorInputJSON string
		componentSpecJSON string
		cmdArgs           []string
		opts              LauncherV2Options
		cm                client_manager.ClientManagerInterface
	}
	tests := []struct {
		name        string
		args        *args
		expectedErr error
	}{
		{
			name: "happy path",
			args: &args{
				executionID:       1,
				executorInputJSON: "{}",
				componentSpecJSON: "{}",
				cmdArgs:           testCmdArgs,
				opts:              testValidLauncherV2Opts,
				cm:                testLauncherV2Deps,
			},
			expectedErr: nil,
		},
		{
			name: "missing executionID",
			args: &args{
				executionID: 0,
			},
			expectedErr: errors.New("must specify execution ID"),
		},
		{
			name: "invalid executorInput",
			args: &args{
				executionID:       1,
				executorInputJSON: "{",
			},
			expectedErr: errors.New("unexpected EOF"),
		},
		{
			name: "invalid componentSpec",
			args: &args{
				executionID:       1,
				executorInputJSON: "{}",
				componentSpecJSON: "{",
			},
			expectedErr: errors.New("unexpected EOF\ncomponentSpec: {"),
		},
		{
			name: "missing cmdArgs",
			args: &args{
				executionID:       1,
				executorInputJSON: "{}",
				componentSpecJSON: "{}",
				cmdArgs:           []string{},
			},
			expectedErr: errors.New("command and arguments are empty"),
		},
		{
			name: "invalid opts",
			args: &args{
				executionID:       1,
				executorInputJSON: "{}",
				componentSpecJSON: "{}",
				cmdArgs:           testCmdArgs,
				opts:              LauncherV2Options{},
			},
			expectedErr: errors.New("invalid launcher options: must specify Namespace"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := test.args
			_, err := NewLauncherV2(context.Background(), args.executionID, args.executorInputJSON, args.componentSpecJSON, args.cmdArgs, &args.opts, args.cm)
			if test.expectedErr != nil {
				assert.ErrorContains(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_retrieve_artifact_path(t *testing.T) {
	customPath := "/var/lib/kubelet/pods/pod-uid/volumes/kubernetes.io~csi/pvc-uuid/mount"
	tests := []struct {
		name         string
		artifact     *pipelinespec.RuntimeArtifact
		expectedPath string
	}{
		{
			"Artifact with no custom path",
			&pipelinespec.RuntimeArtifact{
				Uri: "gs://bucket/path/to/artifact",
			},
			"/gcs/bucket/path/to/artifact",
		},
		{
			"Artifact with custom path",
			&pipelinespec.RuntimeArtifact{
				Uri:        "gs://bucket/path/to/artifact",
				CustomPath: &customPath,
			},
			customPath,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path, err := retrieveArtifactPath(test.artifact)
			assert.Nil(t, err)
			assert.Equal(t, path, test.expectedPath)
		})
	}
}
