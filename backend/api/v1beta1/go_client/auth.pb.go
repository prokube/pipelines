// Copyright 2020 The Kubeflow Authors
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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.31.1
// source: backend/api/v1beta1/auth.proto

package go_client

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Type of resources in pipelines system.
type AuthorizeRequest_Resources int32

const (
	AuthorizeRequest_UNASSIGNED_RESOURCES AuthorizeRequest_Resources = 0
	AuthorizeRequest_VIEWERS              AuthorizeRequest_Resources = 1
)

// Enum value maps for AuthorizeRequest_Resources.
var (
	AuthorizeRequest_Resources_name = map[int32]string{
		0: "UNASSIGNED_RESOURCES",
		1: "VIEWERS",
	}
	AuthorizeRequest_Resources_value = map[string]int32{
		"UNASSIGNED_RESOURCES": 0,
		"VIEWERS":              1,
	}
)

func (x AuthorizeRequest_Resources) Enum() *AuthorizeRequest_Resources {
	p := new(AuthorizeRequest_Resources)
	*p = x
	return p
}

func (x AuthorizeRequest_Resources) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AuthorizeRequest_Resources) Descriptor() protoreflect.EnumDescriptor {
	return file_backend_api_v1beta1_auth_proto_enumTypes[0].Descriptor()
}

func (AuthorizeRequest_Resources) Type() protoreflect.EnumType {
	return &file_backend_api_v1beta1_auth_proto_enumTypes[0]
}

func (x AuthorizeRequest_Resources) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AuthorizeRequest_Resources.Descriptor instead.
func (AuthorizeRequest_Resources) EnumDescriptor() ([]byte, []int) {
	return file_backend_api_v1beta1_auth_proto_rawDescGZIP(), []int{0, 0}
}

// Type of verbs that act on the resources.
type AuthorizeRequest_Verb int32

const (
	AuthorizeRequest_UNASSIGNED_VERB AuthorizeRequest_Verb = 0
	AuthorizeRequest_CREATE          AuthorizeRequest_Verb = 1
	AuthorizeRequest_GET             AuthorizeRequest_Verb = 2
	AuthorizeRequest_DELETE          AuthorizeRequest_Verb = 3
)

// Enum value maps for AuthorizeRequest_Verb.
var (
	AuthorizeRequest_Verb_name = map[int32]string{
		0: "UNASSIGNED_VERB",
		1: "CREATE",
		2: "GET",
		3: "DELETE",
	}
	AuthorizeRequest_Verb_value = map[string]int32{
		"UNASSIGNED_VERB": 0,
		"CREATE":          1,
		"GET":             2,
		"DELETE":          3,
	}
)

func (x AuthorizeRequest_Verb) Enum() *AuthorizeRequest_Verb {
	p := new(AuthorizeRequest_Verb)
	*p = x
	return p
}

func (x AuthorizeRequest_Verb) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AuthorizeRequest_Verb) Descriptor() protoreflect.EnumDescriptor {
	return file_backend_api_v1beta1_auth_proto_enumTypes[1].Descriptor()
}

func (AuthorizeRequest_Verb) Type() protoreflect.EnumType {
	return &file_backend_api_v1beta1_auth_proto_enumTypes[1]
}

func (x AuthorizeRequest_Verb) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AuthorizeRequest_Verb.Descriptor instead.
func (AuthorizeRequest_Verb) EnumDescriptor() ([]byte, []int) {
	return file_backend_api_v1beta1_auth_proto_rawDescGZIP(), []int{0, 1}
}

// Ask for authorization of an access by providing resource's namespace, type
// and verb. User identity is not part of the message, because it is expected
// to be parsed from request headers. Caller should proxy user request's headers.
type AuthorizeRequest struct {
	state         protoimpl.MessageState     `protogen:"open.v1"`
	Namespace     string                     `protobuf:"bytes,1,opt,name=namespace,proto3" json:"namespace,omitempty"`                                      // Namespace the resource belongs to.
	Resources     AuthorizeRequest_Resources `protobuf:"varint,2,opt,name=resources,proto3,enum=api.AuthorizeRequest_Resources" json:"resources,omitempty"` // Resource type asking for authorization.
	Verb          AuthorizeRequest_Verb      `protobuf:"varint,3,opt,name=verb,proto3,enum=api.AuthorizeRequest_Verb" json:"verb,omitempty"`                // Verb on the resource asking for authorization.
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AuthorizeRequest) Reset() {
	*x = AuthorizeRequest{}
	mi := &file_backend_api_v1beta1_auth_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuthorizeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthorizeRequest) ProtoMessage() {}

func (x *AuthorizeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_backend_api_v1beta1_auth_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthorizeRequest.ProtoReflect.Descriptor instead.
func (*AuthorizeRequest) Descriptor() ([]byte, []int) {
	return file_backend_api_v1beta1_auth_proto_rawDescGZIP(), []int{0}
}

func (x *AuthorizeRequest) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *AuthorizeRequest) GetResources() AuthorizeRequest_Resources {
	if x != nil {
		return x.Resources
	}
	return AuthorizeRequest_UNASSIGNED_RESOURCES
}

func (x *AuthorizeRequest) GetVerb() AuthorizeRequest_Verb {
	if x != nil {
		return x.Verb
	}
	return AuthorizeRequest_UNASSIGNED_VERB
}

var File_backend_api_v1beta1_auth_proto protoreflect.FileDescriptor

const file_backend_api_v1beta1_auth_proto_rawDesc = "" +
	"\n" +
	"\x1ebackend/api/v1beta1/auth.proto\x12\x03api\x1a\x1cgoogle/api/annotations.proto\x1a\x1bgoogle/protobuf/empty.proto\x1a.protoc-gen-openapiv2/options/annotations.proto\"\x91\x02\n" +
	"\x10AuthorizeRequest\x12\x1c\n" +
	"\tnamespace\x18\x01 \x01(\tR\tnamespace\x12=\n" +
	"\tresources\x18\x02 \x01(\x0e2\x1f.api.AuthorizeRequest.ResourcesR\tresources\x12.\n" +
	"\x04verb\x18\x03 \x01(\x0e2\x1a.api.AuthorizeRequest.VerbR\x04verb\"2\n" +
	"\tResources\x12\x18\n" +
	"\x14UNASSIGNED_RESOURCES\x10\x00\x12\v\n" +
	"\aVIEWERS\x10\x01\"<\n" +
	"\x04Verb\x12\x13\n" +
	"\x0fUNASSIGNED_VERB\x10\x00\x12\n" +
	"\n" +
	"\x06CREATE\x10\x01\x12\a\n" +
	"\x03GET\x10\x02\x12\n" +
	"\n" +
	"\x06DELETE\x10\x032g\n" +
	"\vAuthService\x12X\n" +
	"\vAuthorizeV1\x12\x15.api.AuthorizeRequest\x1a\x16.google.protobuf.Empty\"\x1a\x82\xd3\xe4\x93\x02\x14\x12\x12/apis/v1beta1/authB\x91\x01\x92AQ*\x02\x01\x02R\x1c\n" +
	"\adefault\x12\x11\x12\x0f\n" +
	"\r\x1a\v.api.StatusZ\x1f\n" +
	"\x1d\n" +
	"\x06Bearer\x12\x13\b\x02\x1a\rauthorization \x02b\f\n" +
	"\n" +
	"\n" +
	"\x06Bearer\x12\x00Z;github.com/kubeflow/pipelines/backend/api/v1beta1/go_clientb\x06proto3"

var (
	file_backend_api_v1beta1_auth_proto_rawDescOnce sync.Once
	file_backend_api_v1beta1_auth_proto_rawDescData []byte
)

func file_backend_api_v1beta1_auth_proto_rawDescGZIP() []byte {
	file_backend_api_v1beta1_auth_proto_rawDescOnce.Do(func() {
		file_backend_api_v1beta1_auth_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_backend_api_v1beta1_auth_proto_rawDesc), len(file_backend_api_v1beta1_auth_proto_rawDesc)))
	})
	return file_backend_api_v1beta1_auth_proto_rawDescData
}

var file_backend_api_v1beta1_auth_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_backend_api_v1beta1_auth_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_backend_api_v1beta1_auth_proto_goTypes = []any{
	(AuthorizeRequest_Resources)(0), // 0: api.AuthorizeRequest.Resources
	(AuthorizeRequest_Verb)(0),      // 1: api.AuthorizeRequest.Verb
	(*AuthorizeRequest)(nil),        // 2: api.AuthorizeRequest
	(*emptypb.Empty)(nil),           // 3: google.protobuf.Empty
}
var file_backend_api_v1beta1_auth_proto_depIdxs = []int32{
	0, // 0: api.AuthorizeRequest.resources:type_name -> api.AuthorizeRequest.Resources
	1, // 1: api.AuthorizeRequest.verb:type_name -> api.AuthorizeRequest.Verb
	2, // 2: api.AuthService.AuthorizeV1:input_type -> api.AuthorizeRequest
	3, // 3: api.AuthService.AuthorizeV1:output_type -> google.protobuf.Empty
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_backend_api_v1beta1_auth_proto_init() }
func file_backend_api_v1beta1_auth_proto_init() {
	if File_backend_api_v1beta1_auth_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_backend_api_v1beta1_auth_proto_rawDesc), len(file_backend_api_v1beta1_auth_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_backend_api_v1beta1_auth_proto_goTypes,
		DependencyIndexes: file_backend_api_v1beta1_auth_proto_depIdxs,
		EnumInfos:         file_backend_api_v1beta1_auth_proto_enumTypes,
		MessageInfos:      file_backend_api_v1beta1_auth_proto_msgTypes,
	}.Build()
	File_backend_api_v1beta1_auth_proto = out.File
	file_backend_api_v1beta1_auth_proto_goTypes = nil
	file_backend_api_v1beta1_auth_proto_depIdxs = nil
}
