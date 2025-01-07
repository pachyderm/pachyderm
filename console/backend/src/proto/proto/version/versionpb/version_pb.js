// source: version/versionpb/version.proto
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

var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
goog.object.extend(proto, google_protobuf_empty_pb);
goog.exportSymbol('proto.versionpb_v2.Version', null, global);
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.versionpb_v2.Version = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.versionpb_v2.Version, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.versionpb_v2.Version.displayName = 'proto.versionpb_v2.Version';
}



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.versionpb_v2.Version.prototype.toObject = function(opt_includeInstance) {
  return proto.versionpb_v2.Version.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.versionpb_v2.Version} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.versionpb_v2.Version.toObject = function(includeInstance, msg) {
  var f, obj = {
    major: jspb.Message.getFieldWithDefault(msg, 1, 0),
    minor: jspb.Message.getFieldWithDefault(msg, 2, 0),
    micro: jspb.Message.getFieldWithDefault(msg, 3, 0),
    additional: jspb.Message.getFieldWithDefault(msg, 4, ""),
    gitCommit: jspb.Message.getFieldWithDefault(msg, 5, ""),
    gitTreeModified: jspb.Message.getFieldWithDefault(msg, 6, ""),
    buildDate: jspb.Message.getFieldWithDefault(msg, 7, ""),
    goVersion: jspb.Message.getFieldWithDefault(msg, 8, ""),
    platform: jspb.Message.getFieldWithDefault(msg, 9, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.versionpb_v2.Version}
 */
proto.versionpb_v2.Version.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.versionpb_v2.Version;
  return proto.versionpb_v2.Version.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.versionpb_v2.Version} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.versionpb_v2.Version}
 */
proto.versionpb_v2.Version.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setMajor(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setMinor(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setMicro(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setAdditional(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setGitCommit(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setGitTreeModified(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setBuildDate(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setGoVersion(value);
      break;
    case 9:
      var value = /** @type {string} */ (reader.readString());
      msg.setPlatform(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.versionpb_v2.Version.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.versionpb_v2.Version.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.versionpb_v2.Version} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.versionpb_v2.Version.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getMajor();
  if (f !== 0) {
    writer.writeUint32(
      1,
      f
    );
  }
  f = message.getMinor();
  if (f !== 0) {
    writer.writeUint32(
      2,
      f
    );
  }
  f = message.getMicro();
  if (f !== 0) {
    writer.writeUint32(
      3,
      f
    );
  }
  f = message.getAdditional();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getGitCommit();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getGitTreeModified();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getBuildDate();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
  f = message.getGoVersion();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
  f = message.getPlatform();
  if (f.length > 0) {
    writer.writeString(
      9,
      f
    );
  }
};


/**
 * optional uint32 major = 1;
 * @return {number}
 */
proto.versionpb_v2.Version.prototype.getMajor = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.versionpb_v2.Version} returns this
 */
proto.versionpb_v2.Version.prototype.setMajor = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional uint32 minor = 2;
 * @return {number}
 */
proto.versionpb_v2.Version.prototype.getMinor = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.versionpb_v2.Version} returns this
 */
proto.versionpb_v2.Version.prototype.setMinor = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional uint32 micro = 3;
 * @return {number}
 */
proto.versionpb_v2.Version.prototype.getMicro = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.versionpb_v2.Version} returns this
 */
proto.versionpb_v2.Version.prototype.setMicro = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional string additional = 4;
 * @return {string}
 */
proto.versionpb_v2.Version.prototype.getAdditional = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.versionpb_v2.Version} returns this
 */
proto.versionpb_v2.Version.prototype.setAdditional = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string git_commit = 5;
 * @return {string}
 */
proto.versionpb_v2.Version.prototype.getGitCommit = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.versionpb_v2.Version} returns this
 */
proto.versionpb_v2.Version.prototype.setGitCommit = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string git_tree_modified = 6;
 * @return {string}
 */
proto.versionpb_v2.Version.prototype.getGitTreeModified = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.versionpb_v2.Version} returns this
 */
proto.versionpb_v2.Version.prototype.setGitTreeModified = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional string build_date = 7;
 * @return {string}
 */
proto.versionpb_v2.Version.prototype.getBuildDate = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.versionpb_v2.Version} returns this
 */
proto.versionpb_v2.Version.prototype.setBuildDate = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};


/**
 * optional string go_version = 8;
 * @return {string}
 */
proto.versionpb_v2.Version.prototype.getGoVersion = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.versionpb_v2.Version} returns this
 */
proto.versionpb_v2.Version.prototype.setGoVersion = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};


/**
 * optional string platform = 9;
 * @return {string}
 */
proto.versionpb_v2.Version.prototype.getPlatform = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 9, ""));
};


/**
 * @param {string} value
 * @return {!proto.versionpb_v2.Version} returns this
 */
proto.versionpb_v2.Version.prototype.setPlatform = function(value) {
  return jspb.Message.setProto3StringField(this, 9, value);
};


goog.object.extend(exports, proto.versionpb_v2);