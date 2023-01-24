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
// Our modifications are to marshal types.BytesValue by marshaling them as a prefix and length
// instead of the full value, and to support gogo.customname.
package main

import (
	"fmt"

	"github.com/pachyderm/pachyderm/etc/proto/protoc-gen-zap/gogoproto"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	zapcorePkg = protogen.GoImportPath("go.uber.org/zap/zapcore")
	fmtPkg     = protogen.GoImportPath("fmt")
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
		g.P("if obj, ok := interface{}(v).(", g.QualifiedGoIdent(zapcorePkg.Ident("ObjectMarshaler")), "); ok {")
		g.P("enc.AppendObject(obj)")
		g.P("} else {")
		g.P("enc.AppendReflected(v)")
		g.P("}")
	case protoreflect.StringKind:
		g.P("enc.AppendString(v)")
	default:
		g.P("enc.AppendReflected(v)")
	}
	g.P("}")
	g.P("return nil")
	g.P("}")
	g.P("enc.AddArray(\"", fname, "\",", g.QualifiedGoIdent(zapcorePkg.Ident("ArrayMarshalerFunc")), "(", fname, "ArrMarshaller))")
	g.P()
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
		g.P("if obj, ok := interface{}(v).(", g.QualifiedGoIdent(zapcorePkg.Ident("ObjectMarshaler")), "); ok {")
		g.P("enc.AddObject(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), obj)")
		g.P("} else {")
		g.P("enc.AddReflected(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
		g.P("}")
	case protoreflect.StringKind:
		g.P("enc.AddString(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	default:
		g.P("enc.AddReflected(", g.QualifiedGoIdent(fmtPkg.Ident("Sprintf")), "(\"%v\", k), v)")
	}
	g.P("}")
	g.P("return nil")
	g.P("}))")
	g.P()
}

func generatePrimitiveField(g *protogen.GeneratedFile, f *protogen.Field) {
	fname := f.Desc.Name()
	var gname string
	if f.Oneof != nil {
		gname = fmt.Sprintf("Get%s()", f.GoName)
	} else {
		gname = f.GoName
	}
	switch f.Desc.Kind() {
	case protoreflect.BoolKind:
		g.P("enc.AddBool(\"", fname, "\", x.", gname, ")")
	case protoreflect.BytesKind:
		g.P("enc.AddBinary(\"", fname, "\", x.", gname, ")")
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
		g.P("if obj, ok := interface{}(x.", gname, ").(", g.QualifiedGoIdent(zapcorePkg.Ident("ObjectMarshaler")), "); ok {")
		g.P("enc.AddObject(\"", fname, "\", obj)")
		g.P("} else {")
		g.P("enc.AddReflected(\"", fname, "\", x.", gname, ")")
		g.P("}")
	case protoreflect.StringKind:
		g.P("enc.AddString(\"", fname, "\", x.", gname, ")")
	default:
		g.P("enc.AddReflected(\"", fname, "\", x.", gname, ")")
	}
	g.P()
}

func generateMessage(g *protogen.GeneratedFile, m *protogen.Message) {
	ident := g.QualifiedGoIdent(m.GoIdent)
	g.P("func (x *", ident, ") MarshalLogObject(enc ", g.QualifiedGoIdent(zapcorePkg.Ident("ObjectEncoder")), ") error {")
	g.P("if x == nil {")
	g.P("return nil")
	g.P("}")
	g.P()
	for _, f := range m.Fields {
		opts := f.Desc.Options().(*descriptorpb.FieldOptions)
		if opts != nil {
			if customName := proto.GetExtension(opts, gogoproto.E_Customname).(string); customName != "" {
				f.GoName = customName
			}
		}
		//m.GoIdent.GoName = gogoName(m.GoIdent.GoName, opts)

		// if isMasked(opts) {
		// 	panic("this works")
		// 	g.P("enc.AddString(\"", f.Desc.Name(), "\", \"[MASKED]\")")
		// 	g.P()
		//} else
		if f.Desc.IsList() {
			generateListField(g, f)
		} else if f.Desc.IsMap() {
			generateMapField(g, f)
		} else {
			generatePrimitiveField(g, f)
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
		for _, file := range plugin.FilesByPath {
			if !file.Generate {
				continue
			}

			generateFile(plugin, file)
		}
		return nil
	})
}
