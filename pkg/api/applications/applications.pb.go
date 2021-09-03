// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.1
// source: api/applications/applications.proto

package applications

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This object represents a single condition for a Kubernetes object.
// It roughly matches the Kubernetes type defined here: https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Condition
type Condition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type      string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Status    string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	Reason    string `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
	Message   string `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`
	Timestamp int32  `protobuf:"varint,5,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *Condition) Reset() {
	*x = Condition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_applications_applications_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Condition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Condition) ProtoMessage() {}

func (x *Condition) ProtoReflect() protoreflect.Message {
	mi := &file_api_applications_applications_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Condition.ProtoReflect.Descriptor instead.
func (*Condition) Descriptor() ([]byte, []int) {
	return file_api_applications_applications_proto_rawDescGZIP(), []int{0}
}

func (x *Condition) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Condition) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *Condition) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *Condition) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Condition) GetTimestamp() int32 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type Application struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name                 string       `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`                                                             // The name of the application
	Path                 string       `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`                                                             // The file path where the k8s yaml files for this application are stored.
	Url                  string       `protobuf:"bytes,3,opt,name=url,proto3" json:"url,omitempty"`                                                               // The git repository URL for this application
	SourceConditions     []*Condition `protobuf:"bytes,4,rep,name=source_conditions,json=sourceConditions,proto3" json:"source_conditions,omitempty"`             // A list of conditions for the Source related to this Application
	DeploymentConditions []*Condition `protobuf:"bytes,5,rep,name=deployment_conditions,json=deploymentConditions,proto3" json:"deployment_conditions,omitempty"` // A list of conditions for the Kustomization or HelmRelease for this application
}

func (x *Application) Reset() {
	*x = Application{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_applications_applications_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Application) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Application) ProtoMessage() {}

func (x *Application) ProtoReflect() protoreflect.Message {
	mi := &file_api_applications_applications_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Application.ProtoReflect.Descriptor instead.
func (*Application) Descriptor() ([]byte, []int) {
	return file_api_applications_applications_proto_rawDescGZIP(), []int{1}
}

func (x *Application) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Application) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Application) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Application) GetSourceConditions() []*Condition {
	if x != nil {
		return x.SourceConditions
	}
	return nil
}

func (x *Application) GetDeploymentConditions() []*Condition {
	if x != nil {
		return x.DeploymentConditions
	}
	return nil
}

type AuthenticateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProviderName string `protobuf:"bytes,1,opt,name=provider_name,json=providerName,proto3" json:"provider_name,omitempty"` // The name of the git provider
	AccessToken  string `protobuf:"bytes,2,opt,name=access_token,json=accessToken,proto3" json:"access_token,omitempty"`    // The token of the git provider
}

func (x *AuthenticateRequest) Reset() {
	*x = AuthenticateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_applications_applications_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthenticateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticateRequest) ProtoMessage() {}

func (x *AuthenticateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_applications_applications_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticateRequest.ProtoReflect.Descriptor instead.
func (*AuthenticateRequest) Descriptor() ([]byte, []int) {
	return file_api_applications_applications_proto_rawDescGZIP(), []int{2}
}

func (x *AuthenticateRequest) GetProviderName() string {
	if x != nil {
		return x.ProviderName
	}
	return ""
}

func (x *AuthenticateRequest) GetAccessToken() string {
	if x != nil {
		return x.AccessToken
	}
	return ""
}

type AuthenticateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"` // The jwt token that was generated using git provider name and git provider token
}

func (x *AuthenticateResponse) Reset() {
	*x = AuthenticateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_applications_applications_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthenticateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticateResponse) ProtoMessage() {}

func (x *AuthenticateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_applications_applications_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticateResponse.ProtoReflect.Descriptor instead.
func (*AuthenticateResponse) Descriptor() ([]byte, []int) {
	return file_api_applications_applications_proto_rawDescGZIP(), []int{3}
}

func (x *AuthenticateResponse) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

type ListApplicationsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Namespace string `protobuf:"bytes,1,opt,name=namespace,proto3" json:"namespace,omitempty"` // The namespace to look for applications
}

func (x *ListApplicationsRequest) Reset() {
	*x = ListApplicationsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_applications_applications_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListApplicationsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListApplicationsRequest) ProtoMessage() {}

func (x *ListApplicationsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_applications_applications_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListApplicationsRequest.ProtoReflect.Descriptor instead.
func (*ListApplicationsRequest) Descriptor() ([]byte, []int) {
	return file_api_applications_applications_proto_rawDescGZIP(), []int{4}
}

func (x *ListApplicationsRequest) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

type ListApplicationsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Applications []*Application `protobuf:"bytes,1,rep,name=applications,proto3" json:"applications,omitempty"` // A list of applications
}

func (x *ListApplicationsResponse) Reset() {
	*x = ListApplicationsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_applications_applications_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListApplicationsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListApplicationsResponse) ProtoMessage() {}

func (x *ListApplicationsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_applications_applications_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListApplicationsResponse.ProtoReflect.Descriptor instead.
func (*ListApplicationsResponse) Descriptor() ([]byte, []int) {
	return file_api_applications_applications_proto_rawDescGZIP(), []int{5}
}

func (x *ListApplicationsResponse) GetApplications() []*Application {
	if x != nil {
		return x.Applications
	}
	return nil
}

type GetApplicationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`           // The name of an application
	Namespace string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"` // The kubernetes namespace of the application. Default is `wego-system`
}

func (x *GetApplicationRequest) Reset() {
	*x = GetApplicationRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_applications_applications_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetApplicationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetApplicationRequest) ProtoMessage() {}

func (x *GetApplicationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_applications_applications_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetApplicationRequest.ProtoReflect.Descriptor instead.
func (*GetApplicationRequest) Descriptor() ([]byte, []int) {
	return file_api_applications_applications_proto_rawDescGZIP(), []int{6}
}

func (x *GetApplicationRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetApplicationRequest) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

type GetApplicationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Application *Application `protobuf:"bytes,1,opt,name=application,proto3" json:"application,omitempty"`
}

func (x *GetApplicationResponse) Reset() {
	*x = GetApplicationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_applications_applications_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetApplicationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetApplicationResponse) ProtoMessage() {}

func (x *GetApplicationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_applications_applications_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetApplicationResponse.ProtoReflect.Descriptor instead.
func (*GetApplicationResponse) Descriptor() ([]byte, []int) {
	return file_api_applications_applications_proto_rawDescGZIP(), []int{7}
}

func (x *GetApplicationResponse) GetApplication() *Application {
	if x != nil {
		return x.Application
	}
	return nil
}

type Commit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CommitHash string `protobuf:"bytes,1,opt,name=commit_hash,json=commitHash,proto3" json:"commit_hash,omitempty"` // The hash of the commit
	Date       string `protobuf:"bytes,2,opt,name=date,proto3" json:"date,omitempty"`                               // The date the commit was made.
	Author     string `protobuf:"bytes,3,opt,name=author,proto3" json:"author,omitempty"`                           // The author of the commit
	Message    string `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`                         // The commit message
}

func (x *Commit) Reset() {
	*x = Commit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_applications_applications_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Commit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Commit) ProtoMessage() {}

func (x *Commit) ProtoReflect() protoreflect.Message {
	mi := &file_api_applications_applications_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Commit.ProtoReflect.Descriptor instead.
func (*Commit) Descriptor() ([]byte, []int) {
	return file_api_applications_applications_proto_rawDescGZIP(), []int{8}
}

func (x *Commit) GetCommitHash() string {
	if x != nil {
		return x.CommitHash
	}
	return ""
}

func (x *Commit) GetDate() string {
	if x != nil {
		return x.Date
	}
	return ""
}

func (x *Commit) GetAuthor() string {
	if x != nil {
		return x.Author
	}
	return ""
}

func (x *Commit) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type ListCommitsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`           // The application name
	Namespace string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"` // The namespace the application is in
	PageSize  int32  `protobuf:"varint,3,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	// Optional. A pagination token returned from a previous call
	// that indicates where this listing should continue from.
	PageToken *int32 `protobuf:"varint,4,opt,name=page_token,json=pageToken,proto3,oneof" json:"page_token,omitempty"`
}

func (x *ListCommitsRequest) Reset() {
	*x = ListCommitsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_applications_applications_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListCommitsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListCommitsRequest) ProtoMessage() {}

func (x *ListCommitsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_applications_applications_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListCommitsRequest.ProtoReflect.Descriptor instead.
func (*ListCommitsRequest) Descriptor() ([]byte, []int) {
	return file_api_applications_applications_proto_rawDescGZIP(), []int{9}
}

func (x *ListCommitsRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ListCommitsRequest) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *ListCommitsRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *ListCommitsRequest) GetPageToken() int32 {
	if x != nil && x.PageToken != nil {
		return *x.PageToken
	}
	return 0
}

type ListCommitsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Commits []*Commit `protobuf:"bytes,1,rep,name=commits,proto3" json:"commits,omitempty"` // A list of commits
	// A pagination token returned from a previous call
	// that indicates from where listing should continue.
	NextPageToken int32 `protobuf:"varint,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
}

func (x *ListCommitsResponse) Reset() {
	*x = ListCommitsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_applications_applications_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListCommitsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListCommitsResponse) ProtoMessage() {}

func (x *ListCommitsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_applications_applications_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListCommitsResponse.ProtoReflect.Descriptor instead.
func (*ListCommitsResponse) Descriptor() ([]byte, []int) {
	return file_api_applications_applications_proto_rawDescGZIP(), []int{10}
}

func (x *ListCommitsResponse) GetCommits() []*Commit {
	if x != nil {
		return x.Commits
	}
	return nil
}

func (x *ListCommitsResponse) GetNextPageToken() int32 {
	if x != nil {
		return x.NextPageToken
	}
	return 0
}

var File_api_applications_applications_proto protoreflect.FileDescriptor

var file_api_applications_applications_proto_rawDesc = []byte{
	0x0a, 0x23, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x77, 0x65, 0x67, 0x6f, 0x5f, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d,
	0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x87, 0x01, 0x0a, 0x09, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x16, 0x0a,
	0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72,
	0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0xdf, 0x01,
	0x0a, 0x0b, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x46, 0x0a, 0x11, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x04, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x19, 0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x10, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12,
	0x4e, 0x0a, 0x15, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x6f,
	0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19,
	0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x14, 0x64, 0x65, 0x70, 0x6c, 0x6f,
	0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22,
	0x5d, 0x0a, 0x13, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64,
	0x65, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70,
	0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x61,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x2c,
	0x0a, 0x14, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x37, 0x0a, 0x17,
	0x4c, 0x69, 0x73, 0x74, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x22, 0x5b, 0x0a, 0x18, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x70, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x3f, 0x0a, 0x0c, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0c, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x22, 0x49, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x22, 0x57, 0x0a,
	0x16, 0x47, 0x65, 0x74, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3d, 0x0a, 0x0b, 0x61, 0x70, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x77,
	0x65, 0x67, 0x6f, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x70,
	0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x61, 0x70, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x6f, 0x0a, 0x06, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74,
	0x12, 0x1f, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x48, 0x61, 0x73,
	0x68, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x96, 0x01, 0x0a, 0x12, 0x4c, 0x69, 0x73, 0x74,
	0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x22, 0x0a,
	0x0a, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x05, 0x48, 0x00, 0x52, 0x09, 0x70, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x88, 0x01,
	0x01, 0x42, 0x0d, 0x0a, 0x0b, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e,
	0x22, 0x6f, 0x0a, 0x13, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x30, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x69,
	0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74,
	0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73, 0x12, 0x26, 0x0a, 0x0f, 0x6e, 0x65, 0x78,
	0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0d, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x32, 0x9c, 0x04, 0x0a, 0x0c, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x12, 0x86, 0x01, 0x0a, 0x0c, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x12, 0x23, 0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e,
	0x74, 0x69, 0x63, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2b,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x25, 0x22, 0x20, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x75, 0x74, 0x68,
	0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65, 0x2f, 0x7b, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64,
	0x65, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x7d, 0x3a, 0x01, 0x2a, 0x12, 0x7f, 0x0a, 0x10, 0x4c,
	0x69, 0x73, 0x74, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12,
	0x27, 0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28, 0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x70,
	0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x18, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x12, 0x10, 0x2f, 0x76, 0x31, 0x2f,
	0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x80, 0x01, 0x0a,
	0x0e, 0x47, 0x65, 0x74, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x25, 0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x70, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1f,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x19, 0x12, 0x17, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x70, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x7b, 0x6e, 0x61, 0x6d, 0x65, 0x7d, 0x12,
	0x7f, 0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73, 0x12, 0x22,
	0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x4c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x23, 0x2e, 0x77, 0x65, 0x67, 0x6f, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x27, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x21, 0x12,
	0x1f, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2f, 0x7b, 0x6e, 0x61, 0x6d, 0x65, 0x7d, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73,
	0x42, 0xce, 0x01, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x77, 0x65, 0x61, 0x76, 0x65, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x2f, 0x77, 0x65, 0x61, 0x76, 0x65,
	0x2d, 0x67, 0x69, 0x74, 0x6f, 0x70, 0x73, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x73, 0x2f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x92,
	0x41, 0x8e, 0x01, 0x12, 0x68, 0x0a, 0x15, 0x57, 0x65, 0x47, 0x6f, 0x20, 0x41, 0x70, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x20, 0x41, 0x50, 0x49, 0x12, 0x4a, 0x54, 0x68,
	0x65, 0x20, 0x57, 0x65, 0x47, 0x6f, 0x20, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x20, 0x41, 0x50, 0x49, 0x20, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x73, 0x20,
	0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x20, 0x66, 0x6f, 0x72, 0x20, 0x57,
	0x65, 0x61, 0x76, 0x65, 0x20, 0x47, 0x69, 0x74, 0x4f, 0x70, 0x73, 0x20, 0x41, 0x70, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x32, 0x03, 0x30, 0x2e, 0x31, 0x32, 0x10, 0x61,
	0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x3a,
	0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f,
	0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_applications_applications_proto_rawDescOnce sync.Once
	file_api_applications_applications_proto_rawDescData = file_api_applications_applications_proto_rawDesc
)

func file_api_applications_applications_proto_rawDescGZIP() []byte {
	file_api_applications_applications_proto_rawDescOnce.Do(func() {
		file_api_applications_applications_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_applications_applications_proto_rawDescData)
	})
	return file_api_applications_applications_proto_rawDescData
}

var file_api_applications_applications_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_api_applications_applications_proto_goTypes = []interface{}{
	(*Condition)(nil),                // 0: wego_server.v1.Condition
	(*Application)(nil),              // 1: wego_server.v1.Application
	(*AuthenticateRequest)(nil),      // 2: wego_server.v1.AuthenticateRequest
	(*AuthenticateResponse)(nil),     // 3: wego_server.v1.AuthenticateResponse
	(*ListApplicationsRequest)(nil),  // 4: wego_server.v1.ListApplicationsRequest
	(*ListApplicationsResponse)(nil), // 5: wego_server.v1.ListApplicationsResponse
	(*GetApplicationRequest)(nil),    // 6: wego_server.v1.GetApplicationRequest
	(*GetApplicationResponse)(nil),   // 7: wego_server.v1.GetApplicationResponse
	(*Commit)(nil),                   // 8: wego_server.v1.Commit
	(*ListCommitsRequest)(nil),       // 9: wego_server.v1.ListCommitsRequest
	(*ListCommitsResponse)(nil),      // 10: wego_server.v1.ListCommitsResponse
}
var file_api_applications_applications_proto_depIdxs = []int32{
	0,  // 0: wego_server.v1.Application.source_conditions:type_name -> wego_server.v1.Condition
	0,  // 1: wego_server.v1.Application.deployment_conditions:type_name -> wego_server.v1.Condition
	1,  // 2: wego_server.v1.ListApplicationsResponse.applications:type_name -> wego_server.v1.Application
	1,  // 3: wego_server.v1.GetApplicationResponse.application:type_name -> wego_server.v1.Application
	8,  // 4: wego_server.v1.ListCommitsResponse.commits:type_name -> wego_server.v1.Commit
	2,  // 5: wego_server.v1.Applications.Authenticate:input_type -> wego_server.v1.AuthenticateRequest
	4,  // 6: wego_server.v1.Applications.ListApplications:input_type -> wego_server.v1.ListApplicationsRequest
	6,  // 7: wego_server.v1.Applications.GetApplication:input_type -> wego_server.v1.GetApplicationRequest
	9,  // 8: wego_server.v1.Applications.ListCommits:input_type -> wego_server.v1.ListCommitsRequest
	3,  // 9: wego_server.v1.Applications.Authenticate:output_type -> wego_server.v1.AuthenticateResponse
	5,  // 10: wego_server.v1.Applications.ListApplications:output_type -> wego_server.v1.ListApplicationsResponse
	7,  // 11: wego_server.v1.Applications.GetApplication:output_type -> wego_server.v1.GetApplicationResponse
	10, // 12: wego_server.v1.Applications.ListCommits:output_type -> wego_server.v1.ListCommitsResponse
	9,  // [9:13] is the sub-list for method output_type
	5,  // [5:9] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_api_applications_applications_proto_init() }
func file_api_applications_applications_proto_init() {
	if File_api_applications_applications_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_applications_applications_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Condition); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_applications_applications_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Application); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_applications_applications_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthenticateRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_applications_applications_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthenticateResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_applications_applications_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListApplicationsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_applications_applications_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListApplicationsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_applications_applications_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetApplicationRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_applications_applications_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetApplicationResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_applications_applications_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Commit); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_applications_applications_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListCommitsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_applications_applications_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListCommitsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_api_applications_applications_proto_msgTypes[9].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_applications_applications_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_applications_applications_proto_goTypes,
		DependencyIndexes: file_api_applications_applications_proto_depIdxs,
		MessageInfos:      file_api_applications_applications_proto_msgTypes,
	}.Build()
	File_api_applications_applications_proto = out.File
	file_api_applications_applications_proto_rawDesc = nil
	file_api_applications_applications_proto_goTypes = nil
	file_api_applications_applications_proto_depIdxs = nil
}
