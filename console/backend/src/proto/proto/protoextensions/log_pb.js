// source: protoextensions/log.proto
/**
 * @fileoverview
 * @enhanceable
 * @suppress {missingRequire} reports error on implicit type usages.
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!
/* eslint-disable */
// @ts-nocheck

var jspb = require('google-protobuf');
var goog = jspb;
var global = (function() {
  if (this) { return this; }
  if (typeof window !== 'undefined') { return window; }
  if (typeof global !== 'undefined') { return global; }
  if (typeof self !== 'undefined') { return self; }
  return Function('return this')();
}.call(null));

var google_protobuf_descriptor_pb = require('google-protobuf/google/protobuf/descriptor_pb.js');
goog.object.extend(proto, google_protobuf_descriptor_pb);
goog.exportSymbol('proto.log.half', null, global);
goog.exportSymbol('proto.log.mask', null, global);

/**
 * A tuple of {field number, class constructor} for the extension
 * field named `mask`.
 * @type {!jspb.ExtensionFieldInfo<boolean>}
 */
proto.log.mask = new jspb.ExtensionFieldInfo(
    50001,
    {mask: 0},
    null,
     /** @type {?function((boolean|undefined),!jspb.Message=): !Object} */ (
         null),
    0);

google_protobuf_descriptor_pb.FieldOptions.extensionsBinary[50001] = new jspb.ExtensionFieldBinaryInfo(
    proto.log.mask,
    jspb.BinaryReader.prototype.readBool,
    jspb.BinaryWriter.prototype.writeBool,
    undefined,
    undefined,
    false);
// This registers the extension field with the extended class, so that
// toObject() will function correctly.
google_protobuf_descriptor_pb.FieldOptions.extensions[50001] = proto.log.mask;


/**
 * A tuple of {field number, class constructor} for the extension
 * field named `half`.
 * @type {!jspb.ExtensionFieldInfo<boolean>}
 */
proto.log.half = new jspb.ExtensionFieldInfo(
    50002,
    {half: 0},
    null,
     /** @type {?function((boolean|undefined),!jspb.Message=): !Object} */ (
         null),
    0);

google_protobuf_descriptor_pb.FieldOptions.extensionsBinary[50002] = new jspb.ExtensionFieldBinaryInfo(
    proto.log.half,
    jspb.BinaryReader.prototype.readBool,
    jspb.BinaryWriter.prototype.writeBool,
    undefined,
    undefined,
    false);
// This registers the extension field with the extended class, so that
// toObject() will function correctly.
google_protobuf_descriptor_pb.FieldOptions.extensions[50002] = proto.log.half;

goog.object.extend(exports, proto.log);
