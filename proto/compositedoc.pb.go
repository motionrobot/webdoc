// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/motionrobot/webdoc/proto/compositedoc.proto

package proto

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ImageElement_ImageSource int32

const (
	ImageElement_UNKNOWN          ImageElement_ImageSource = 0
	ImageElement_IMG_TAG          ImageElement_ImageSource = 1
	ImageElement_PICTURE_TAG      ImageElement_ImageSource = 2
	ImageElement_SCRIPT_LD_JSON   ImageElement_ImageSource = 3
	ImageElement_META_OG          ImageElement_ImageSource = 4
	ImageElement_META_TWITTER     ImageElement_ImageSource = 5
	ImageElement_NOSCRIPT_IMG_TAG ImageElement_ImageSource = 6
)

var ImageElement_ImageSource_name = map[int32]string{
	0: "UNKNOWN",
	1: "IMG_TAG",
	2: "PICTURE_TAG",
	3: "SCRIPT_LD_JSON",
	4: "META_OG",
	5: "META_TWITTER",
	6: "NOSCRIPT_IMG_TAG",
}

var ImageElement_ImageSource_value = map[string]int32{
	"UNKNOWN":          0,
	"IMG_TAG":          1,
	"PICTURE_TAG":      2,
	"SCRIPT_LD_JSON":   3,
	"META_OG":          4,
	"META_TWITTER":     5,
	"NOSCRIPT_IMG_TAG": 6,
}

func (x ImageElement_ImageSource) String() string {
	return proto.EnumName(ImageElement_ImageSource_name, int32(x))
}

func (ImageElement_ImageSource) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_079b79b92abe9697, []int{2, 0}
}

type ImageSrcEle struct {
	Url                  string   `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	SizeDesc             string   `protobuf:"bytes,2,opt,name=size_desc,json=sizeDesc,proto3" json:"size_desc,omitempty"`
	Media                string   `protobuf:"bytes,3,opt,name=media,proto3" json:"media,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ImageSrcEle) Reset()         { *m = ImageSrcEle{} }
func (m *ImageSrcEle) String() string { return proto.CompactTextString(m) }
func (*ImageSrcEle) ProtoMessage()    {}
func (*ImageSrcEle) Descriptor() ([]byte, []int) {
	return fileDescriptor_079b79b92abe9697, []int{0}
}

func (m *ImageSrcEle) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ImageSrcEle.Unmarshal(m, b)
}
func (m *ImageSrcEle) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ImageSrcEle.Marshal(b, m, deterministic)
}
func (m *ImageSrcEle) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ImageSrcEle.Merge(m, src)
}
func (m *ImageSrcEle) XXX_Size() int {
	return xxx_messageInfo_ImageSrcEle.Size(m)
}
func (m *ImageSrcEle) XXX_DiscardUnknown() {
	xxx_messageInfo_ImageSrcEle.DiscardUnknown(m)
}

var xxx_messageInfo_ImageSrcEle proto.InternalMessageInfo

func (m *ImageSrcEle) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *ImageSrcEle) GetSizeDesc() string {
	if m != nil {
		return m.SizeDesc
	}
	return ""
}

func (m *ImageSrcEle) GetMedia() string {
	if m != nil {
		return m.Media
	}
	return ""
}

type ImageGroupInfo struct {
	Type                 string         `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Media                string         `protobuf:"bytes,2,opt,name=media,proto3" json:"media,omitempty"`
	ImageSources         []*ImageSrcEle `protobuf:"bytes,3,rep,name=image_sources,json=imageSources,proto3" json:"image_sources,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *ImageGroupInfo) Reset()         { *m = ImageGroupInfo{} }
func (m *ImageGroupInfo) String() string { return proto.CompactTextString(m) }
func (*ImageGroupInfo) ProtoMessage()    {}
func (*ImageGroupInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_079b79b92abe9697, []int{1}
}

func (m *ImageGroupInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ImageGroupInfo.Unmarshal(m, b)
}
func (m *ImageGroupInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ImageGroupInfo.Marshal(b, m, deterministic)
}
func (m *ImageGroupInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ImageGroupInfo.Merge(m, src)
}
func (m *ImageGroupInfo) XXX_Size() int {
	return xxx_messageInfo_ImageGroupInfo.Size(m)
}
func (m *ImageGroupInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ImageGroupInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ImageGroupInfo proto.InternalMessageInfo

func (m *ImageGroupInfo) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *ImageGroupInfo) GetMedia() string {
	if m != nil {
		return m.Media
	}
	return ""
}

func (m *ImageGroupInfo) GetImageSources() []*ImageSrcEle {
	if m != nil {
		return m.ImageSources
	}
	return nil
}

type ImageElement struct {
	Url                  string                     `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Text                 string                     `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	Alt                  string                     `protobuf:"bytes,3,opt,name=alt,proto3" json:"alt,omitempty"`
	Width                int32                      `protobuf:"varint,4,opt,name=width,proto3" json:"width,omitempty"`
	Height               int32                      `protobuf:"varint,5,opt,name=height,proto3" json:"height,omitempty"`
	Sources              []ImageElement_ImageSource `protobuf:"varint,6,rep,packed,name=sources,proto3,enum=extractor.ImageElement_ImageSource" json:"sources,omitempty"`
	ImageGroups          []*ImageGroupInfo          `protobuf:"bytes,10,rep,name=image_groups,json=imageGroups,proto3" json:"image_groups,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                   `json:"-"`
	XXX_unrecognized     []byte                     `json:"-"`
	XXX_sizecache        int32                      `json:"-"`
}

func (m *ImageElement) Reset()         { *m = ImageElement{} }
func (m *ImageElement) String() string { return proto.CompactTextString(m) }
func (*ImageElement) ProtoMessage()    {}
func (*ImageElement) Descriptor() ([]byte, []int) {
	return fileDescriptor_079b79b92abe9697, []int{2}
}

func (m *ImageElement) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ImageElement.Unmarshal(m, b)
}
func (m *ImageElement) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ImageElement.Marshal(b, m, deterministic)
}
func (m *ImageElement) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ImageElement.Merge(m, src)
}
func (m *ImageElement) XXX_Size() int {
	return xxx_messageInfo_ImageElement.Size(m)
}
func (m *ImageElement) XXX_DiscardUnknown() {
	xxx_messageInfo_ImageElement.DiscardUnknown(m)
}

var xxx_messageInfo_ImageElement proto.InternalMessageInfo

func (m *ImageElement) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *ImageElement) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func (m *ImageElement) GetAlt() string {
	if m != nil {
		return m.Alt
	}
	return ""
}

func (m *ImageElement) GetWidth() int32 {
	if m != nil {
		return m.Width
	}
	return 0
}

func (m *ImageElement) GetHeight() int32 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *ImageElement) GetSources() []ImageElement_ImageSource {
	if m != nil {
		return m.Sources
	}
	return nil
}

func (m *ImageElement) GetImageGroups() []*ImageGroupInfo {
	if m != nil {
		return m.ImageGroups
	}
	return nil
}

type CompositeDoc struct {
	Url                  string          `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Html                 string          `protobuf:"bytes,2,opt,name=html,proto3" json:"html,omitempty"`
	Images               []*ImageElement `protobuf:"bytes,3,rep,name=images,proto3" json:"images,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *CompositeDoc) Reset()         { *m = CompositeDoc{} }
func (m *CompositeDoc) String() string { return proto.CompactTextString(m) }
func (*CompositeDoc) ProtoMessage()    {}
func (*CompositeDoc) Descriptor() ([]byte, []int) {
	return fileDescriptor_079b79b92abe9697, []int{3}
}

func (m *CompositeDoc) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CompositeDoc.Unmarshal(m, b)
}
func (m *CompositeDoc) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CompositeDoc.Marshal(b, m, deterministic)
}
func (m *CompositeDoc) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CompositeDoc.Merge(m, src)
}
func (m *CompositeDoc) XXX_Size() int {
	return xxx_messageInfo_CompositeDoc.Size(m)
}
func (m *CompositeDoc) XXX_DiscardUnknown() {
	xxx_messageInfo_CompositeDoc.DiscardUnknown(m)
}

var xxx_messageInfo_CompositeDoc proto.InternalMessageInfo

func (m *CompositeDoc) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *CompositeDoc) GetHtml() string {
	if m != nil {
		return m.Html
	}
	return ""
}

func (m *CompositeDoc) GetImages() []*ImageElement {
	if m != nil {
		return m.Images
	}
	return nil
}

type ImageCrawlInfo struct {
	WebUrl               string        `protobuf:"bytes,1,opt,name=web_url,json=webUrl,proto3" json:"web_url,omitempty"`
	ImgEle               *ImageElement `protobuf:"bytes,2,opt,name=img_ele,json=imgEle,proto3" json:"img_ele,omitempty"`
	CachedFile           string        `protobuf:"bytes,3,opt,name=cached_file,json=cachedFile,proto3" json:"cached_file,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *ImageCrawlInfo) Reset()         { *m = ImageCrawlInfo{} }
func (m *ImageCrawlInfo) String() string { return proto.CompactTextString(m) }
func (*ImageCrawlInfo) ProtoMessage()    {}
func (*ImageCrawlInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_079b79b92abe9697, []int{4}
}

func (m *ImageCrawlInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ImageCrawlInfo.Unmarshal(m, b)
}
func (m *ImageCrawlInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ImageCrawlInfo.Marshal(b, m, deterministic)
}
func (m *ImageCrawlInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ImageCrawlInfo.Merge(m, src)
}
func (m *ImageCrawlInfo) XXX_Size() int {
	return xxx_messageInfo_ImageCrawlInfo.Size(m)
}
func (m *ImageCrawlInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ImageCrawlInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ImageCrawlInfo proto.InternalMessageInfo

func (m *ImageCrawlInfo) GetWebUrl() string {
	if m != nil {
		return m.WebUrl
	}
	return ""
}

func (m *ImageCrawlInfo) GetImgEle() *ImageElement {
	if m != nil {
		return m.ImgEle
	}
	return nil
}

func (m *ImageCrawlInfo) GetCachedFile() string {
	if m != nil {
		return m.CachedFile
	}
	return ""
}

func init() {
	proto.RegisterEnum("extractor.ImageElement_ImageSource", ImageElement_ImageSource_name, ImageElement_ImageSource_value)
	proto.RegisterType((*ImageSrcEle)(nil), "extractor.ImageSrcEle")
	proto.RegisterType((*ImageGroupInfo)(nil), "extractor.ImageGroupInfo")
	proto.RegisterType((*ImageElement)(nil), "extractor.ImageElement")
	proto.RegisterType((*CompositeDoc)(nil), "extractor.CompositeDoc")
	proto.RegisterType((*ImageCrawlInfo)(nil), "extractor.ImageCrawlInfo")
}

func init() {
	proto.RegisterFile("github.com/motionrobot/webdoc/proto/compositedoc.proto", fileDescriptor_079b79b92abe9697)
}

var fileDescriptor_079b79b92abe9697 = []byte{
	// 519 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0x25, 0x71, 0xe2, 0x90, 0x71, 0x08, 0xd6, 0xaa, 0x6a, 0x8d, 0x38, 0x10, 0xb9, 0x42, 0xca,
	0x29, 0x41, 0x45, 0xe2, 0x02, 0x1c, 0x4a, 0x12, 0x22, 0x03, 0x4d, 0x2a, 0xc7, 0x51, 0x25, 0x2e,
	0x96, 0xbd, 0x9e, 0xda, 0x2b, 0xd9, 0xd9, 0xc8, 0xde, 0xc8, 0x05, 0x71, 0xe1, 0xa3, 0xb9, 0xa3,
	0xdd, 0xd8, 0x69, 0x55, 0x15, 0xd4, 0x93, 0x67, 0xde, 0xce, 0x7b, 0xfb, 0x66, 0x76, 0x0c, 0xef,
	0x62, 0x26, 0x92, 0x5d, 0x38, 0xa2, 0x3c, 0x1b, 0x67, 0x5c, 0x30, 0xbe, 0xc9, 0x79, 0xc8, 0xc5,
	0xb8, 0xc4, 0x30, 0xe2, 0x74, 0xbc, 0xcd, 0xb9, 0xe0, 0x63, 0xca, 0xb3, 0x2d, 0x2f, 0x98, 0xc0,
	0x88, 0xd3, 0x91, 0x82, 0x48, 0x17, 0x6f, 0x44, 0x1e, 0x50, 0xc1, 0x73, 0xdb, 0x05, 0xc3, 0xc9,
	0x82, 0x18, 0x57, 0x39, 0x9d, 0xa5, 0x48, 0x4c, 0xd0, 0x76, 0x79, 0x6a, 0x35, 0x06, 0x8d, 0x61,
	0xd7, 0x95, 0x21, 0x79, 0x09, 0xdd, 0x82, 0xfd, 0x44, 0x3f, 0xc2, 0x82, 0x5a, 0x4d, 0x85, 0x3f,
	0x95, 0xc0, 0x14, 0x0b, 0x4a, 0x8e, 0xa0, 0x9d, 0x61, 0xc4, 0x02, 0x4b, 0x53, 0x07, 0xfb, 0xc4,
	0x2e, 0xa1, 0xaf, 0x34, 0xe7, 0x39, 0xdf, 0x6d, 0x9d, 0xcd, 0x35, 0x27, 0x04, 0x5a, 0xe2, 0xc7,
	0x16, 0x2b, 0x5d, 0x15, 0xdf, 0x72, 0x9b, 0x77, 0xb8, 0xe4, 0x3d, 0x3c, 0x63, 0x92, 0xeb, 0x17,
	0x7c, 0x97, 0x53, 0x2c, 0x2c, 0x6d, 0xa0, 0x0d, 0x8d, 0xb3, 0xe3, 0xd1, 0xc1, 0xf2, 0xe8, 0x8e,
	0x5f, 0xb7, 0xa7, 0x8a, 0x57, 0xfb, 0x5a, 0xfb, 0x4f, 0x13, 0x7a, 0xea, 0x74, 0x96, 0x62, 0x86,
	0x1b, 0xf1, 0x40, 0x3b, 0xd2, 0x09, 0xde, 0x88, 0xea, 0x52, 0x15, 0xcb, 0xaa, 0x20, 0x15, 0x55,
	0x0f, 0x32, 0x94, 0xde, 0x4a, 0x16, 0x89, 0xc4, 0x6a, 0x0d, 0x1a, 0xc3, 0xb6, 0xbb, 0x4f, 0xc8,
	0x31, 0xe8, 0x09, 0xb2, 0x38, 0x11, 0x56, 0x5b, 0xc1, 0x55, 0x46, 0x3e, 0x42, 0xa7, 0x76, 0xab,
	0x0f, 0xb4, 0x61, 0xff, 0xec, 0xf4, 0xbe, 0xdb, 0xca, 0x4f, 0x65, 0x5d, 0xd5, 0xba, 0x35, 0x87,
	0x7c, 0x80, 0x7d, 0x17, 0x7e, 0x2c, 0xe7, 0x55, 0x58, 0xa0, 0x3a, 0x7e, 0x71, 0x5f, 0xe3, 0x30,
	0x4d, 0xd7, 0x60, 0x87, 0xbc, 0xb0, 0x7f, 0x37, 0xea, 0x17, 0x54, 0x72, 0xc4, 0x80, 0xce, 0x7a,
	0xf1, 0x75, 0xb1, 0xbc, 0x5a, 0x98, 0x4f, 0x64, 0xe2, 0x5c, 0xcc, 0x7d, 0xef, 0x7c, 0x6e, 0x36,
	0xc8, 0x73, 0x30, 0x2e, 0x9d, 0x89, 0xb7, 0x76, 0x67, 0x0a, 0x68, 0x12, 0x02, 0xfd, 0xd5, 0xc4,
	0x75, 0x2e, 0x3d, 0xff, 0xdb, 0xd4, 0xff, 0xb2, 0x5a, 0x2e, 0x4c, 0x4d, 0x32, 0x2e, 0x66, 0xde,
	0xb9, 0xbf, 0x9c, 0x9b, 0x2d, 0x62, 0x42, 0x4f, 0x25, 0xde, 0x95, 0xe3, 0x79, 0x33, 0xd7, 0x6c,
	0x93, 0x23, 0x30, 0x17, 0xcb, 0x8a, 0x54, 0x2b, 0xeb, 0x36, 0x42, 0x6f, 0x52, 0x6f, 0xd9, 0x94,
	0xd3, 0x87, 0xc7, 0x9e, 0x88, 0x2c, 0xad, 0xc7, 0x2e, 0x63, 0x32, 0x06, 0x5d, 0x35, 0x52, 0xbf,
	0xf1, 0xc9, 0x3f, 0xa6, 0xe6, 0x56, 0x65, 0xf6, 0xaf, 0x6a, 0xaf, 0x26, 0x79, 0x50, 0xa6, 0x6a,
	0xaf, 0x4e, 0xa0, 0x53, 0x62, 0xe8, 0xdf, 0x5e, 0xa6, 0x97, 0x18, 0xae, 0xf3, 0x94, 0xbc, 0x81,
	0x0e, 0xcb, 0x62, 0x1f, 0x53, 0x54, 0x57, 0xfe, 0x5f, 0x3c, 0x96, 0x9b, 0xff, 0x0a, 0x0c, 0x1a,
	0xd0, 0x04, 0x23, 0xff, 0x9a, 0xa5, 0x58, 0x2d, 0x03, 0xec, 0xa1, 0xcf, 0x2c, 0xc5, 0x4f, 0xaf,
	0xbf, 0x9f, 0x3e, 0xe2, 0x77, 0x0b, 0x75, 0xf5, 0x79, 0xfb, 0x37, 0x00, 0x00, 0xff, 0xff, 0x0a,
	0xb4, 0xdc, 0x62, 0x9c, 0x03, 0x00, 0x00,
}