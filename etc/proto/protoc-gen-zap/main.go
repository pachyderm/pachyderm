// Command protoc-gen-zap generates MarshalLogObject methods for protocol buffer messages; allowing
// them to be printed with zap.Object().
//
// This code is copied from github.com/kei2100/protoc-gen-marshal-zap@0bd9d0de7073722759981352ac71da50273319d2.
//
// It includes this copyright message:
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Our modifications are to marshal BytesValue by marshaling them as a prefix and length
// instead of the full value.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/etc/proto/protoc-gen-zap/protoextensions"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	zapcorePkg   = protogen.GoImportPath("go.uber.org/zap/zapcore")
	extensionPkg = protogen.GoImportPath("github.com/pachyderm/pachyderm/v2/src/protoextensions")
	fmtPkg       = protogen.GoImportPath("fmt")
)

func generateListField(g *protogen.GeneratedFile, f *protogen.Field) {
	fname := f.Desc.Name()
	g.P(fname, "ArrMarshaller := func(enc ", g.QualifiedGoIdent(zapcorePkg.Ident("ArrayEncoder")), ") error {")
	g.P("for _, v := range x.", f.GoName, " {")
	switch f.Desc.Kind() {
	case protoreflect.BoolKind:
		g.P("enc.AppendBool(v)")
	case protoreflect.BytesKind:
		g.P("enc.AppendByteString(v)")
	case protoreflect.DoubleKind:
		g.P("enc.AppendFloat64(v)")
	case protoreflect.EnumKind:
		g.P("enc.AppendString(v.String())")
	case protoreflect.Fixed32Kind, protoreflect.Uint32Kind:
		g.P("enc.AppendUint32(v)")
	case protoreflect.Fixed64Kind, protoreflect.Uint64Kind:
		g.P("enc.AppendUint64(v)")
	case protoreflect.FloatKind:
		g.P("enc.AppendFloat32(v)")
	case protoreflect.Int32Kind, protoreflect.Sfixed32Kind, protoreflect.Sint32Kind:
		g.P("enc.AppendInt32(v)")
	case protoreflect.Int64Kind, protoreflect.Sfixed64Kind, protoreflect.Sint64Kind:
		g.P("enc.AppendInt64(v)")
	case protoreflect.GroupKind:
		g.P("enc.AppendReflected(v)")
	case protoreflect.MessageKind:
		// We don't have any protos that have 'repeated google.protobuf.*' fields, so no
		// special support exists to handle them here.
		if isPachydermProto(string(f.Desc.Message().FullName())) {
			g.P("enc.AppendObject(v)")
		} else {
			fmt.Fprintf(os.Stderr, "** OBJECT FALLBACK ON %v\n", f.Desc.Message().FullName())
			g.P("if obj, ok := interface{}(v).(", g.QualifiedGoIdent(zapcorePkg.Ident("ObjectMarshaler")), "); ok {")
			g.P("enc.AppendObject(obj)")
			g.P("} else {")
			g.P("enc.AppendReflected(v)")
			g.P("}")
		}
	case protoreflect.StringKind:
		g.P("enc.AppendString(v)")
	default:
		g.P("enc.AppendReflected(v)")
	}
	g.P("}")
	g.P("return nil")
	g.P("}")
	g.P("enc.AddArray(\"", fname, "\",", g.QualifiedGoIdent(zapcorePkg.Ident("ArrayMarshalerFunc")), "(", fname, "ArrMarshaller))")
}

func generateMapField(g *protogen.GeneratedFile, f *protogen.Field) {
	fname := f.Desc.Name()
	g.P("enc.AddObject(\"", fname, "\", ", g.QualifiedGoIdent(zapcorePkg.Ident("ObjectMarshalerFunc")), "(func(enc ", g.QualifiedGoIdent(zapcorePkg.Ident("ObjectEncoder")), ") error {")
	g.P("for k, v := range x.", f.GoName, " {")
	switch f.Desc.MapValue().Kind() {
	case protoreflect.BoolKind:
		g.P("enc.AddBool(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	case protoreflect.BytesKind:
		g.P("enc.AddBinary(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	case protoreflect.DoubleKind:
		g.P("enc.AddFloat64(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	case protoreflect.EnumKind:
		g.P("enc.AddString(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v.String())")
	case protoreflect.Fixed32Kind, protoreflect.Uint32Kind:
		g.P("enc.AddUint32(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	case protoreflect.Fixed64Kind, protoreflect.Uint64Kind:
		g.P("enc.AddUint64(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	case protoreflect.FloatKind:
		g.P("enc.AddFloat32(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	case protoreflect.Int32Kind, protoreflect.Sfixed32Kind, protoreflect.Sint32Kind:
		g.P("enc.AddInt32(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	case protoreflect.Int64Kind, protoreflect.Sfixed64Kind, protoreflect.Sint64Kind:
		g.P("enc.AddInt64(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	case protoreflect.GroupKind:
		g.P("enc.AddReflected(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	case protoreflect.MessageKind:
		// We don't have any protos that use google.protobuf.* map keys or values, so no
		// special support for those is here.
		if isPachydermProto(string(f.Desc.Message().FullName())) {
			g.P("enc.AddObject(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
		} else {
			fmt.Fprintf(os.Stderr, "** OBJECT FALLBACK ON %v\n", f.Desc.Message().FullName())
			g.P("if obj, ok := interface{}(v).(", g.QualifiedGoIdent(zapcorePkg.Ident("ObjectMarshaler")), "); ok {")
			g.P("enc.AddObject(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), obj)")
			g.P("} else {")
			g.P("enc.AddReflected(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
			g.P("}")
		}
	case protoreflect.StringKind:
		g.P("enc.AddString(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	default:
		g.P("enc.AddReflected(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	}
	g.P("}")
	g.P("return nil")
	g.P("}))")
}

func generatePrimitiveField(g *protogen.GeneratedFile, f *protogen.Field, opts *descriptorpb.FieldOptions) {
	fname := f.Desc.Name()
	var gname string
	if f.Oneof != nil {
		gname = fmt.Sprintf("Get%s()", f.GoName)
	} else {
		gname = f.GoName
	}

	half := proto.GetExtension(opts, protoextensions.E_Half).(bool)
	if half && f.Desc.Kind() != protoreflect.StringKind {
		fmt.Fprintf(os.Stderr, "field %v: can only apply log.half to fields of type string (not %v)", f.Desc.Name(), f.Desc.Kind().String())
		os.Exit(1)
	}

	switch f.Desc.Kind() {
	case protoreflect.BoolKind:
		g.P("enc.AddBool(\"", fname, "\", x.", gname, ")")
	case protoreflect.BytesKind:
		g.P(g.QualifiedGoIdent(extensionPkg.Ident("AddBytes")), `(enc, "`, fname, `", x.`, gname, `)`)
	case protoreflect.DoubleKind:
		g.P("enc.AddFloat64(\"", fname, "\", x.", gname, ")")
	case protoreflect.EnumKind:
		g.P("enc.AddString(\"", fname, "\", x.", gname, ".String())")
	case protoreflect.Fixed32Kind, protoreflect.Uint32Kind:
		g.P("enc.AddUint32(\"", fname, "\", x.", gname, ")")
	case protoreflect.Fixed64Kind, protoreflect.Uint64Kind:
		g.P("enc.AddUint64(\"", fname, "\", x.", gname, ")")
	case protoreflect.FloatKind:
		g.P("enc.AddFloat32(\"", fname, "\", x.", gname, ")")
	case protoreflect.Int32Kind, protoreflect.Sfixed32Kind, protoreflect.Sint32Kind:
		g.P("enc.AddInt32(\"", fname, "\", x.", gname, ")")
	case protoreflect.Int64Kind, protoreflect.Sfixed64Kind, protoreflect.Sint64Kind:
		g.P("enc.AddInt64(\"", fname, "\", x.", gname, ")")
	case protoreflect.GroupKind:
		g.P("enc.AddReflected(\"", fname, "\", x.", gname, ")")
	case protoreflect.MessageKind:
		switch f.Desc.Message().FullName() {
		case "google.protobuf.Empty":
			// do nothing
		case "google.protobuf.Timestamp":
			g.P(g.QualifiedGoIdent(extensionPkg.Ident("AddTimestamp")), `(enc, "`, fname, `", x.`, gname, `)`)
		case "google.protobuf.Duration":
			g.P(g.QualifiedGoIdent(extensionPkg.Ident("AddDuration")), `(enc, "`, fname, `", x.`, gname, `)`)
		case "google.protobuf.BytesValue":
			g.P(g.QualifiedGoIdent(extensionPkg.Ident("AddBytesValue")), `(enc, "`, fname, `", x.`, gname, `)`)
		case "google.protobuf.Any":
			g.P(g.QualifiedGoIdent(extensionPkg.Ident("AddAny")), `(enc, "`, fname, `", x.`, gname, `)`)
		case "google.protobuf.Int64Value":
			g.P(g.QualifiedGoIdent(extensionPkg.Ident("AddInt64Value")), `(enc, "`, fname, `", x.`, gname, `)`)
		default:
			// Avoid reflection for our own types, which we know will have MarshalLogObject.
			if isPachydermProto(string(f.Desc.Message().FullName())) {
				g.P("enc.AddObject(\"", fname, "\", x.", gname, ")")
			} else {
				fmt.Fprintf(os.Stderr, "** OBJECT FALLBACK ON %v\n", f.Desc.Message().FullName())

				g.P("if obj, ok := interface{}(x.", gname, ").(", g.QualifiedGoIdent(zapcorePkg.Ident("ObjectMarshaler")), "); ok {")
				g.P("enc.AddObject(\"", fname, "\", obj)")
				g.P("} else {")
				g.P("enc.AddReflected(\"", fname, "\", x.", gname, ")")
				g.P("}")
			}
		}
	case protoreflect.StringKind:
		if half {
			g.P(g.QualifiedGoIdent(extensionPkg.Ident("AddHalfString")), `(enc, "`, fname, `", x.`, gname, `)`)
		} else {
			g.P("enc.AddString(\"", fname, "\", x.", gname, ")")
		}
	default:
		g.P("enc.AddReflected(\"", fname, "\", x.", gname, ")")
	}
}

func isPachydermProto(fullName string) bool {
	if strings.Contains(fullName, "_v2.") {
		return true
	}
	return strings.HasPrefix(fullName, "datum.") ||
		strings.HasPrefix(fullName, "pfsload.") ||
		strings.HasPrefix(fullName, "pfsserver.") ||
		strings.HasPrefix(fullName, "taskapi.")
}

func generateMessage(g *protogen.GeneratedFile, m *protogen.Message) {
	ident := g.QualifiedGoIdent(m.GoIdent)
	g.P("func (x *", ident, ") MarshalLogObject(enc ", g.QualifiedGoIdent(zapcorePkg.Ident("ObjectEncoder")), ") error {")
	g.P("if x == nil {")
	g.P("return nil")
	g.P("}")
	for _, f := range m.Fields {
		opts := f.Desc.Options().(*descriptorpb.FieldOptions)
		if proto.GetExtension(opts, protoextensions.E_Mask).(bool) {
			g.P("enc.AddString(\"", f.Desc.Name(), "\", \"[MASKED]\")")
		} else if f.Desc.IsList() {
			generateListField(g, f)
		} else if f.Desc.IsMap() {
			generateMapField(g, f)
		} else {
			generatePrimitiveField(g, f, opts)
		}
	}
	g.P("return nil")
	g.P("}")
	g.P()
	for _, submsg := range m.Messages {
		if submsg.Desc.IsMapEntry() {
			continue
		}
		generateMessage(g, submsg)
	}
}

func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Messages) == 0 {
		return nil
	}

	filename := fmt.Sprintf("%s.pb.zap.go", file.GeneratedFilenamePrefix)
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-zap (etc/proto/protoc-gen-zap). DO NOT EDIT.")
	g.P("//")
	g.P("// source: ", file.Desc.Path())
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()

	for _, m := range file.Messages {
		generateMessage(g, m)
	}

	return g
}

func main() {
	protogen.Options{}.Run(func(plugin *protogen.Plugin) error {
		// Optional fields are represented internally as oneof fields;
		// generatePrimitiveField handles these appropriately.
		plugin.SupportedFeatures |= uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, file := range plugin.FilesByPath {
			if !file.Generate {
				continue
			}

			generateFile(plugin, file)
		}
		return nil
	})
}
