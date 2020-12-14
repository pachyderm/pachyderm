// source: client/pfs/pfs.proto
/**
 * @fileoverview
 * @enhanceable
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!
/* eslint-disable */
// @ts-nocheck

var jspb = require('google-protobuf');
var goog = jspb;
var global = Function('return this')();

var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
goog.object.extend(proto, google_protobuf_empty_pb);
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
goog.object.extend(proto, google_protobuf_timestamp_pb);
var google_protobuf_wrappers_pb = require('google-protobuf/google/protobuf/wrappers_pb.js');
goog.object.extend(proto, google_protobuf_wrappers_pb);
var gogoproto_gogo_pb = require('../../gogoproto/gogo_pb.js');
goog.object.extend(proto, gogoproto_gogo_pb);
var client_auth_auth_pb = require('../../client/auth/auth_pb.js');
goog.object.extend(proto, client_auth_auth_pb);
goog.exportSymbol('proto.pfs.Block', null, global);
goog.exportSymbol('proto.pfs.BlockRef', null, global);
goog.exportSymbol('proto.pfs.Branch', null, global);
goog.exportSymbol('proto.pfs.BranchInfo', null, global);
goog.exportSymbol('proto.pfs.BranchInfos', null, global);
goog.exportSymbol('proto.pfs.BuildCommitRequest', null, global);
goog.exportSymbol('proto.pfs.ByteRange', null, global);
goog.exportSymbol('proto.pfs.CheckObjectRequest', null, global);
goog.exportSymbol('proto.pfs.CheckObjectResponse', null, global);
goog.exportSymbol('proto.pfs.ClearCommitRequestV2', null, global);
goog.exportSymbol('proto.pfs.Commit', null, global);
goog.exportSymbol('proto.pfs.CommitInfo', null, global);
goog.exportSymbol('proto.pfs.CommitInfos', null, global);
goog.exportSymbol('proto.pfs.CommitOrigin', null, global);
goog.exportSymbol('proto.pfs.CommitProvenance', null, global);
goog.exportSymbol('proto.pfs.CommitRange', null, global);
goog.exportSymbol('proto.pfs.CommitState', null, global);
goog.exportSymbol('proto.pfs.Compaction', null, global);
goog.exportSymbol('proto.pfs.CopyFileRequest', null, global);
goog.exportSymbol('proto.pfs.CreateBranchRequest', null, global);
goog.exportSymbol('proto.pfs.CreateObjectRequest', null, global);
goog.exportSymbol('proto.pfs.CreateRepoRequest', null, global);
goog.exportSymbol('proto.pfs.CreateTmpFileSetResponse', null, global);
goog.exportSymbol('proto.pfs.DeleteBranchRequest', null, global);
goog.exportSymbol('proto.pfs.DeleteCommitRequest', null, global);
goog.exportSymbol('proto.pfs.DeleteFileRequest', null, global);
goog.exportSymbol('proto.pfs.DeleteFilesRequestV2', null, global);
goog.exportSymbol('proto.pfs.DeleteObjDirectRequest', null, global);
goog.exportSymbol('proto.pfs.DeleteObjectsRequest', null, global);
goog.exportSymbol('proto.pfs.DeleteObjectsResponse', null, global);
goog.exportSymbol('proto.pfs.DeleteRepoRequest', null, global);
goog.exportSymbol('proto.pfs.DeleteTagsRequest', null, global);
goog.exportSymbol('proto.pfs.DeleteTagsResponse', null, global);
goog.exportSymbol('proto.pfs.Delimiter', null, global);
goog.exportSymbol('proto.pfs.DiffFileRequest', null, global);
goog.exportSymbol('proto.pfs.DiffFileResponse', null, global);
goog.exportSymbol('proto.pfs.DiffFileResponseV2', null, global);
goog.exportSymbol('proto.pfs.File', null, global);
goog.exportSymbol('proto.pfs.FileInfo', null, global);
goog.exportSymbol('proto.pfs.FileInfos', null, global);
goog.exportSymbol('proto.pfs.FileOperationRequestV2', null, global);
goog.exportSymbol('proto.pfs.FileOperationRequestV2.OperationCase', null, global);
goog.exportSymbol('proto.pfs.FileType', null, global);
goog.exportSymbol('proto.pfs.FinishCommitRequest', null, global);
goog.exportSymbol('proto.pfs.FlushCommitRequest', null, global);
goog.exportSymbol('proto.pfs.FsckRequest', null, global);
goog.exportSymbol('proto.pfs.FsckResponse', null, global);
goog.exportSymbol('proto.pfs.GetBlockRequest', null, global);
goog.exportSymbol('proto.pfs.GetBlocksRequest', null, global);
goog.exportSymbol('proto.pfs.GetFileRequest', null, global);
goog.exportSymbol('proto.pfs.GetObjDirectRequest', null, global);
goog.exportSymbol('proto.pfs.GetObjectsRequest', null, global);
goog.exportSymbol('proto.pfs.GetTarRequestV2', null, global);
goog.exportSymbol('proto.pfs.GlobFileRequest', null, global);
goog.exportSymbol('proto.pfs.InspectBranchRequest', null, global);
goog.exportSymbol('proto.pfs.InspectCommitRequest', null, global);
goog.exportSymbol('proto.pfs.InspectFileRequest', null, global);
goog.exportSymbol('proto.pfs.InspectRepoRequest', null, global);
goog.exportSymbol('proto.pfs.ListBlockRequest', null, global);
goog.exportSymbol('proto.pfs.ListBranchRequest', null, global);
goog.exportSymbol('proto.pfs.ListCommitRequest', null, global);
goog.exportSymbol('proto.pfs.ListFileRequest', null, global);
goog.exportSymbol('proto.pfs.ListObjectsRequest', null, global);
goog.exportSymbol('proto.pfs.ListRepoRequest', null, global);
goog.exportSymbol('proto.pfs.ListRepoResponse', null, global);
goog.exportSymbol('proto.pfs.ListTagsRequest', null, global);
goog.exportSymbol('proto.pfs.ListTagsResponse', null, global);
goog.exportSymbol('proto.pfs.Object', null, global);
goog.exportSymbol('proto.pfs.ObjectIndex', null, global);
goog.exportSymbol('proto.pfs.ObjectInfo', null, global);
goog.exportSymbol('proto.pfs.Objects', null, global);
goog.exportSymbol('proto.pfs.OriginKind', null, global);
goog.exportSymbol('proto.pfs.OverwriteIndex', null, global);
goog.exportSymbol('proto.pfs.PathRange', null, global);
goog.exportSymbol('proto.pfs.PutBlockRequest', null, global);
goog.exportSymbol('proto.pfs.PutFileRecord', null, global);
goog.exportSymbol('proto.pfs.PutFileRecords', null, global);
goog.exportSymbol('proto.pfs.PutFileRequest', null, global);
goog.exportSymbol('proto.pfs.PutObjDirectRequest', null, global);
goog.exportSymbol('proto.pfs.PutObjectRequest', null, global);
goog.exportSymbol('proto.pfs.PutTarRequestV2', null, global);
goog.exportSymbol('proto.pfs.RenewTmpFileSetRequest', null, global);
goog.exportSymbol('proto.pfs.Repo', null, global);
goog.exportSymbol('proto.pfs.RepoAuthInfo', null, global);
goog.exportSymbol('proto.pfs.RepoInfo', null, global);
goog.exportSymbol('proto.pfs.Shard', null, global);
goog.exportSymbol('proto.pfs.StartCommitRequest', null, global);
goog.exportSymbol('proto.pfs.SubscribeCommitRequest', null, global);
goog.exportSymbol('proto.pfs.Tag', null, global);
goog.exportSymbol('proto.pfs.TagObjectRequest', null, global);
goog.exportSymbol('proto.pfs.Trigger', null, global);
goog.exportSymbol('proto.pfs.WalkFileRequest', null, global);
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
proto.pfs.Repo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.Repo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.Repo.displayName = 'proto.pfs.Repo';
}
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
proto.pfs.Branch = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.Branch, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.Branch.displayName = 'proto.pfs.Branch';
}
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
proto.pfs.File = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.File, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.File.displayName = 'proto.pfs.File';
}
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
proto.pfs.Block = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.Block, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.Block.displayName = 'proto.pfs.Block';
}
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
proto.pfs.Object = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.Object, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.Object.displayName = 'proto.pfs.Object';
}
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
proto.pfs.Tag = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.Tag, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.Tag.displayName = 'proto.pfs.Tag';
}
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
proto.pfs.RepoInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.RepoInfo.repeatedFields_, null);
};
goog.inherits(proto.pfs.RepoInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.RepoInfo.displayName = 'proto.pfs.RepoInfo';
}
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
proto.pfs.RepoAuthInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.RepoAuthInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.RepoAuthInfo.displayName = 'proto.pfs.RepoAuthInfo';
}
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
proto.pfs.BranchInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.BranchInfo.repeatedFields_, null);
};
goog.inherits(proto.pfs.BranchInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.BranchInfo.displayName = 'proto.pfs.BranchInfo';
}
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
proto.pfs.BranchInfos = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.BranchInfos.repeatedFields_, null);
};
goog.inherits(proto.pfs.BranchInfos, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.BranchInfos.displayName = 'proto.pfs.BranchInfos';
}
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
proto.pfs.Trigger = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.Trigger, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.Trigger.displayName = 'proto.pfs.Trigger';
}
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
proto.pfs.CommitOrigin = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.CommitOrigin, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CommitOrigin.displayName = 'proto.pfs.CommitOrigin';
}
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
proto.pfs.Commit = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.Commit, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.Commit.displayName = 'proto.pfs.Commit';
}
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
proto.pfs.CommitRange = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.CommitRange, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CommitRange.displayName = 'proto.pfs.CommitRange';
}
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
proto.pfs.CommitProvenance = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.CommitProvenance, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CommitProvenance.displayName = 'proto.pfs.CommitProvenance';
}
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
proto.pfs.CommitInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.CommitInfo.repeatedFields_, null);
};
goog.inherits(proto.pfs.CommitInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CommitInfo.displayName = 'proto.pfs.CommitInfo';
}
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
proto.pfs.FileInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.FileInfo.repeatedFields_, null);
};
goog.inherits(proto.pfs.FileInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.FileInfo.displayName = 'proto.pfs.FileInfo';
}
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
proto.pfs.ByteRange = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ByteRange, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ByteRange.displayName = 'proto.pfs.ByteRange';
}
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
proto.pfs.BlockRef = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.BlockRef, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.BlockRef.displayName = 'proto.pfs.BlockRef';
}
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
proto.pfs.ObjectInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ObjectInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ObjectInfo.displayName = 'proto.pfs.ObjectInfo';
}
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
proto.pfs.Compaction = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.Compaction.repeatedFields_, null);
};
goog.inherits(proto.pfs.Compaction, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.Compaction.displayName = 'proto.pfs.Compaction';
}
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
proto.pfs.Shard = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.Shard, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.Shard.displayName = 'proto.pfs.Shard';
}
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
proto.pfs.PathRange = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.PathRange, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.PathRange.displayName = 'proto.pfs.PathRange';
}
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
proto.pfs.CreateRepoRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.CreateRepoRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CreateRepoRequest.displayName = 'proto.pfs.CreateRepoRequest';
}
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
proto.pfs.InspectRepoRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.InspectRepoRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.InspectRepoRequest.displayName = 'proto.pfs.InspectRepoRequest';
}
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
proto.pfs.ListRepoRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ListRepoRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ListRepoRequest.displayName = 'proto.pfs.ListRepoRequest';
}
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
proto.pfs.ListRepoResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.ListRepoResponse.repeatedFields_, null);
};
goog.inherits(proto.pfs.ListRepoResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ListRepoResponse.displayName = 'proto.pfs.ListRepoResponse';
}
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
proto.pfs.DeleteRepoRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.DeleteRepoRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DeleteRepoRequest.displayName = 'proto.pfs.DeleteRepoRequest';
}
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
proto.pfs.StartCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.StartCommitRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs.StartCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.StartCommitRequest.displayName = 'proto.pfs.StartCommitRequest';
}
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
proto.pfs.BuildCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.BuildCommitRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs.BuildCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.BuildCommitRequest.displayName = 'proto.pfs.BuildCommitRequest';
}
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
proto.pfs.FinishCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.FinishCommitRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs.FinishCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.FinishCommitRequest.displayName = 'proto.pfs.FinishCommitRequest';
}
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
proto.pfs.InspectCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.InspectCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.InspectCommitRequest.displayName = 'proto.pfs.InspectCommitRequest';
}
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
proto.pfs.ListCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ListCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ListCommitRequest.displayName = 'proto.pfs.ListCommitRequest';
}
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
proto.pfs.CommitInfos = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.CommitInfos.repeatedFields_, null);
};
goog.inherits(proto.pfs.CommitInfos, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CommitInfos.displayName = 'proto.pfs.CommitInfos';
}
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
proto.pfs.CreateBranchRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.CreateBranchRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs.CreateBranchRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CreateBranchRequest.displayName = 'proto.pfs.CreateBranchRequest';
}
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
proto.pfs.InspectBranchRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.InspectBranchRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.InspectBranchRequest.displayName = 'proto.pfs.InspectBranchRequest';
}
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
proto.pfs.ListBranchRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ListBranchRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ListBranchRequest.displayName = 'proto.pfs.ListBranchRequest';
}
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
proto.pfs.DeleteBranchRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.DeleteBranchRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DeleteBranchRequest.displayName = 'proto.pfs.DeleteBranchRequest';
}
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
proto.pfs.DeleteCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.DeleteCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DeleteCommitRequest.displayName = 'proto.pfs.DeleteCommitRequest';
}
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
proto.pfs.FlushCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.FlushCommitRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs.FlushCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.FlushCommitRequest.displayName = 'proto.pfs.FlushCommitRequest';
}
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
proto.pfs.SubscribeCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.SubscribeCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.SubscribeCommitRequest.displayName = 'proto.pfs.SubscribeCommitRequest';
}
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
proto.pfs.GetFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.GetFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.GetFileRequest.displayName = 'proto.pfs.GetFileRequest';
}
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
proto.pfs.OverwriteIndex = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.OverwriteIndex, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.OverwriteIndex.displayName = 'proto.pfs.OverwriteIndex';
}
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
proto.pfs.PutFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.PutFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.PutFileRequest.displayName = 'proto.pfs.PutFileRequest';
}
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
proto.pfs.PutFileRecord = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.PutFileRecord, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.PutFileRecord.displayName = 'proto.pfs.PutFileRecord';
}
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
proto.pfs.PutFileRecords = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.PutFileRecords.repeatedFields_, null);
};
goog.inherits(proto.pfs.PutFileRecords, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.PutFileRecords.displayName = 'proto.pfs.PutFileRecords';
}
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
proto.pfs.CopyFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.CopyFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CopyFileRequest.displayName = 'proto.pfs.CopyFileRequest';
}
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
proto.pfs.InspectFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.InspectFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.InspectFileRequest.displayName = 'proto.pfs.InspectFileRequest';
}
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
proto.pfs.ListFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ListFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ListFileRequest.displayName = 'proto.pfs.ListFileRequest';
}
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
proto.pfs.WalkFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.WalkFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.WalkFileRequest.displayName = 'proto.pfs.WalkFileRequest';
}
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
proto.pfs.GlobFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.GlobFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.GlobFileRequest.displayName = 'proto.pfs.GlobFileRequest';
}
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
proto.pfs.FileInfos = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.FileInfos.repeatedFields_, null);
};
goog.inherits(proto.pfs.FileInfos, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.FileInfos.displayName = 'proto.pfs.FileInfos';
}
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
proto.pfs.DiffFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.DiffFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DiffFileRequest.displayName = 'proto.pfs.DiffFileRequest';
}
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
proto.pfs.DiffFileResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.DiffFileResponse.repeatedFields_, null);
};
goog.inherits(proto.pfs.DiffFileResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DiffFileResponse.displayName = 'proto.pfs.DiffFileResponse';
}
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
proto.pfs.DeleteFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.DeleteFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DeleteFileRequest.displayName = 'proto.pfs.DeleteFileRequest';
}
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
proto.pfs.FsckRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.FsckRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.FsckRequest.displayName = 'proto.pfs.FsckRequest';
}
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
proto.pfs.FsckResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.FsckResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.FsckResponse.displayName = 'proto.pfs.FsckResponse';
}
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
proto.pfs.FileOperationRequestV2 = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, proto.pfs.FileOperationRequestV2.oneofGroups_);
};
goog.inherits(proto.pfs.FileOperationRequestV2, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.FileOperationRequestV2.displayName = 'proto.pfs.FileOperationRequestV2';
}
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
proto.pfs.PutTarRequestV2 = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.PutTarRequestV2, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.PutTarRequestV2.displayName = 'proto.pfs.PutTarRequestV2';
}
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
proto.pfs.DeleteFilesRequestV2 = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.DeleteFilesRequestV2.repeatedFields_, null);
};
goog.inherits(proto.pfs.DeleteFilesRequestV2, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DeleteFilesRequestV2.displayName = 'proto.pfs.DeleteFilesRequestV2';
}
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
proto.pfs.GetTarRequestV2 = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.GetTarRequestV2, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.GetTarRequestV2.displayName = 'proto.pfs.GetTarRequestV2';
}
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
proto.pfs.DiffFileResponseV2 = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.DiffFileResponseV2, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DiffFileResponseV2.displayName = 'proto.pfs.DiffFileResponseV2';
}
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
proto.pfs.CreateTmpFileSetResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.CreateTmpFileSetResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CreateTmpFileSetResponse.displayName = 'proto.pfs.CreateTmpFileSetResponse';
}
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
proto.pfs.RenewTmpFileSetRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.RenewTmpFileSetRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.RenewTmpFileSetRequest.displayName = 'proto.pfs.RenewTmpFileSetRequest';
}
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
proto.pfs.ClearCommitRequestV2 = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ClearCommitRequestV2, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ClearCommitRequestV2.displayName = 'proto.pfs.ClearCommitRequestV2';
}
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
proto.pfs.PutObjectRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.PutObjectRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs.PutObjectRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.PutObjectRequest.displayName = 'proto.pfs.PutObjectRequest';
}
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
proto.pfs.CreateObjectRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.CreateObjectRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CreateObjectRequest.displayName = 'proto.pfs.CreateObjectRequest';
}
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
proto.pfs.GetObjectsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.GetObjectsRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs.GetObjectsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.GetObjectsRequest.displayName = 'proto.pfs.GetObjectsRequest';
}
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
proto.pfs.PutBlockRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.PutBlockRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.PutBlockRequest.displayName = 'proto.pfs.PutBlockRequest';
}
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
proto.pfs.GetBlockRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.GetBlockRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.GetBlockRequest.displayName = 'proto.pfs.GetBlockRequest';
}
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
proto.pfs.GetBlocksRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.GetBlocksRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs.GetBlocksRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.GetBlocksRequest.displayName = 'proto.pfs.GetBlocksRequest';
}
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
proto.pfs.ListBlockRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ListBlockRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ListBlockRequest.displayName = 'proto.pfs.ListBlockRequest';
}
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
proto.pfs.TagObjectRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.TagObjectRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs.TagObjectRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.TagObjectRequest.displayName = 'proto.pfs.TagObjectRequest';
}
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
proto.pfs.ListObjectsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ListObjectsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ListObjectsRequest.displayName = 'proto.pfs.ListObjectsRequest';
}
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
proto.pfs.ListTagsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ListTagsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ListTagsRequest.displayName = 'proto.pfs.ListTagsRequest';
}
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
proto.pfs.ListTagsResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ListTagsResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ListTagsResponse.displayName = 'proto.pfs.ListTagsResponse';
}
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
proto.pfs.DeleteObjectsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.DeleteObjectsRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs.DeleteObjectsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DeleteObjectsRequest.displayName = 'proto.pfs.DeleteObjectsRequest';
}
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
proto.pfs.DeleteObjectsResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.DeleteObjectsResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DeleteObjectsResponse.displayName = 'proto.pfs.DeleteObjectsResponse';
}
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
proto.pfs.DeleteTagsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.DeleteTagsRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs.DeleteTagsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DeleteTagsRequest.displayName = 'proto.pfs.DeleteTagsRequest';
}
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
proto.pfs.DeleteTagsResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.DeleteTagsResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DeleteTagsResponse.displayName = 'proto.pfs.DeleteTagsResponse';
}
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
proto.pfs.CheckObjectRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.CheckObjectRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CheckObjectRequest.displayName = 'proto.pfs.CheckObjectRequest';
}
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
proto.pfs.CheckObjectResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.CheckObjectResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.CheckObjectResponse.displayName = 'proto.pfs.CheckObjectResponse';
}
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
proto.pfs.Objects = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs.Objects.repeatedFields_, null);
};
goog.inherits(proto.pfs.Objects, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.Objects.displayName = 'proto.pfs.Objects';
}
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
proto.pfs.PutObjDirectRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.PutObjDirectRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.PutObjDirectRequest.displayName = 'proto.pfs.PutObjDirectRequest';
}
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
proto.pfs.GetObjDirectRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.GetObjDirectRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.GetObjDirectRequest.displayName = 'proto.pfs.GetObjDirectRequest';
}
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
proto.pfs.DeleteObjDirectRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.DeleteObjDirectRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.DeleteObjDirectRequest.displayName = 'proto.pfs.DeleteObjDirectRequest';
}
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
proto.pfs.ObjectIndex = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs.ObjectIndex, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs.ObjectIndex.displayName = 'proto.pfs.ObjectIndex';
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
proto.pfs.Repo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.Repo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.Repo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Repo.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, "")
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
 * @return {!proto.pfs.Repo}
 */
proto.pfs.Repo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.Repo;
  return proto.pfs.Repo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.Repo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.Repo}
 */
proto.pfs.Repo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
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
proto.pfs.Repo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.Repo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.Repo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Repo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pfs.Repo.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.Repo} returns this
 */
proto.pfs.Repo.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





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
proto.pfs.Branch.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.Branch.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.Branch} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Branch.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs.Repo.toObject(includeInstance, f),
    name: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pfs.Branch}
 */
proto.pfs.Branch.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.Branch;
  return proto.pfs.Branch.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.Branch} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.Branch}
 */
proto.pfs.Branch.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Repo;
      reader.readMessage(value,proto.pfs.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
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
proto.pfs.Branch.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.Branch.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.Branch} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Branch.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Repo.serializeBinaryToWriter
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs.Repo}
 */
proto.pfs.Branch.prototype.getRepo = function() {
  return /** @type{?proto.pfs.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Repo, 1));
};


/**
 * @param {?proto.pfs.Repo|undefined} value
 * @return {!proto.pfs.Branch} returns this
*/
proto.pfs.Branch.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.Branch} returns this
 */
proto.pfs.Branch.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.Branch.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.pfs.Branch.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.Branch} returns this
 */
proto.pfs.Branch.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





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
proto.pfs.File.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.File.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.File} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.File.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs.Commit.toObject(includeInstance, f),
    path: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pfs.File}
 */
proto.pfs.File.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.File;
  return proto.pfs.File.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.File} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.File}
 */
proto.pfs.File.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setPath(value);
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
proto.pfs.File.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.File.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.File} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.File.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.File.prototype.getCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.File} returns this
*/
proto.pfs.File.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.File} returns this
 */
proto.pfs.File.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.File.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string path = 2;
 * @return {string}
 */
proto.pfs.File.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.File} returns this
 */
proto.pfs.File.prototype.setPath = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





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
proto.pfs.Block.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.Block.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.Block} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Block.toObject = function(includeInstance, msg) {
  var f, obj = {
    hash: jspb.Message.getFieldWithDefault(msg, 1, "")
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
 * @return {!proto.pfs.Block}
 */
proto.pfs.Block.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.Block;
  return proto.pfs.Block.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.Block} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.Block}
 */
proto.pfs.Block.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setHash(value);
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
proto.pfs.Block.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.Block.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.Block} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Block.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getHash();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string hash = 1;
 * @return {string}
 */
proto.pfs.Block.prototype.getHash = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.Block} returns this
 */
proto.pfs.Block.prototype.setHash = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





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
proto.pfs.Object.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.Object.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.Object} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Object.toObject = function(includeInstance, msg) {
  var f, obj = {
    hash: jspb.Message.getFieldWithDefault(msg, 1, "")
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
 * @return {!proto.pfs.Object}
 */
proto.pfs.Object.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.Object;
  return proto.pfs.Object.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.Object} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.Object}
 */
proto.pfs.Object.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setHash(value);
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
proto.pfs.Object.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.Object.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.Object} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Object.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getHash();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string hash = 1;
 * @return {string}
 */
proto.pfs.Object.prototype.getHash = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.Object} returns this
 */
proto.pfs.Object.prototype.setHash = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





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
proto.pfs.Tag.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.Tag.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.Tag} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Tag.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, "")
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
 * @return {!proto.pfs.Tag}
 */
proto.pfs.Tag.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.Tag;
  return proto.pfs.Tag.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.Tag} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.Tag}
 */
proto.pfs.Tag.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
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
proto.pfs.Tag.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.Tag.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.Tag} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Tag.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pfs.Tag.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.Tag} returns this
 */
proto.pfs.Tag.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.RepoInfo.repeatedFields_ = [7];



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
proto.pfs.RepoInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.RepoInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.RepoInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.RepoInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs.Repo.toObject(includeInstance, f),
    created: (f = msg.getCreated()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 3, 0),
    description: jspb.Message.getFieldWithDefault(msg, 5, ""),
    branchesList: jspb.Message.toObjectList(msg.getBranchesList(),
    proto.pfs.Branch.toObject, includeInstance),
    authInfo: (f = msg.getAuthInfo()) && proto.pfs.RepoAuthInfo.toObject(includeInstance, f),
    tombstone: jspb.Message.getBooleanFieldWithDefault(msg, 8, false)
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
 * @return {!proto.pfs.RepoInfo}
 */
proto.pfs.RepoInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.RepoInfo;
  return proto.pfs.RepoInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.RepoInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.RepoInfo}
 */
proto.pfs.RepoInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Repo;
      reader.readMessage(value,proto.pfs.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
      break;
    case 2:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCreated(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setSizeBytes(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 7:
      var value = new proto.pfs.Branch;
      reader.readMessage(value,proto.pfs.Branch.deserializeBinaryFromReader);
      msg.addBranches(value);
      break;
    case 6:
      var value = new proto.pfs.RepoAuthInfo;
      reader.readMessage(value,proto.pfs.RepoAuthInfo.deserializeBinaryFromReader);
      msg.setAuthInfo(value);
      break;
    case 8:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setTombstone(value);
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
proto.pfs.RepoInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.RepoInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.RepoInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.RepoInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Repo.serializeBinaryToWriter
    );
  }
  f = message.getCreated();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getSizeBytes();
  if (f !== 0) {
    writer.writeUint64(
      3,
      f
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getBranchesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      7,
      f,
      proto.pfs.Branch.serializeBinaryToWriter
    );
  }
  f = message.getAuthInfo();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      proto.pfs.RepoAuthInfo.serializeBinaryToWriter
    );
  }
  f = message.getTombstone();
  if (f) {
    writer.writeBool(
      8,
      f
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs.Repo}
 */
proto.pfs.RepoInfo.prototype.getRepo = function() {
  return /** @type{?proto.pfs.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Repo, 1));
};


/**
 * @param {?proto.pfs.Repo|undefined} value
 * @return {!proto.pfs.RepoInfo} returns this
*/
proto.pfs.RepoInfo.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.RepoInfo} returns this
 */
proto.pfs.RepoInfo.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.RepoInfo.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional google.protobuf.Timestamp created = 2;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pfs.RepoInfo.prototype.getCreated = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 2));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pfs.RepoInfo} returns this
*/
proto.pfs.RepoInfo.prototype.setCreated = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.RepoInfo} returns this
 */
proto.pfs.RepoInfo.prototype.clearCreated = function() {
  return this.setCreated(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.RepoInfo.prototype.hasCreated = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional uint64 size_bytes = 3;
 * @return {number}
 */
proto.pfs.RepoInfo.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.RepoInfo} returns this
 */
proto.pfs.RepoInfo.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional string description = 5;
 * @return {string}
 */
proto.pfs.RepoInfo.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.RepoInfo} returns this
 */
proto.pfs.RepoInfo.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * repeated Branch branches = 7;
 * @return {!Array<!proto.pfs.Branch>}
 */
proto.pfs.RepoInfo.prototype.getBranchesList = function() {
  return /** @type{!Array<!proto.pfs.Branch>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Branch, 7));
};


/**
 * @param {!Array<!proto.pfs.Branch>} value
 * @return {!proto.pfs.RepoInfo} returns this
*/
proto.pfs.RepoInfo.prototype.setBranchesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 7, value);
};


/**
 * @param {!proto.pfs.Branch=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Branch}
 */
proto.pfs.RepoInfo.prototype.addBranches = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 7, opt_value, proto.pfs.Branch, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.RepoInfo} returns this
 */
proto.pfs.RepoInfo.prototype.clearBranchesList = function() {
  return this.setBranchesList([]);
};


/**
 * optional RepoAuthInfo auth_info = 6;
 * @return {?proto.pfs.RepoAuthInfo}
 */
proto.pfs.RepoInfo.prototype.getAuthInfo = function() {
  return /** @type{?proto.pfs.RepoAuthInfo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.RepoAuthInfo, 6));
};


/**
 * @param {?proto.pfs.RepoAuthInfo|undefined} value
 * @return {!proto.pfs.RepoInfo} returns this
*/
proto.pfs.RepoInfo.prototype.setAuthInfo = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.RepoInfo} returns this
 */
proto.pfs.RepoInfo.prototype.clearAuthInfo = function() {
  return this.setAuthInfo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.RepoInfo.prototype.hasAuthInfo = function() {
  return jspb.Message.getField(this, 6) != null;
};


/**
 * optional bool tombstone = 8;
 * @return {boolean}
 */
proto.pfs.RepoInfo.prototype.getTombstone = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 8, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.RepoInfo} returns this
 */
proto.pfs.RepoInfo.prototype.setTombstone = function(value) {
  return jspb.Message.setProto3BooleanField(this, 8, value);
};





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
proto.pfs.RepoAuthInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.RepoAuthInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.RepoAuthInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.RepoAuthInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    accessLevel: jspb.Message.getFieldWithDefault(msg, 1, 0)
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
 * @return {!proto.pfs.RepoAuthInfo}
 */
proto.pfs.RepoAuthInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.RepoAuthInfo;
  return proto.pfs.RepoAuthInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.RepoAuthInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.RepoAuthInfo}
 */
proto.pfs.RepoAuthInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!proto.auth.Scope} */ (reader.readEnum());
      msg.setAccessLevel(value);
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
proto.pfs.RepoAuthInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.RepoAuthInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.RepoAuthInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.RepoAuthInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAccessLevel();
  if (f !== 0.0) {
    writer.writeEnum(
      1,
      f
    );
  }
};


/**
 * optional auth.Scope access_level = 1;
 * @return {!proto.auth.Scope}
 */
proto.pfs.RepoAuthInfo.prototype.getAccessLevel = function() {
  return /** @type {!proto.auth.Scope} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {!proto.auth.Scope} value
 * @return {!proto.pfs.RepoAuthInfo} returns this
 */
proto.pfs.RepoAuthInfo.prototype.setAccessLevel = function(value) {
  return jspb.Message.setProto3EnumField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.BranchInfo.repeatedFields_ = [3,5,6];



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
proto.pfs.BranchInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.BranchInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.BranchInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.BranchInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    branch: (f = msg.getBranch()) && proto.pfs.Branch.toObject(includeInstance, f),
    head: (f = msg.getHead()) && proto.pfs.Commit.toObject(includeInstance, f),
    provenanceList: jspb.Message.toObjectList(msg.getProvenanceList(),
    proto.pfs.Branch.toObject, includeInstance),
    subvenanceList: jspb.Message.toObjectList(msg.getSubvenanceList(),
    proto.pfs.Branch.toObject, includeInstance),
    directProvenanceList: jspb.Message.toObjectList(msg.getDirectProvenanceList(),
    proto.pfs.Branch.toObject, includeInstance),
    trigger: (f = msg.getTrigger()) && proto.pfs.Trigger.toObject(includeInstance, f),
    name: jspb.Message.getFieldWithDefault(msg, 1, "")
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
 * @return {!proto.pfs.BranchInfo}
 */
proto.pfs.BranchInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.BranchInfo;
  return proto.pfs.BranchInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.BranchInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.BranchInfo}
 */
proto.pfs.BranchInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 4:
      var value = new proto.pfs.Branch;
      reader.readMessage(value,proto.pfs.Branch.deserializeBinaryFromReader);
      msg.setBranch(value);
      break;
    case 2:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setHead(value);
      break;
    case 3:
      var value = new proto.pfs.Branch;
      reader.readMessage(value,proto.pfs.Branch.deserializeBinaryFromReader);
      msg.addProvenance(value);
      break;
    case 5:
      var value = new proto.pfs.Branch;
      reader.readMessage(value,proto.pfs.Branch.deserializeBinaryFromReader);
      msg.addSubvenance(value);
      break;
    case 6:
      var value = new proto.pfs.Branch;
      reader.readMessage(value,proto.pfs.Branch.deserializeBinaryFromReader);
      msg.addDirectProvenance(value);
      break;
    case 7:
      var value = new proto.pfs.Trigger;
      reader.readMessage(value,proto.pfs.Trigger.deserializeBinaryFromReader);
      msg.setTrigger(value);
      break;
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
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
proto.pfs.BranchInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.BranchInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.BranchInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.BranchInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pfs.Branch.serializeBinaryToWriter
    );
  }
  f = message.getHead();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      3,
      f,
      proto.pfs.Branch.serializeBinaryToWriter
    );
  }
  f = message.getSubvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      5,
      f,
      proto.pfs.Branch.serializeBinaryToWriter
    );
  }
  f = message.getDirectProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      6,
      f,
      proto.pfs.Branch.serializeBinaryToWriter
    );
  }
  f = message.getTrigger();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      proto.pfs.Trigger.serializeBinaryToWriter
    );
  }
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional Branch branch = 4;
 * @return {?proto.pfs.Branch}
 */
proto.pfs.BranchInfo.prototype.getBranch = function() {
  return /** @type{?proto.pfs.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Branch, 4));
};


/**
 * @param {?proto.pfs.Branch|undefined} value
 * @return {!proto.pfs.BranchInfo} returns this
*/
proto.pfs.BranchInfo.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.BranchInfo} returns this
 */
proto.pfs.BranchInfo.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.BranchInfo.prototype.hasBranch = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional Commit head = 2;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.BranchInfo.prototype.getHead = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 2));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.BranchInfo} returns this
*/
proto.pfs.BranchInfo.prototype.setHead = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.BranchInfo} returns this
 */
proto.pfs.BranchInfo.prototype.clearHead = function() {
  return this.setHead(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.BranchInfo.prototype.hasHead = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * repeated Branch provenance = 3;
 * @return {!Array<!proto.pfs.Branch>}
 */
proto.pfs.BranchInfo.prototype.getProvenanceList = function() {
  return /** @type{!Array<!proto.pfs.Branch>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Branch, 3));
};


/**
 * @param {!Array<!proto.pfs.Branch>} value
 * @return {!proto.pfs.BranchInfo} returns this
*/
proto.pfs.BranchInfo.prototype.setProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 3, value);
};


/**
 * @param {!proto.pfs.Branch=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Branch}
 */
proto.pfs.BranchInfo.prototype.addProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 3, opt_value, proto.pfs.Branch, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.BranchInfo} returns this
 */
proto.pfs.BranchInfo.prototype.clearProvenanceList = function() {
  return this.setProvenanceList([]);
};


/**
 * repeated Branch subvenance = 5;
 * @return {!Array<!proto.pfs.Branch>}
 */
proto.pfs.BranchInfo.prototype.getSubvenanceList = function() {
  return /** @type{!Array<!proto.pfs.Branch>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Branch, 5));
};


/**
 * @param {!Array<!proto.pfs.Branch>} value
 * @return {!proto.pfs.BranchInfo} returns this
*/
proto.pfs.BranchInfo.prototype.setSubvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.pfs.Branch=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Branch}
 */
proto.pfs.BranchInfo.prototype.addSubvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.pfs.Branch, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.BranchInfo} returns this
 */
proto.pfs.BranchInfo.prototype.clearSubvenanceList = function() {
  return this.setSubvenanceList([]);
};


/**
 * repeated Branch direct_provenance = 6;
 * @return {!Array<!proto.pfs.Branch>}
 */
proto.pfs.BranchInfo.prototype.getDirectProvenanceList = function() {
  return /** @type{!Array<!proto.pfs.Branch>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Branch, 6));
};


/**
 * @param {!Array<!proto.pfs.Branch>} value
 * @return {!proto.pfs.BranchInfo} returns this
*/
proto.pfs.BranchInfo.prototype.setDirectProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 6, value);
};


/**
 * @param {!proto.pfs.Branch=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Branch}
 */
proto.pfs.BranchInfo.prototype.addDirectProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 6, opt_value, proto.pfs.Branch, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.BranchInfo} returns this
 */
proto.pfs.BranchInfo.prototype.clearDirectProvenanceList = function() {
  return this.setDirectProvenanceList([]);
};


/**
 * optional Trigger trigger = 7;
 * @return {?proto.pfs.Trigger}
 */
proto.pfs.BranchInfo.prototype.getTrigger = function() {
  return /** @type{?proto.pfs.Trigger} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Trigger, 7));
};


/**
 * @param {?proto.pfs.Trigger|undefined} value
 * @return {!proto.pfs.BranchInfo} returns this
*/
proto.pfs.BranchInfo.prototype.setTrigger = function(value) {
  return jspb.Message.setWrapperField(this, 7, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.BranchInfo} returns this
 */
proto.pfs.BranchInfo.prototype.clearTrigger = function() {
  return this.setTrigger(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.BranchInfo.prototype.hasTrigger = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pfs.BranchInfo.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.BranchInfo} returns this
 */
proto.pfs.BranchInfo.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.BranchInfos.repeatedFields_ = [1];



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
proto.pfs.BranchInfos.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.BranchInfos.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.BranchInfos} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.BranchInfos.toObject = function(includeInstance, msg) {
  var f, obj = {
    branchInfoList: jspb.Message.toObjectList(msg.getBranchInfoList(),
    proto.pfs.BranchInfo.toObject, includeInstance)
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
 * @return {!proto.pfs.BranchInfos}
 */
proto.pfs.BranchInfos.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.BranchInfos;
  return proto.pfs.BranchInfos.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.BranchInfos} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.BranchInfos}
 */
proto.pfs.BranchInfos.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.BranchInfo;
      reader.readMessage(value,proto.pfs.BranchInfo.deserializeBinaryFromReader);
      msg.addBranchInfo(value);
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
proto.pfs.BranchInfos.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.BranchInfos.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.BranchInfos} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.BranchInfos.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBranchInfoList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs.BranchInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated BranchInfo branch_info = 1;
 * @return {!Array<!proto.pfs.BranchInfo>}
 */
proto.pfs.BranchInfos.prototype.getBranchInfoList = function() {
  return /** @type{!Array<!proto.pfs.BranchInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.BranchInfo, 1));
};


/**
 * @param {!Array<!proto.pfs.BranchInfo>} value
 * @return {!proto.pfs.BranchInfos} returns this
*/
proto.pfs.BranchInfos.prototype.setBranchInfoList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.BranchInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.BranchInfo}
 */
proto.pfs.BranchInfos.prototype.addBranchInfo = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.BranchInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.BranchInfos} returns this
 */
proto.pfs.BranchInfos.prototype.clearBranchInfoList = function() {
  return this.setBranchInfoList([]);
};





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
proto.pfs.Trigger.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.Trigger.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.Trigger} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Trigger.toObject = function(includeInstance, msg) {
  var f, obj = {
    branch: jspb.Message.getFieldWithDefault(msg, 1, ""),
    all: jspb.Message.getBooleanFieldWithDefault(msg, 2, false),
    cronSpec: jspb.Message.getFieldWithDefault(msg, 3, ""),
    size: jspb.Message.getFieldWithDefault(msg, 4, ""),
    commits: jspb.Message.getFieldWithDefault(msg, 5, 0)
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
 * @return {!proto.pfs.Trigger}
 */
proto.pfs.Trigger.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.Trigger;
  return proto.pfs.Trigger.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.Trigger} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.Trigger}
 */
proto.pfs.Trigger.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setBranch(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAll(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setCronSpec(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setSize(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setCommits(value);
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
proto.pfs.Trigger.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.Trigger.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.Trigger} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Trigger.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBranch();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getAll();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
  f = message.getCronSpec();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getSize();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getCommits();
  if (f !== 0) {
    writer.writeInt64(
      5,
      f
    );
  }
};


/**
 * optional string branch = 1;
 * @return {string}
 */
proto.pfs.Trigger.prototype.getBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.Trigger} returns this
 */
proto.pfs.Trigger.prototype.setBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional bool all = 2;
 * @return {boolean}
 */
proto.pfs.Trigger.prototype.getAll = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.Trigger} returns this
 */
proto.pfs.Trigger.prototype.setAll = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};


/**
 * optional string cron_spec = 3;
 * @return {string}
 */
proto.pfs.Trigger.prototype.getCronSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.Trigger} returns this
 */
proto.pfs.Trigger.prototype.setCronSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string size = 4;
 * @return {string}
 */
proto.pfs.Trigger.prototype.getSize = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.Trigger} returns this
 */
proto.pfs.Trigger.prototype.setSize = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional int64 commits = 5;
 * @return {number}
 */
proto.pfs.Trigger.prototype.getCommits = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.Trigger} returns this
 */
proto.pfs.Trigger.prototype.setCommits = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};





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
proto.pfs.CommitOrigin.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CommitOrigin.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CommitOrigin} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CommitOrigin.toObject = function(includeInstance, msg) {
  var f, obj = {
    kind: jspb.Message.getFieldWithDefault(msg, 1, 0)
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
 * @return {!proto.pfs.CommitOrigin}
 */
proto.pfs.CommitOrigin.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CommitOrigin;
  return proto.pfs.CommitOrigin.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CommitOrigin} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CommitOrigin}
 */
proto.pfs.CommitOrigin.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!proto.pfs.OriginKind} */ (reader.readEnum());
      msg.setKind(value);
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
proto.pfs.CommitOrigin.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CommitOrigin.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CommitOrigin} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CommitOrigin.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getKind();
  if (f !== 0.0) {
    writer.writeEnum(
      1,
      f
    );
  }
};


/**
 * optional OriginKind kind = 1;
 * @return {!proto.pfs.OriginKind}
 */
proto.pfs.CommitOrigin.prototype.getKind = function() {
  return /** @type {!proto.pfs.OriginKind} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {!proto.pfs.OriginKind} value
 * @return {!proto.pfs.CommitOrigin} returns this
 */
proto.pfs.CommitOrigin.prototype.setKind = function(value) {
  return jspb.Message.setProto3EnumField(this, 1, value);
};





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
proto.pfs.Commit.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.Commit.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.Commit} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Commit.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs.Repo.toObject(includeInstance, f),
    id: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pfs.Commit}
 */
proto.pfs.Commit.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.Commit;
  return proto.pfs.Commit.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.Commit} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.Commit}
 */
proto.pfs.Commit.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Repo;
      reader.readMessage(value,proto.pfs.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
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
proto.pfs.Commit.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.Commit.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.Commit} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Commit.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Repo.serializeBinaryToWriter
    );
  }
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs.Repo}
 */
proto.pfs.Commit.prototype.getRepo = function() {
  return /** @type{?proto.pfs.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Repo, 1));
};


/**
 * @param {?proto.pfs.Repo|undefined} value
 * @return {!proto.pfs.Commit} returns this
*/
proto.pfs.Commit.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.Commit} returns this
 */
proto.pfs.Commit.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.Commit.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string id = 2;
 * @return {string}
 */
proto.pfs.Commit.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.Commit} returns this
 */
proto.pfs.Commit.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





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
proto.pfs.CommitRange.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CommitRange.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CommitRange} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CommitRange.toObject = function(includeInstance, msg) {
  var f, obj = {
    lower: (f = msg.getLower()) && proto.pfs.Commit.toObject(includeInstance, f),
    upper: (f = msg.getUpper()) && proto.pfs.Commit.toObject(includeInstance, f)
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
 * @return {!proto.pfs.CommitRange}
 */
proto.pfs.CommitRange.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CommitRange;
  return proto.pfs.CommitRange.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CommitRange} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CommitRange}
 */
proto.pfs.CommitRange.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setLower(value);
      break;
    case 2:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setUpper(value);
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
proto.pfs.CommitRange.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CommitRange.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CommitRange} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CommitRange.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getLower();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getUpper();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit lower = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.CommitRange.prototype.getLower = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.CommitRange} returns this
*/
proto.pfs.CommitRange.prototype.setLower = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitRange} returns this
 */
proto.pfs.CommitRange.prototype.clearLower = function() {
  return this.setLower(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitRange.prototype.hasLower = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Commit upper = 2;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.CommitRange.prototype.getUpper = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 2));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.CommitRange} returns this
*/
proto.pfs.CommitRange.prototype.setUpper = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitRange} returns this
 */
proto.pfs.CommitRange.prototype.clearUpper = function() {
  return this.setUpper(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitRange.prototype.hasUpper = function() {
  return jspb.Message.getField(this, 2) != null;
};





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
proto.pfs.CommitProvenance.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CommitProvenance.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CommitProvenance} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CommitProvenance.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs.Commit.toObject(includeInstance, f),
    branch: (f = msg.getBranch()) && proto.pfs.Branch.toObject(includeInstance, f)
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
 * @return {!proto.pfs.CommitProvenance}
 */
proto.pfs.CommitProvenance.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CommitProvenance;
  return proto.pfs.CommitProvenance.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CommitProvenance} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CommitProvenance}
 */
proto.pfs.CommitProvenance.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = new proto.pfs.Branch;
      reader.readMessage(value,proto.pfs.Branch.deserializeBinaryFromReader);
      msg.setBranch(value);
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
proto.pfs.CommitProvenance.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CommitProvenance.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CommitProvenance} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CommitProvenance.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.Branch.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.CommitProvenance.prototype.getCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.CommitProvenance} returns this
*/
proto.pfs.CommitProvenance.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitProvenance} returns this
 */
proto.pfs.CommitProvenance.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitProvenance.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Branch branch = 2;
 * @return {?proto.pfs.Branch}
 */
proto.pfs.CommitProvenance.prototype.getBranch = function() {
  return /** @type{?proto.pfs.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Branch, 2));
};


/**
 * @param {?proto.pfs.Branch|undefined} value
 * @return {!proto.pfs.CommitProvenance} returns this
*/
proto.pfs.CommitProvenance.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitProvenance} returns this
 */
proto.pfs.CommitProvenance.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitProvenance.prototype.hasBranch = function() {
  return jspb.Message.getField(this, 2) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.CommitInfo.repeatedFields_ = [11,16,9,13];



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
proto.pfs.CommitInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CommitInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CommitInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CommitInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs.Commit.toObject(includeInstance, f),
    branch: (f = msg.getBranch()) && proto.pfs.Branch.toObject(includeInstance, f),
    origin: (f = msg.getOrigin()) && proto.pfs.CommitOrigin.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 8, ""),
    parentCommit: (f = msg.getParentCommit()) && proto.pfs.Commit.toObject(includeInstance, f),
    childCommitsList: jspb.Message.toObjectList(msg.getChildCommitsList(),
    proto.pfs.Commit.toObject, includeInstance),
    started: (f = msg.getStarted()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    finished: (f = msg.getFinished()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 5, 0),
    provenanceList: jspb.Message.toObjectList(msg.getProvenanceList(),
    proto.pfs.CommitProvenance.toObject, includeInstance),
    readyProvenance: jspb.Message.getFieldWithDefault(msg, 12, 0),
    subvenanceList: jspb.Message.toObjectList(msg.getSubvenanceList(),
    proto.pfs.CommitRange.toObject, includeInstance),
    tree: (f = msg.getTree()) && proto.pfs.Object.toObject(includeInstance, f),
    treesList: jspb.Message.toObjectList(msg.getTreesList(),
    proto.pfs.Object.toObject, includeInstance),
    datums: (f = msg.getDatums()) && proto.pfs.Object.toObject(includeInstance, f),
    subvenantCommitsSuccess: jspb.Message.getFieldWithDefault(msg, 18, 0),
    subvenantCommitsFailure: jspb.Message.getFieldWithDefault(msg, 19, 0),
    subvenantCommitsTotal: jspb.Message.getFieldWithDefault(msg, 20, 0)
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
 * @return {!proto.pfs.CommitInfo}
 */
proto.pfs.CommitInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CommitInfo;
  return proto.pfs.CommitInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CommitInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CommitInfo}
 */
proto.pfs.CommitInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 15:
      var value = new proto.pfs.Branch;
      reader.readMessage(value,proto.pfs.Branch.deserializeBinaryFromReader);
      msg.setBranch(value);
      break;
    case 17:
      var value = new proto.pfs.CommitOrigin;
      reader.readMessage(value,proto.pfs.CommitOrigin.deserializeBinaryFromReader);
      msg.setOrigin(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 2:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setParentCommit(value);
      break;
    case 11:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.addChildCommits(value);
      break;
    case 3:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setStarted(value);
      break;
    case 4:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setFinished(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setSizeBytes(value);
      break;
    case 16:
      var value = new proto.pfs.CommitProvenance;
      reader.readMessage(value,proto.pfs.CommitProvenance.deserializeBinaryFromReader);
      msg.addProvenance(value);
      break;
    case 12:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setReadyProvenance(value);
      break;
    case 9:
      var value = new proto.pfs.CommitRange;
      reader.readMessage(value,proto.pfs.CommitRange.deserializeBinaryFromReader);
      msg.addSubvenance(value);
      break;
    case 7:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.setTree(value);
      break;
    case 13:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.addTrees(value);
      break;
    case 14:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.setDatums(value);
      break;
    case 18:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setSubvenantCommitsSuccess(value);
      break;
    case 19:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setSubvenantCommitsFailure(value);
      break;
    case 20:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setSubvenantCommitsTotal(value);
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
proto.pfs.CommitInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CommitInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CommitInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CommitInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      15,
      f,
      proto.pfs.Branch.serializeBinaryToWriter
    );
  }
  f = message.getOrigin();
  if (f != null) {
    writer.writeMessage(
      17,
      f,
      proto.pfs.CommitOrigin.serializeBinaryToWriter
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
  f = message.getParentCommit();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getChildCommitsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      11,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getStarted();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getFinished();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getSizeBytes();
  if (f !== 0) {
    writer.writeUint64(
      5,
      f
    );
  }
  f = message.getProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      16,
      f,
      proto.pfs.CommitProvenance.serializeBinaryToWriter
    );
  }
  f = message.getReadyProvenance();
  if (f !== 0) {
    writer.writeInt64(
      12,
      f
    );
  }
  f = message.getSubvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      9,
      f,
      proto.pfs.CommitRange.serializeBinaryToWriter
    );
  }
  f = message.getTree();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getTreesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      13,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getDatums();
  if (f != null) {
    writer.writeMessage(
      14,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getSubvenantCommitsSuccess();
  if (f !== 0) {
    writer.writeInt64(
      18,
      f
    );
  }
  f = message.getSubvenantCommitsFailure();
  if (f !== 0) {
    writer.writeInt64(
      19,
      f
    );
  }
  f = message.getSubvenantCommitsTotal();
  if (f !== 0) {
    writer.writeInt64(
      20,
      f
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.CommitInfo.prototype.getCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitInfo.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Branch branch = 15;
 * @return {?proto.pfs.Branch}
 */
proto.pfs.CommitInfo.prototype.getBranch = function() {
  return /** @type{?proto.pfs.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Branch, 15));
};


/**
 * @param {?proto.pfs.Branch|undefined} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 15, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitInfo.prototype.hasBranch = function() {
  return jspb.Message.getField(this, 15) != null;
};


/**
 * optional CommitOrigin origin = 17;
 * @return {?proto.pfs.CommitOrigin}
 */
proto.pfs.CommitInfo.prototype.getOrigin = function() {
  return /** @type{?proto.pfs.CommitOrigin} */ (
    jspb.Message.getWrapperField(this, proto.pfs.CommitOrigin, 17));
};


/**
 * @param {?proto.pfs.CommitOrigin|undefined} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setOrigin = function(value) {
  return jspb.Message.setWrapperField(this, 17, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearOrigin = function() {
  return this.setOrigin(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitInfo.prototype.hasOrigin = function() {
  return jspb.Message.getField(this, 17) != null;
};


/**
 * optional string description = 8;
 * @return {string}
 */
proto.pfs.CommitInfo.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};


/**
 * optional Commit parent_commit = 2;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.CommitInfo.prototype.getParentCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 2));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setParentCommit = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearParentCommit = function() {
  return this.setParentCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitInfo.prototype.hasParentCommit = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * repeated Commit child_commits = 11;
 * @return {!Array<!proto.pfs.Commit>}
 */
proto.pfs.CommitInfo.prototype.getChildCommitsList = function() {
  return /** @type{!Array<!proto.pfs.Commit>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Commit, 11));
};


/**
 * @param {!Array<!proto.pfs.Commit>} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setChildCommitsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 11, value);
};


/**
 * @param {!proto.pfs.Commit=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Commit}
 */
proto.pfs.CommitInfo.prototype.addChildCommits = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 11, opt_value, proto.pfs.Commit, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearChildCommitsList = function() {
  return this.setChildCommitsList([]);
};


/**
 * optional google.protobuf.Timestamp started = 3;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pfs.CommitInfo.prototype.getStarted = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 3));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setStarted = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearStarted = function() {
  return this.setStarted(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitInfo.prototype.hasStarted = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional google.protobuf.Timestamp finished = 4;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pfs.CommitInfo.prototype.getFinished = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 4));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setFinished = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearFinished = function() {
  return this.setFinished(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitInfo.prototype.hasFinished = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional uint64 size_bytes = 5;
 * @return {number}
 */
proto.pfs.CommitInfo.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};


/**
 * repeated CommitProvenance provenance = 16;
 * @return {!Array<!proto.pfs.CommitProvenance>}
 */
proto.pfs.CommitInfo.prototype.getProvenanceList = function() {
  return /** @type{!Array<!proto.pfs.CommitProvenance>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.CommitProvenance, 16));
};


/**
 * @param {!Array<!proto.pfs.CommitProvenance>} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 16, value);
};


/**
 * @param {!proto.pfs.CommitProvenance=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.CommitProvenance}
 */
proto.pfs.CommitInfo.prototype.addProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 16, opt_value, proto.pfs.CommitProvenance, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearProvenanceList = function() {
  return this.setProvenanceList([]);
};


/**
 * optional int64 ready_provenance = 12;
 * @return {number}
 */
proto.pfs.CommitInfo.prototype.getReadyProvenance = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 12, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.setReadyProvenance = function(value) {
  return jspb.Message.setProto3IntField(this, 12, value);
};


/**
 * repeated CommitRange subvenance = 9;
 * @return {!Array<!proto.pfs.CommitRange>}
 */
proto.pfs.CommitInfo.prototype.getSubvenanceList = function() {
  return /** @type{!Array<!proto.pfs.CommitRange>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.CommitRange, 9));
};


/**
 * @param {!Array<!proto.pfs.CommitRange>} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setSubvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 9, value);
};


/**
 * @param {!proto.pfs.CommitRange=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.CommitRange}
 */
proto.pfs.CommitInfo.prototype.addSubvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 9, opt_value, proto.pfs.CommitRange, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearSubvenanceList = function() {
  return this.setSubvenanceList([]);
};


/**
 * optional Object tree = 7;
 * @return {?proto.pfs.Object}
 */
proto.pfs.CommitInfo.prototype.getTree = function() {
  return /** @type{?proto.pfs.Object} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Object, 7));
};


/**
 * @param {?proto.pfs.Object|undefined} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setTree = function(value) {
  return jspb.Message.setWrapperField(this, 7, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearTree = function() {
  return this.setTree(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitInfo.prototype.hasTree = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * repeated Object trees = 13;
 * @return {!Array<!proto.pfs.Object>}
 */
proto.pfs.CommitInfo.prototype.getTreesList = function() {
  return /** @type{!Array<!proto.pfs.Object>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Object, 13));
};


/**
 * @param {!Array<!proto.pfs.Object>} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setTreesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 13, value);
};


/**
 * @param {!proto.pfs.Object=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Object}
 */
proto.pfs.CommitInfo.prototype.addTrees = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 13, opt_value, proto.pfs.Object, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearTreesList = function() {
  return this.setTreesList([]);
};


/**
 * optional Object datums = 14;
 * @return {?proto.pfs.Object}
 */
proto.pfs.CommitInfo.prototype.getDatums = function() {
  return /** @type{?proto.pfs.Object} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Object, 14));
};


/**
 * @param {?proto.pfs.Object|undefined} value
 * @return {!proto.pfs.CommitInfo} returns this
*/
proto.pfs.CommitInfo.prototype.setDatums = function(value) {
  return jspb.Message.setWrapperField(this, 14, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.clearDatums = function() {
  return this.setDatums(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CommitInfo.prototype.hasDatums = function() {
  return jspb.Message.getField(this, 14) != null;
};


/**
 * optional int64 subvenant_commits_success = 18;
 * @return {number}
 */
proto.pfs.CommitInfo.prototype.getSubvenantCommitsSuccess = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 18, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.setSubvenantCommitsSuccess = function(value) {
  return jspb.Message.setProto3IntField(this, 18, value);
};


/**
 * optional int64 subvenant_commits_failure = 19;
 * @return {number}
 */
proto.pfs.CommitInfo.prototype.getSubvenantCommitsFailure = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 19, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.setSubvenantCommitsFailure = function(value) {
  return jspb.Message.setProto3IntField(this, 19, value);
};


/**
 * optional int64 subvenant_commits_total = 20;
 * @return {number}
 */
proto.pfs.CommitInfo.prototype.getSubvenantCommitsTotal = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 20, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.CommitInfo} returns this
 */
proto.pfs.CommitInfo.prototype.setSubvenantCommitsTotal = function(value) {
  return jspb.Message.setProto3IntField(this, 20, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.FileInfo.repeatedFields_ = [6,8,9];



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
proto.pfs.FileInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.FileInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.FileInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FileInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs.File.toObject(includeInstance, f),
    fileType: jspb.Message.getFieldWithDefault(msg, 2, 0),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 3, 0),
    committed: (f = msg.getCommitted()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    childrenList: (f = jspb.Message.getRepeatedField(msg, 6)) == null ? undefined : f,
    objectsList: jspb.Message.toObjectList(msg.getObjectsList(),
    proto.pfs.Object.toObject, includeInstance),
    blockrefsList: jspb.Message.toObjectList(msg.getBlockrefsList(),
    proto.pfs.BlockRef.toObject, includeInstance),
    hash: msg.getHash_asB64()
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
 * @return {!proto.pfs.FileInfo}
 */
proto.pfs.FileInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.FileInfo;
  return proto.pfs.FileInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.FileInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.FileInfo}
 */
proto.pfs.FileInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setFile(value);
      break;
    case 2:
      var value = /** @type {!proto.pfs.FileType} */ (reader.readEnum());
      msg.setFileType(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setSizeBytes(value);
      break;
    case 10:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCommitted(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.addChildren(value);
      break;
    case 8:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.addObjects(value);
      break;
    case 9:
      var value = new proto.pfs.BlockRef;
      reader.readMessage(value,proto.pfs.BlockRef.deserializeBinaryFromReader);
      msg.addBlockrefs(value);
      break;
    case 7:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setHash(value);
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
proto.pfs.FileInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.FileInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.FileInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FileInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
  f = message.getFileType();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
  f = message.getSizeBytes();
  if (f !== 0) {
    writer.writeUint64(
      3,
      f
    );
  }
  f = message.getCommitted();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getChildrenList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      6,
      f
    );
  }
  f = message.getObjectsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      8,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getBlockrefsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      9,
      f,
      proto.pfs.BlockRef.serializeBinaryToWriter
    );
  }
  f = message.getHash_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      7,
      f
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs.File}
 */
proto.pfs.FileInfo.prototype.getFile = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 1));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.FileInfo} returns this
*/
proto.pfs.FileInfo.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.FileInfo} returns this
 */
proto.pfs.FileInfo.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.FileInfo.prototype.hasFile = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional FileType file_type = 2;
 * @return {!proto.pfs.FileType}
 */
proto.pfs.FileInfo.prototype.getFileType = function() {
  return /** @type {!proto.pfs.FileType} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.pfs.FileType} value
 * @return {!proto.pfs.FileInfo} returns this
 */
proto.pfs.FileInfo.prototype.setFileType = function(value) {
  return jspb.Message.setProto3EnumField(this, 2, value);
};


/**
 * optional uint64 size_bytes = 3;
 * @return {number}
 */
proto.pfs.FileInfo.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.FileInfo} returns this
 */
proto.pfs.FileInfo.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional google.protobuf.Timestamp committed = 10;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pfs.FileInfo.prototype.getCommitted = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 10));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pfs.FileInfo} returns this
*/
proto.pfs.FileInfo.prototype.setCommitted = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.FileInfo} returns this
 */
proto.pfs.FileInfo.prototype.clearCommitted = function() {
  return this.setCommitted(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.FileInfo.prototype.hasCommitted = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * repeated string children = 6;
 * @return {!Array<string>}
 */
proto.pfs.FileInfo.prototype.getChildrenList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 6));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pfs.FileInfo} returns this
 */
proto.pfs.FileInfo.prototype.setChildrenList = function(value) {
  return jspb.Message.setField(this, 6, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pfs.FileInfo} returns this
 */
proto.pfs.FileInfo.prototype.addChildren = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 6, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.FileInfo} returns this
 */
proto.pfs.FileInfo.prototype.clearChildrenList = function() {
  return this.setChildrenList([]);
};


/**
 * repeated Object objects = 8;
 * @return {!Array<!proto.pfs.Object>}
 */
proto.pfs.FileInfo.prototype.getObjectsList = function() {
  return /** @type{!Array<!proto.pfs.Object>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Object, 8));
};


/**
 * @param {!Array<!proto.pfs.Object>} value
 * @return {!proto.pfs.FileInfo} returns this
*/
proto.pfs.FileInfo.prototype.setObjectsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 8, value);
};


/**
 * @param {!proto.pfs.Object=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Object}
 */
proto.pfs.FileInfo.prototype.addObjects = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 8, opt_value, proto.pfs.Object, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.FileInfo} returns this
 */
proto.pfs.FileInfo.prototype.clearObjectsList = function() {
  return this.setObjectsList([]);
};


/**
 * repeated BlockRef blockRefs = 9;
 * @return {!Array<!proto.pfs.BlockRef>}
 */
proto.pfs.FileInfo.prototype.getBlockrefsList = function() {
  return /** @type{!Array<!proto.pfs.BlockRef>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.BlockRef, 9));
};


/**
 * @param {!Array<!proto.pfs.BlockRef>} value
 * @return {!proto.pfs.FileInfo} returns this
*/
proto.pfs.FileInfo.prototype.setBlockrefsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 9, value);
};


/**
 * @param {!proto.pfs.BlockRef=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.BlockRef}
 */
proto.pfs.FileInfo.prototype.addBlockrefs = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 9, opt_value, proto.pfs.BlockRef, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.FileInfo} returns this
 */
proto.pfs.FileInfo.prototype.clearBlockrefsList = function() {
  return this.setBlockrefsList([]);
};


/**
 * optional bytes hash = 7;
 * @return {string}
 */
proto.pfs.FileInfo.prototype.getHash = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * optional bytes hash = 7;
 * This is a type-conversion wrapper around `getHash()`
 * @return {string}
 */
proto.pfs.FileInfo.prototype.getHash_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getHash()));
};


/**
 * optional bytes hash = 7;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getHash()`
 * @return {!Uint8Array}
 */
proto.pfs.FileInfo.prototype.getHash_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getHash()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pfs.FileInfo} returns this
 */
proto.pfs.FileInfo.prototype.setHash = function(value) {
  return jspb.Message.setProto3BytesField(this, 7, value);
};





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
proto.pfs.ByteRange.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ByteRange.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ByteRange} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ByteRange.toObject = function(includeInstance, msg) {
  var f, obj = {
    lower: jspb.Message.getFieldWithDefault(msg, 1, 0),
    upper: jspb.Message.getFieldWithDefault(msg, 2, 0)
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
 * @return {!proto.pfs.ByteRange}
 */
proto.pfs.ByteRange.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ByteRange;
  return proto.pfs.ByteRange.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ByteRange} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ByteRange}
 */
proto.pfs.ByteRange.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setLower(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setUpper(value);
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
proto.pfs.ByteRange.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ByteRange.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ByteRange} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ByteRange.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getLower();
  if (f !== 0) {
    writer.writeUint64(
      1,
      f
    );
  }
  f = message.getUpper();
  if (f !== 0) {
    writer.writeUint64(
      2,
      f
    );
  }
};


/**
 * optional uint64 lower = 1;
 * @return {number}
 */
proto.pfs.ByteRange.prototype.getLower = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.ByteRange} returns this
 */
proto.pfs.ByteRange.prototype.setLower = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional uint64 upper = 2;
 * @return {number}
 */
proto.pfs.ByteRange.prototype.getUpper = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.ByteRange} returns this
 */
proto.pfs.ByteRange.prototype.setUpper = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};





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
proto.pfs.BlockRef.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.BlockRef.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.BlockRef} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.BlockRef.toObject = function(includeInstance, msg) {
  var f, obj = {
    block: (f = msg.getBlock()) && proto.pfs.Block.toObject(includeInstance, f),
    range: (f = msg.getRange()) && proto.pfs.ByteRange.toObject(includeInstance, f)
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
 * @return {!proto.pfs.BlockRef}
 */
proto.pfs.BlockRef.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.BlockRef;
  return proto.pfs.BlockRef.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.BlockRef} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.BlockRef}
 */
proto.pfs.BlockRef.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Block;
      reader.readMessage(value,proto.pfs.Block.deserializeBinaryFromReader);
      msg.setBlock(value);
      break;
    case 2:
      var value = new proto.pfs.ByteRange;
      reader.readMessage(value,proto.pfs.ByteRange.deserializeBinaryFromReader);
      msg.setRange(value);
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
proto.pfs.BlockRef.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.BlockRef.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.BlockRef} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.BlockRef.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBlock();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Block.serializeBinaryToWriter
    );
  }
  f = message.getRange();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.ByteRange.serializeBinaryToWriter
    );
  }
};


/**
 * optional Block block = 1;
 * @return {?proto.pfs.Block}
 */
proto.pfs.BlockRef.prototype.getBlock = function() {
  return /** @type{?proto.pfs.Block} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Block, 1));
};


/**
 * @param {?proto.pfs.Block|undefined} value
 * @return {!proto.pfs.BlockRef} returns this
*/
proto.pfs.BlockRef.prototype.setBlock = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.BlockRef} returns this
 */
proto.pfs.BlockRef.prototype.clearBlock = function() {
  return this.setBlock(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.BlockRef.prototype.hasBlock = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional ByteRange range = 2;
 * @return {?proto.pfs.ByteRange}
 */
proto.pfs.BlockRef.prototype.getRange = function() {
  return /** @type{?proto.pfs.ByteRange} */ (
    jspb.Message.getWrapperField(this, proto.pfs.ByteRange, 2));
};


/**
 * @param {?proto.pfs.ByteRange|undefined} value
 * @return {!proto.pfs.BlockRef} returns this
*/
proto.pfs.BlockRef.prototype.setRange = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.BlockRef} returns this
 */
proto.pfs.BlockRef.prototype.clearRange = function() {
  return this.setRange(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.BlockRef.prototype.hasRange = function() {
  return jspb.Message.getField(this, 2) != null;
};





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
proto.pfs.ObjectInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ObjectInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ObjectInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ObjectInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    object: (f = msg.getObject()) && proto.pfs.Object.toObject(includeInstance, f),
    blockRef: (f = msg.getBlockRef()) && proto.pfs.BlockRef.toObject(includeInstance, f)
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
 * @return {!proto.pfs.ObjectInfo}
 */
proto.pfs.ObjectInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ObjectInfo;
  return proto.pfs.ObjectInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ObjectInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ObjectInfo}
 */
proto.pfs.ObjectInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.setObject(value);
      break;
    case 2:
      var value = new proto.pfs.BlockRef;
      reader.readMessage(value,proto.pfs.BlockRef.deserializeBinaryFromReader);
      msg.setBlockRef(value);
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
proto.pfs.ObjectInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ObjectInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ObjectInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ObjectInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getObject();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getBlockRef();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.BlockRef.serializeBinaryToWriter
    );
  }
};


/**
 * optional Object object = 1;
 * @return {?proto.pfs.Object}
 */
proto.pfs.ObjectInfo.prototype.getObject = function() {
  return /** @type{?proto.pfs.Object} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Object, 1));
};


/**
 * @param {?proto.pfs.Object|undefined} value
 * @return {!proto.pfs.ObjectInfo} returns this
*/
proto.pfs.ObjectInfo.prototype.setObject = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.ObjectInfo} returns this
 */
proto.pfs.ObjectInfo.prototype.clearObject = function() {
  return this.setObject(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.ObjectInfo.prototype.hasObject = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional BlockRef block_ref = 2;
 * @return {?proto.pfs.BlockRef}
 */
proto.pfs.ObjectInfo.prototype.getBlockRef = function() {
  return /** @type{?proto.pfs.BlockRef} */ (
    jspb.Message.getWrapperField(this, proto.pfs.BlockRef, 2));
};


/**
 * @param {?proto.pfs.BlockRef|undefined} value
 * @return {!proto.pfs.ObjectInfo} returns this
*/
proto.pfs.ObjectInfo.prototype.setBlockRef = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.ObjectInfo} returns this
 */
proto.pfs.ObjectInfo.prototype.clearBlockRef = function() {
  return this.setBlockRef(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.ObjectInfo.prototype.hasBlockRef = function() {
  return jspb.Message.getField(this, 2) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.Compaction.repeatedFields_ = [2];



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
proto.pfs.Compaction.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.Compaction.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.Compaction} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Compaction.toObject = function(includeInstance, msg) {
  var f, obj = {
    inputPrefixesList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f
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
 * @return {!proto.pfs.Compaction}
 */
proto.pfs.Compaction.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.Compaction;
  return proto.pfs.Compaction.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.Compaction} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.Compaction}
 */
proto.pfs.Compaction.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addInputPrefixes(value);
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
proto.pfs.Compaction.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.Compaction.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.Compaction} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Compaction.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getInputPrefixesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
};


/**
 * repeated string input_prefixes = 2;
 * @return {!Array<string>}
 */
proto.pfs.Compaction.prototype.getInputPrefixesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pfs.Compaction} returns this
 */
proto.pfs.Compaction.prototype.setInputPrefixesList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pfs.Compaction} returns this
 */
proto.pfs.Compaction.prototype.addInputPrefixes = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.Compaction} returns this
 */
proto.pfs.Compaction.prototype.clearInputPrefixesList = function() {
  return this.setInputPrefixesList([]);
};





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
proto.pfs.Shard.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.Shard.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.Shard} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Shard.toObject = function(includeInstance, msg) {
  var f, obj = {
    compaction: (f = msg.getCompaction()) && proto.pfs.Compaction.toObject(includeInstance, f),
    range: (f = msg.getRange()) && proto.pfs.PathRange.toObject(includeInstance, f),
    outputPath: jspb.Message.getFieldWithDefault(msg, 3, "")
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
 * @return {!proto.pfs.Shard}
 */
proto.pfs.Shard.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.Shard;
  return proto.pfs.Shard.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.Shard} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.Shard}
 */
proto.pfs.Shard.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Compaction;
      reader.readMessage(value,proto.pfs.Compaction.deserializeBinaryFromReader);
      msg.setCompaction(value);
      break;
    case 2:
      var value = new proto.pfs.PathRange;
      reader.readMessage(value,proto.pfs.PathRange.deserializeBinaryFromReader);
      msg.setRange(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setOutputPath(value);
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
proto.pfs.Shard.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.Shard.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.Shard} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Shard.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCompaction();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Compaction.serializeBinaryToWriter
    );
  }
  f = message.getRange();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.PathRange.serializeBinaryToWriter
    );
  }
  f = message.getOutputPath();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional Compaction compaction = 1;
 * @return {?proto.pfs.Compaction}
 */
proto.pfs.Shard.prototype.getCompaction = function() {
  return /** @type{?proto.pfs.Compaction} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Compaction, 1));
};


/**
 * @param {?proto.pfs.Compaction|undefined} value
 * @return {!proto.pfs.Shard} returns this
*/
proto.pfs.Shard.prototype.setCompaction = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.Shard} returns this
 */
proto.pfs.Shard.prototype.clearCompaction = function() {
  return this.setCompaction(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.Shard.prototype.hasCompaction = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional PathRange range = 2;
 * @return {?proto.pfs.PathRange}
 */
proto.pfs.Shard.prototype.getRange = function() {
  return /** @type{?proto.pfs.PathRange} */ (
    jspb.Message.getWrapperField(this, proto.pfs.PathRange, 2));
};


/**
 * @param {?proto.pfs.PathRange|undefined} value
 * @return {!proto.pfs.Shard} returns this
*/
proto.pfs.Shard.prototype.setRange = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.Shard} returns this
 */
proto.pfs.Shard.prototype.clearRange = function() {
  return this.setRange(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.Shard.prototype.hasRange = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional string output_path = 3;
 * @return {string}
 */
proto.pfs.Shard.prototype.getOutputPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.Shard} returns this
 */
proto.pfs.Shard.prototype.setOutputPath = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};





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
proto.pfs.PathRange.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.PathRange.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.PathRange} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PathRange.toObject = function(includeInstance, msg) {
  var f, obj = {
    lower: jspb.Message.getFieldWithDefault(msg, 1, ""),
    upper: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pfs.PathRange}
 */
proto.pfs.PathRange.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.PathRange;
  return proto.pfs.PathRange.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.PathRange} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.PathRange}
 */
proto.pfs.PathRange.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setLower(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setUpper(value);
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
proto.pfs.PathRange.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.PathRange.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.PathRange} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PathRange.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getLower();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getUpper();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string lower = 1;
 * @return {string}
 */
proto.pfs.PathRange.prototype.getLower = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.PathRange} returns this
 */
proto.pfs.PathRange.prototype.setLower = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string upper = 2;
 * @return {string}
 */
proto.pfs.PathRange.prototype.getUpper = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.PathRange} returns this
 */
proto.pfs.PathRange.prototype.setUpper = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





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
proto.pfs.CreateRepoRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CreateRepoRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CreateRepoRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CreateRepoRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs.Repo.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 3, ""),
    update: jspb.Message.getBooleanFieldWithDefault(msg, 4, false)
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
 * @return {!proto.pfs.CreateRepoRequest}
 */
proto.pfs.CreateRepoRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CreateRepoRequest;
  return proto.pfs.CreateRepoRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CreateRepoRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CreateRepoRequest}
 */
proto.pfs.CreateRepoRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Repo;
      reader.readMessage(value,proto.pfs.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setUpdate(value);
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
proto.pfs.CreateRepoRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CreateRepoRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CreateRepoRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CreateRepoRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Repo.serializeBinaryToWriter
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getUpdate();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs.Repo}
 */
proto.pfs.CreateRepoRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Repo, 1));
};


/**
 * @param {?proto.pfs.Repo|undefined} value
 * @return {!proto.pfs.CreateRepoRequest} returns this
*/
proto.pfs.CreateRepoRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CreateRepoRequest} returns this
 */
proto.pfs.CreateRepoRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CreateRepoRequest.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string description = 3;
 * @return {string}
 */
proto.pfs.CreateRepoRequest.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.CreateRepoRequest} returns this
 */
proto.pfs.CreateRepoRequest.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional bool update = 4;
 * @return {boolean}
 */
proto.pfs.CreateRepoRequest.prototype.getUpdate = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.CreateRepoRequest} returns this
 */
proto.pfs.CreateRepoRequest.prototype.setUpdate = function(value) {
  return jspb.Message.setProto3BooleanField(this, 4, value);
};





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
proto.pfs.InspectRepoRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.InspectRepoRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.InspectRepoRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.InspectRepoRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs.Repo.toObject(includeInstance, f)
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
 * @return {!proto.pfs.InspectRepoRequest}
 */
proto.pfs.InspectRepoRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.InspectRepoRequest;
  return proto.pfs.InspectRepoRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.InspectRepoRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.InspectRepoRequest}
 */
proto.pfs.InspectRepoRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Repo;
      reader.readMessage(value,proto.pfs.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
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
proto.pfs.InspectRepoRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.InspectRepoRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.InspectRepoRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.InspectRepoRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Repo.serializeBinaryToWriter
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs.Repo}
 */
proto.pfs.InspectRepoRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Repo, 1));
};


/**
 * @param {?proto.pfs.Repo|undefined} value
 * @return {!proto.pfs.InspectRepoRequest} returns this
*/
proto.pfs.InspectRepoRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.InspectRepoRequest} returns this
 */
proto.pfs.InspectRepoRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.InspectRepoRequest.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};





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
proto.pfs.ListRepoRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ListRepoRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ListRepoRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListRepoRequest.toObject = function(includeInstance, msg) {
  var f, obj = {

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
 * @return {!proto.pfs.ListRepoRequest}
 */
proto.pfs.ListRepoRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ListRepoRequest;
  return proto.pfs.ListRepoRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ListRepoRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ListRepoRequest}
 */
proto.pfs.ListRepoRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
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
proto.pfs.ListRepoRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ListRepoRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ListRepoRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListRepoRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.ListRepoResponse.repeatedFields_ = [1];



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
proto.pfs.ListRepoResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ListRepoResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ListRepoResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListRepoResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    repoInfoList: jspb.Message.toObjectList(msg.getRepoInfoList(),
    proto.pfs.RepoInfo.toObject, includeInstance)
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
 * @return {!proto.pfs.ListRepoResponse}
 */
proto.pfs.ListRepoResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ListRepoResponse;
  return proto.pfs.ListRepoResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ListRepoResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ListRepoResponse}
 */
proto.pfs.ListRepoResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.RepoInfo;
      reader.readMessage(value,proto.pfs.RepoInfo.deserializeBinaryFromReader);
      msg.addRepoInfo(value);
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
proto.pfs.ListRepoResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ListRepoResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ListRepoResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListRepoResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepoInfoList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs.RepoInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated RepoInfo repo_info = 1;
 * @return {!Array<!proto.pfs.RepoInfo>}
 */
proto.pfs.ListRepoResponse.prototype.getRepoInfoList = function() {
  return /** @type{!Array<!proto.pfs.RepoInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.RepoInfo, 1));
};


/**
 * @param {!Array<!proto.pfs.RepoInfo>} value
 * @return {!proto.pfs.ListRepoResponse} returns this
*/
proto.pfs.ListRepoResponse.prototype.setRepoInfoList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.RepoInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.RepoInfo}
 */
proto.pfs.ListRepoResponse.prototype.addRepoInfo = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.RepoInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.ListRepoResponse} returns this
 */
proto.pfs.ListRepoResponse.prototype.clearRepoInfoList = function() {
  return this.setRepoInfoList([]);
};





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
proto.pfs.DeleteRepoRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DeleteRepoRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DeleteRepoRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteRepoRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs.Repo.toObject(includeInstance, f),
    force: jspb.Message.getBooleanFieldWithDefault(msg, 2, false),
    all: jspb.Message.getBooleanFieldWithDefault(msg, 3, false),
    splitTransaction: jspb.Message.getBooleanFieldWithDefault(msg, 4, false)
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
 * @return {!proto.pfs.DeleteRepoRequest}
 */
proto.pfs.DeleteRepoRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DeleteRepoRequest;
  return proto.pfs.DeleteRepoRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DeleteRepoRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DeleteRepoRequest}
 */
proto.pfs.DeleteRepoRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Repo;
      reader.readMessage(value,proto.pfs.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setForce(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAll(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSplitTransaction(value);
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
proto.pfs.DeleteRepoRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DeleteRepoRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DeleteRepoRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteRepoRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Repo.serializeBinaryToWriter
    );
  }
  f = message.getForce();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
  f = message.getAll();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
  f = message.getSplitTransaction();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs.Repo}
 */
proto.pfs.DeleteRepoRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Repo, 1));
};


/**
 * @param {?proto.pfs.Repo|undefined} value
 * @return {!proto.pfs.DeleteRepoRequest} returns this
*/
proto.pfs.DeleteRepoRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.DeleteRepoRequest} returns this
 */
proto.pfs.DeleteRepoRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.DeleteRepoRequest.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool force = 2;
 * @return {boolean}
 */
proto.pfs.DeleteRepoRequest.prototype.getForce = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.DeleteRepoRequest} returns this
 */
proto.pfs.DeleteRepoRequest.prototype.setForce = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};


/**
 * optional bool all = 3;
 * @return {boolean}
 */
proto.pfs.DeleteRepoRequest.prototype.getAll = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.DeleteRepoRequest} returns this
 */
proto.pfs.DeleteRepoRequest.prototype.setAll = function(value) {
  return jspb.Message.setProto3BooleanField(this, 3, value);
};


/**
 * optional bool split_transaction = 4;
 * @return {boolean}
 */
proto.pfs.DeleteRepoRequest.prototype.getSplitTransaction = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.DeleteRepoRequest} returns this
 */
proto.pfs.DeleteRepoRequest.prototype.setSplitTransaction = function(value) {
  return jspb.Message.setProto3BooleanField(this, 4, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.StartCommitRequest.repeatedFields_ = [5];



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
proto.pfs.StartCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.StartCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.StartCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.StartCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    parent: (f = msg.getParent()) && proto.pfs.Commit.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 4, ""),
    branch: jspb.Message.getFieldWithDefault(msg, 3, ""),
    provenanceList: jspb.Message.toObjectList(msg.getProvenanceList(),
    proto.pfs.CommitProvenance.toObject, includeInstance)
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
 * @return {!proto.pfs.StartCommitRequest}
 */
proto.pfs.StartCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.StartCommitRequest;
  return proto.pfs.StartCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.StartCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.StartCommitRequest}
 */
proto.pfs.StartCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setParent(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setBranch(value);
      break;
    case 5:
      var value = new proto.pfs.CommitProvenance;
      reader.readMessage(value,proto.pfs.CommitProvenance.deserializeBinaryFromReader);
      msg.addProvenance(value);
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
proto.pfs.StartCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.StartCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.StartCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.StartCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getParent();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getBranch();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      5,
      f,
      proto.pfs.CommitProvenance.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit parent = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.StartCommitRequest.prototype.getParent = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.StartCommitRequest} returns this
*/
proto.pfs.StartCommitRequest.prototype.setParent = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.StartCommitRequest} returns this
 */
proto.pfs.StartCommitRequest.prototype.clearParent = function() {
  return this.setParent(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.StartCommitRequest.prototype.hasParent = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string description = 4;
 * @return {string}
 */
proto.pfs.StartCommitRequest.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.StartCommitRequest} returns this
 */
proto.pfs.StartCommitRequest.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string branch = 3;
 * @return {string}
 */
proto.pfs.StartCommitRequest.prototype.getBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.StartCommitRequest} returns this
 */
proto.pfs.StartCommitRequest.prototype.setBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * repeated CommitProvenance provenance = 5;
 * @return {!Array<!proto.pfs.CommitProvenance>}
 */
proto.pfs.StartCommitRequest.prototype.getProvenanceList = function() {
  return /** @type{!Array<!proto.pfs.CommitProvenance>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.CommitProvenance, 5));
};


/**
 * @param {!Array<!proto.pfs.CommitProvenance>} value
 * @return {!proto.pfs.StartCommitRequest} returns this
*/
proto.pfs.StartCommitRequest.prototype.setProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.pfs.CommitProvenance=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.CommitProvenance}
 */
proto.pfs.StartCommitRequest.prototype.addProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.pfs.CommitProvenance, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.StartCommitRequest} returns this
 */
proto.pfs.StartCommitRequest.prototype.clearProvenanceList = function() {
  return this.setProvenanceList([]);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.BuildCommitRequest.repeatedFields_ = [6,7];



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
proto.pfs.BuildCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.BuildCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.BuildCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.BuildCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    parent: (f = msg.getParent()) && proto.pfs.Commit.toObject(includeInstance, f),
    branch: jspb.Message.getFieldWithDefault(msg, 4, ""),
    origin: (f = msg.getOrigin()) && proto.pfs.CommitOrigin.toObject(includeInstance, f),
    provenanceList: jspb.Message.toObjectList(msg.getProvenanceList(),
    proto.pfs.CommitProvenance.toObject, includeInstance),
    tree: (f = msg.getTree()) && proto.pfs.Object.toObject(includeInstance, f),
    treesList: jspb.Message.toObjectList(msg.getTreesList(),
    proto.pfs.Object.toObject, includeInstance),
    datums: (f = msg.getDatums()) && proto.pfs.Object.toObject(includeInstance, f),
    id: jspb.Message.getFieldWithDefault(msg, 5, ""),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 9, 0),
    started: (f = msg.getStarted()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    finished: (f = msg.getFinished()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f)
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
 * @return {!proto.pfs.BuildCommitRequest}
 */
proto.pfs.BuildCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.BuildCommitRequest;
  return proto.pfs.BuildCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.BuildCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.BuildCommitRequest}
 */
proto.pfs.BuildCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setParent(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setBranch(value);
      break;
    case 12:
      var value = new proto.pfs.CommitOrigin;
      reader.readMessage(value,proto.pfs.CommitOrigin.deserializeBinaryFromReader);
      msg.setOrigin(value);
      break;
    case 6:
      var value = new proto.pfs.CommitProvenance;
      reader.readMessage(value,proto.pfs.CommitProvenance.deserializeBinaryFromReader);
      msg.addProvenance(value);
      break;
    case 3:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.setTree(value);
      break;
    case 7:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.addTrees(value);
      break;
    case 8:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.setDatums(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setSizeBytes(value);
      break;
    case 10:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setStarted(value);
      break;
    case 11:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setFinished(value);
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
proto.pfs.BuildCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.BuildCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.BuildCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.BuildCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getParent();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getBranch();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getOrigin();
  if (f != null) {
    writer.writeMessage(
      12,
      f,
      proto.pfs.CommitOrigin.serializeBinaryToWriter
    );
  }
  f = message.getProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      6,
      f,
      proto.pfs.CommitProvenance.serializeBinaryToWriter
    );
  }
  f = message.getTree();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getTreesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      7,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getDatums();
  if (f != null) {
    writer.writeMessage(
      8,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getSizeBytes();
  if (f !== 0) {
    writer.writeUint64(
      9,
      f
    );
  }
  f = message.getStarted();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getFinished();
  if (f != null) {
    writer.writeMessage(
      11,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit parent = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.BuildCommitRequest.prototype.getParent = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.BuildCommitRequest} returns this
*/
proto.pfs.BuildCommitRequest.prototype.setParent = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.BuildCommitRequest} returns this
 */
proto.pfs.BuildCommitRequest.prototype.clearParent = function() {
  return this.setParent(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.BuildCommitRequest.prototype.hasParent = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string branch = 4;
 * @return {string}
 */
proto.pfs.BuildCommitRequest.prototype.getBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.BuildCommitRequest} returns this
 */
proto.pfs.BuildCommitRequest.prototype.setBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional CommitOrigin origin = 12;
 * @return {?proto.pfs.CommitOrigin}
 */
proto.pfs.BuildCommitRequest.prototype.getOrigin = function() {
  return /** @type{?proto.pfs.CommitOrigin} */ (
    jspb.Message.getWrapperField(this, proto.pfs.CommitOrigin, 12));
};


/**
 * @param {?proto.pfs.CommitOrigin|undefined} value
 * @return {!proto.pfs.BuildCommitRequest} returns this
*/
proto.pfs.BuildCommitRequest.prototype.setOrigin = function(value) {
  return jspb.Message.setWrapperField(this, 12, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.BuildCommitRequest} returns this
 */
proto.pfs.BuildCommitRequest.prototype.clearOrigin = function() {
  return this.setOrigin(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.BuildCommitRequest.prototype.hasOrigin = function() {
  return jspb.Message.getField(this, 12) != null;
};


/**
 * repeated CommitProvenance provenance = 6;
 * @return {!Array<!proto.pfs.CommitProvenance>}
 */
proto.pfs.BuildCommitRequest.prototype.getProvenanceList = function() {
  return /** @type{!Array<!proto.pfs.CommitProvenance>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.CommitProvenance, 6));
};


/**
 * @param {!Array<!proto.pfs.CommitProvenance>} value
 * @return {!proto.pfs.BuildCommitRequest} returns this
*/
proto.pfs.BuildCommitRequest.prototype.setProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 6, value);
};


/**
 * @param {!proto.pfs.CommitProvenance=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.CommitProvenance}
 */
proto.pfs.BuildCommitRequest.prototype.addProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 6, opt_value, proto.pfs.CommitProvenance, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.BuildCommitRequest} returns this
 */
proto.pfs.BuildCommitRequest.prototype.clearProvenanceList = function() {
  return this.setProvenanceList([]);
};


/**
 * optional Object tree = 3;
 * @return {?proto.pfs.Object}
 */
proto.pfs.BuildCommitRequest.prototype.getTree = function() {
  return /** @type{?proto.pfs.Object} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Object, 3));
};


/**
 * @param {?proto.pfs.Object|undefined} value
 * @return {!proto.pfs.BuildCommitRequest} returns this
*/
proto.pfs.BuildCommitRequest.prototype.setTree = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.BuildCommitRequest} returns this
 */
proto.pfs.BuildCommitRequest.prototype.clearTree = function() {
  return this.setTree(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.BuildCommitRequest.prototype.hasTree = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * repeated Object trees = 7;
 * @return {!Array<!proto.pfs.Object>}
 */
proto.pfs.BuildCommitRequest.prototype.getTreesList = function() {
  return /** @type{!Array<!proto.pfs.Object>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Object, 7));
};


/**
 * @param {!Array<!proto.pfs.Object>} value
 * @return {!proto.pfs.BuildCommitRequest} returns this
*/
proto.pfs.BuildCommitRequest.prototype.setTreesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 7, value);
};


/**
 * @param {!proto.pfs.Object=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Object}
 */
proto.pfs.BuildCommitRequest.prototype.addTrees = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 7, opt_value, proto.pfs.Object, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.BuildCommitRequest} returns this
 */
proto.pfs.BuildCommitRequest.prototype.clearTreesList = function() {
  return this.setTreesList([]);
};


/**
 * optional Object datums = 8;
 * @return {?proto.pfs.Object}
 */
proto.pfs.BuildCommitRequest.prototype.getDatums = function() {
  return /** @type{?proto.pfs.Object} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Object, 8));
};


/**
 * @param {?proto.pfs.Object|undefined} value
 * @return {!proto.pfs.BuildCommitRequest} returns this
*/
proto.pfs.BuildCommitRequest.prototype.setDatums = function(value) {
  return jspb.Message.setWrapperField(this, 8, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.BuildCommitRequest} returns this
 */
proto.pfs.BuildCommitRequest.prototype.clearDatums = function() {
  return this.setDatums(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.BuildCommitRequest.prototype.hasDatums = function() {
  return jspb.Message.getField(this, 8) != null;
};


/**
 * optional string ID = 5;
 * @return {string}
 */
proto.pfs.BuildCommitRequest.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.BuildCommitRequest} returns this
 */
proto.pfs.BuildCommitRequest.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional uint64 size_bytes = 9;
 * @return {number}
 */
proto.pfs.BuildCommitRequest.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.BuildCommitRequest} returns this
 */
proto.pfs.BuildCommitRequest.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 9, value);
};


/**
 * optional google.protobuf.Timestamp started = 10;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pfs.BuildCommitRequest.prototype.getStarted = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 10));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pfs.BuildCommitRequest} returns this
*/
proto.pfs.BuildCommitRequest.prototype.setStarted = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.BuildCommitRequest} returns this
 */
proto.pfs.BuildCommitRequest.prototype.clearStarted = function() {
  return this.setStarted(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.BuildCommitRequest.prototype.hasStarted = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * optional google.protobuf.Timestamp finished = 11;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pfs.BuildCommitRequest.prototype.getFinished = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 11));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pfs.BuildCommitRequest} returns this
*/
proto.pfs.BuildCommitRequest.prototype.setFinished = function(value) {
  return jspb.Message.setWrapperField(this, 11, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.BuildCommitRequest} returns this
 */
proto.pfs.BuildCommitRequest.prototype.clearFinished = function() {
  return this.setFinished(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.BuildCommitRequest.prototype.hasFinished = function() {
  return jspb.Message.getField(this, 11) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.FinishCommitRequest.repeatedFields_ = [5];



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
proto.pfs.FinishCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.FinishCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.FinishCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FinishCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs.Commit.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 2, ""),
    tree: (f = msg.getTree()) && proto.pfs.Object.toObject(includeInstance, f),
    treesList: jspb.Message.toObjectList(msg.getTreesList(),
    proto.pfs.Object.toObject, includeInstance),
    datums: (f = msg.getDatums()) && proto.pfs.Object.toObject(includeInstance, f),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 6, 0),
    empty: jspb.Message.getBooleanFieldWithDefault(msg, 4, false)
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
 * @return {!proto.pfs.FinishCommitRequest}
 */
proto.pfs.FinishCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.FinishCommitRequest;
  return proto.pfs.FinishCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.FinishCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.FinishCommitRequest}
 */
proto.pfs.FinishCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 3:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.setTree(value);
      break;
    case 5:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.addTrees(value);
      break;
    case 7:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.setDatums(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setSizeBytes(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEmpty(value);
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
proto.pfs.FinishCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.FinishCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.FinishCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FinishCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getTree();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getTreesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      5,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getDatums();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getSizeBytes();
  if (f !== 0) {
    writer.writeUint64(
      6,
      f
    );
  }
  f = message.getEmpty();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.FinishCommitRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.FinishCommitRequest} returns this
*/
proto.pfs.FinishCommitRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.FinishCommitRequest} returns this
 */
proto.pfs.FinishCommitRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.FinishCommitRequest.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string description = 2;
 * @return {string}
 */
proto.pfs.FinishCommitRequest.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.FinishCommitRequest} returns this
 */
proto.pfs.FinishCommitRequest.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional Object tree = 3;
 * @return {?proto.pfs.Object}
 */
proto.pfs.FinishCommitRequest.prototype.getTree = function() {
  return /** @type{?proto.pfs.Object} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Object, 3));
};


/**
 * @param {?proto.pfs.Object|undefined} value
 * @return {!proto.pfs.FinishCommitRequest} returns this
*/
proto.pfs.FinishCommitRequest.prototype.setTree = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.FinishCommitRequest} returns this
 */
proto.pfs.FinishCommitRequest.prototype.clearTree = function() {
  return this.setTree(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.FinishCommitRequest.prototype.hasTree = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * repeated Object trees = 5;
 * @return {!Array<!proto.pfs.Object>}
 */
proto.pfs.FinishCommitRequest.prototype.getTreesList = function() {
  return /** @type{!Array<!proto.pfs.Object>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Object, 5));
};


/**
 * @param {!Array<!proto.pfs.Object>} value
 * @return {!proto.pfs.FinishCommitRequest} returns this
*/
proto.pfs.FinishCommitRequest.prototype.setTreesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.pfs.Object=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Object}
 */
proto.pfs.FinishCommitRequest.prototype.addTrees = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.pfs.Object, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.FinishCommitRequest} returns this
 */
proto.pfs.FinishCommitRequest.prototype.clearTreesList = function() {
  return this.setTreesList([]);
};


/**
 * optional Object datums = 7;
 * @return {?proto.pfs.Object}
 */
proto.pfs.FinishCommitRequest.prototype.getDatums = function() {
  return /** @type{?proto.pfs.Object} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Object, 7));
};


/**
 * @param {?proto.pfs.Object|undefined} value
 * @return {!proto.pfs.FinishCommitRequest} returns this
*/
proto.pfs.FinishCommitRequest.prototype.setDatums = function(value) {
  return jspb.Message.setWrapperField(this, 7, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.FinishCommitRequest} returns this
 */
proto.pfs.FinishCommitRequest.prototype.clearDatums = function() {
  return this.setDatums(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.FinishCommitRequest.prototype.hasDatums = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * optional uint64 size_bytes = 6;
 * @return {number}
 */
proto.pfs.FinishCommitRequest.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.FinishCommitRequest} returns this
 */
proto.pfs.FinishCommitRequest.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 6, value);
};


/**
 * optional bool empty = 4;
 * @return {boolean}
 */
proto.pfs.FinishCommitRequest.prototype.getEmpty = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.FinishCommitRequest} returns this
 */
proto.pfs.FinishCommitRequest.prototype.setEmpty = function(value) {
  return jspb.Message.setProto3BooleanField(this, 4, value);
};





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
proto.pfs.InspectCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.InspectCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.InspectCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.InspectCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs.Commit.toObject(includeInstance, f),
    blockState: jspb.Message.getFieldWithDefault(msg, 2, 0)
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
 * @return {!proto.pfs.InspectCommitRequest}
 */
proto.pfs.InspectCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.InspectCommitRequest;
  return proto.pfs.InspectCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.InspectCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.InspectCommitRequest}
 */
proto.pfs.InspectCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = /** @type {!proto.pfs.CommitState} */ (reader.readEnum());
      msg.setBlockState(value);
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
proto.pfs.InspectCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.InspectCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.InspectCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.InspectCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getBlockState();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.InspectCommitRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.InspectCommitRequest} returns this
*/
proto.pfs.InspectCommitRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.InspectCommitRequest} returns this
 */
proto.pfs.InspectCommitRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.InspectCommitRequest.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional CommitState block_state = 2;
 * @return {!proto.pfs.CommitState}
 */
proto.pfs.InspectCommitRequest.prototype.getBlockState = function() {
  return /** @type {!proto.pfs.CommitState} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.pfs.CommitState} value
 * @return {!proto.pfs.InspectCommitRequest} returns this
 */
proto.pfs.InspectCommitRequest.prototype.setBlockState = function(value) {
  return jspb.Message.setProto3EnumField(this, 2, value);
};





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
proto.pfs.ListCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ListCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ListCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs.Repo.toObject(includeInstance, f),
    from: (f = msg.getFrom()) && proto.pfs.Commit.toObject(includeInstance, f),
    to: (f = msg.getTo()) && proto.pfs.Commit.toObject(includeInstance, f),
    number: jspb.Message.getFieldWithDefault(msg, 4, 0),
    reverse: jspb.Message.getBooleanFieldWithDefault(msg, 5, false)
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
 * @return {!proto.pfs.ListCommitRequest}
 */
proto.pfs.ListCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ListCommitRequest;
  return proto.pfs.ListCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ListCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ListCommitRequest}
 */
proto.pfs.ListCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Repo;
      reader.readMessage(value,proto.pfs.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
      break;
    case 2:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setFrom(value);
      break;
    case 3:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setTo(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setNumber(value);
      break;
    case 5:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setReverse(value);
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
proto.pfs.ListCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ListCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ListCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Repo.serializeBinaryToWriter
    );
  }
  f = message.getFrom();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getTo();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getNumber();
  if (f !== 0) {
    writer.writeUint64(
      4,
      f
    );
  }
  f = message.getReverse();
  if (f) {
    writer.writeBool(
      5,
      f
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs.Repo}
 */
proto.pfs.ListCommitRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Repo, 1));
};


/**
 * @param {?proto.pfs.Repo|undefined} value
 * @return {!proto.pfs.ListCommitRequest} returns this
*/
proto.pfs.ListCommitRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.ListCommitRequest} returns this
 */
proto.pfs.ListCommitRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.ListCommitRequest.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Commit from = 2;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.ListCommitRequest.prototype.getFrom = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 2));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.ListCommitRequest} returns this
*/
proto.pfs.ListCommitRequest.prototype.setFrom = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.ListCommitRequest} returns this
 */
proto.pfs.ListCommitRequest.prototype.clearFrom = function() {
  return this.setFrom(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.ListCommitRequest.prototype.hasFrom = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional Commit to = 3;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.ListCommitRequest.prototype.getTo = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 3));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.ListCommitRequest} returns this
*/
proto.pfs.ListCommitRequest.prototype.setTo = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.ListCommitRequest} returns this
 */
proto.pfs.ListCommitRequest.prototype.clearTo = function() {
  return this.setTo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.ListCommitRequest.prototype.hasTo = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional uint64 number = 4;
 * @return {number}
 */
proto.pfs.ListCommitRequest.prototype.getNumber = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.ListCommitRequest} returns this
 */
proto.pfs.ListCommitRequest.prototype.setNumber = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional bool reverse = 5;
 * @return {boolean}
 */
proto.pfs.ListCommitRequest.prototype.getReverse = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 5, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.ListCommitRequest} returns this
 */
proto.pfs.ListCommitRequest.prototype.setReverse = function(value) {
  return jspb.Message.setProto3BooleanField(this, 5, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.CommitInfos.repeatedFields_ = [1];



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
proto.pfs.CommitInfos.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CommitInfos.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CommitInfos} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CommitInfos.toObject = function(includeInstance, msg) {
  var f, obj = {
    commitInfoList: jspb.Message.toObjectList(msg.getCommitInfoList(),
    proto.pfs.CommitInfo.toObject, includeInstance)
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
 * @return {!proto.pfs.CommitInfos}
 */
proto.pfs.CommitInfos.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CommitInfos;
  return proto.pfs.CommitInfos.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CommitInfos} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CommitInfos}
 */
proto.pfs.CommitInfos.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.CommitInfo;
      reader.readMessage(value,proto.pfs.CommitInfo.deserializeBinaryFromReader);
      msg.addCommitInfo(value);
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
proto.pfs.CommitInfos.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CommitInfos.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CommitInfos} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CommitInfos.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommitInfoList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs.CommitInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated CommitInfo commit_info = 1;
 * @return {!Array<!proto.pfs.CommitInfo>}
 */
proto.pfs.CommitInfos.prototype.getCommitInfoList = function() {
  return /** @type{!Array<!proto.pfs.CommitInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.CommitInfo, 1));
};


/**
 * @param {!Array<!proto.pfs.CommitInfo>} value
 * @return {!proto.pfs.CommitInfos} returns this
*/
proto.pfs.CommitInfos.prototype.setCommitInfoList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.CommitInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.CommitInfo}
 */
proto.pfs.CommitInfos.prototype.addCommitInfo = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.CommitInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.CommitInfos} returns this
 */
proto.pfs.CommitInfos.prototype.clearCommitInfoList = function() {
  return this.setCommitInfoList([]);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.CreateBranchRequest.repeatedFields_ = [4];



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
proto.pfs.CreateBranchRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CreateBranchRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CreateBranchRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CreateBranchRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    head: (f = msg.getHead()) && proto.pfs.Commit.toObject(includeInstance, f),
    sBranch: jspb.Message.getFieldWithDefault(msg, 2, ""),
    branch: (f = msg.getBranch()) && proto.pfs.Branch.toObject(includeInstance, f),
    provenanceList: jspb.Message.toObjectList(msg.getProvenanceList(),
    proto.pfs.Branch.toObject, includeInstance),
    trigger: (f = msg.getTrigger()) && proto.pfs.Trigger.toObject(includeInstance, f)
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
 * @return {!proto.pfs.CreateBranchRequest}
 */
proto.pfs.CreateBranchRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CreateBranchRequest;
  return proto.pfs.CreateBranchRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CreateBranchRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CreateBranchRequest}
 */
proto.pfs.CreateBranchRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setHead(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setSBranch(value);
      break;
    case 3:
      var value = new proto.pfs.Branch;
      reader.readMessage(value,proto.pfs.Branch.deserializeBinaryFromReader);
      msg.setBranch(value);
      break;
    case 4:
      var value = new proto.pfs.Branch;
      reader.readMessage(value,proto.pfs.Branch.deserializeBinaryFromReader);
      msg.addProvenance(value);
      break;
    case 5:
      var value = new proto.pfs.Trigger;
      reader.readMessage(value,proto.pfs.Trigger.deserializeBinaryFromReader);
      msg.setTrigger(value);
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
proto.pfs.CreateBranchRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CreateBranchRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CreateBranchRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CreateBranchRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getHead();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getSBranch();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pfs.Branch.serializeBinaryToWriter
    );
  }
  f = message.getProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      4,
      f,
      proto.pfs.Branch.serializeBinaryToWriter
    );
  }
  f = message.getTrigger();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.pfs.Trigger.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit head = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.CreateBranchRequest.prototype.getHead = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.CreateBranchRequest} returns this
*/
proto.pfs.CreateBranchRequest.prototype.setHead = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CreateBranchRequest} returns this
 */
proto.pfs.CreateBranchRequest.prototype.clearHead = function() {
  return this.setHead(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CreateBranchRequest.prototype.hasHead = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string s_branch = 2;
 * @return {string}
 */
proto.pfs.CreateBranchRequest.prototype.getSBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.CreateBranchRequest} returns this
 */
proto.pfs.CreateBranchRequest.prototype.setSBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional Branch branch = 3;
 * @return {?proto.pfs.Branch}
 */
proto.pfs.CreateBranchRequest.prototype.getBranch = function() {
  return /** @type{?proto.pfs.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Branch, 3));
};


/**
 * @param {?proto.pfs.Branch|undefined} value
 * @return {!proto.pfs.CreateBranchRequest} returns this
*/
proto.pfs.CreateBranchRequest.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CreateBranchRequest} returns this
 */
proto.pfs.CreateBranchRequest.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CreateBranchRequest.prototype.hasBranch = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * repeated Branch provenance = 4;
 * @return {!Array<!proto.pfs.Branch>}
 */
proto.pfs.CreateBranchRequest.prototype.getProvenanceList = function() {
  return /** @type{!Array<!proto.pfs.Branch>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Branch, 4));
};


/**
 * @param {!Array<!proto.pfs.Branch>} value
 * @return {!proto.pfs.CreateBranchRequest} returns this
*/
proto.pfs.CreateBranchRequest.prototype.setProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 4, value);
};


/**
 * @param {!proto.pfs.Branch=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Branch}
 */
proto.pfs.CreateBranchRequest.prototype.addProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 4, opt_value, proto.pfs.Branch, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.CreateBranchRequest} returns this
 */
proto.pfs.CreateBranchRequest.prototype.clearProvenanceList = function() {
  return this.setProvenanceList([]);
};


/**
 * optional Trigger trigger = 5;
 * @return {?proto.pfs.Trigger}
 */
proto.pfs.CreateBranchRequest.prototype.getTrigger = function() {
  return /** @type{?proto.pfs.Trigger} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Trigger, 5));
};


/**
 * @param {?proto.pfs.Trigger|undefined} value
 * @return {!proto.pfs.CreateBranchRequest} returns this
*/
proto.pfs.CreateBranchRequest.prototype.setTrigger = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CreateBranchRequest} returns this
 */
proto.pfs.CreateBranchRequest.prototype.clearTrigger = function() {
  return this.setTrigger(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CreateBranchRequest.prototype.hasTrigger = function() {
  return jspb.Message.getField(this, 5) != null;
};





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
proto.pfs.InspectBranchRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.InspectBranchRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.InspectBranchRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.InspectBranchRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    branch: (f = msg.getBranch()) && proto.pfs.Branch.toObject(includeInstance, f)
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
 * @return {!proto.pfs.InspectBranchRequest}
 */
proto.pfs.InspectBranchRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.InspectBranchRequest;
  return proto.pfs.InspectBranchRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.InspectBranchRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.InspectBranchRequest}
 */
proto.pfs.InspectBranchRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Branch;
      reader.readMessage(value,proto.pfs.Branch.deserializeBinaryFromReader);
      msg.setBranch(value);
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
proto.pfs.InspectBranchRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.InspectBranchRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.InspectBranchRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.InspectBranchRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Branch.serializeBinaryToWriter
    );
  }
};


/**
 * optional Branch branch = 1;
 * @return {?proto.pfs.Branch}
 */
proto.pfs.InspectBranchRequest.prototype.getBranch = function() {
  return /** @type{?proto.pfs.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Branch, 1));
};


/**
 * @param {?proto.pfs.Branch|undefined} value
 * @return {!proto.pfs.InspectBranchRequest} returns this
*/
proto.pfs.InspectBranchRequest.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.InspectBranchRequest} returns this
 */
proto.pfs.InspectBranchRequest.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.InspectBranchRequest.prototype.hasBranch = function() {
  return jspb.Message.getField(this, 1) != null;
};





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
proto.pfs.ListBranchRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ListBranchRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ListBranchRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListBranchRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs.Repo.toObject(includeInstance, f),
    reverse: jspb.Message.getBooleanFieldWithDefault(msg, 2, false)
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
 * @return {!proto.pfs.ListBranchRequest}
 */
proto.pfs.ListBranchRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ListBranchRequest;
  return proto.pfs.ListBranchRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ListBranchRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ListBranchRequest}
 */
proto.pfs.ListBranchRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Repo;
      reader.readMessage(value,proto.pfs.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setReverse(value);
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
proto.pfs.ListBranchRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ListBranchRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ListBranchRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListBranchRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Repo.serializeBinaryToWriter
    );
  }
  f = message.getReverse();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs.Repo}
 */
proto.pfs.ListBranchRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Repo, 1));
};


/**
 * @param {?proto.pfs.Repo|undefined} value
 * @return {!proto.pfs.ListBranchRequest} returns this
*/
proto.pfs.ListBranchRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.ListBranchRequest} returns this
 */
proto.pfs.ListBranchRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.ListBranchRequest.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool reverse = 2;
 * @return {boolean}
 */
proto.pfs.ListBranchRequest.prototype.getReverse = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.ListBranchRequest} returns this
 */
proto.pfs.ListBranchRequest.prototype.setReverse = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};





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
proto.pfs.DeleteBranchRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DeleteBranchRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DeleteBranchRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteBranchRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    branch: (f = msg.getBranch()) && proto.pfs.Branch.toObject(includeInstance, f),
    force: jspb.Message.getBooleanFieldWithDefault(msg, 2, false)
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
 * @return {!proto.pfs.DeleteBranchRequest}
 */
proto.pfs.DeleteBranchRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DeleteBranchRequest;
  return proto.pfs.DeleteBranchRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DeleteBranchRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DeleteBranchRequest}
 */
proto.pfs.DeleteBranchRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Branch;
      reader.readMessage(value,proto.pfs.Branch.deserializeBinaryFromReader);
      msg.setBranch(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setForce(value);
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
proto.pfs.DeleteBranchRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DeleteBranchRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DeleteBranchRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteBranchRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Branch.serializeBinaryToWriter
    );
  }
  f = message.getForce();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
};


/**
 * optional Branch branch = 1;
 * @return {?proto.pfs.Branch}
 */
proto.pfs.DeleteBranchRequest.prototype.getBranch = function() {
  return /** @type{?proto.pfs.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Branch, 1));
};


/**
 * @param {?proto.pfs.Branch|undefined} value
 * @return {!proto.pfs.DeleteBranchRequest} returns this
*/
proto.pfs.DeleteBranchRequest.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.DeleteBranchRequest} returns this
 */
proto.pfs.DeleteBranchRequest.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.DeleteBranchRequest.prototype.hasBranch = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool force = 2;
 * @return {boolean}
 */
proto.pfs.DeleteBranchRequest.prototype.getForce = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.DeleteBranchRequest} returns this
 */
proto.pfs.DeleteBranchRequest.prototype.setForce = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};





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
proto.pfs.DeleteCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DeleteCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DeleteCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs.Commit.toObject(includeInstance, f)
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
 * @return {!proto.pfs.DeleteCommitRequest}
 */
proto.pfs.DeleteCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DeleteCommitRequest;
  return proto.pfs.DeleteCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DeleteCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DeleteCommitRequest}
 */
proto.pfs.DeleteCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
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
proto.pfs.DeleteCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DeleteCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DeleteCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.DeleteCommitRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.DeleteCommitRequest} returns this
*/
proto.pfs.DeleteCommitRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.DeleteCommitRequest} returns this
 */
proto.pfs.DeleteCommitRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.DeleteCommitRequest.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.FlushCommitRequest.repeatedFields_ = [1,2];



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
proto.pfs.FlushCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.FlushCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.FlushCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FlushCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commitsList: jspb.Message.toObjectList(msg.getCommitsList(),
    proto.pfs.Commit.toObject, includeInstance),
    toReposList: jspb.Message.toObjectList(msg.getToReposList(),
    proto.pfs.Repo.toObject, includeInstance)
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
 * @return {!proto.pfs.FlushCommitRequest}
 */
proto.pfs.FlushCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.FlushCommitRequest;
  return proto.pfs.FlushCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.FlushCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.FlushCommitRequest}
 */
proto.pfs.FlushCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.addCommits(value);
      break;
    case 2:
      var value = new proto.pfs.Repo;
      reader.readMessage(value,proto.pfs.Repo.deserializeBinaryFromReader);
      msg.addToRepos(value);
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
proto.pfs.FlushCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.FlushCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.FlushCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FlushCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommitsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getToReposList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.pfs.Repo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Commit commits = 1;
 * @return {!Array<!proto.pfs.Commit>}
 */
proto.pfs.FlushCommitRequest.prototype.getCommitsList = function() {
  return /** @type{!Array<!proto.pfs.Commit>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {!Array<!proto.pfs.Commit>} value
 * @return {!proto.pfs.FlushCommitRequest} returns this
*/
proto.pfs.FlushCommitRequest.prototype.setCommitsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.Commit=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Commit}
 */
proto.pfs.FlushCommitRequest.prototype.addCommits = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.Commit, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.FlushCommitRequest} returns this
 */
proto.pfs.FlushCommitRequest.prototype.clearCommitsList = function() {
  return this.setCommitsList([]);
};


/**
 * repeated Repo to_repos = 2;
 * @return {!Array<!proto.pfs.Repo>}
 */
proto.pfs.FlushCommitRequest.prototype.getToReposList = function() {
  return /** @type{!Array<!proto.pfs.Repo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Repo, 2));
};


/**
 * @param {!Array<!proto.pfs.Repo>} value
 * @return {!proto.pfs.FlushCommitRequest} returns this
*/
proto.pfs.FlushCommitRequest.prototype.setToReposList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pfs.Repo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Repo}
 */
proto.pfs.FlushCommitRequest.prototype.addToRepos = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pfs.Repo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.FlushCommitRequest} returns this
 */
proto.pfs.FlushCommitRequest.prototype.clearToReposList = function() {
  return this.setToReposList([]);
};





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
proto.pfs.SubscribeCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.SubscribeCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.SubscribeCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.SubscribeCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs.Repo.toObject(includeInstance, f),
    branch: jspb.Message.getFieldWithDefault(msg, 2, ""),
    prov: (f = msg.getProv()) && proto.pfs.CommitProvenance.toObject(includeInstance, f),
    from: (f = msg.getFrom()) && proto.pfs.Commit.toObject(includeInstance, f),
    state: jspb.Message.getFieldWithDefault(msg, 4, 0)
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
 * @return {!proto.pfs.SubscribeCommitRequest}
 */
proto.pfs.SubscribeCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.SubscribeCommitRequest;
  return proto.pfs.SubscribeCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.SubscribeCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.SubscribeCommitRequest}
 */
proto.pfs.SubscribeCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Repo;
      reader.readMessage(value,proto.pfs.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setBranch(value);
      break;
    case 5:
      var value = new proto.pfs.CommitProvenance;
      reader.readMessage(value,proto.pfs.CommitProvenance.deserializeBinaryFromReader);
      msg.setProv(value);
      break;
    case 3:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setFrom(value);
      break;
    case 4:
      var value = /** @type {!proto.pfs.CommitState} */ (reader.readEnum());
      msg.setState(value);
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
proto.pfs.SubscribeCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.SubscribeCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.SubscribeCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.SubscribeCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Repo.serializeBinaryToWriter
    );
  }
  f = message.getBranch();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getProv();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.pfs.CommitProvenance.serializeBinaryToWriter
    );
  }
  f = message.getFrom();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      4,
      f
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs.Repo}
 */
proto.pfs.SubscribeCommitRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Repo, 1));
};


/**
 * @param {?proto.pfs.Repo|undefined} value
 * @return {!proto.pfs.SubscribeCommitRequest} returns this
*/
proto.pfs.SubscribeCommitRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.SubscribeCommitRequest} returns this
 */
proto.pfs.SubscribeCommitRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.SubscribeCommitRequest.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string branch = 2;
 * @return {string}
 */
proto.pfs.SubscribeCommitRequest.prototype.getBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.SubscribeCommitRequest} returns this
 */
proto.pfs.SubscribeCommitRequest.prototype.setBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional CommitProvenance prov = 5;
 * @return {?proto.pfs.CommitProvenance}
 */
proto.pfs.SubscribeCommitRequest.prototype.getProv = function() {
  return /** @type{?proto.pfs.CommitProvenance} */ (
    jspb.Message.getWrapperField(this, proto.pfs.CommitProvenance, 5));
};


/**
 * @param {?proto.pfs.CommitProvenance|undefined} value
 * @return {!proto.pfs.SubscribeCommitRequest} returns this
*/
proto.pfs.SubscribeCommitRequest.prototype.setProv = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.SubscribeCommitRequest} returns this
 */
proto.pfs.SubscribeCommitRequest.prototype.clearProv = function() {
  return this.setProv(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.SubscribeCommitRequest.prototype.hasProv = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional Commit from = 3;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.SubscribeCommitRequest.prototype.getFrom = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 3));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.SubscribeCommitRequest} returns this
*/
proto.pfs.SubscribeCommitRequest.prototype.setFrom = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.SubscribeCommitRequest} returns this
 */
proto.pfs.SubscribeCommitRequest.prototype.clearFrom = function() {
  return this.setFrom(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.SubscribeCommitRequest.prototype.hasFrom = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional CommitState state = 4;
 * @return {!proto.pfs.CommitState}
 */
proto.pfs.SubscribeCommitRequest.prototype.getState = function() {
  return /** @type {!proto.pfs.CommitState} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {!proto.pfs.CommitState} value
 * @return {!proto.pfs.SubscribeCommitRequest} returns this
 */
proto.pfs.SubscribeCommitRequest.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 4, value);
};





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
proto.pfs.GetFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.GetFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.GetFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs.File.toObject(includeInstance, f),
    offsetBytes: jspb.Message.getFieldWithDefault(msg, 2, 0),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 3, 0)
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
 * @return {!proto.pfs.GetFileRequest}
 */
proto.pfs.GetFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.GetFileRequest;
  return proto.pfs.GetFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.GetFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.GetFileRequest}
 */
proto.pfs.GetFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setFile(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setOffsetBytes(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setSizeBytes(value);
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
proto.pfs.GetFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.GetFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.GetFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
  f = message.getOffsetBytes();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
  f = message.getSizeBytes();
  if (f !== 0) {
    writer.writeInt64(
      3,
      f
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs.File}
 */
proto.pfs.GetFileRequest.prototype.getFile = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 1));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.GetFileRequest} returns this
*/
proto.pfs.GetFileRequest.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.GetFileRequest} returns this
 */
proto.pfs.GetFileRequest.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.GetFileRequest.prototype.hasFile = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional int64 offset_bytes = 2;
 * @return {number}
 */
proto.pfs.GetFileRequest.prototype.getOffsetBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.GetFileRequest} returns this
 */
proto.pfs.GetFileRequest.prototype.setOffsetBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional int64 size_bytes = 3;
 * @return {number}
 */
proto.pfs.GetFileRequest.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.GetFileRequest} returns this
 */
proto.pfs.GetFileRequest.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};





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
proto.pfs.OverwriteIndex.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.OverwriteIndex.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.OverwriteIndex} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.OverwriteIndex.toObject = function(includeInstance, msg) {
  var f, obj = {
    index: jspb.Message.getFieldWithDefault(msg, 1, 0)
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
 * @return {!proto.pfs.OverwriteIndex}
 */
proto.pfs.OverwriteIndex.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.OverwriteIndex;
  return proto.pfs.OverwriteIndex.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.OverwriteIndex} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.OverwriteIndex}
 */
proto.pfs.OverwriteIndex.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setIndex(value);
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
proto.pfs.OverwriteIndex.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.OverwriteIndex.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.OverwriteIndex} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.OverwriteIndex.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getIndex();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
};


/**
 * optional int64 index = 1;
 * @return {number}
 */
proto.pfs.OverwriteIndex.prototype.getIndex = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.OverwriteIndex} returns this
 */
proto.pfs.OverwriteIndex.prototype.setIndex = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};





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
proto.pfs.PutFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.PutFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.PutFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs.File.toObject(includeInstance, f),
    value: msg.getValue_asB64(),
    url: jspb.Message.getFieldWithDefault(msg, 5, ""),
    recursive: jspb.Message.getBooleanFieldWithDefault(msg, 6, false),
    delimiter: jspb.Message.getFieldWithDefault(msg, 7, 0),
    targetFileDatums: jspb.Message.getFieldWithDefault(msg, 8, 0),
    targetFileBytes: jspb.Message.getFieldWithDefault(msg, 9, 0),
    headerRecords: jspb.Message.getFieldWithDefault(msg, 11, 0),
    overwriteIndex: (f = msg.getOverwriteIndex()) && proto.pfs.OverwriteIndex.toObject(includeInstance, f),
    pb_delete: jspb.Message.getBooleanFieldWithDefault(msg, 12, false)
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
 * @return {!proto.pfs.PutFileRequest}
 */
proto.pfs.PutFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.PutFileRequest;
  return proto.pfs.PutFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.PutFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.PutFileRequest}
 */
proto.pfs.PutFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setFile(value);
      break;
    case 3:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setValue(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setUrl(value);
      break;
    case 6:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setRecursive(value);
      break;
    case 7:
      var value = /** @type {!proto.pfs.Delimiter} */ (reader.readEnum());
      msg.setDelimiter(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTargetFileDatums(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTargetFileBytes(value);
      break;
    case 11:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setHeaderRecords(value);
      break;
    case 10:
      var value = new proto.pfs.OverwriteIndex;
      reader.readMessage(value,proto.pfs.OverwriteIndex.deserializeBinaryFromReader);
      msg.setOverwriteIndex(value);
      break;
    case 12:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDelete(value);
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
proto.pfs.PutFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.PutFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.PutFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
  f = message.getValue_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      3,
      f
    );
  }
  f = message.getUrl();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getRecursive();
  if (f) {
    writer.writeBool(
      6,
      f
    );
  }
  f = message.getDelimiter();
  if (f !== 0.0) {
    writer.writeEnum(
      7,
      f
    );
  }
  f = message.getTargetFileDatums();
  if (f !== 0) {
    writer.writeInt64(
      8,
      f
    );
  }
  f = message.getTargetFileBytes();
  if (f !== 0) {
    writer.writeInt64(
      9,
      f
    );
  }
  f = message.getHeaderRecords();
  if (f !== 0) {
    writer.writeInt64(
      11,
      f
    );
  }
  f = message.getOverwriteIndex();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      proto.pfs.OverwriteIndex.serializeBinaryToWriter
    );
  }
  f = message.getDelete();
  if (f) {
    writer.writeBool(
      12,
      f
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs.File}
 */
proto.pfs.PutFileRequest.prototype.getFile = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 1));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.PutFileRequest} returns this
*/
proto.pfs.PutFileRequest.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.PutFileRequest} returns this
 */
proto.pfs.PutFileRequest.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.PutFileRequest.prototype.hasFile = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bytes value = 3;
 * @return {string}
 */
proto.pfs.PutFileRequest.prototype.getValue = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * optional bytes value = 3;
 * This is a type-conversion wrapper around `getValue()`
 * @return {string}
 */
proto.pfs.PutFileRequest.prototype.getValue_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getValue()));
};


/**
 * optional bytes value = 3;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getValue()`
 * @return {!Uint8Array}
 */
proto.pfs.PutFileRequest.prototype.getValue_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getValue()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pfs.PutFileRequest} returns this
 */
proto.pfs.PutFileRequest.prototype.setValue = function(value) {
  return jspb.Message.setProto3BytesField(this, 3, value);
};


/**
 * optional string url = 5;
 * @return {string}
 */
proto.pfs.PutFileRequest.prototype.getUrl = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.PutFileRequest} returns this
 */
proto.pfs.PutFileRequest.prototype.setUrl = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional bool recursive = 6;
 * @return {boolean}
 */
proto.pfs.PutFileRequest.prototype.getRecursive = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 6, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.PutFileRequest} returns this
 */
proto.pfs.PutFileRequest.prototype.setRecursive = function(value) {
  return jspb.Message.setProto3BooleanField(this, 6, value);
};


/**
 * optional Delimiter delimiter = 7;
 * @return {!proto.pfs.Delimiter}
 */
proto.pfs.PutFileRequest.prototype.getDelimiter = function() {
  return /** @type {!proto.pfs.Delimiter} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {!proto.pfs.Delimiter} value
 * @return {!proto.pfs.PutFileRequest} returns this
 */
proto.pfs.PutFileRequest.prototype.setDelimiter = function(value) {
  return jspb.Message.setProto3EnumField(this, 7, value);
};


/**
 * optional int64 target_file_datums = 8;
 * @return {number}
 */
proto.pfs.PutFileRequest.prototype.getTargetFileDatums = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.PutFileRequest} returns this
 */
proto.pfs.PutFileRequest.prototype.setTargetFileDatums = function(value) {
  return jspb.Message.setProto3IntField(this, 8, value);
};


/**
 * optional int64 target_file_bytes = 9;
 * @return {number}
 */
proto.pfs.PutFileRequest.prototype.getTargetFileBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.PutFileRequest} returns this
 */
proto.pfs.PutFileRequest.prototype.setTargetFileBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 9, value);
};


/**
 * optional int64 header_records = 11;
 * @return {number}
 */
proto.pfs.PutFileRequest.prototype.getHeaderRecords = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 11, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.PutFileRequest} returns this
 */
proto.pfs.PutFileRequest.prototype.setHeaderRecords = function(value) {
  return jspb.Message.setProto3IntField(this, 11, value);
};


/**
 * optional OverwriteIndex overwrite_index = 10;
 * @return {?proto.pfs.OverwriteIndex}
 */
proto.pfs.PutFileRequest.prototype.getOverwriteIndex = function() {
  return /** @type{?proto.pfs.OverwriteIndex} */ (
    jspb.Message.getWrapperField(this, proto.pfs.OverwriteIndex, 10));
};


/**
 * @param {?proto.pfs.OverwriteIndex|undefined} value
 * @return {!proto.pfs.PutFileRequest} returns this
*/
proto.pfs.PutFileRequest.prototype.setOverwriteIndex = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.PutFileRequest} returns this
 */
proto.pfs.PutFileRequest.prototype.clearOverwriteIndex = function() {
  return this.setOverwriteIndex(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.PutFileRequest.prototype.hasOverwriteIndex = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * optional bool delete = 12;
 * @return {boolean}
 */
proto.pfs.PutFileRequest.prototype.getDelete = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 12, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.PutFileRequest} returns this
 */
proto.pfs.PutFileRequest.prototype.setDelete = function(value) {
  return jspb.Message.setProto3BooleanField(this, 12, value);
};





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
proto.pfs.PutFileRecord.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.PutFileRecord.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.PutFileRecord} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutFileRecord.toObject = function(includeInstance, msg) {
  var f, obj = {
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 1, 0),
    objectHash: jspb.Message.getFieldWithDefault(msg, 2, ""),
    overwriteIndex: (f = msg.getOverwriteIndex()) && proto.pfs.OverwriteIndex.toObject(includeInstance, f),
    blockRef: (f = msg.getBlockRef()) && proto.pfs.BlockRef.toObject(includeInstance, f)
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
 * @return {!proto.pfs.PutFileRecord}
 */
proto.pfs.PutFileRecord.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.PutFileRecord;
  return proto.pfs.PutFileRecord.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.PutFileRecord} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.PutFileRecord}
 */
proto.pfs.PutFileRecord.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setSizeBytes(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setObjectHash(value);
      break;
    case 3:
      var value = new proto.pfs.OverwriteIndex;
      reader.readMessage(value,proto.pfs.OverwriteIndex.deserializeBinaryFromReader);
      msg.setOverwriteIndex(value);
      break;
    case 4:
      var value = new proto.pfs.BlockRef;
      reader.readMessage(value,proto.pfs.BlockRef.deserializeBinaryFromReader);
      msg.setBlockRef(value);
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
proto.pfs.PutFileRecord.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.PutFileRecord.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.PutFileRecord} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutFileRecord.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSizeBytes();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getObjectHash();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getOverwriteIndex();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pfs.OverwriteIndex.serializeBinaryToWriter
    );
  }
  f = message.getBlockRef();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pfs.BlockRef.serializeBinaryToWriter
    );
  }
};


/**
 * optional int64 size_bytes = 1;
 * @return {number}
 */
proto.pfs.PutFileRecord.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.PutFileRecord} returns this
 */
proto.pfs.PutFileRecord.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string object_hash = 2;
 * @return {string}
 */
proto.pfs.PutFileRecord.prototype.getObjectHash = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.PutFileRecord} returns this
 */
proto.pfs.PutFileRecord.prototype.setObjectHash = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional OverwriteIndex overwrite_index = 3;
 * @return {?proto.pfs.OverwriteIndex}
 */
proto.pfs.PutFileRecord.prototype.getOverwriteIndex = function() {
  return /** @type{?proto.pfs.OverwriteIndex} */ (
    jspb.Message.getWrapperField(this, proto.pfs.OverwriteIndex, 3));
};


/**
 * @param {?proto.pfs.OverwriteIndex|undefined} value
 * @return {!proto.pfs.PutFileRecord} returns this
*/
proto.pfs.PutFileRecord.prototype.setOverwriteIndex = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.PutFileRecord} returns this
 */
proto.pfs.PutFileRecord.prototype.clearOverwriteIndex = function() {
  return this.setOverwriteIndex(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.PutFileRecord.prototype.hasOverwriteIndex = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional BlockRef block_ref = 4;
 * @return {?proto.pfs.BlockRef}
 */
proto.pfs.PutFileRecord.prototype.getBlockRef = function() {
  return /** @type{?proto.pfs.BlockRef} */ (
    jspb.Message.getWrapperField(this, proto.pfs.BlockRef, 4));
};


/**
 * @param {?proto.pfs.BlockRef|undefined} value
 * @return {!proto.pfs.PutFileRecord} returns this
*/
proto.pfs.PutFileRecord.prototype.setBlockRef = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.PutFileRecord} returns this
 */
proto.pfs.PutFileRecord.prototype.clearBlockRef = function() {
  return this.setBlockRef(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.PutFileRecord.prototype.hasBlockRef = function() {
  return jspb.Message.getField(this, 4) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.PutFileRecords.repeatedFields_ = [2];



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
proto.pfs.PutFileRecords.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.PutFileRecords.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.PutFileRecords} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutFileRecords.toObject = function(includeInstance, msg) {
  var f, obj = {
    split: jspb.Message.getBooleanFieldWithDefault(msg, 1, false),
    recordsList: jspb.Message.toObjectList(msg.getRecordsList(),
    proto.pfs.PutFileRecord.toObject, includeInstance),
    tombstone: jspb.Message.getBooleanFieldWithDefault(msg, 3, false),
    header: (f = msg.getHeader()) && proto.pfs.PutFileRecord.toObject(includeInstance, f),
    footer: (f = msg.getFooter()) && proto.pfs.PutFileRecord.toObject(includeInstance, f)
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
 * @return {!proto.pfs.PutFileRecords}
 */
proto.pfs.PutFileRecords.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.PutFileRecords;
  return proto.pfs.PutFileRecords.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.PutFileRecords} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.PutFileRecords}
 */
proto.pfs.PutFileRecords.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSplit(value);
      break;
    case 2:
      var value = new proto.pfs.PutFileRecord;
      reader.readMessage(value,proto.pfs.PutFileRecord.deserializeBinaryFromReader);
      msg.addRecords(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setTombstone(value);
      break;
    case 4:
      var value = new proto.pfs.PutFileRecord;
      reader.readMessage(value,proto.pfs.PutFileRecord.deserializeBinaryFromReader);
      msg.setHeader(value);
      break;
    case 5:
      var value = new proto.pfs.PutFileRecord;
      reader.readMessage(value,proto.pfs.PutFileRecord.deserializeBinaryFromReader);
      msg.setFooter(value);
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
proto.pfs.PutFileRecords.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.PutFileRecords.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.PutFileRecords} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutFileRecords.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSplit();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
  f = message.getRecordsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.pfs.PutFileRecord.serializeBinaryToWriter
    );
  }
  f = message.getTombstone();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
  f = message.getHeader();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pfs.PutFileRecord.serializeBinaryToWriter
    );
  }
  f = message.getFooter();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.pfs.PutFileRecord.serializeBinaryToWriter
    );
  }
};


/**
 * optional bool split = 1;
 * @return {boolean}
 */
proto.pfs.PutFileRecords.prototype.getSplit = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.PutFileRecords} returns this
 */
proto.pfs.PutFileRecords.prototype.setSplit = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};


/**
 * repeated PutFileRecord records = 2;
 * @return {!Array<!proto.pfs.PutFileRecord>}
 */
proto.pfs.PutFileRecords.prototype.getRecordsList = function() {
  return /** @type{!Array<!proto.pfs.PutFileRecord>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.PutFileRecord, 2));
};


/**
 * @param {!Array<!proto.pfs.PutFileRecord>} value
 * @return {!proto.pfs.PutFileRecords} returns this
*/
proto.pfs.PutFileRecords.prototype.setRecordsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pfs.PutFileRecord=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.PutFileRecord}
 */
proto.pfs.PutFileRecords.prototype.addRecords = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pfs.PutFileRecord, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.PutFileRecords} returns this
 */
proto.pfs.PutFileRecords.prototype.clearRecordsList = function() {
  return this.setRecordsList([]);
};


/**
 * optional bool tombstone = 3;
 * @return {boolean}
 */
proto.pfs.PutFileRecords.prototype.getTombstone = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.PutFileRecords} returns this
 */
proto.pfs.PutFileRecords.prototype.setTombstone = function(value) {
  return jspb.Message.setProto3BooleanField(this, 3, value);
};


/**
 * optional PutFileRecord header = 4;
 * @return {?proto.pfs.PutFileRecord}
 */
proto.pfs.PutFileRecords.prototype.getHeader = function() {
  return /** @type{?proto.pfs.PutFileRecord} */ (
    jspb.Message.getWrapperField(this, proto.pfs.PutFileRecord, 4));
};


/**
 * @param {?proto.pfs.PutFileRecord|undefined} value
 * @return {!proto.pfs.PutFileRecords} returns this
*/
proto.pfs.PutFileRecords.prototype.setHeader = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.PutFileRecords} returns this
 */
proto.pfs.PutFileRecords.prototype.clearHeader = function() {
  return this.setHeader(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.PutFileRecords.prototype.hasHeader = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional PutFileRecord footer = 5;
 * @return {?proto.pfs.PutFileRecord}
 */
proto.pfs.PutFileRecords.prototype.getFooter = function() {
  return /** @type{?proto.pfs.PutFileRecord} */ (
    jspb.Message.getWrapperField(this, proto.pfs.PutFileRecord, 5));
};


/**
 * @param {?proto.pfs.PutFileRecord|undefined} value
 * @return {!proto.pfs.PutFileRecords} returns this
*/
proto.pfs.PutFileRecords.prototype.setFooter = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.PutFileRecords} returns this
 */
proto.pfs.PutFileRecords.prototype.clearFooter = function() {
  return this.setFooter(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.PutFileRecords.prototype.hasFooter = function() {
  return jspb.Message.getField(this, 5) != null;
};





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
proto.pfs.CopyFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CopyFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CopyFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CopyFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    src: (f = msg.getSrc()) && proto.pfs.File.toObject(includeInstance, f),
    dst: (f = msg.getDst()) && proto.pfs.File.toObject(includeInstance, f),
    overwrite: jspb.Message.getBooleanFieldWithDefault(msg, 3, false)
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
 * @return {!proto.pfs.CopyFileRequest}
 */
proto.pfs.CopyFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CopyFileRequest;
  return proto.pfs.CopyFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CopyFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CopyFileRequest}
 */
proto.pfs.CopyFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setSrc(value);
      break;
    case 2:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setDst(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setOverwrite(value);
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
proto.pfs.CopyFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CopyFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CopyFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CopyFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSrc();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
  f = message.getDst();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
  f = message.getOverwrite();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
};


/**
 * optional File src = 1;
 * @return {?proto.pfs.File}
 */
proto.pfs.CopyFileRequest.prototype.getSrc = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 1));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.CopyFileRequest} returns this
*/
proto.pfs.CopyFileRequest.prototype.setSrc = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CopyFileRequest} returns this
 */
proto.pfs.CopyFileRequest.prototype.clearSrc = function() {
  return this.setSrc(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CopyFileRequest.prototype.hasSrc = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional File dst = 2;
 * @return {?proto.pfs.File}
 */
proto.pfs.CopyFileRequest.prototype.getDst = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 2));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.CopyFileRequest} returns this
*/
proto.pfs.CopyFileRequest.prototype.setDst = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CopyFileRequest} returns this
 */
proto.pfs.CopyFileRequest.prototype.clearDst = function() {
  return this.setDst(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CopyFileRequest.prototype.hasDst = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional bool overwrite = 3;
 * @return {boolean}
 */
proto.pfs.CopyFileRequest.prototype.getOverwrite = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.CopyFileRequest} returns this
 */
proto.pfs.CopyFileRequest.prototype.setOverwrite = function(value) {
  return jspb.Message.setProto3BooleanField(this, 3, value);
};





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
proto.pfs.InspectFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.InspectFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.InspectFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.InspectFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs.File.toObject(includeInstance, f)
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
 * @return {!proto.pfs.InspectFileRequest}
 */
proto.pfs.InspectFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.InspectFileRequest;
  return proto.pfs.InspectFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.InspectFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.InspectFileRequest}
 */
proto.pfs.InspectFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setFile(value);
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
proto.pfs.InspectFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.InspectFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.InspectFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.InspectFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs.File}
 */
proto.pfs.InspectFileRequest.prototype.getFile = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 1));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.InspectFileRequest} returns this
*/
proto.pfs.InspectFileRequest.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.InspectFileRequest} returns this
 */
proto.pfs.InspectFileRequest.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.InspectFileRequest.prototype.hasFile = function() {
  return jspb.Message.getField(this, 1) != null;
};





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
proto.pfs.ListFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ListFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ListFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs.File.toObject(includeInstance, f),
    full: jspb.Message.getBooleanFieldWithDefault(msg, 2, false),
    history: jspb.Message.getFieldWithDefault(msg, 3, 0)
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
 * @return {!proto.pfs.ListFileRequest}
 */
proto.pfs.ListFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ListFileRequest;
  return proto.pfs.ListFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ListFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ListFileRequest}
 */
proto.pfs.ListFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setFile(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setFull(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setHistory(value);
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
proto.pfs.ListFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ListFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ListFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
  f = message.getFull();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
  f = message.getHistory();
  if (f !== 0) {
    writer.writeInt64(
      3,
      f
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs.File}
 */
proto.pfs.ListFileRequest.prototype.getFile = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 1));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.ListFileRequest} returns this
*/
proto.pfs.ListFileRequest.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.ListFileRequest} returns this
 */
proto.pfs.ListFileRequest.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.ListFileRequest.prototype.hasFile = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool full = 2;
 * @return {boolean}
 */
proto.pfs.ListFileRequest.prototype.getFull = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.ListFileRequest} returns this
 */
proto.pfs.ListFileRequest.prototype.setFull = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};


/**
 * optional int64 history = 3;
 * @return {number}
 */
proto.pfs.ListFileRequest.prototype.getHistory = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.ListFileRequest} returns this
 */
proto.pfs.ListFileRequest.prototype.setHistory = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};





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
proto.pfs.WalkFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.WalkFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.WalkFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.WalkFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs.File.toObject(includeInstance, f)
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
 * @return {!proto.pfs.WalkFileRequest}
 */
proto.pfs.WalkFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.WalkFileRequest;
  return proto.pfs.WalkFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.WalkFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.WalkFileRequest}
 */
proto.pfs.WalkFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setFile(value);
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
proto.pfs.WalkFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.WalkFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.WalkFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.WalkFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs.File}
 */
proto.pfs.WalkFileRequest.prototype.getFile = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 1));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.WalkFileRequest} returns this
*/
proto.pfs.WalkFileRequest.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.WalkFileRequest} returns this
 */
proto.pfs.WalkFileRequest.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.WalkFileRequest.prototype.hasFile = function() {
  return jspb.Message.getField(this, 1) != null;
};





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
proto.pfs.GlobFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.GlobFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.GlobFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GlobFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs.Commit.toObject(includeInstance, f),
    pattern: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pfs.GlobFileRequest}
 */
proto.pfs.GlobFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.GlobFileRequest;
  return proto.pfs.GlobFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.GlobFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.GlobFileRequest}
 */
proto.pfs.GlobFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setPattern(value);
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
proto.pfs.GlobFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.GlobFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.GlobFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GlobFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getPattern();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.GlobFileRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.GlobFileRequest} returns this
*/
proto.pfs.GlobFileRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.GlobFileRequest} returns this
 */
proto.pfs.GlobFileRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.GlobFileRequest.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string pattern = 2;
 * @return {string}
 */
proto.pfs.GlobFileRequest.prototype.getPattern = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.GlobFileRequest} returns this
 */
proto.pfs.GlobFileRequest.prototype.setPattern = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.FileInfos.repeatedFields_ = [1];



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
proto.pfs.FileInfos.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.FileInfos.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.FileInfos} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FileInfos.toObject = function(includeInstance, msg) {
  var f, obj = {
    fileInfoList: jspb.Message.toObjectList(msg.getFileInfoList(),
    proto.pfs.FileInfo.toObject, includeInstance)
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
 * @return {!proto.pfs.FileInfos}
 */
proto.pfs.FileInfos.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.FileInfos;
  return proto.pfs.FileInfos.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.FileInfos} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.FileInfos}
 */
proto.pfs.FileInfos.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.FileInfo;
      reader.readMessage(value,proto.pfs.FileInfo.deserializeBinaryFromReader);
      msg.addFileInfo(value);
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
proto.pfs.FileInfos.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.FileInfos.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.FileInfos} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FileInfos.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFileInfoList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs.FileInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated FileInfo file_info = 1;
 * @return {!Array<!proto.pfs.FileInfo>}
 */
proto.pfs.FileInfos.prototype.getFileInfoList = function() {
  return /** @type{!Array<!proto.pfs.FileInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.FileInfo, 1));
};


/**
 * @param {!Array<!proto.pfs.FileInfo>} value
 * @return {!proto.pfs.FileInfos} returns this
*/
proto.pfs.FileInfos.prototype.setFileInfoList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.FileInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.FileInfo}
 */
proto.pfs.FileInfos.prototype.addFileInfo = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.FileInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.FileInfos} returns this
 */
proto.pfs.FileInfos.prototype.clearFileInfoList = function() {
  return this.setFileInfoList([]);
};





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
proto.pfs.DiffFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DiffFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DiffFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DiffFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    newFile: (f = msg.getNewFile()) && proto.pfs.File.toObject(includeInstance, f),
    oldFile: (f = msg.getOldFile()) && proto.pfs.File.toObject(includeInstance, f),
    shallow: jspb.Message.getBooleanFieldWithDefault(msg, 3, false)
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
 * @return {!proto.pfs.DiffFileRequest}
 */
proto.pfs.DiffFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DiffFileRequest;
  return proto.pfs.DiffFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DiffFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DiffFileRequest}
 */
proto.pfs.DiffFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setNewFile(value);
      break;
    case 2:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setOldFile(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setShallow(value);
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
proto.pfs.DiffFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DiffFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DiffFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DiffFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getNewFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
  f = message.getOldFile();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
  f = message.getShallow();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
};


/**
 * optional File new_file = 1;
 * @return {?proto.pfs.File}
 */
proto.pfs.DiffFileRequest.prototype.getNewFile = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 1));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.DiffFileRequest} returns this
*/
proto.pfs.DiffFileRequest.prototype.setNewFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.DiffFileRequest} returns this
 */
proto.pfs.DiffFileRequest.prototype.clearNewFile = function() {
  return this.setNewFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.DiffFileRequest.prototype.hasNewFile = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional File old_file = 2;
 * @return {?proto.pfs.File}
 */
proto.pfs.DiffFileRequest.prototype.getOldFile = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 2));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.DiffFileRequest} returns this
*/
proto.pfs.DiffFileRequest.prototype.setOldFile = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.DiffFileRequest} returns this
 */
proto.pfs.DiffFileRequest.prototype.clearOldFile = function() {
  return this.setOldFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.DiffFileRequest.prototype.hasOldFile = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional bool shallow = 3;
 * @return {boolean}
 */
proto.pfs.DiffFileRequest.prototype.getShallow = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.DiffFileRequest} returns this
 */
proto.pfs.DiffFileRequest.prototype.setShallow = function(value) {
  return jspb.Message.setProto3BooleanField(this, 3, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.DiffFileResponse.repeatedFields_ = [1,2];



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
proto.pfs.DiffFileResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DiffFileResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DiffFileResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DiffFileResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    newFilesList: jspb.Message.toObjectList(msg.getNewFilesList(),
    proto.pfs.FileInfo.toObject, includeInstance),
    oldFilesList: jspb.Message.toObjectList(msg.getOldFilesList(),
    proto.pfs.FileInfo.toObject, includeInstance)
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
 * @return {!proto.pfs.DiffFileResponse}
 */
proto.pfs.DiffFileResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DiffFileResponse;
  return proto.pfs.DiffFileResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DiffFileResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DiffFileResponse}
 */
proto.pfs.DiffFileResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.FileInfo;
      reader.readMessage(value,proto.pfs.FileInfo.deserializeBinaryFromReader);
      msg.addNewFiles(value);
      break;
    case 2:
      var value = new proto.pfs.FileInfo;
      reader.readMessage(value,proto.pfs.FileInfo.deserializeBinaryFromReader);
      msg.addOldFiles(value);
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
proto.pfs.DiffFileResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DiffFileResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DiffFileResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DiffFileResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getNewFilesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs.FileInfo.serializeBinaryToWriter
    );
  }
  f = message.getOldFilesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.pfs.FileInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated FileInfo new_files = 1;
 * @return {!Array<!proto.pfs.FileInfo>}
 */
proto.pfs.DiffFileResponse.prototype.getNewFilesList = function() {
  return /** @type{!Array<!proto.pfs.FileInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.FileInfo, 1));
};


/**
 * @param {!Array<!proto.pfs.FileInfo>} value
 * @return {!proto.pfs.DiffFileResponse} returns this
*/
proto.pfs.DiffFileResponse.prototype.setNewFilesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.FileInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.FileInfo}
 */
proto.pfs.DiffFileResponse.prototype.addNewFiles = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.FileInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.DiffFileResponse} returns this
 */
proto.pfs.DiffFileResponse.prototype.clearNewFilesList = function() {
  return this.setNewFilesList([]);
};


/**
 * repeated FileInfo old_files = 2;
 * @return {!Array<!proto.pfs.FileInfo>}
 */
proto.pfs.DiffFileResponse.prototype.getOldFilesList = function() {
  return /** @type{!Array<!proto.pfs.FileInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.FileInfo, 2));
};


/**
 * @param {!Array<!proto.pfs.FileInfo>} value
 * @return {!proto.pfs.DiffFileResponse} returns this
*/
proto.pfs.DiffFileResponse.prototype.setOldFilesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pfs.FileInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.FileInfo}
 */
proto.pfs.DiffFileResponse.prototype.addOldFiles = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pfs.FileInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.DiffFileResponse} returns this
 */
proto.pfs.DiffFileResponse.prototype.clearOldFilesList = function() {
  return this.setOldFilesList([]);
};





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
proto.pfs.DeleteFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DeleteFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DeleteFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs.File.toObject(includeInstance, f)
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
 * @return {!proto.pfs.DeleteFileRequest}
 */
proto.pfs.DeleteFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DeleteFileRequest;
  return proto.pfs.DeleteFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DeleteFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DeleteFileRequest}
 */
proto.pfs.DeleteFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setFile(value);
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
proto.pfs.DeleteFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DeleteFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DeleteFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs.File}
 */
proto.pfs.DeleteFileRequest.prototype.getFile = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 1));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.DeleteFileRequest} returns this
*/
proto.pfs.DeleteFileRequest.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.DeleteFileRequest} returns this
 */
proto.pfs.DeleteFileRequest.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.DeleteFileRequest.prototype.hasFile = function() {
  return jspb.Message.getField(this, 1) != null;
};





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
proto.pfs.FsckRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.FsckRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.FsckRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FsckRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    fix: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
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
 * @return {!proto.pfs.FsckRequest}
 */
proto.pfs.FsckRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.FsckRequest;
  return proto.pfs.FsckRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.FsckRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.FsckRequest}
 */
proto.pfs.FsckRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setFix(value);
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
proto.pfs.FsckRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.FsckRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.FsckRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FsckRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFix();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool fix = 1;
 * @return {boolean}
 */
proto.pfs.FsckRequest.prototype.getFix = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.FsckRequest} returns this
 */
proto.pfs.FsckRequest.prototype.setFix = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};





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
proto.pfs.FsckResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.FsckResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.FsckResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FsckResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    fix: jspb.Message.getFieldWithDefault(msg, 1, ""),
    error: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pfs.FsckResponse}
 */
proto.pfs.FsckResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.FsckResponse;
  return proto.pfs.FsckResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.FsckResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.FsckResponse}
 */
proto.pfs.FsckResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setFix(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setError(value);
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
proto.pfs.FsckResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.FsckResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.FsckResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FsckResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFix();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getError();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string fix = 1;
 * @return {string}
 */
proto.pfs.FsckResponse.prototype.getFix = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.FsckResponse} returns this
 */
proto.pfs.FsckResponse.prototype.setFix = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string error = 2;
 * @return {string}
 */
proto.pfs.FsckResponse.prototype.getError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.FsckResponse} returns this
 */
proto.pfs.FsckResponse.prototype.setError = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};



/**
 * Oneof group definitions for this message. Each group defines the field
 * numbers belonging to that group. When of these fields' value is set, all
 * other fields in the group are cleared. During deserialization, if multiple
 * fields are encountered for a group, only the last value seen will be kept.
 * @private {!Array<!Array<number>>}
 * @const
 */
proto.pfs.FileOperationRequestV2.oneofGroups_ = [[2,3]];

/**
 * @enum {number}
 */
proto.pfs.FileOperationRequestV2.OperationCase = {
  OPERATION_NOT_SET: 0,
  PUT_TAR: 2,
  DELETE_FILES: 3
};

/**
 * @return {proto.pfs.FileOperationRequestV2.OperationCase}
 */
proto.pfs.FileOperationRequestV2.prototype.getOperationCase = function() {
  return /** @type {proto.pfs.FileOperationRequestV2.OperationCase} */(jspb.Message.computeOneofCase(this, proto.pfs.FileOperationRequestV2.oneofGroups_[0]));
};



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
proto.pfs.FileOperationRequestV2.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.FileOperationRequestV2.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.FileOperationRequestV2} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FileOperationRequestV2.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs.Commit.toObject(includeInstance, f),
    putTar: (f = msg.getPutTar()) && proto.pfs.PutTarRequestV2.toObject(includeInstance, f),
    deleteFiles: (f = msg.getDeleteFiles()) && proto.pfs.DeleteFilesRequestV2.toObject(includeInstance, f)
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
 * @return {!proto.pfs.FileOperationRequestV2}
 */
proto.pfs.FileOperationRequestV2.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.FileOperationRequestV2;
  return proto.pfs.FileOperationRequestV2.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.FileOperationRequestV2} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.FileOperationRequestV2}
 */
proto.pfs.FileOperationRequestV2.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = new proto.pfs.PutTarRequestV2;
      reader.readMessage(value,proto.pfs.PutTarRequestV2.deserializeBinaryFromReader);
      msg.setPutTar(value);
      break;
    case 3:
      var value = new proto.pfs.DeleteFilesRequestV2;
      reader.readMessage(value,proto.pfs.DeleteFilesRequestV2.deserializeBinaryFromReader);
      msg.setDeleteFiles(value);
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
proto.pfs.FileOperationRequestV2.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.FileOperationRequestV2.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.FileOperationRequestV2} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.FileOperationRequestV2.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
  f = message.getPutTar();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.PutTarRequestV2.serializeBinaryToWriter
    );
  }
  f = message.getDeleteFiles();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pfs.DeleteFilesRequestV2.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.FileOperationRequestV2.prototype.getCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.FileOperationRequestV2} returns this
*/
proto.pfs.FileOperationRequestV2.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.FileOperationRequestV2} returns this
 */
proto.pfs.FileOperationRequestV2.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.FileOperationRequestV2.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional PutTarRequestV2 put_tar = 2;
 * @return {?proto.pfs.PutTarRequestV2}
 */
proto.pfs.FileOperationRequestV2.prototype.getPutTar = function() {
  return /** @type{?proto.pfs.PutTarRequestV2} */ (
    jspb.Message.getWrapperField(this, proto.pfs.PutTarRequestV2, 2));
};


/**
 * @param {?proto.pfs.PutTarRequestV2|undefined} value
 * @return {!proto.pfs.FileOperationRequestV2} returns this
*/
proto.pfs.FileOperationRequestV2.prototype.setPutTar = function(value) {
  return jspb.Message.setOneofWrapperField(this, 2, proto.pfs.FileOperationRequestV2.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.FileOperationRequestV2} returns this
 */
proto.pfs.FileOperationRequestV2.prototype.clearPutTar = function() {
  return this.setPutTar(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.FileOperationRequestV2.prototype.hasPutTar = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional DeleteFilesRequestV2 delete_files = 3;
 * @return {?proto.pfs.DeleteFilesRequestV2}
 */
proto.pfs.FileOperationRequestV2.prototype.getDeleteFiles = function() {
  return /** @type{?proto.pfs.DeleteFilesRequestV2} */ (
    jspb.Message.getWrapperField(this, proto.pfs.DeleteFilesRequestV2, 3));
};


/**
 * @param {?proto.pfs.DeleteFilesRequestV2|undefined} value
 * @return {!proto.pfs.FileOperationRequestV2} returns this
*/
proto.pfs.FileOperationRequestV2.prototype.setDeleteFiles = function(value) {
  return jspb.Message.setOneofWrapperField(this, 3, proto.pfs.FileOperationRequestV2.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.FileOperationRequestV2} returns this
 */
proto.pfs.FileOperationRequestV2.prototype.clearDeleteFiles = function() {
  return this.setDeleteFiles(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.FileOperationRequestV2.prototype.hasDeleteFiles = function() {
  return jspb.Message.getField(this, 3) != null;
};





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
proto.pfs.PutTarRequestV2.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.PutTarRequestV2.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.PutTarRequestV2} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutTarRequestV2.toObject = function(includeInstance, msg) {
  var f, obj = {
    overwrite: jspb.Message.getBooleanFieldWithDefault(msg, 1, false),
    tag: jspb.Message.getFieldWithDefault(msg, 2, ""),
    data: msg.getData_asB64()
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
 * @return {!proto.pfs.PutTarRequestV2}
 */
proto.pfs.PutTarRequestV2.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.PutTarRequestV2;
  return proto.pfs.PutTarRequestV2.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.PutTarRequestV2} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.PutTarRequestV2}
 */
proto.pfs.PutTarRequestV2.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setOverwrite(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setTag(value);
      break;
    case 3:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setData(value);
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
proto.pfs.PutTarRequestV2.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.PutTarRequestV2.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.PutTarRequestV2} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutTarRequestV2.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getOverwrite();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
  f = message.getTag();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getData_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      3,
      f
    );
  }
};


/**
 * optional bool overwrite = 1;
 * @return {boolean}
 */
proto.pfs.PutTarRequestV2.prototype.getOverwrite = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.PutTarRequestV2} returns this
 */
proto.pfs.PutTarRequestV2.prototype.setOverwrite = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};


/**
 * optional string tag = 2;
 * @return {string}
 */
proto.pfs.PutTarRequestV2.prototype.getTag = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.PutTarRequestV2} returns this
 */
proto.pfs.PutTarRequestV2.prototype.setTag = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional bytes data = 3;
 * @return {string}
 */
proto.pfs.PutTarRequestV2.prototype.getData = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * optional bytes data = 3;
 * This is a type-conversion wrapper around `getData()`
 * @return {string}
 */
proto.pfs.PutTarRequestV2.prototype.getData_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getData()));
};


/**
 * optional bytes data = 3;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getData()`
 * @return {!Uint8Array}
 */
proto.pfs.PutTarRequestV2.prototype.getData_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getData()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pfs.PutTarRequestV2} returns this
 */
proto.pfs.PutTarRequestV2.prototype.setData = function(value) {
  return jspb.Message.setProto3BytesField(this, 3, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.DeleteFilesRequestV2.repeatedFields_ = [1];



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
proto.pfs.DeleteFilesRequestV2.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DeleteFilesRequestV2.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DeleteFilesRequestV2} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteFilesRequestV2.toObject = function(includeInstance, msg) {
  var f, obj = {
    filesList: (f = jspb.Message.getRepeatedField(msg, 1)) == null ? undefined : f,
    tag: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pfs.DeleteFilesRequestV2}
 */
proto.pfs.DeleteFilesRequestV2.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DeleteFilesRequestV2;
  return proto.pfs.DeleteFilesRequestV2.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DeleteFilesRequestV2} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DeleteFilesRequestV2}
 */
proto.pfs.DeleteFilesRequestV2.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.addFiles(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setTag(value);
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
proto.pfs.DeleteFilesRequestV2.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DeleteFilesRequestV2.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DeleteFilesRequestV2} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteFilesRequestV2.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFilesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      1,
      f
    );
  }
  f = message.getTag();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * repeated string files = 1;
 * @return {!Array<string>}
 */
proto.pfs.DeleteFilesRequestV2.prototype.getFilesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 1));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pfs.DeleteFilesRequestV2} returns this
 */
proto.pfs.DeleteFilesRequestV2.prototype.setFilesList = function(value) {
  return jspb.Message.setField(this, 1, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pfs.DeleteFilesRequestV2} returns this
 */
proto.pfs.DeleteFilesRequestV2.prototype.addFiles = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 1, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.DeleteFilesRequestV2} returns this
 */
proto.pfs.DeleteFilesRequestV2.prototype.clearFilesList = function() {
  return this.setFilesList([]);
};


/**
 * optional string tag = 2;
 * @return {string}
 */
proto.pfs.DeleteFilesRequestV2.prototype.getTag = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.DeleteFilesRequestV2} returns this
 */
proto.pfs.DeleteFilesRequestV2.prototype.setTag = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





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
proto.pfs.GetTarRequestV2.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.GetTarRequestV2.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.GetTarRequestV2} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetTarRequestV2.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs.File.toObject(includeInstance, f)
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
 * @return {!proto.pfs.GetTarRequestV2}
 */
proto.pfs.GetTarRequestV2.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.GetTarRequestV2;
  return proto.pfs.GetTarRequestV2.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.GetTarRequestV2} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.GetTarRequestV2}
 */
proto.pfs.GetTarRequestV2.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.File;
      reader.readMessage(value,proto.pfs.File.deserializeBinaryFromReader);
      msg.setFile(value);
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
proto.pfs.GetTarRequestV2.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.GetTarRequestV2.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.GetTarRequestV2} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetTarRequestV2.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.File.serializeBinaryToWriter
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs.File}
 */
proto.pfs.GetTarRequestV2.prototype.getFile = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs.File, 1));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pfs.GetTarRequestV2} returns this
*/
proto.pfs.GetTarRequestV2.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.GetTarRequestV2} returns this
 */
proto.pfs.GetTarRequestV2.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.GetTarRequestV2.prototype.hasFile = function() {
  return jspb.Message.getField(this, 1) != null;
};





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
proto.pfs.DiffFileResponseV2.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DiffFileResponseV2.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DiffFileResponseV2} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DiffFileResponseV2.toObject = function(includeInstance, msg) {
  var f, obj = {
    oldFile: (f = msg.getOldFile()) && proto.pfs.FileInfo.toObject(includeInstance, f),
    newFile: (f = msg.getNewFile()) && proto.pfs.FileInfo.toObject(includeInstance, f)
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
 * @return {!proto.pfs.DiffFileResponseV2}
 */
proto.pfs.DiffFileResponseV2.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DiffFileResponseV2;
  return proto.pfs.DiffFileResponseV2.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DiffFileResponseV2} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DiffFileResponseV2}
 */
proto.pfs.DiffFileResponseV2.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.FileInfo;
      reader.readMessage(value,proto.pfs.FileInfo.deserializeBinaryFromReader);
      msg.setOldFile(value);
      break;
    case 2:
      var value = new proto.pfs.FileInfo;
      reader.readMessage(value,proto.pfs.FileInfo.deserializeBinaryFromReader);
      msg.setNewFile(value);
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
proto.pfs.DiffFileResponseV2.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DiffFileResponseV2.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DiffFileResponseV2} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DiffFileResponseV2.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getOldFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.FileInfo.serializeBinaryToWriter
    );
  }
  f = message.getNewFile();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.FileInfo.serializeBinaryToWriter
    );
  }
};


/**
 * optional FileInfo old_file = 1;
 * @return {?proto.pfs.FileInfo}
 */
proto.pfs.DiffFileResponseV2.prototype.getOldFile = function() {
  return /** @type{?proto.pfs.FileInfo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.FileInfo, 1));
};


/**
 * @param {?proto.pfs.FileInfo|undefined} value
 * @return {!proto.pfs.DiffFileResponseV2} returns this
*/
proto.pfs.DiffFileResponseV2.prototype.setOldFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.DiffFileResponseV2} returns this
 */
proto.pfs.DiffFileResponseV2.prototype.clearOldFile = function() {
  return this.setOldFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.DiffFileResponseV2.prototype.hasOldFile = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional FileInfo new_file = 2;
 * @return {?proto.pfs.FileInfo}
 */
proto.pfs.DiffFileResponseV2.prototype.getNewFile = function() {
  return /** @type{?proto.pfs.FileInfo} */ (
    jspb.Message.getWrapperField(this, proto.pfs.FileInfo, 2));
};


/**
 * @param {?proto.pfs.FileInfo|undefined} value
 * @return {!proto.pfs.DiffFileResponseV2} returns this
*/
proto.pfs.DiffFileResponseV2.prototype.setNewFile = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.DiffFileResponseV2} returns this
 */
proto.pfs.DiffFileResponseV2.prototype.clearNewFile = function() {
  return this.setNewFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.DiffFileResponseV2.prototype.hasNewFile = function() {
  return jspb.Message.getField(this, 2) != null;
};





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
proto.pfs.CreateTmpFileSetResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CreateTmpFileSetResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CreateTmpFileSetResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CreateTmpFileSetResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    filesetId: jspb.Message.getFieldWithDefault(msg, 1, "")
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
 * @return {!proto.pfs.CreateTmpFileSetResponse}
 */
proto.pfs.CreateTmpFileSetResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CreateTmpFileSetResponse;
  return proto.pfs.CreateTmpFileSetResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CreateTmpFileSetResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CreateTmpFileSetResponse}
 */
proto.pfs.CreateTmpFileSetResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setFilesetId(value);
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
proto.pfs.CreateTmpFileSetResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CreateTmpFileSetResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CreateTmpFileSetResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CreateTmpFileSetResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFilesetId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string fileset_id = 1;
 * @return {string}
 */
proto.pfs.CreateTmpFileSetResponse.prototype.getFilesetId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.CreateTmpFileSetResponse} returns this
 */
proto.pfs.CreateTmpFileSetResponse.prototype.setFilesetId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





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
proto.pfs.RenewTmpFileSetRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.RenewTmpFileSetRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.RenewTmpFileSetRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.RenewTmpFileSetRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    filesetId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    ttlSeconds: jspb.Message.getFieldWithDefault(msg, 2, 0)
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
 * @return {!proto.pfs.RenewTmpFileSetRequest}
 */
proto.pfs.RenewTmpFileSetRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.RenewTmpFileSetRequest;
  return proto.pfs.RenewTmpFileSetRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.RenewTmpFileSetRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.RenewTmpFileSetRequest}
 */
proto.pfs.RenewTmpFileSetRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setFilesetId(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTtlSeconds(value);
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
proto.pfs.RenewTmpFileSetRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.RenewTmpFileSetRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.RenewTmpFileSetRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.RenewTmpFileSetRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFilesetId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getTtlSeconds();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
};


/**
 * optional string fileset_id = 1;
 * @return {string}
 */
proto.pfs.RenewTmpFileSetRequest.prototype.getFilesetId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.RenewTmpFileSetRequest} returns this
 */
proto.pfs.RenewTmpFileSetRequest.prototype.setFilesetId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional int64 ttl_seconds = 2;
 * @return {number}
 */
proto.pfs.RenewTmpFileSetRequest.prototype.getTtlSeconds = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.RenewTmpFileSetRequest} returns this
 */
proto.pfs.RenewTmpFileSetRequest.prototype.setTtlSeconds = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};





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
proto.pfs.ClearCommitRequestV2.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ClearCommitRequestV2.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ClearCommitRequestV2} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ClearCommitRequestV2.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs.Commit.toObject(includeInstance, f)
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
 * @return {!proto.pfs.ClearCommitRequestV2}
 */
proto.pfs.ClearCommitRequestV2.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ClearCommitRequestV2;
  return proto.pfs.ClearCommitRequestV2.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ClearCommitRequestV2} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ClearCommitRequestV2}
 */
proto.pfs.ClearCommitRequestV2.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Commit;
      reader.readMessage(value,proto.pfs.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
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
proto.pfs.ClearCommitRequestV2.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ClearCommitRequestV2.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ClearCommitRequestV2} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ClearCommitRequestV2.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Commit.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pfs.ClearCommitRequestV2.prototype.getCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pfs.ClearCommitRequestV2} returns this
*/
proto.pfs.ClearCommitRequestV2.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.ClearCommitRequestV2} returns this
 */
proto.pfs.ClearCommitRequestV2.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.ClearCommitRequestV2.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.PutObjectRequest.repeatedFields_ = [2];



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
proto.pfs.PutObjectRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.PutObjectRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.PutObjectRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutObjectRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    value: msg.getValue_asB64(),
    tagsList: jspb.Message.toObjectList(msg.getTagsList(),
    proto.pfs.Tag.toObject, includeInstance),
    block: (f = msg.getBlock()) && proto.pfs.Block.toObject(includeInstance, f)
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
 * @return {!proto.pfs.PutObjectRequest}
 */
proto.pfs.PutObjectRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.PutObjectRequest;
  return proto.pfs.PutObjectRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.PutObjectRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.PutObjectRequest}
 */
proto.pfs.PutObjectRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setValue(value);
      break;
    case 2:
      var value = new proto.pfs.Tag;
      reader.readMessage(value,proto.pfs.Tag.deserializeBinaryFromReader);
      msg.addTags(value);
      break;
    case 3:
      var value = new proto.pfs.Block;
      reader.readMessage(value,proto.pfs.Block.deserializeBinaryFromReader);
      msg.setBlock(value);
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
proto.pfs.PutObjectRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.PutObjectRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.PutObjectRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutObjectRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getValue_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      1,
      f
    );
  }
  f = message.getTagsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.pfs.Tag.serializeBinaryToWriter
    );
  }
  f = message.getBlock();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pfs.Block.serializeBinaryToWriter
    );
  }
};


/**
 * optional bytes value = 1;
 * @return {string}
 */
proto.pfs.PutObjectRequest.prototype.getValue = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * optional bytes value = 1;
 * This is a type-conversion wrapper around `getValue()`
 * @return {string}
 */
proto.pfs.PutObjectRequest.prototype.getValue_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getValue()));
};


/**
 * optional bytes value = 1;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getValue()`
 * @return {!Uint8Array}
 */
proto.pfs.PutObjectRequest.prototype.getValue_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getValue()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pfs.PutObjectRequest} returns this
 */
proto.pfs.PutObjectRequest.prototype.setValue = function(value) {
  return jspb.Message.setProto3BytesField(this, 1, value);
};


/**
 * repeated Tag tags = 2;
 * @return {!Array<!proto.pfs.Tag>}
 */
proto.pfs.PutObjectRequest.prototype.getTagsList = function() {
  return /** @type{!Array<!proto.pfs.Tag>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Tag, 2));
};


/**
 * @param {!Array<!proto.pfs.Tag>} value
 * @return {!proto.pfs.PutObjectRequest} returns this
*/
proto.pfs.PutObjectRequest.prototype.setTagsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pfs.Tag=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Tag}
 */
proto.pfs.PutObjectRequest.prototype.addTags = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pfs.Tag, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.PutObjectRequest} returns this
 */
proto.pfs.PutObjectRequest.prototype.clearTagsList = function() {
  return this.setTagsList([]);
};


/**
 * optional Block block = 3;
 * @return {?proto.pfs.Block}
 */
proto.pfs.PutObjectRequest.prototype.getBlock = function() {
  return /** @type{?proto.pfs.Block} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Block, 3));
};


/**
 * @param {?proto.pfs.Block|undefined} value
 * @return {!proto.pfs.PutObjectRequest} returns this
*/
proto.pfs.PutObjectRequest.prototype.setBlock = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.PutObjectRequest} returns this
 */
proto.pfs.PutObjectRequest.prototype.clearBlock = function() {
  return this.setBlock(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.PutObjectRequest.prototype.hasBlock = function() {
  return jspb.Message.getField(this, 3) != null;
};





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
proto.pfs.CreateObjectRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CreateObjectRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CreateObjectRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CreateObjectRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    object: (f = msg.getObject()) && proto.pfs.Object.toObject(includeInstance, f),
    blockRef: (f = msg.getBlockRef()) && proto.pfs.BlockRef.toObject(includeInstance, f)
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
 * @return {!proto.pfs.CreateObjectRequest}
 */
proto.pfs.CreateObjectRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CreateObjectRequest;
  return proto.pfs.CreateObjectRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CreateObjectRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CreateObjectRequest}
 */
proto.pfs.CreateObjectRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.setObject(value);
      break;
    case 2:
      var value = new proto.pfs.BlockRef;
      reader.readMessage(value,proto.pfs.BlockRef.deserializeBinaryFromReader);
      msg.setBlockRef(value);
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
proto.pfs.CreateObjectRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CreateObjectRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CreateObjectRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CreateObjectRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getObject();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getBlockRef();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.BlockRef.serializeBinaryToWriter
    );
  }
};


/**
 * optional Object object = 1;
 * @return {?proto.pfs.Object}
 */
proto.pfs.CreateObjectRequest.prototype.getObject = function() {
  return /** @type{?proto.pfs.Object} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Object, 1));
};


/**
 * @param {?proto.pfs.Object|undefined} value
 * @return {!proto.pfs.CreateObjectRequest} returns this
*/
proto.pfs.CreateObjectRequest.prototype.setObject = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CreateObjectRequest} returns this
 */
proto.pfs.CreateObjectRequest.prototype.clearObject = function() {
  return this.setObject(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CreateObjectRequest.prototype.hasObject = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional BlockRef block_ref = 2;
 * @return {?proto.pfs.BlockRef}
 */
proto.pfs.CreateObjectRequest.prototype.getBlockRef = function() {
  return /** @type{?proto.pfs.BlockRef} */ (
    jspb.Message.getWrapperField(this, proto.pfs.BlockRef, 2));
};


/**
 * @param {?proto.pfs.BlockRef|undefined} value
 * @return {!proto.pfs.CreateObjectRequest} returns this
*/
proto.pfs.CreateObjectRequest.prototype.setBlockRef = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CreateObjectRequest} returns this
 */
proto.pfs.CreateObjectRequest.prototype.clearBlockRef = function() {
  return this.setBlockRef(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CreateObjectRequest.prototype.hasBlockRef = function() {
  return jspb.Message.getField(this, 2) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.GetObjectsRequest.repeatedFields_ = [1];



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
proto.pfs.GetObjectsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.GetObjectsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.GetObjectsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetObjectsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    objectsList: jspb.Message.toObjectList(msg.getObjectsList(),
    proto.pfs.Object.toObject, includeInstance),
    offsetBytes: jspb.Message.getFieldWithDefault(msg, 2, 0),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 3, 0),
    totalSize: jspb.Message.getFieldWithDefault(msg, 4, 0)
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
 * @return {!proto.pfs.GetObjectsRequest}
 */
proto.pfs.GetObjectsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.GetObjectsRequest;
  return proto.pfs.GetObjectsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.GetObjectsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.GetObjectsRequest}
 */
proto.pfs.GetObjectsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.addObjects(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setOffsetBytes(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setSizeBytes(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setTotalSize(value);
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
proto.pfs.GetObjectsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.GetObjectsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.GetObjectsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetObjectsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getObjectsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getOffsetBytes();
  if (f !== 0) {
    writer.writeUint64(
      2,
      f
    );
  }
  f = message.getSizeBytes();
  if (f !== 0) {
    writer.writeUint64(
      3,
      f
    );
  }
  f = message.getTotalSize();
  if (f !== 0) {
    writer.writeUint64(
      4,
      f
    );
  }
};


/**
 * repeated Object objects = 1;
 * @return {!Array<!proto.pfs.Object>}
 */
proto.pfs.GetObjectsRequest.prototype.getObjectsList = function() {
  return /** @type{!Array<!proto.pfs.Object>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Object, 1));
};


/**
 * @param {!Array<!proto.pfs.Object>} value
 * @return {!proto.pfs.GetObjectsRequest} returns this
*/
proto.pfs.GetObjectsRequest.prototype.setObjectsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.Object=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Object}
 */
proto.pfs.GetObjectsRequest.prototype.addObjects = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.Object, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.GetObjectsRequest} returns this
 */
proto.pfs.GetObjectsRequest.prototype.clearObjectsList = function() {
  return this.setObjectsList([]);
};


/**
 * optional uint64 offset_bytes = 2;
 * @return {number}
 */
proto.pfs.GetObjectsRequest.prototype.getOffsetBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.GetObjectsRequest} returns this
 */
proto.pfs.GetObjectsRequest.prototype.setOffsetBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional uint64 size_bytes = 3;
 * @return {number}
 */
proto.pfs.GetObjectsRequest.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.GetObjectsRequest} returns this
 */
proto.pfs.GetObjectsRequest.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional uint64 total_size = 4;
 * @return {number}
 */
proto.pfs.GetObjectsRequest.prototype.getTotalSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.GetObjectsRequest} returns this
 */
proto.pfs.GetObjectsRequest.prototype.setTotalSize = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};





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
proto.pfs.PutBlockRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.PutBlockRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.PutBlockRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutBlockRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    block: (f = msg.getBlock()) && proto.pfs.Block.toObject(includeInstance, f),
    value: msg.getValue_asB64()
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
 * @return {!proto.pfs.PutBlockRequest}
 */
proto.pfs.PutBlockRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.PutBlockRequest;
  return proto.pfs.PutBlockRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.PutBlockRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.PutBlockRequest}
 */
proto.pfs.PutBlockRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Block;
      reader.readMessage(value,proto.pfs.Block.deserializeBinaryFromReader);
      msg.setBlock(value);
      break;
    case 2:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setValue(value);
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
proto.pfs.PutBlockRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.PutBlockRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.PutBlockRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutBlockRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBlock();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Block.serializeBinaryToWriter
    );
  }
  f = message.getValue_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      2,
      f
    );
  }
};


/**
 * optional Block block = 1;
 * @return {?proto.pfs.Block}
 */
proto.pfs.PutBlockRequest.prototype.getBlock = function() {
  return /** @type{?proto.pfs.Block} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Block, 1));
};


/**
 * @param {?proto.pfs.Block|undefined} value
 * @return {!proto.pfs.PutBlockRequest} returns this
*/
proto.pfs.PutBlockRequest.prototype.setBlock = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.PutBlockRequest} returns this
 */
proto.pfs.PutBlockRequest.prototype.clearBlock = function() {
  return this.setBlock(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.PutBlockRequest.prototype.hasBlock = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bytes value = 2;
 * @return {string}
 */
proto.pfs.PutBlockRequest.prototype.getValue = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * optional bytes value = 2;
 * This is a type-conversion wrapper around `getValue()`
 * @return {string}
 */
proto.pfs.PutBlockRequest.prototype.getValue_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getValue()));
};


/**
 * optional bytes value = 2;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getValue()`
 * @return {!Uint8Array}
 */
proto.pfs.PutBlockRequest.prototype.getValue_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getValue()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pfs.PutBlockRequest} returns this
 */
proto.pfs.PutBlockRequest.prototype.setValue = function(value) {
  return jspb.Message.setProto3BytesField(this, 2, value);
};





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
proto.pfs.GetBlockRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.GetBlockRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.GetBlockRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetBlockRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    block: (f = msg.getBlock()) && proto.pfs.Block.toObject(includeInstance, f)
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
 * @return {!proto.pfs.GetBlockRequest}
 */
proto.pfs.GetBlockRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.GetBlockRequest;
  return proto.pfs.GetBlockRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.GetBlockRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.GetBlockRequest}
 */
proto.pfs.GetBlockRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Block;
      reader.readMessage(value,proto.pfs.Block.deserializeBinaryFromReader);
      msg.setBlock(value);
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
proto.pfs.GetBlockRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.GetBlockRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.GetBlockRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetBlockRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBlock();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Block.serializeBinaryToWriter
    );
  }
};


/**
 * optional Block block = 1;
 * @return {?proto.pfs.Block}
 */
proto.pfs.GetBlockRequest.prototype.getBlock = function() {
  return /** @type{?proto.pfs.Block} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Block, 1));
};


/**
 * @param {?proto.pfs.Block|undefined} value
 * @return {!proto.pfs.GetBlockRequest} returns this
*/
proto.pfs.GetBlockRequest.prototype.setBlock = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.GetBlockRequest} returns this
 */
proto.pfs.GetBlockRequest.prototype.clearBlock = function() {
  return this.setBlock(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.GetBlockRequest.prototype.hasBlock = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.GetBlocksRequest.repeatedFields_ = [1];



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
proto.pfs.GetBlocksRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.GetBlocksRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.GetBlocksRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetBlocksRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    blockrefsList: jspb.Message.toObjectList(msg.getBlockrefsList(),
    proto.pfs.BlockRef.toObject, includeInstance),
    offsetBytes: jspb.Message.getFieldWithDefault(msg, 2, 0),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 3, 0),
    totalSize: jspb.Message.getFieldWithDefault(msg, 4, 0)
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
 * @return {!proto.pfs.GetBlocksRequest}
 */
proto.pfs.GetBlocksRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.GetBlocksRequest;
  return proto.pfs.GetBlocksRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.GetBlocksRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.GetBlocksRequest}
 */
proto.pfs.GetBlocksRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.BlockRef;
      reader.readMessage(value,proto.pfs.BlockRef.deserializeBinaryFromReader);
      msg.addBlockrefs(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setOffsetBytes(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setSizeBytes(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setTotalSize(value);
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
proto.pfs.GetBlocksRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.GetBlocksRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.GetBlocksRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetBlocksRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBlockrefsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs.BlockRef.serializeBinaryToWriter
    );
  }
  f = message.getOffsetBytes();
  if (f !== 0) {
    writer.writeUint64(
      2,
      f
    );
  }
  f = message.getSizeBytes();
  if (f !== 0) {
    writer.writeUint64(
      3,
      f
    );
  }
  f = message.getTotalSize();
  if (f !== 0) {
    writer.writeUint64(
      4,
      f
    );
  }
};


/**
 * repeated BlockRef blockRefs = 1;
 * @return {!Array<!proto.pfs.BlockRef>}
 */
proto.pfs.GetBlocksRequest.prototype.getBlockrefsList = function() {
  return /** @type{!Array<!proto.pfs.BlockRef>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.BlockRef, 1));
};


/**
 * @param {!Array<!proto.pfs.BlockRef>} value
 * @return {!proto.pfs.GetBlocksRequest} returns this
*/
proto.pfs.GetBlocksRequest.prototype.setBlockrefsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.BlockRef=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.BlockRef}
 */
proto.pfs.GetBlocksRequest.prototype.addBlockrefs = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.BlockRef, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.GetBlocksRequest} returns this
 */
proto.pfs.GetBlocksRequest.prototype.clearBlockrefsList = function() {
  return this.setBlockrefsList([]);
};


/**
 * optional uint64 offset_bytes = 2;
 * @return {number}
 */
proto.pfs.GetBlocksRequest.prototype.getOffsetBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.GetBlocksRequest} returns this
 */
proto.pfs.GetBlocksRequest.prototype.setOffsetBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional uint64 size_bytes = 3;
 * @return {number}
 */
proto.pfs.GetBlocksRequest.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.GetBlocksRequest} returns this
 */
proto.pfs.GetBlocksRequest.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional uint64 total_size = 4;
 * @return {number}
 */
proto.pfs.GetBlocksRequest.prototype.getTotalSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs.GetBlocksRequest} returns this
 */
proto.pfs.GetBlocksRequest.prototype.setTotalSize = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};





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
proto.pfs.ListBlockRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ListBlockRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ListBlockRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListBlockRequest.toObject = function(includeInstance, msg) {
  var f, obj = {

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
 * @return {!proto.pfs.ListBlockRequest}
 */
proto.pfs.ListBlockRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ListBlockRequest;
  return proto.pfs.ListBlockRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ListBlockRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ListBlockRequest}
 */
proto.pfs.ListBlockRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
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
proto.pfs.ListBlockRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ListBlockRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ListBlockRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListBlockRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.TagObjectRequest.repeatedFields_ = [2];



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
proto.pfs.TagObjectRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.TagObjectRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.TagObjectRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.TagObjectRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    object: (f = msg.getObject()) && proto.pfs.Object.toObject(includeInstance, f),
    tagsList: jspb.Message.toObjectList(msg.getTagsList(),
    proto.pfs.Tag.toObject, includeInstance)
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
 * @return {!proto.pfs.TagObjectRequest}
 */
proto.pfs.TagObjectRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.TagObjectRequest;
  return proto.pfs.TagObjectRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.TagObjectRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.TagObjectRequest}
 */
proto.pfs.TagObjectRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.setObject(value);
      break;
    case 2:
      var value = new proto.pfs.Tag;
      reader.readMessage(value,proto.pfs.Tag.deserializeBinaryFromReader);
      msg.addTags(value);
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
proto.pfs.TagObjectRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.TagObjectRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.TagObjectRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.TagObjectRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getObject();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
  f = message.getTagsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.pfs.Tag.serializeBinaryToWriter
    );
  }
};


/**
 * optional Object object = 1;
 * @return {?proto.pfs.Object}
 */
proto.pfs.TagObjectRequest.prototype.getObject = function() {
  return /** @type{?proto.pfs.Object} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Object, 1));
};


/**
 * @param {?proto.pfs.Object|undefined} value
 * @return {!proto.pfs.TagObjectRequest} returns this
*/
proto.pfs.TagObjectRequest.prototype.setObject = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.TagObjectRequest} returns this
 */
proto.pfs.TagObjectRequest.prototype.clearObject = function() {
  return this.setObject(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.TagObjectRequest.prototype.hasObject = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * repeated Tag tags = 2;
 * @return {!Array<!proto.pfs.Tag>}
 */
proto.pfs.TagObjectRequest.prototype.getTagsList = function() {
  return /** @type{!Array<!proto.pfs.Tag>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Tag, 2));
};


/**
 * @param {!Array<!proto.pfs.Tag>} value
 * @return {!proto.pfs.TagObjectRequest} returns this
*/
proto.pfs.TagObjectRequest.prototype.setTagsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pfs.Tag=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Tag}
 */
proto.pfs.TagObjectRequest.prototype.addTags = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pfs.Tag, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.TagObjectRequest} returns this
 */
proto.pfs.TagObjectRequest.prototype.clearTagsList = function() {
  return this.setTagsList([]);
};





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
proto.pfs.ListObjectsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ListObjectsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ListObjectsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListObjectsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {

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
 * @return {!proto.pfs.ListObjectsRequest}
 */
proto.pfs.ListObjectsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ListObjectsRequest;
  return proto.pfs.ListObjectsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ListObjectsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ListObjectsRequest}
 */
proto.pfs.ListObjectsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
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
proto.pfs.ListObjectsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ListObjectsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ListObjectsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListObjectsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};





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
proto.pfs.ListTagsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ListTagsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ListTagsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListTagsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    prefix: jspb.Message.getFieldWithDefault(msg, 1, ""),
    includeObject: jspb.Message.getBooleanFieldWithDefault(msg, 2, false)
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
 * @return {!proto.pfs.ListTagsRequest}
 */
proto.pfs.ListTagsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ListTagsRequest;
  return proto.pfs.ListTagsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ListTagsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ListTagsRequest}
 */
proto.pfs.ListTagsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setPrefix(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setIncludeObject(value);
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
proto.pfs.ListTagsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ListTagsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ListTagsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListTagsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPrefix();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getIncludeObject();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
};


/**
 * optional string prefix = 1;
 * @return {string}
 */
proto.pfs.ListTagsRequest.prototype.getPrefix = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.ListTagsRequest} returns this
 */
proto.pfs.ListTagsRequest.prototype.setPrefix = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional bool include_object = 2;
 * @return {boolean}
 */
proto.pfs.ListTagsRequest.prototype.getIncludeObject = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.ListTagsRequest} returns this
 */
proto.pfs.ListTagsRequest.prototype.setIncludeObject = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};





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
proto.pfs.ListTagsResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ListTagsResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ListTagsResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListTagsResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    tag: (f = msg.getTag()) && proto.pfs.Tag.toObject(includeInstance, f),
    object: (f = msg.getObject()) && proto.pfs.Object.toObject(includeInstance, f)
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
 * @return {!proto.pfs.ListTagsResponse}
 */
proto.pfs.ListTagsResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ListTagsResponse;
  return proto.pfs.ListTagsResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ListTagsResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ListTagsResponse}
 */
proto.pfs.ListTagsResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Tag;
      reader.readMessage(value,proto.pfs.Tag.deserializeBinaryFromReader);
      msg.setTag(value);
      break;
    case 2:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.setObject(value);
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
proto.pfs.ListTagsResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ListTagsResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ListTagsResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ListTagsResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTag();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Tag.serializeBinaryToWriter
    );
  }
  f = message.getObject();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
};


/**
 * optional Tag tag = 1;
 * @return {?proto.pfs.Tag}
 */
proto.pfs.ListTagsResponse.prototype.getTag = function() {
  return /** @type{?proto.pfs.Tag} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Tag, 1));
};


/**
 * @param {?proto.pfs.Tag|undefined} value
 * @return {!proto.pfs.ListTagsResponse} returns this
*/
proto.pfs.ListTagsResponse.prototype.setTag = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.ListTagsResponse} returns this
 */
proto.pfs.ListTagsResponse.prototype.clearTag = function() {
  return this.setTag(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.ListTagsResponse.prototype.hasTag = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Object object = 2;
 * @return {?proto.pfs.Object}
 */
proto.pfs.ListTagsResponse.prototype.getObject = function() {
  return /** @type{?proto.pfs.Object} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Object, 2));
};


/**
 * @param {?proto.pfs.Object|undefined} value
 * @return {!proto.pfs.ListTagsResponse} returns this
*/
proto.pfs.ListTagsResponse.prototype.setObject = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.ListTagsResponse} returns this
 */
proto.pfs.ListTagsResponse.prototype.clearObject = function() {
  return this.setObject(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.ListTagsResponse.prototype.hasObject = function() {
  return jspb.Message.getField(this, 2) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.DeleteObjectsRequest.repeatedFields_ = [1];



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
proto.pfs.DeleteObjectsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DeleteObjectsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DeleteObjectsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteObjectsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    objectsList: jspb.Message.toObjectList(msg.getObjectsList(),
    proto.pfs.Object.toObject, includeInstance)
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
 * @return {!proto.pfs.DeleteObjectsRequest}
 */
proto.pfs.DeleteObjectsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DeleteObjectsRequest;
  return proto.pfs.DeleteObjectsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DeleteObjectsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DeleteObjectsRequest}
 */
proto.pfs.DeleteObjectsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.addObjects(value);
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
proto.pfs.DeleteObjectsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DeleteObjectsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DeleteObjectsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteObjectsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getObjectsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Object objects = 1;
 * @return {!Array<!proto.pfs.Object>}
 */
proto.pfs.DeleteObjectsRequest.prototype.getObjectsList = function() {
  return /** @type{!Array<!proto.pfs.Object>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Object, 1));
};


/**
 * @param {!Array<!proto.pfs.Object>} value
 * @return {!proto.pfs.DeleteObjectsRequest} returns this
*/
proto.pfs.DeleteObjectsRequest.prototype.setObjectsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.Object=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Object}
 */
proto.pfs.DeleteObjectsRequest.prototype.addObjects = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.Object, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.DeleteObjectsRequest} returns this
 */
proto.pfs.DeleteObjectsRequest.prototype.clearObjectsList = function() {
  return this.setObjectsList([]);
};





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
proto.pfs.DeleteObjectsResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DeleteObjectsResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DeleteObjectsResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteObjectsResponse.toObject = function(includeInstance, msg) {
  var f, obj = {

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
 * @return {!proto.pfs.DeleteObjectsResponse}
 */
proto.pfs.DeleteObjectsResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DeleteObjectsResponse;
  return proto.pfs.DeleteObjectsResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DeleteObjectsResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DeleteObjectsResponse}
 */
proto.pfs.DeleteObjectsResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
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
proto.pfs.DeleteObjectsResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DeleteObjectsResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DeleteObjectsResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteObjectsResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.DeleteTagsRequest.repeatedFields_ = [1];



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
proto.pfs.DeleteTagsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DeleteTagsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DeleteTagsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteTagsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    tagsList: jspb.Message.toObjectList(msg.getTagsList(),
    proto.pfs.Tag.toObject, includeInstance)
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
 * @return {!proto.pfs.DeleteTagsRequest}
 */
proto.pfs.DeleteTagsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DeleteTagsRequest;
  return proto.pfs.DeleteTagsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DeleteTagsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DeleteTagsRequest}
 */
proto.pfs.DeleteTagsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Tag;
      reader.readMessage(value,proto.pfs.Tag.deserializeBinaryFromReader);
      msg.addTags(value);
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
proto.pfs.DeleteTagsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DeleteTagsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DeleteTagsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteTagsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTagsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs.Tag.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Tag tags = 1;
 * @return {!Array<!proto.pfs.Tag>}
 */
proto.pfs.DeleteTagsRequest.prototype.getTagsList = function() {
  return /** @type{!Array<!proto.pfs.Tag>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Tag, 1));
};


/**
 * @param {!Array<!proto.pfs.Tag>} value
 * @return {!proto.pfs.DeleteTagsRequest} returns this
*/
proto.pfs.DeleteTagsRequest.prototype.setTagsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.Tag=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Tag}
 */
proto.pfs.DeleteTagsRequest.prototype.addTags = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.Tag, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.DeleteTagsRequest} returns this
 */
proto.pfs.DeleteTagsRequest.prototype.clearTagsList = function() {
  return this.setTagsList([]);
};





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
proto.pfs.DeleteTagsResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DeleteTagsResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DeleteTagsResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteTagsResponse.toObject = function(includeInstance, msg) {
  var f, obj = {

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
 * @return {!proto.pfs.DeleteTagsResponse}
 */
proto.pfs.DeleteTagsResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DeleteTagsResponse;
  return proto.pfs.DeleteTagsResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DeleteTagsResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DeleteTagsResponse}
 */
proto.pfs.DeleteTagsResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
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
proto.pfs.DeleteTagsResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DeleteTagsResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DeleteTagsResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteTagsResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};





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
proto.pfs.CheckObjectRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CheckObjectRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CheckObjectRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CheckObjectRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    object: (f = msg.getObject()) && proto.pfs.Object.toObject(includeInstance, f)
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
 * @return {!proto.pfs.CheckObjectRequest}
 */
proto.pfs.CheckObjectRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CheckObjectRequest;
  return proto.pfs.CheckObjectRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CheckObjectRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CheckObjectRequest}
 */
proto.pfs.CheckObjectRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.setObject(value);
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
proto.pfs.CheckObjectRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CheckObjectRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CheckObjectRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CheckObjectRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getObject();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
};


/**
 * optional Object object = 1;
 * @return {?proto.pfs.Object}
 */
proto.pfs.CheckObjectRequest.prototype.getObject = function() {
  return /** @type{?proto.pfs.Object} */ (
    jspb.Message.getWrapperField(this, proto.pfs.Object, 1));
};


/**
 * @param {?proto.pfs.Object|undefined} value
 * @return {!proto.pfs.CheckObjectRequest} returns this
*/
proto.pfs.CheckObjectRequest.prototype.setObject = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs.CheckObjectRequest} returns this
 */
proto.pfs.CheckObjectRequest.prototype.clearObject = function() {
  return this.setObject(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs.CheckObjectRequest.prototype.hasObject = function() {
  return jspb.Message.getField(this, 1) != null;
};





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
proto.pfs.CheckObjectResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.CheckObjectResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.CheckObjectResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CheckObjectResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    exists: jspb.Message.getBooleanFieldWithDefault(msg, 1, false)
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
 * @return {!proto.pfs.CheckObjectResponse}
 */
proto.pfs.CheckObjectResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.CheckObjectResponse;
  return proto.pfs.CheckObjectResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.CheckObjectResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.CheckObjectResponse}
 */
proto.pfs.CheckObjectResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setExists(value);
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
proto.pfs.CheckObjectResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.CheckObjectResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.CheckObjectResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.CheckObjectResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getExists();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
};


/**
 * optional bool exists = 1;
 * @return {boolean}
 */
proto.pfs.CheckObjectResponse.prototype.getExists = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs.CheckObjectResponse} returns this
 */
proto.pfs.CheckObjectResponse.prototype.setExists = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs.Objects.repeatedFields_ = [1];



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
proto.pfs.Objects.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.Objects.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.Objects} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Objects.toObject = function(includeInstance, msg) {
  var f, obj = {
    objectsList: jspb.Message.toObjectList(msg.getObjectsList(),
    proto.pfs.Object.toObject, includeInstance)
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
 * @return {!proto.pfs.Objects}
 */
proto.pfs.Objects.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.Objects;
  return proto.pfs.Objects.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.Objects} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.Objects}
 */
proto.pfs.Objects.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs.Object;
      reader.readMessage(value,proto.pfs.Object.deserializeBinaryFromReader);
      msg.addObjects(value);
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
proto.pfs.Objects.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.Objects.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.Objects} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.Objects.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getObjectsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs.Object.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Object objects = 1;
 * @return {!Array<!proto.pfs.Object>}
 */
proto.pfs.Objects.prototype.getObjectsList = function() {
  return /** @type{!Array<!proto.pfs.Object>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs.Object, 1));
};


/**
 * @param {!Array<!proto.pfs.Object>} value
 * @return {!proto.pfs.Objects} returns this
*/
proto.pfs.Objects.prototype.setObjectsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.Object=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Object}
 */
proto.pfs.Objects.prototype.addObjects = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.Object, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs.Objects} returns this
 */
proto.pfs.Objects.prototype.clearObjectsList = function() {
  return this.setObjectsList([]);
};





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
proto.pfs.PutObjDirectRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.PutObjDirectRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.PutObjDirectRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutObjDirectRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    obj: jspb.Message.getFieldWithDefault(msg, 1, ""),
    value: msg.getValue_asB64()
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
 * @return {!proto.pfs.PutObjDirectRequest}
 */
proto.pfs.PutObjDirectRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.PutObjDirectRequest;
  return proto.pfs.PutObjDirectRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.PutObjDirectRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.PutObjDirectRequest}
 */
proto.pfs.PutObjDirectRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setObj(value);
      break;
    case 2:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setValue(value);
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
proto.pfs.PutObjDirectRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.PutObjDirectRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.PutObjDirectRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.PutObjDirectRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getObj();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getValue_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      2,
      f
    );
  }
};


/**
 * optional string obj = 1;
 * @return {string}
 */
proto.pfs.PutObjDirectRequest.prototype.getObj = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.PutObjDirectRequest} returns this
 */
proto.pfs.PutObjDirectRequest.prototype.setObj = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional bytes value = 2;
 * @return {string}
 */
proto.pfs.PutObjDirectRequest.prototype.getValue = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * optional bytes value = 2;
 * This is a type-conversion wrapper around `getValue()`
 * @return {string}
 */
proto.pfs.PutObjDirectRequest.prototype.getValue_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getValue()));
};


/**
 * optional bytes value = 2;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getValue()`
 * @return {!Uint8Array}
 */
proto.pfs.PutObjDirectRequest.prototype.getValue_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getValue()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pfs.PutObjDirectRequest} returns this
 */
proto.pfs.PutObjDirectRequest.prototype.setValue = function(value) {
  return jspb.Message.setProto3BytesField(this, 2, value);
};





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
proto.pfs.GetObjDirectRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.GetObjDirectRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.GetObjDirectRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetObjDirectRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    obj: jspb.Message.getFieldWithDefault(msg, 1, "")
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
 * @return {!proto.pfs.GetObjDirectRequest}
 */
proto.pfs.GetObjDirectRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.GetObjDirectRequest;
  return proto.pfs.GetObjDirectRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.GetObjDirectRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.GetObjDirectRequest}
 */
proto.pfs.GetObjDirectRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setObj(value);
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
proto.pfs.GetObjDirectRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.GetObjDirectRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.GetObjDirectRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.GetObjDirectRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getObj();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string obj = 1;
 * @return {string}
 */
proto.pfs.GetObjDirectRequest.prototype.getObj = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.GetObjDirectRequest} returns this
 */
proto.pfs.GetObjDirectRequest.prototype.setObj = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};





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
proto.pfs.DeleteObjDirectRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.DeleteObjDirectRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.DeleteObjDirectRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteObjDirectRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    object: jspb.Message.getFieldWithDefault(msg, 1, ""),
    prefix: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pfs.DeleteObjDirectRequest}
 */
proto.pfs.DeleteObjDirectRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.DeleteObjDirectRequest;
  return proto.pfs.DeleteObjDirectRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.DeleteObjDirectRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.DeleteObjDirectRequest}
 */
proto.pfs.DeleteObjDirectRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setObject(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setPrefix(value);
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
proto.pfs.DeleteObjDirectRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.DeleteObjDirectRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.DeleteObjDirectRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.DeleteObjDirectRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getObject();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getPrefix();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string object = 1;
 * @return {string}
 */
proto.pfs.DeleteObjDirectRequest.prototype.getObject = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.DeleteObjDirectRequest} returns this
 */
proto.pfs.DeleteObjDirectRequest.prototype.setObject = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string prefix = 2;
 * @return {string}
 */
proto.pfs.DeleteObjDirectRequest.prototype.getPrefix = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs.DeleteObjDirectRequest} returns this
 */
proto.pfs.DeleteObjDirectRequest.prototype.setPrefix = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};





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
proto.pfs.ObjectIndex.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs.ObjectIndex.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs.ObjectIndex} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ObjectIndex.toObject = function(includeInstance, msg) {
  var f, obj = {
    objectsMap: (f = msg.getObjectsMap()) ? f.toObject(includeInstance, proto.pfs.BlockRef.toObject) : [],
    tagsMap: (f = msg.getTagsMap()) ? f.toObject(includeInstance, proto.pfs.Object.toObject) : []
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
 * @return {!proto.pfs.ObjectIndex}
 */
proto.pfs.ObjectIndex.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs.ObjectIndex;
  return proto.pfs.ObjectIndex.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs.ObjectIndex} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs.ObjectIndex}
 */
proto.pfs.ObjectIndex.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getObjectsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.pfs.BlockRef.deserializeBinaryFromReader, "", new proto.pfs.BlockRef());
         });
      break;
    case 2:
      var value = msg.getTagsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.pfs.Object.deserializeBinaryFromReader, "", new proto.pfs.Object());
         });
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
proto.pfs.ObjectIndex.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs.ObjectIndex.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs.ObjectIndex} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs.ObjectIndex.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getObjectsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.pfs.BlockRef.serializeBinaryToWriter);
  }
  f = message.getTagsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(2, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.pfs.Object.serializeBinaryToWriter);
  }
};


/**
 * map<string, BlockRef> objects = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.pfs.BlockRef>}
 */
proto.pfs.ObjectIndex.prototype.getObjectsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.pfs.BlockRef>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      proto.pfs.BlockRef));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pfs.ObjectIndex} returns this
 */
proto.pfs.ObjectIndex.prototype.clearObjectsMap = function() {
  this.getObjectsMap().clear();
  return this;};


/**
 * map<string, Object> tags = 2;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.pfs.Object>}
 */
proto.pfs.ObjectIndex.prototype.getTagsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.pfs.Object>} */ (
      jspb.Message.getMapField(this, 2, opt_noLazyCreate,
      proto.pfs.Object));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pfs.ObjectIndex} returns this
 */
proto.pfs.ObjectIndex.prototype.clearTagsMap = function() {
  this.getTagsMap().clear();
  return this;};


/**
 * @enum {number}
 */
proto.pfs.OriginKind = {
  USER: 0,
  AUTO: 1,
  FSCK: 2
};

/**
 * @enum {number}
 */
proto.pfs.FileType = {
  RESERVED: 0,
  FILE: 1,
  DIR: 2
};

/**
 * @enum {number}
 */
proto.pfs.CommitState = {
  STARTED: 0,
  READY: 1,
  FINISHED: 2
};

/**
 * @enum {number}
 */
proto.pfs.Delimiter = {
  NONE: 0,
  JSON: 1,
  LINE: 2,
  SQL: 3,
  CSV: 4
};

goog.object.extend(exports, proto.pfs);
