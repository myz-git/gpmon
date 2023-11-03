//grpc/proto/dbstatus.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v4.24.4
// source: proto/dbstatus.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type DatabaseStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ip          string                 `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Port        int32                  `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"` // 确保与ClientInfo中的port类型匹配
	Dbtype      string                 `protobuf:"bytes,3,opt,name=dbtype,proto3" json:"dbtype,omitempty"`
	Dbnm        string                 `protobuf:"bytes,4,opt,name=dbnm,proto3" json:"dbnm,omitempty"`
	CheckNm     string                 `protobuf:"bytes,5,opt,name=checkNm,proto3" json:"checkNm,omitempty"`         // 检查名称
	CheckResult string                 `protobuf:"bytes,6,opt,name=checkResult,proto3" json:"checkResult,omitempty"` // 检查结果
	Details     string                 `protobuf:"bytes,7,opt,name=details,proto3" json:"details,omitempty"`
	Timestamp   *timestamppb.Timestamp `protobuf:"bytes,8,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *DatabaseStatus) Reset() {
	*x = DatabaseStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_dbstatus_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DatabaseStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DatabaseStatus) ProtoMessage() {}

func (x *DatabaseStatus) ProtoReflect() protoreflect.Message {
	mi := &file_proto_dbstatus_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DatabaseStatus.ProtoReflect.Descriptor instead.
func (*DatabaseStatus) Descriptor() ([]byte, []int) {
	return file_proto_dbstatus_proto_rawDescGZIP(), []int{0}
}

func (x *DatabaseStatus) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *DatabaseStatus) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *DatabaseStatus) GetDbtype() string {
	if x != nil {
		return x.Dbtype
	}
	return ""
}

func (x *DatabaseStatus) GetDbnm() string {
	if x != nil {
		return x.Dbnm
	}
	return ""
}

func (x *DatabaseStatus) GetCheckNm() string {
	if x != nil {
		return x.CheckNm
	}
	return ""
}

func (x *DatabaseStatus) GetCheckResult() string {
	if x != nil {
		return x.CheckResult
	}
	return ""
}

func (x *DatabaseStatus) GetDetails() string {
	if x != nil {
		return x.Details
	}
	return ""
}

func (x *DatabaseStatus) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

type DatabaseStatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *DatabaseStatusResponse) Reset() {
	*x = DatabaseStatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_dbstatus_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DatabaseStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DatabaseStatusResponse) ProtoMessage() {}

func (x *DatabaseStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_dbstatus_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DatabaseStatusResponse.ProtoReflect.Descriptor instead.
func (*DatabaseStatusResponse) Descriptor() ([]byte, []int) {
	return file_proto_dbstatus_proto_rawDescGZIP(), []int{1}
}

func (x *DatabaseStatusResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type ClientInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DbType string `protobuf:"bytes,1,opt,name=DbType,proto3" json:"DbType,omitempty"` // 使用dbtype作为请求参数
}

func (x *ClientInfoRequest) Reset() {
	*x = ClientInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_dbstatus_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientInfoRequest) ProtoMessage() {}

func (x *ClientInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_dbstatus_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientInfoRequest.ProtoReflect.Descriptor instead.
func (*ClientInfoRequest) Descriptor() ([]byte, []int) {
	return file_proto_dbstatus_proto_rawDescGZIP(), []int{2}
}

func (x *ClientInfoRequest) GetDbType() string {
	if x != nil {
		return x.DbType
	}
	return ""
}

type ClientInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientInfos []*ClientInfo `protobuf:"bytes,1,rep,name=clientInfos,proto3" json:"clientInfos,omitempty"`
}

func (x *ClientInfoResponse) Reset() {
	*x = ClientInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_dbstatus_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientInfoResponse) ProtoMessage() {}

func (x *ClientInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_dbstatus_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientInfoResponse.ProtoReflect.Descriptor instead.
func (*ClientInfoResponse) Descriptor() ([]byte, []int) {
	return file_proto_dbstatus_proto_rawDescGZIP(), []int{3}
}

func (x *ClientInfoResponse) GetClientInfos() []*ClientInfo {
	if x != nil {
		return x.ClientInfos
	}
	return nil
}

type ClientInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ip       string `protobuf:"bytes,1,opt,name=Ip,proto3" json:"Ip,omitempty"`
	Port     int32  `protobuf:"varint,2,opt,name=Port,proto3" json:"Port,omitempty"`
	DbType   string `protobuf:"bytes,3,opt,name=DbType,proto3" json:"DbType,omitempty"`
	DbName   string `protobuf:"bytes,4,opt,name=DbName,proto3" json:"DbName,omitempty"`
	DbUser   string `protobuf:"bytes,5,opt,name=DbUser,proto3" json:"DbUser,omitempty"`
	UserPwd  string `protobuf:"bytes,6,opt,name=UserPwd,proto3" json:"UserPwd,omitempty"`
	IsEnable bool   `protobuf:"varint,7,opt,name=IsEnable,proto3" json:"IsEnable,omitempty"`
}

func (x *ClientInfo) Reset() {
	*x = ClientInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_dbstatus_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientInfo) ProtoMessage() {}

func (x *ClientInfo) ProtoReflect() protoreflect.Message {
	mi := &file_proto_dbstatus_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientInfo.ProtoReflect.Descriptor instead.
func (*ClientInfo) Descriptor() ([]byte, []int) {
	return file_proto_dbstatus_proto_rawDescGZIP(), []int{4}
}

func (x *ClientInfo) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *ClientInfo) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *ClientInfo) GetDbType() string {
	if x != nil {
		return x.DbType
	}
	return ""
}

func (x *ClientInfo) GetDbName() string {
	if x != nil {
		return x.DbName
	}
	return ""
}

func (x *ClientInfo) GetDbUser() string {
	if x != nil {
		return x.DbUser
	}
	return ""
}

func (x *ClientInfo) GetUserPwd() string {
	if x != nil {
		return x.UserPwd
	}
	return ""
}

func (x *ClientInfo) GetIsEnable() bool {
	if x != nil {
		return x.IsEnable
	}
	return false
}

var File_proto_dbstatus_proto protoreflect.FileDescriptor

var file_proto_dbstatus_proto_rawDesc = []byte{
	0x0a, 0x14, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x64, 0x62, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf0,
	0x01, 0x0a, 0x0e, 0x44, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x70, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x62, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x64, 0x62, 0x74, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x64, 0x62, 0x6e, 0x6d, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x62, 0x6e,
	0x6d, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x4e, 0x6d, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x4e, 0x6d, 0x12, 0x20, 0x0a, 0x0b, 0x63,
	0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0b, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x18, 0x0a,
	0x07, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x12, 0x38, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x22, 0x32, 0x0a, 0x16, 0x44, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x2b, 0x0a, 0x11, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x44, 0x62,
	0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x44, 0x62, 0x54, 0x79,
	0x70, 0x65, 0x22, 0x49, 0x0a, 0x12, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x33, 0x0a, 0x0b, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x0b, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x22, 0xae, 0x01,
	0x0a, 0x0a, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x0e, 0x0a, 0x02,
	0x49, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x49, 0x70, 0x12, 0x12, 0x0a, 0x04,
	0x50, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x50, 0x6f, 0x72, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x44, 0x62, 0x54, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x44, 0x62, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x44, 0x62, 0x4e, 0x61,
	0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x44, 0x62, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x44, 0x62, 0x55, 0x73, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x44, 0x62, 0x55, 0x73, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x55, 0x73, 0x65, 0x72,
	0x50, 0x77, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x55, 0x73, 0x65, 0x72, 0x50,
	0x77, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x49, 0x73, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x49, 0x73, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x32, 0x5b,
	0x0a, 0x15, 0x44, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x42, 0x0a, 0x0a, 0x53, 0x65, 0x6e, 0x64, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x44, 0x61,
	0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x1a, 0x1d, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x59, 0x0a, 0x11, 0x43,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x44, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x6e, 0x66,
	0x6f, 0x12, 0x18, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_dbstatus_proto_rawDescOnce sync.Once
	file_proto_dbstatus_proto_rawDescData = file_proto_dbstatus_proto_rawDesc
)

func file_proto_dbstatus_proto_rawDescGZIP() []byte {
	file_proto_dbstatus_proto_rawDescOnce.Do(func() {
		file_proto_dbstatus_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_dbstatus_proto_rawDescData)
	})
	return file_proto_dbstatus_proto_rawDescData
}

var file_proto_dbstatus_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_proto_dbstatus_proto_goTypes = []interface{}{
	(*DatabaseStatus)(nil),         // 0: proto.DatabaseStatus
	(*DatabaseStatusResponse)(nil), // 1: proto.DatabaseStatusResponse
	(*ClientInfoRequest)(nil),      // 2: proto.ClientInfoRequest
	(*ClientInfoResponse)(nil),     // 3: proto.ClientInfoResponse
	(*ClientInfo)(nil),             // 4: proto.ClientInfo
	(*timestamppb.Timestamp)(nil),  // 5: google.protobuf.Timestamp
}
var file_proto_dbstatus_proto_depIdxs = []int32{
	5, // 0: proto.DatabaseStatus.timestamp:type_name -> google.protobuf.Timestamp
	4, // 1: proto.ClientInfoResponse.clientInfos:type_name -> proto.ClientInfo
	0, // 2: proto.DatabaseStatusService.SendStatus:input_type -> proto.DatabaseStatus
	2, // 3: proto.ClientInfoService.GetClientInfo:input_type -> proto.ClientInfoRequest
	1, // 4: proto.DatabaseStatusService.SendStatus:output_type -> proto.DatabaseStatusResponse
	3, // 5: proto.ClientInfoService.GetClientInfo:output_type -> proto.ClientInfoResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proto_dbstatus_proto_init() }
func file_proto_dbstatus_proto_init() {
	if File_proto_dbstatus_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_dbstatus_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DatabaseStatus); i {
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
		file_proto_dbstatus_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DatabaseStatusResponse); i {
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
		file_proto_dbstatus_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientInfoRequest); i {
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
		file_proto_dbstatus_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientInfoResponse); i {
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
		file_proto_dbstatus_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientInfo); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_dbstatus_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_proto_dbstatus_proto_goTypes,
		DependencyIndexes: file_proto_dbstatus_proto_depIdxs,
		MessageInfos:      file_proto_dbstatus_proto_msgTypes,
	}.Build()
	File_proto_dbstatus_proto = out.File
	file_proto_dbstatus_proto_rawDesc = nil
	file_proto_dbstatus_proto_goTypes = nil
	file_proto_dbstatus_proto_depIdxs = nil
}
