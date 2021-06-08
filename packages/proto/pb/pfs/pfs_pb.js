// source: pfs/pfs.proto
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
var global = Function('return this')();

var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
goog.object.extend(proto, google_protobuf_empty_pb);
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
goog.object.extend(proto, google_protobuf_timestamp_pb);
var google_protobuf_wrappers_pb = require('google-protobuf/google/protobuf/wrappers_pb.js');
goog.object.extend(proto, google_protobuf_wrappers_pb);
var gogoproto_gogo_pb = require('../gogoproto/gogo_pb.js');
goog.object.extend(proto, gogoproto_gogo_pb);
var auth_auth_pb = require('../auth/auth_pb.js');
goog.object.extend(proto, auth_auth_pb);
goog.exportSymbol('proto.pfs_v2.ActivateAuthRequest', null, global);
goog.exportSymbol('proto.pfs_v2.ActivateAuthResponse', null, global);
goog.exportSymbol('proto.pfs_v2.AddFilesetRequest', null, global);
goog.exportSymbol('proto.pfs_v2.Branch', null, global);
goog.exportSymbol('proto.pfs_v2.BranchInfo', null, global);
goog.exportSymbol('proto.pfs_v2.BranchInfos', null, global);
goog.exportSymbol('proto.pfs_v2.ClearCommitRequest', null, global);
goog.exportSymbol('proto.pfs_v2.Commit', null, global);
goog.exportSymbol('proto.pfs_v2.CommitInfo', null, global);
goog.exportSymbol('proto.pfs_v2.CommitInfos', null, global);
goog.exportSymbol('proto.pfs_v2.CommitOrigin', null, global);
goog.exportSymbol('proto.pfs_v2.CommitProvenance', null, global);
goog.exportSymbol('proto.pfs_v2.CommitRange', null, global);
goog.exportSymbol('proto.pfs_v2.CommitState', null, global);
goog.exportSymbol('proto.pfs_v2.Commitset', null, global);
goog.exportSymbol('proto.pfs_v2.CopyFile', null, global);
goog.exportSymbol('proto.pfs_v2.CreateBranchRequest', null, global);
goog.exportSymbol('proto.pfs_v2.CreateFilesetResponse', null, global);
goog.exportSymbol('proto.pfs_v2.CreateRepoRequest', null, global);
goog.exportSymbol('proto.pfs_v2.DeleteBranchRequest', null, global);
goog.exportSymbol('proto.pfs_v2.DeleteFile', null, global);
goog.exportSymbol('proto.pfs_v2.DeleteRepoRequest', null, global);
goog.exportSymbol('proto.pfs_v2.Delimiter', null, global);
goog.exportSymbol('proto.pfs_v2.DiffFileRequest', null, global);
goog.exportSymbol('proto.pfs_v2.DiffFileResponse', null, global);
goog.exportSymbol('proto.pfs_v2.File', null, global);
goog.exportSymbol('proto.pfs_v2.FileInfo', null, global);
goog.exportSymbol('proto.pfs_v2.FileType', null, global);
goog.exportSymbol('proto.pfs_v2.FinishCommitRequest', null, global);
goog.exportSymbol('proto.pfs_v2.FlushCommitRequest', null, global);
goog.exportSymbol('proto.pfs_v2.FsckRequest', null, global);
goog.exportSymbol('proto.pfs_v2.FsckResponse', null, global);
goog.exportSymbol('proto.pfs_v2.GetFileRequest', null, global);
goog.exportSymbol('proto.pfs_v2.GetFilesetRequest', null, global);
goog.exportSymbol('proto.pfs_v2.GlobFileRequest', null, global);
goog.exportSymbol('proto.pfs_v2.InspectBranchRequest', null, global);
goog.exportSymbol('proto.pfs_v2.InspectCommitRequest', null, global);
goog.exportSymbol('proto.pfs_v2.InspectFileRequest', null, global);
goog.exportSymbol('proto.pfs_v2.InspectRepoRequest', null, global);
goog.exportSymbol('proto.pfs_v2.ListBranchRequest', null, global);
goog.exportSymbol('proto.pfs_v2.ListCommitRequest', null, global);
goog.exportSymbol('proto.pfs_v2.ListFileRequest', null, global);
goog.exportSymbol('proto.pfs_v2.ListRepoRequest', null, global);
goog.exportSymbol('proto.pfs_v2.ListRepoResponse', null, global);
goog.exportSymbol('proto.pfs_v2.ModifyFileRequest', null, global);
goog.exportSymbol('proto.pfs_v2.ModifyFileRequest.ModificationCase', null, global);
goog.exportSymbol('proto.pfs_v2.OriginKind', null, global);
goog.exportSymbol('proto.pfs_v2.PutFile', null, global);
goog.exportSymbol('proto.pfs_v2.PutFile.SourceCase', null, global);
goog.exportSymbol('proto.pfs_v2.RawFileSource', null, global);
goog.exportSymbol('proto.pfs_v2.RenewFilesetRequest', null, global);
goog.exportSymbol('proto.pfs_v2.Repo', null, global);
goog.exportSymbol('proto.pfs_v2.RepoAuthInfo', null, global);
goog.exportSymbol('proto.pfs_v2.RepoInfo', null, global);
goog.exportSymbol('proto.pfs_v2.RunLoadTestRequest', null, global);
goog.exportSymbol('proto.pfs_v2.RunLoadTestResponse', null, global);
goog.exportSymbol('proto.pfs_v2.SquashCommitRequest', null, global);
goog.exportSymbol('proto.pfs_v2.StartCommitRequest', null, global);
goog.exportSymbol('proto.pfs_v2.StoredCommitset', null, global);
goog.exportSymbol('proto.pfs_v2.SubscribeCommitRequest', null, global);
goog.exportSymbol('proto.pfs_v2.TarFileSource', null, global);
goog.exportSymbol('proto.pfs_v2.Trigger', null, global);
goog.exportSymbol('proto.pfs_v2.URLFileSource', null, global);
goog.exportSymbol('proto.pfs_v2.WalkFileRequest', null, global);
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
proto.pfs_v2.Repo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.Repo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.Repo.displayName = 'proto.pfs_v2.Repo';
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
proto.pfs_v2.Branch = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.Branch, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.Branch.displayName = 'proto.pfs_v2.Branch';
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
proto.pfs_v2.File = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.File, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.File.displayName = 'proto.pfs_v2.File';
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
proto.pfs_v2.RepoInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.RepoInfo.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.RepoInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.RepoInfo.displayName = 'proto.pfs_v2.RepoInfo';
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
proto.pfs_v2.RepoAuthInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.RepoAuthInfo.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.RepoAuthInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.RepoAuthInfo.displayName = 'proto.pfs_v2.RepoAuthInfo';
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
proto.pfs_v2.BranchInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.BranchInfo.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.BranchInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.BranchInfo.displayName = 'proto.pfs_v2.BranchInfo';
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
proto.pfs_v2.BranchInfos = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.BranchInfos.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.BranchInfos, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.BranchInfos.displayName = 'proto.pfs_v2.BranchInfos';
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
proto.pfs_v2.Trigger = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.Trigger, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.Trigger.displayName = 'proto.pfs_v2.Trigger';
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
proto.pfs_v2.CommitOrigin = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.CommitOrigin, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.CommitOrigin.displayName = 'proto.pfs_v2.CommitOrigin';
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
proto.pfs_v2.Commit = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.Commit, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.Commit.displayName = 'proto.pfs_v2.Commit';
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
proto.pfs_v2.CommitRange = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.CommitRange, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.CommitRange.displayName = 'proto.pfs_v2.CommitRange';
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
proto.pfs_v2.CommitProvenance = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.CommitProvenance, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.CommitProvenance.displayName = 'proto.pfs_v2.CommitProvenance';
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
proto.pfs_v2.CommitInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.CommitInfo.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.CommitInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.CommitInfo.displayName = 'proto.pfs_v2.CommitInfo';
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
proto.pfs_v2.StoredCommitset = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.StoredCommitset.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.StoredCommitset, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.StoredCommitset.displayName = 'proto.pfs_v2.StoredCommitset';
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
proto.pfs_v2.Commitset = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.Commitset.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.Commitset, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.Commitset.displayName = 'proto.pfs_v2.Commitset';
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
proto.pfs_v2.FileInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.FileInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.FileInfo.displayName = 'proto.pfs_v2.FileInfo';
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
proto.pfs_v2.CreateRepoRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.CreateRepoRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.CreateRepoRequest.displayName = 'proto.pfs_v2.CreateRepoRequest';
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
proto.pfs_v2.InspectRepoRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.InspectRepoRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.InspectRepoRequest.displayName = 'proto.pfs_v2.InspectRepoRequest';
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
proto.pfs_v2.ListRepoRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.ListRepoRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.ListRepoRequest.displayName = 'proto.pfs_v2.ListRepoRequest';
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
proto.pfs_v2.ListRepoResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.ListRepoResponse.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.ListRepoResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.ListRepoResponse.displayName = 'proto.pfs_v2.ListRepoResponse';
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
proto.pfs_v2.DeleteRepoRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.DeleteRepoRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.DeleteRepoRequest.displayName = 'proto.pfs_v2.DeleteRepoRequest';
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
proto.pfs_v2.StartCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.StartCommitRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.StartCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.StartCommitRequest.displayName = 'proto.pfs_v2.StartCommitRequest';
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
proto.pfs_v2.FinishCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.FinishCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.FinishCommitRequest.displayName = 'proto.pfs_v2.FinishCommitRequest';
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
proto.pfs_v2.InspectCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.InspectCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.InspectCommitRequest.displayName = 'proto.pfs_v2.InspectCommitRequest';
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
proto.pfs_v2.ListCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.ListCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.ListCommitRequest.displayName = 'proto.pfs_v2.ListCommitRequest';
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
proto.pfs_v2.CommitInfos = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.CommitInfos.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.CommitInfos, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.CommitInfos.displayName = 'proto.pfs_v2.CommitInfos';
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
proto.pfs_v2.SquashCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.SquashCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.SquashCommitRequest.displayName = 'proto.pfs_v2.SquashCommitRequest';
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
proto.pfs_v2.FlushCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.FlushCommitRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.FlushCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.FlushCommitRequest.displayName = 'proto.pfs_v2.FlushCommitRequest';
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
proto.pfs_v2.SubscribeCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.SubscribeCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.SubscribeCommitRequest.displayName = 'proto.pfs_v2.SubscribeCommitRequest';
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
proto.pfs_v2.ClearCommitRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.ClearCommitRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.ClearCommitRequest.displayName = 'proto.pfs_v2.ClearCommitRequest';
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
proto.pfs_v2.CreateBranchRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pfs_v2.CreateBranchRequest.repeatedFields_, null);
};
goog.inherits(proto.pfs_v2.CreateBranchRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.CreateBranchRequest.displayName = 'proto.pfs_v2.CreateBranchRequest';
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
proto.pfs_v2.InspectBranchRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.InspectBranchRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.InspectBranchRequest.displayName = 'proto.pfs_v2.InspectBranchRequest';
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
proto.pfs_v2.ListBranchRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.ListBranchRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.ListBranchRequest.displayName = 'proto.pfs_v2.ListBranchRequest';
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
proto.pfs_v2.DeleteBranchRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.DeleteBranchRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.DeleteBranchRequest.displayName = 'proto.pfs_v2.DeleteBranchRequest';
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
proto.pfs_v2.PutFile = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, proto.pfs_v2.PutFile.oneofGroups_);
};
goog.inherits(proto.pfs_v2.PutFile, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.PutFile.displayName = 'proto.pfs_v2.PutFile';
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
proto.pfs_v2.RawFileSource = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.RawFileSource, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.RawFileSource.displayName = 'proto.pfs_v2.RawFileSource';
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
proto.pfs_v2.TarFileSource = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.TarFileSource, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.TarFileSource.displayName = 'proto.pfs_v2.TarFileSource';
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
proto.pfs_v2.URLFileSource = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.URLFileSource, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.URLFileSource.displayName = 'proto.pfs_v2.URLFileSource';
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
proto.pfs_v2.DeleteFile = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.DeleteFile, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.DeleteFile.displayName = 'proto.pfs_v2.DeleteFile';
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
proto.pfs_v2.CopyFile = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.CopyFile, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.CopyFile.displayName = 'proto.pfs_v2.CopyFile';
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
proto.pfs_v2.ModifyFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, proto.pfs_v2.ModifyFileRequest.oneofGroups_);
};
goog.inherits(proto.pfs_v2.ModifyFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.ModifyFileRequest.displayName = 'proto.pfs_v2.ModifyFileRequest';
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
proto.pfs_v2.GetFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.GetFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.GetFileRequest.displayName = 'proto.pfs_v2.GetFileRequest';
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
proto.pfs_v2.InspectFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.InspectFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.InspectFileRequest.displayName = 'proto.pfs_v2.InspectFileRequest';
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
proto.pfs_v2.ListFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.ListFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.ListFileRequest.displayName = 'proto.pfs_v2.ListFileRequest';
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
proto.pfs_v2.WalkFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.WalkFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.WalkFileRequest.displayName = 'proto.pfs_v2.WalkFileRequest';
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
proto.pfs_v2.GlobFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.GlobFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.GlobFileRequest.displayName = 'proto.pfs_v2.GlobFileRequest';
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
proto.pfs_v2.DiffFileRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.DiffFileRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.DiffFileRequest.displayName = 'proto.pfs_v2.DiffFileRequest';
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
proto.pfs_v2.DiffFileResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.DiffFileResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.DiffFileResponse.displayName = 'proto.pfs_v2.DiffFileResponse';
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
proto.pfs_v2.FsckRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.FsckRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.FsckRequest.displayName = 'proto.pfs_v2.FsckRequest';
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
proto.pfs_v2.FsckResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.FsckResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.FsckResponse.displayName = 'proto.pfs_v2.FsckResponse';
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
proto.pfs_v2.CreateFilesetResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.CreateFilesetResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.CreateFilesetResponse.displayName = 'proto.pfs_v2.CreateFilesetResponse';
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
proto.pfs_v2.GetFilesetRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.GetFilesetRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.GetFilesetRequest.displayName = 'proto.pfs_v2.GetFilesetRequest';
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
proto.pfs_v2.AddFilesetRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.AddFilesetRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.AddFilesetRequest.displayName = 'proto.pfs_v2.AddFilesetRequest';
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
proto.pfs_v2.RenewFilesetRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.RenewFilesetRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.RenewFilesetRequest.displayName = 'proto.pfs_v2.RenewFilesetRequest';
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
proto.pfs_v2.ActivateAuthRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.ActivateAuthRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.ActivateAuthRequest.displayName = 'proto.pfs_v2.ActivateAuthRequest';
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
proto.pfs_v2.ActivateAuthResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.ActivateAuthResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.ActivateAuthResponse.displayName = 'proto.pfs_v2.ActivateAuthResponse';
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
proto.pfs_v2.RunLoadTestRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.RunLoadTestRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.RunLoadTestRequest.displayName = 'proto.pfs_v2.RunLoadTestRequest';
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
proto.pfs_v2.RunLoadTestResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pfs_v2.RunLoadTestResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pfs_v2.RunLoadTestResponse.displayName = 'proto.pfs_v2.RunLoadTestResponse';
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
proto.pfs_v2.Repo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.Repo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.Repo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.Repo.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    type: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pfs_v2.Repo}
 */
proto.pfs_v2.Repo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.Repo;
  return proto.pfs_v2.Repo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.Repo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.Repo}
 */
proto.pfs_v2.Repo.deserializeBinaryFromReader = function(msg, reader) {
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
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setType(value);
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
proto.pfs_v2.Repo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.Repo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.Repo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.Repo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getType();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pfs_v2.Repo.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.Repo} returns this
 */
proto.pfs_v2.Repo.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string type = 2;
 * @return {string}
 */
proto.pfs_v2.Repo.prototype.getType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.Repo} returns this
 */
proto.pfs_v2.Repo.prototype.setType = function(value) {
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
proto.pfs_v2.Branch.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.Branch.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.Branch} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.Branch.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs_v2.Repo.toObject(includeInstance, f),
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
 * @return {!proto.pfs_v2.Branch}
 */
proto.pfs_v2.Branch.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.Branch;
  return proto.pfs_v2.Branch.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.Branch} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.Branch}
 */
proto.pfs_v2.Branch.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Repo;
      reader.readMessage(value,proto.pfs_v2.Repo.deserializeBinaryFromReader);
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
proto.pfs_v2.Branch.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.Branch.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.Branch} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.Branch.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Repo.serializeBinaryToWriter
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
 * @return {?proto.pfs_v2.Repo}
 */
proto.pfs_v2.Branch.prototype.getRepo = function() {
  return /** @type{?proto.pfs_v2.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Repo, 1));
};


/**
 * @param {?proto.pfs_v2.Repo|undefined} value
 * @return {!proto.pfs_v2.Branch} returns this
*/
proto.pfs_v2.Branch.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.Branch} returns this
 */
proto.pfs_v2.Branch.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.Branch.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string name = 2;
 * @return {string}
 */
proto.pfs_v2.Branch.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.Branch} returns this
 */
proto.pfs_v2.Branch.prototype.setName = function(value) {
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
proto.pfs_v2.File.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.File.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.File} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.File.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    path: jspb.Message.getFieldWithDefault(msg, 2, ""),
    tag: jspb.Message.getFieldWithDefault(msg, 3, "")
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
 * @return {!proto.pfs_v2.File}
 */
proto.pfs_v2.File.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.File;
  return proto.pfs_v2.File.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.File} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.File}
 */
proto.pfs_v2.File.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setPath(value);
      break;
    case 3:
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
proto.pfs_v2.File.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.File.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.File} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.File.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getTag();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.File.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.File} returns this
*/
proto.pfs_v2.File.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.File} returns this
 */
proto.pfs_v2.File.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.File.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string path = 2;
 * @return {string}
 */
proto.pfs_v2.File.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.File} returns this
 */
proto.pfs_v2.File.prototype.setPath = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string tag = 3;
 * @return {string}
 */
proto.pfs_v2.File.prototype.getTag = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.File} returns this
 */
proto.pfs_v2.File.prototype.setTag = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.RepoInfo.repeatedFields_ = [5];



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
proto.pfs_v2.RepoInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.RepoInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.RepoInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RepoInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs_v2.Repo.toObject(includeInstance, f),
    created: (f = msg.getCreated()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 3, 0),
    description: jspb.Message.getFieldWithDefault(msg, 4, ""),
    branchesList: jspb.Message.toObjectList(msg.getBranchesList(),
    proto.pfs_v2.Branch.toObject, includeInstance),
    authInfo: (f = msg.getAuthInfo()) && proto.pfs_v2.RepoAuthInfo.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.RepoInfo}
 */
proto.pfs_v2.RepoInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.RepoInfo;
  return proto.pfs_v2.RepoInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.RepoInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.RepoInfo}
 */
proto.pfs_v2.RepoInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Repo;
      reader.readMessage(value,proto.pfs_v2.Repo.deserializeBinaryFromReader);
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
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 5:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
      msg.addBranches(value);
      break;
    case 6:
      var value = new proto.pfs_v2.RepoAuthInfo;
      reader.readMessage(value,proto.pfs_v2.RepoAuthInfo.deserializeBinaryFromReader);
      msg.setAuthInfo(value);
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
proto.pfs_v2.RepoInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.RepoInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.RepoInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RepoInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Repo.serializeBinaryToWriter
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
      4,
      f
    );
  }
  f = message.getBranchesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      5,
      f,
      proto.pfs_v2.Branch.serializeBinaryToWriter
    );
  }
  f = message.getAuthInfo();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      proto.pfs_v2.RepoAuthInfo.serializeBinaryToWriter
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs_v2.Repo}
 */
proto.pfs_v2.RepoInfo.prototype.getRepo = function() {
  return /** @type{?proto.pfs_v2.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Repo, 1));
};


/**
 * @param {?proto.pfs_v2.Repo|undefined} value
 * @return {!proto.pfs_v2.RepoInfo} returns this
*/
proto.pfs_v2.RepoInfo.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.RepoInfo} returns this
 */
proto.pfs_v2.RepoInfo.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.RepoInfo.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional google.protobuf.Timestamp created = 2;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pfs_v2.RepoInfo.prototype.getCreated = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 2));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pfs_v2.RepoInfo} returns this
*/
proto.pfs_v2.RepoInfo.prototype.setCreated = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.RepoInfo} returns this
 */
proto.pfs_v2.RepoInfo.prototype.clearCreated = function() {
  return this.setCreated(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.RepoInfo.prototype.hasCreated = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional uint64 size_bytes = 3;
 * @return {number}
 */
proto.pfs_v2.RepoInfo.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.RepoInfo} returns this
 */
proto.pfs_v2.RepoInfo.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional string description = 4;
 * @return {string}
 */
proto.pfs_v2.RepoInfo.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.RepoInfo} returns this
 */
proto.pfs_v2.RepoInfo.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * repeated Branch branches = 5;
 * @return {!Array<!proto.pfs_v2.Branch>}
 */
proto.pfs_v2.RepoInfo.prototype.getBranchesList = function() {
  return /** @type{!Array<!proto.pfs_v2.Branch>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.Branch, 5));
};


/**
 * @param {!Array<!proto.pfs_v2.Branch>} value
 * @return {!proto.pfs_v2.RepoInfo} returns this
*/
proto.pfs_v2.RepoInfo.prototype.setBranchesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.pfs_v2.Branch=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.Branch}
 */
proto.pfs_v2.RepoInfo.prototype.addBranches = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.pfs_v2.Branch, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.RepoInfo} returns this
 */
proto.pfs_v2.RepoInfo.prototype.clearBranchesList = function() {
  return this.setBranchesList([]);
};


/**
 * optional RepoAuthInfo auth_info = 6;
 * @return {?proto.pfs_v2.RepoAuthInfo}
 */
proto.pfs_v2.RepoInfo.prototype.getAuthInfo = function() {
  return /** @type{?proto.pfs_v2.RepoAuthInfo} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.RepoAuthInfo, 6));
};


/**
 * @param {?proto.pfs_v2.RepoAuthInfo|undefined} value
 * @return {!proto.pfs_v2.RepoInfo} returns this
*/
proto.pfs_v2.RepoInfo.prototype.setAuthInfo = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.RepoInfo} returns this
 */
proto.pfs_v2.RepoInfo.prototype.clearAuthInfo = function() {
  return this.setAuthInfo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.RepoInfo.prototype.hasAuthInfo = function() {
  return jspb.Message.getField(this, 6) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.RepoAuthInfo.repeatedFields_ = [1,2];



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
proto.pfs_v2.RepoAuthInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.RepoAuthInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.RepoAuthInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RepoAuthInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    permissionsList: (f = jspb.Message.getRepeatedField(msg, 1)) == null ? undefined : f,
    rolesList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f
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
 * @return {!proto.pfs_v2.RepoAuthInfo}
 */
proto.pfs_v2.RepoAuthInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.RepoAuthInfo;
  return proto.pfs_v2.RepoAuthInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.RepoAuthInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.RepoAuthInfo}
 */
proto.pfs_v2.RepoAuthInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var values = /** @type {!Array<!proto.auth_v2.Permission>} */ (reader.isDelimited() ? reader.readPackedEnum() : [reader.readEnum()]);
      for (var i = 0; i < values.length; i++) {
        msg.addPermissions(values[i]);
      }
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addRoles(value);
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
proto.pfs_v2.RepoAuthInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.RepoAuthInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.RepoAuthInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RepoAuthInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPermissionsList();
  if (f.length > 0) {
    writer.writePackedEnum(
      1,
      f
    );
  }
  f = message.getRolesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
};


/**
 * repeated auth_v2.Permission permissions = 1;
 * @return {!Array<!proto.auth_v2.Permission>}
 */
proto.pfs_v2.RepoAuthInfo.prototype.getPermissionsList = function() {
  return /** @type {!Array<!proto.auth_v2.Permission>} */ (jspb.Message.getRepeatedField(this, 1));
};


/**
 * @param {!Array<!proto.auth_v2.Permission>} value
 * @return {!proto.pfs_v2.RepoAuthInfo} returns this
 */
proto.pfs_v2.RepoAuthInfo.prototype.setPermissionsList = function(value) {
  return jspb.Message.setField(this, 1, value || []);
};


/**
 * @param {!proto.auth_v2.Permission} value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.RepoAuthInfo} returns this
 */
proto.pfs_v2.RepoAuthInfo.prototype.addPermissions = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 1, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.RepoAuthInfo} returns this
 */
proto.pfs_v2.RepoAuthInfo.prototype.clearPermissionsList = function() {
  return this.setPermissionsList([]);
};


/**
 * repeated string roles = 2;
 * @return {!Array<string>}
 */
proto.pfs_v2.RepoAuthInfo.prototype.getRolesList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pfs_v2.RepoAuthInfo} returns this
 */
proto.pfs_v2.RepoAuthInfo.prototype.setRolesList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.RepoAuthInfo} returns this
 */
proto.pfs_v2.RepoAuthInfo.prototype.addRoles = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.RepoAuthInfo} returns this
 */
proto.pfs_v2.RepoAuthInfo.prototype.clearRolesList = function() {
  return this.setRolesList([]);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.BranchInfo.repeatedFields_ = [3,4,5];



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
proto.pfs_v2.BranchInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.BranchInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.BranchInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.BranchInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    branch: (f = msg.getBranch()) && proto.pfs_v2.Branch.toObject(includeInstance, f),
    head: (f = msg.getHead()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    provenanceList: jspb.Message.toObjectList(msg.getProvenanceList(),
    proto.pfs_v2.Branch.toObject, includeInstance),
    subvenanceList: jspb.Message.toObjectList(msg.getSubvenanceList(),
    proto.pfs_v2.Branch.toObject, includeInstance),
    directProvenanceList: jspb.Message.toObjectList(msg.getDirectProvenanceList(),
    proto.pfs_v2.Branch.toObject, includeInstance),
    trigger: (f = msg.getTrigger()) && proto.pfs_v2.Trigger.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.BranchInfo}
 */
proto.pfs_v2.BranchInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.BranchInfo;
  return proto.pfs_v2.BranchInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.BranchInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.BranchInfo}
 */
proto.pfs_v2.BranchInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
      msg.setBranch(value);
      break;
    case 2:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setHead(value);
      break;
    case 3:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
      msg.addProvenance(value);
      break;
    case 4:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
      msg.addSubvenance(value);
      break;
    case 5:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
      msg.addDirectProvenance(value);
      break;
    case 6:
      var value = new proto.pfs_v2.Trigger;
      reader.readMessage(value,proto.pfs_v2.Trigger.deserializeBinaryFromReader);
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
proto.pfs_v2.BranchInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.BranchInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.BranchInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.BranchInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Branch.serializeBinaryToWriter
    );
  }
  f = message.getHead();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      3,
      f,
      proto.pfs_v2.Branch.serializeBinaryToWriter
    );
  }
  f = message.getSubvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      4,
      f,
      proto.pfs_v2.Branch.serializeBinaryToWriter
    );
  }
  f = message.getDirectProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      5,
      f,
      proto.pfs_v2.Branch.serializeBinaryToWriter
    );
  }
  f = message.getTrigger();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      proto.pfs_v2.Trigger.serializeBinaryToWriter
    );
  }
};


/**
 * optional Branch branch = 1;
 * @return {?proto.pfs_v2.Branch}
 */
proto.pfs_v2.BranchInfo.prototype.getBranch = function() {
  return /** @type{?proto.pfs_v2.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Branch, 1));
};


/**
 * @param {?proto.pfs_v2.Branch|undefined} value
 * @return {!proto.pfs_v2.BranchInfo} returns this
*/
proto.pfs_v2.BranchInfo.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.BranchInfo} returns this
 */
proto.pfs_v2.BranchInfo.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.BranchInfo.prototype.hasBranch = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Commit head = 2;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.BranchInfo.prototype.getHead = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 2));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.BranchInfo} returns this
*/
proto.pfs_v2.BranchInfo.prototype.setHead = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.BranchInfo} returns this
 */
proto.pfs_v2.BranchInfo.prototype.clearHead = function() {
  return this.setHead(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.BranchInfo.prototype.hasHead = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * repeated Branch provenance = 3;
 * @return {!Array<!proto.pfs_v2.Branch>}
 */
proto.pfs_v2.BranchInfo.prototype.getProvenanceList = function() {
  return /** @type{!Array<!proto.pfs_v2.Branch>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.Branch, 3));
};


/**
 * @param {!Array<!proto.pfs_v2.Branch>} value
 * @return {!proto.pfs_v2.BranchInfo} returns this
*/
proto.pfs_v2.BranchInfo.prototype.setProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 3, value);
};


/**
 * @param {!proto.pfs_v2.Branch=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.Branch}
 */
proto.pfs_v2.BranchInfo.prototype.addProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 3, opt_value, proto.pfs_v2.Branch, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.BranchInfo} returns this
 */
proto.pfs_v2.BranchInfo.prototype.clearProvenanceList = function() {
  return this.setProvenanceList([]);
};


/**
 * repeated Branch subvenance = 4;
 * @return {!Array<!proto.pfs_v2.Branch>}
 */
proto.pfs_v2.BranchInfo.prototype.getSubvenanceList = function() {
  return /** @type{!Array<!proto.pfs_v2.Branch>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.Branch, 4));
};


/**
 * @param {!Array<!proto.pfs_v2.Branch>} value
 * @return {!proto.pfs_v2.BranchInfo} returns this
*/
proto.pfs_v2.BranchInfo.prototype.setSubvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 4, value);
};


/**
 * @param {!proto.pfs_v2.Branch=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.Branch}
 */
proto.pfs_v2.BranchInfo.prototype.addSubvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 4, opt_value, proto.pfs_v2.Branch, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.BranchInfo} returns this
 */
proto.pfs_v2.BranchInfo.prototype.clearSubvenanceList = function() {
  return this.setSubvenanceList([]);
};


/**
 * repeated Branch direct_provenance = 5;
 * @return {!Array<!proto.pfs_v2.Branch>}
 */
proto.pfs_v2.BranchInfo.prototype.getDirectProvenanceList = function() {
  return /** @type{!Array<!proto.pfs_v2.Branch>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.Branch, 5));
};


/**
 * @param {!Array<!proto.pfs_v2.Branch>} value
 * @return {!proto.pfs_v2.BranchInfo} returns this
*/
proto.pfs_v2.BranchInfo.prototype.setDirectProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.pfs_v2.Branch=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.Branch}
 */
proto.pfs_v2.BranchInfo.prototype.addDirectProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.pfs_v2.Branch, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.BranchInfo} returns this
 */
proto.pfs_v2.BranchInfo.prototype.clearDirectProvenanceList = function() {
  return this.setDirectProvenanceList([]);
};


/**
 * optional Trigger trigger = 6;
 * @return {?proto.pfs_v2.Trigger}
 */
proto.pfs_v2.BranchInfo.prototype.getTrigger = function() {
  return /** @type{?proto.pfs_v2.Trigger} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Trigger, 6));
};


/**
 * @param {?proto.pfs_v2.Trigger|undefined} value
 * @return {!proto.pfs_v2.BranchInfo} returns this
*/
proto.pfs_v2.BranchInfo.prototype.setTrigger = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.BranchInfo} returns this
 */
proto.pfs_v2.BranchInfo.prototype.clearTrigger = function() {
  return this.setTrigger(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.BranchInfo.prototype.hasTrigger = function() {
  return jspb.Message.getField(this, 6) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.BranchInfos.repeatedFields_ = [1];



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
proto.pfs_v2.BranchInfos.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.BranchInfos.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.BranchInfos} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.BranchInfos.toObject = function(includeInstance, msg) {
  var f, obj = {
    branchInfoList: jspb.Message.toObjectList(msg.getBranchInfoList(),
    proto.pfs_v2.BranchInfo.toObject, includeInstance)
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
 * @return {!proto.pfs_v2.BranchInfos}
 */
proto.pfs_v2.BranchInfos.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.BranchInfos;
  return proto.pfs_v2.BranchInfos.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.BranchInfos} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.BranchInfos}
 */
proto.pfs_v2.BranchInfos.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.BranchInfo;
      reader.readMessage(value,proto.pfs_v2.BranchInfo.deserializeBinaryFromReader);
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
proto.pfs_v2.BranchInfos.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.BranchInfos.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.BranchInfos} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.BranchInfos.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBranchInfoList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs_v2.BranchInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated BranchInfo branch_info = 1;
 * @return {!Array<!proto.pfs_v2.BranchInfo>}
 */
proto.pfs_v2.BranchInfos.prototype.getBranchInfoList = function() {
  return /** @type{!Array<!proto.pfs_v2.BranchInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.BranchInfo, 1));
};


/**
 * @param {!Array<!proto.pfs_v2.BranchInfo>} value
 * @return {!proto.pfs_v2.BranchInfos} returns this
*/
proto.pfs_v2.BranchInfos.prototype.setBranchInfoList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs_v2.BranchInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.BranchInfo}
 */
proto.pfs_v2.BranchInfos.prototype.addBranchInfo = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs_v2.BranchInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.BranchInfos} returns this
 */
proto.pfs_v2.BranchInfos.prototype.clearBranchInfoList = function() {
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
proto.pfs_v2.Trigger.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.Trigger.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.Trigger} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.Trigger.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pfs_v2.Trigger}
 */
proto.pfs_v2.Trigger.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.Trigger;
  return proto.pfs_v2.Trigger.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.Trigger} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.Trigger}
 */
proto.pfs_v2.Trigger.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pfs_v2.Trigger.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.Trigger.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.Trigger} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.Trigger.serializeBinaryToWriter = function(message, writer) {
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
proto.pfs_v2.Trigger.prototype.getBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.Trigger} returns this
 */
proto.pfs_v2.Trigger.prototype.setBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional bool all = 2;
 * @return {boolean}
 */
proto.pfs_v2.Trigger.prototype.getAll = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.Trigger} returns this
 */
proto.pfs_v2.Trigger.prototype.setAll = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};


/**
 * optional string cron_spec = 3;
 * @return {string}
 */
proto.pfs_v2.Trigger.prototype.getCronSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.Trigger} returns this
 */
proto.pfs_v2.Trigger.prototype.setCronSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string size = 4;
 * @return {string}
 */
proto.pfs_v2.Trigger.prototype.getSize = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.Trigger} returns this
 */
proto.pfs_v2.Trigger.prototype.setSize = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional int64 commits = 5;
 * @return {number}
 */
proto.pfs_v2.Trigger.prototype.getCommits = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.Trigger} returns this
 */
proto.pfs_v2.Trigger.prototype.setCommits = function(value) {
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
proto.pfs_v2.CommitOrigin.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.CommitOrigin.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.CommitOrigin} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CommitOrigin.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pfs_v2.CommitOrigin}
 */
proto.pfs_v2.CommitOrigin.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.CommitOrigin;
  return proto.pfs_v2.CommitOrigin.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.CommitOrigin} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.CommitOrigin}
 */
proto.pfs_v2.CommitOrigin.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!proto.pfs_v2.OriginKind} */ (reader.readEnum());
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
proto.pfs_v2.CommitOrigin.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.CommitOrigin.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.CommitOrigin} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CommitOrigin.serializeBinaryToWriter = function(message, writer) {
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
 * @return {!proto.pfs_v2.OriginKind}
 */
proto.pfs_v2.CommitOrigin.prototype.getKind = function() {
  return /** @type {!proto.pfs_v2.OriginKind} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {!proto.pfs_v2.OriginKind} value
 * @return {!proto.pfs_v2.CommitOrigin} returns this
 */
proto.pfs_v2.CommitOrigin.prototype.setKind = function(value) {
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
proto.pfs_v2.Commit.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.Commit.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.Commit} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.Commit.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    branch: (f = msg.getBranch()) && proto.pfs_v2.Branch.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.Commit}
 */
proto.pfs_v2.Commit.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.Commit;
  return proto.pfs_v2.Commit.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.Commit} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.Commit}
 */
proto.pfs_v2.Commit.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
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
proto.pfs_v2.Commit.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.Commit.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.Commit} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.Commit.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs_v2.Branch.serializeBinaryToWriter
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.pfs_v2.Commit.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.Commit} returns this
 */
proto.pfs_v2.Commit.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional Branch branch = 2;
 * @return {?proto.pfs_v2.Branch}
 */
proto.pfs_v2.Commit.prototype.getBranch = function() {
  return /** @type{?proto.pfs_v2.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Branch, 2));
};


/**
 * @param {?proto.pfs_v2.Branch|undefined} value
 * @return {!proto.pfs_v2.Commit} returns this
*/
proto.pfs_v2.Commit.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.Commit} returns this
 */
proto.pfs_v2.Commit.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.Commit.prototype.hasBranch = function() {
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
proto.pfs_v2.CommitRange.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.CommitRange.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.CommitRange} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CommitRange.toObject = function(includeInstance, msg) {
  var f, obj = {
    lower: (f = msg.getLower()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    upper: (f = msg.getUpper()) && proto.pfs_v2.Commit.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.CommitRange}
 */
proto.pfs_v2.CommitRange.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.CommitRange;
  return proto.pfs_v2.CommitRange.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.CommitRange} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.CommitRange}
 */
proto.pfs_v2.CommitRange.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setLower(value);
      break;
    case 2:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
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
proto.pfs_v2.CommitRange.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.CommitRange.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.CommitRange} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CommitRange.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getLower();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getUpper();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit lower = 1;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.CommitRange.prototype.getLower = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.CommitRange} returns this
*/
proto.pfs_v2.CommitRange.prototype.setLower = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CommitRange} returns this
 */
proto.pfs_v2.CommitRange.prototype.clearLower = function() {
  return this.setLower(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CommitRange.prototype.hasLower = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Commit upper = 2;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.CommitRange.prototype.getUpper = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 2));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.CommitRange} returns this
*/
proto.pfs_v2.CommitRange.prototype.setUpper = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CommitRange} returns this
 */
proto.pfs_v2.CommitRange.prototype.clearUpper = function() {
  return this.setUpper(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CommitRange.prototype.hasUpper = function() {
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
proto.pfs_v2.CommitProvenance.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.CommitProvenance.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.CommitProvenance} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CommitProvenance.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.CommitProvenance}
 */
proto.pfs_v2.CommitProvenance.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.CommitProvenance;
  return proto.pfs_v2.CommitProvenance.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.CommitProvenance} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.CommitProvenance}
 */
proto.pfs_v2.CommitProvenance.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
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
proto.pfs_v2.CommitProvenance.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.CommitProvenance.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.CommitProvenance} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CommitProvenance.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.CommitProvenance.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.CommitProvenance} returns this
*/
proto.pfs_v2.CommitProvenance.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CommitProvenance} returns this
 */
proto.pfs_v2.CommitProvenance.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CommitProvenance.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.CommitInfo.repeatedFields_ = [5,9,11];



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
proto.pfs_v2.CommitInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.CommitInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.CommitInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CommitInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    origin: (f = msg.getOrigin()) && proto.pfs_v2.CommitOrigin.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 3, ""),
    parentCommit: (f = msg.getParentCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    childCommitsList: jspb.Message.toObjectList(msg.getChildCommitsList(),
    proto.pfs_v2.Commit.toObject, includeInstance),
    started: (f = msg.getStarted()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    finished: (f = msg.getFinished()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 8, 0),
    provenanceList: jspb.Message.toObjectList(msg.getProvenanceList(),
    proto.pfs_v2.CommitProvenance.toObject, includeInstance),
    readyProvenance: jspb.Message.getFieldWithDefault(msg, 10, 0),
    subvenanceList: jspb.Message.toObjectList(msg.getSubvenanceList(),
    proto.pfs_v2.CommitRange.toObject, includeInstance),
    subvenantCommitsSuccess: jspb.Message.getFieldWithDefault(msg, 12, 0),
    subvenantCommitsFailure: jspb.Message.getFieldWithDefault(msg, 13, 0),
    subvenantCommitsTotal: jspb.Message.getFieldWithDefault(msg, 14, 0)
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
 * @return {!proto.pfs_v2.CommitInfo}
 */
proto.pfs_v2.CommitInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.CommitInfo;
  return proto.pfs_v2.CommitInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.CommitInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.CommitInfo}
 */
proto.pfs_v2.CommitInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = new proto.pfs_v2.CommitOrigin;
      reader.readMessage(value,proto.pfs_v2.CommitOrigin.deserializeBinaryFromReader);
      msg.setOrigin(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 4:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setParentCommit(value);
      break;
    case 5:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.addChildCommits(value);
      break;
    case 6:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setStarted(value);
      break;
    case 7:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setFinished(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setSizeBytes(value);
      break;
    case 9:
      var value = new proto.pfs_v2.CommitProvenance;
      reader.readMessage(value,proto.pfs_v2.CommitProvenance.deserializeBinaryFromReader);
      msg.addProvenance(value);
      break;
    case 10:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setReadyProvenance(value);
      break;
    case 11:
      var value = new proto.pfs_v2.CommitRange;
      reader.readMessage(value,proto.pfs_v2.CommitRange.deserializeBinaryFromReader);
      msg.addSubvenance(value);
      break;
    case 12:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setSubvenantCommitsSuccess(value);
      break;
    case 13:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setSubvenantCommitsFailure(value);
      break;
    case 14:
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
proto.pfs_v2.CommitInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.CommitInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.CommitInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CommitInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getOrigin();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs_v2.CommitOrigin.serializeBinaryToWriter
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getParentCommit();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getChildCommitsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      5,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getStarted();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getFinished();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getSizeBytes();
  if (f !== 0) {
    writer.writeUint64(
      8,
      f
    );
  }
  f = message.getProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      9,
      f,
      proto.pfs_v2.CommitProvenance.serializeBinaryToWriter
    );
  }
  f = message.getReadyProvenance();
  if (f !== 0) {
    writer.writeInt64(
      10,
      f
    );
  }
  f = message.getSubvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      11,
      f,
      proto.pfs_v2.CommitRange.serializeBinaryToWriter
    );
  }
  f = message.getSubvenantCommitsSuccess();
  if (f !== 0) {
    writer.writeInt64(
      12,
      f
    );
  }
  f = message.getSubvenantCommitsFailure();
  if (f !== 0) {
    writer.writeInt64(
      13,
      f
    );
  }
  f = message.getSubvenantCommitsTotal();
  if (f !== 0) {
    writer.writeInt64(
      14,
      f
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.CommitInfo.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
*/
proto.pfs_v2.CommitInfo.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CommitInfo.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional CommitOrigin origin = 2;
 * @return {?proto.pfs_v2.CommitOrigin}
 */
proto.pfs_v2.CommitInfo.prototype.getOrigin = function() {
  return /** @type{?proto.pfs_v2.CommitOrigin} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.CommitOrigin, 2));
};


/**
 * @param {?proto.pfs_v2.CommitOrigin|undefined} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
*/
proto.pfs_v2.CommitInfo.prototype.setOrigin = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.clearOrigin = function() {
  return this.setOrigin(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CommitInfo.prototype.hasOrigin = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional string description = 3;
 * @return {string}
 */
proto.pfs_v2.CommitInfo.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional Commit parent_commit = 4;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.CommitInfo.prototype.getParentCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 4));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
*/
proto.pfs_v2.CommitInfo.prototype.setParentCommit = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.clearParentCommit = function() {
  return this.setParentCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CommitInfo.prototype.hasParentCommit = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * repeated Commit child_commits = 5;
 * @return {!Array<!proto.pfs_v2.Commit>}
 */
proto.pfs_v2.CommitInfo.prototype.getChildCommitsList = function() {
  return /** @type{!Array<!proto.pfs_v2.Commit>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.Commit, 5));
};


/**
 * @param {!Array<!proto.pfs_v2.Commit>} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
*/
proto.pfs_v2.CommitInfo.prototype.setChildCommitsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.pfs_v2.Commit=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.Commit}
 */
proto.pfs_v2.CommitInfo.prototype.addChildCommits = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.pfs_v2.Commit, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.clearChildCommitsList = function() {
  return this.setChildCommitsList([]);
};


/**
 * optional google.protobuf.Timestamp started = 6;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pfs_v2.CommitInfo.prototype.getStarted = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 6));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
*/
proto.pfs_v2.CommitInfo.prototype.setStarted = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.clearStarted = function() {
  return this.setStarted(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CommitInfo.prototype.hasStarted = function() {
  return jspb.Message.getField(this, 6) != null;
};


/**
 * optional google.protobuf.Timestamp finished = 7;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pfs_v2.CommitInfo.prototype.getFinished = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 7));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
*/
proto.pfs_v2.CommitInfo.prototype.setFinished = function(value) {
  return jspb.Message.setWrapperField(this, 7, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.clearFinished = function() {
  return this.setFinished(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CommitInfo.prototype.hasFinished = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * optional uint64 size_bytes = 8;
 * @return {number}
 */
proto.pfs_v2.CommitInfo.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 8, value);
};


/**
 * repeated CommitProvenance provenance = 9;
 * @return {!Array<!proto.pfs_v2.CommitProvenance>}
 */
proto.pfs_v2.CommitInfo.prototype.getProvenanceList = function() {
  return /** @type{!Array<!proto.pfs_v2.CommitProvenance>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.CommitProvenance, 9));
};


/**
 * @param {!Array<!proto.pfs_v2.CommitProvenance>} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
*/
proto.pfs_v2.CommitInfo.prototype.setProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 9, value);
};


/**
 * @param {!proto.pfs_v2.CommitProvenance=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.CommitProvenance}
 */
proto.pfs_v2.CommitInfo.prototype.addProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 9, opt_value, proto.pfs_v2.CommitProvenance, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.clearProvenanceList = function() {
  return this.setProvenanceList([]);
};


/**
 * optional int64 ready_provenance = 10;
 * @return {number}
 */
proto.pfs_v2.CommitInfo.prototype.getReadyProvenance = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 10, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.setReadyProvenance = function(value) {
  return jspb.Message.setProto3IntField(this, 10, value);
};


/**
 * repeated CommitRange subvenance = 11;
 * @return {!Array<!proto.pfs_v2.CommitRange>}
 */
proto.pfs_v2.CommitInfo.prototype.getSubvenanceList = function() {
  return /** @type{!Array<!proto.pfs_v2.CommitRange>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.CommitRange, 11));
};


/**
 * @param {!Array<!proto.pfs_v2.CommitRange>} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
*/
proto.pfs_v2.CommitInfo.prototype.setSubvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 11, value);
};


/**
 * @param {!proto.pfs_v2.CommitRange=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.CommitRange}
 */
proto.pfs_v2.CommitInfo.prototype.addSubvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 11, opt_value, proto.pfs_v2.CommitRange, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.clearSubvenanceList = function() {
  return this.setSubvenanceList([]);
};


/**
 * optional int64 subvenant_commits_success = 12;
 * @return {number}
 */
proto.pfs_v2.CommitInfo.prototype.getSubvenantCommitsSuccess = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 12, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.setSubvenantCommitsSuccess = function(value) {
  return jspb.Message.setProto3IntField(this, 12, value);
};


/**
 * optional int64 subvenant_commits_failure = 13;
 * @return {number}
 */
proto.pfs_v2.CommitInfo.prototype.getSubvenantCommitsFailure = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 13, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.setSubvenantCommitsFailure = function(value) {
  return jspb.Message.setProto3IntField(this, 13, value);
};


/**
 * optional int64 subvenant_commits_total = 14;
 * @return {number}
 */
proto.pfs_v2.CommitInfo.prototype.getSubvenantCommitsTotal = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 14, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.CommitInfo} returns this
 */
proto.pfs_v2.CommitInfo.prototype.setSubvenantCommitsTotal = function(value) {
  return jspb.Message.setProto3IntField(this, 14, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.StoredCommitset.repeatedFields_ = [3];



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
proto.pfs_v2.StoredCommitset.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.StoredCommitset.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.StoredCommitset} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.StoredCommitset.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    origin: (f = msg.getOrigin()) && proto.pfs_v2.CommitOrigin.toObject(includeInstance, f),
    commitsList: jspb.Message.toObjectList(msg.getCommitsList(),
    proto.pfs_v2.Commit.toObject, includeInstance)
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
 * @return {!proto.pfs_v2.StoredCommitset}
 */
proto.pfs_v2.StoredCommitset.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.StoredCommitset;
  return proto.pfs_v2.StoredCommitset.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.StoredCommitset} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.StoredCommitset}
 */
proto.pfs_v2.StoredCommitset.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = new proto.pfs_v2.CommitOrigin;
      reader.readMessage(value,proto.pfs_v2.CommitOrigin.deserializeBinaryFromReader);
      msg.setOrigin(value);
      break;
    case 3:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.addCommits(value);
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
proto.pfs_v2.StoredCommitset.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.StoredCommitset.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.StoredCommitset} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.StoredCommitset.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getOrigin();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs_v2.CommitOrigin.serializeBinaryToWriter
    );
  }
  f = message.getCommitsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      3,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.pfs_v2.StoredCommitset.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.StoredCommitset} returns this
 */
proto.pfs_v2.StoredCommitset.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional CommitOrigin origin = 2;
 * @return {?proto.pfs_v2.CommitOrigin}
 */
proto.pfs_v2.StoredCommitset.prototype.getOrigin = function() {
  return /** @type{?proto.pfs_v2.CommitOrigin} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.CommitOrigin, 2));
};


/**
 * @param {?proto.pfs_v2.CommitOrigin|undefined} value
 * @return {!proto.pfs_v2.StoredCommitset} returns this
*/
proto.pfs_v2.StoredCommitset.prototype.setOrigin = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.StoredCommitset} returns this
 */
proto.pfs_v2.StoredCommitset.prototype.clearOrigin = function() {
  return this.setOrigin(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.StoredCommitset.prototype.hasOrigin = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * repeated Commit commits = 3;
 * @return {!Array<!proto.pfs_v2.Commit>}
 */
proto.pfs_v2.StoredCommitset.prototype.getCommitsList = function() {
  return /** @type{!Array<!proto.pfs_v2.Commit>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.Commit, 3));
};


/**
 * @param {!Array<!proto.pfs_v2.Commit>} value
 * @return {!proto.pfs_v2.StoredCommitset} returns this
*/
proto.pfs_v2.StoredCommitset.prototype.setCommitsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 3, value);
};


/**
 * @param {!proto.pfs_v2.Commit=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.Commit}
 */
proto.pfs_v2.StoredCommitset.prototype.addCommits = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 3, opt_value, proto.pfs_v2.Commit, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.StoredCommitset} returns this
 */
proto.pfs_v2.StoredCommitset.prototype.clearCommitsList = function() {
  return this.setCommitsList([]);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.Commitset.repeatedFields_ = [3];



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
proto.pfs_v2.Commitset.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.Commitset.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.Commitset} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.Commitset.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    origin: (f = msg.getOrigin()) && proto.pfs_v2.CommitOrigin.toObject(includeInstance, f),
    commitsList: jspb.Message.toObjectList(msg.getCommitsList(),
    proto.pfs_v2.CommitInfo.toObject, includeInstance)
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
 * @return {!proto.pfs_v2.Commitset}
 */
proto.pfs_v2.Commitset.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.Commitset;
  return proto.pfs_v2.Commitset.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.Commitset} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.Commitset}
 */
proto.pfs_v2.Commitset.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = new proto.pfs_v2.CommitOrigin;
      reader.readMessage(value,proto.pfs_v2.CommitOrigin.deserializeBinaryFromReader);
      msg.setOrigin(value);
      break;
    case 3:
      var value = new proto.pfs_v2.CommitInfo;
      reader.readMessage(value,proto.pfs_v2.CommitInfo.deserializeBinaryFromReader);
      msg.addCommits(value);
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
proto.pfs_v2.Commitset.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.Commitset.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.Commitset} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.Commitset.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getOrigin();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs_v2.CommitOrigin.serializeBinaryToWriter
    );
  }
  f = message.getCommitsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      3,
      f,
      proto.pfs_v2.CommitInfo.serializeBinaryToWriter
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.pfs_v2.Commitset.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.Commitset} returns this
 */
proto.pfs_v2.Commitset.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional CommitOrigin origin = 2;
 * @return {?proto.pfs_v2.CommitOrigin}
 */
proto.pfs_v2.Commitset.prototype.getOrigin = function() {
  return /** @type{?proto.pfs_v2.CommitOrigin} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.CommitOrigin, 2));
};


/**
 * @param {?proto.pfs_v2.CommitOrigin|undefined} value
 * @return {!proto.pfs_v2.Commitset} returns this
*/
proto.pfs_v2.Commitset.prototype.setOrigin = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.Commitset} returns this
 */
proto.pfs_v2.Commitset.prototype.clearOrigin = function() {
  return this.setOrigin(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.Commitset.prototype.hasOrigin = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * repeated CommitInfo commits = 3;
 * @return {!Array<!proto.pfs_v2.CommitInfo>}
 */
proto.pfs_v2.Commitset.prototype.getCommitsList = function() {
  return /** @type{!Array<!proto.pfs_v2.CommitInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.CommitInfo, 3));
};


/**
 * @param {!Array<!proto.pfs_v2.CommitInfo>} value
 * @return {!proto.pfs_v2.Commitset} returns this
*/
proto.pfs_v2.Commitset.prototype.setCommitsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 3, value);
};


/**
 * @param {!proto.pfs_v2.CommitInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.CommitInfo}
 */
proto.pfs_v2.Commitset.prototype.addCommits = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 3, opt_value, proto.pfs_v2.CommitInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.Commitset} returns this
 */
proto.pfs_v2.Commitset.prototype.clearCommitsList = function() {
  return this.setCommitsList([]);
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
proto.pfs_v2.FileInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.FileInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.FileInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.FileInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs_v2.File.toObject(includeInstance, f),
    fileType: jspb.Message.getFieldWithDefault(msg, 2, 0),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 3, 0),
    committed: (f = msg.getCommitted()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
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
 * @return {!proto.pfs_v2.FileInfo}
 */
proto.pfs_v2.FileInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.FileInfo;
  return proto.pfs_v2.FileInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.FileInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.FileInfo}
 */
proto.pfs_v2.FileInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.File;
      reader.readMessage(value,proto.pfs_v2.File.deserializeBinaryFromReader);
      msg.setFile(value);
      break;
    case 2:
      var value = /** @type {!proto.pfs_v2.FileType} */ (reader.readEnum());
      msg.setFileType(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setSizeBytes(value);
      break;
    case 4:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCommitted(value);
      break;
    case 5:
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
proto.pfs_v2.FileInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.FileInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.FileInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.FileInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.File.serializeBinaryToWriter
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
      4,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getHash_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      5,
      f
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs_v2.File}
 */
proto.pfs_v2.FileInfo.prototype.getFile = function() {
  return /** @type{?proto.pfs_v2.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.File, 1));
};


/**
 * @param {?proto.pfs_v2.File|undefined} value
 * @return {!proto.pfs_v2.FileInfo} returns this
*/
proto.pfs_v2.FileInfo.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.FileInfo} returns this
 */
proto.pfs_v2.FileInfo.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.FileInfo.prototype.hasFile = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional FileType file_type = 2;
 * @return {!proto.pfs_v2.FileType}
 */
proto.pfs_v2.FileInfo.prototype.getFileType = function() {
  return /** @type {!proto.pfs_v2.FileType} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.pfs_v2.FileType} value
 * @return {!proto.pfs_v2.FileInfo} returns this
 */
proto.pfs_v2.FileInfo.prototype.setFileType = function(value) {
  return jspb.Message.setProto3EnumField(this, 2, value);
};


/**
 * optional uint64 size_bytes = 3;
 * @return {number}
 */
proto.pfs_v2.FileInfo.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.FileInfo} returns this
 */
proto.pfs_v2.FileInfo.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional google.protobuf.Timestamp committed = 4;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pfs_v2.FileInfo.prototype.getCommitted = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 4));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pfs_v2.FileInfo} returns this
*/
proto.pfs_v2.FileInfo.prototype.setCommitted = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.FileInfo} returns this
 */
proto.pfs_v2.FileInfo.prototype.clearCommitted = function() {
  return this.setCommitted(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.FileInfo.prototype.hasCommitted = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional bytes hash = 5;
 * @return {string}
 */
proto.pfs_v2.FileInfo.prototype.getHash = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * optional bytes hash = 5;
 * This is a type-conversion wrapper around `getHash()`
 * @return {string}
 */
proto.pfs_v2.FileInfo.prototype.getHash_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getHash()));
};


/**
 * optional bytes hash = 5;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getHash()`
 * @return {!Uint8Array}
 */
proto.pfs_v2.FileInfo.prototype.getHash_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getHash()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pfs_v2.FileInfo} returns this
 */
proto.pfs_v2.FileInfo.prototype.setHash = function(value) {
  return jspb.Message.setProto3BytesField(this, 5, value);
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
proto.pfs_v2.CreateRepoRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.CreateRepoRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.CreateRepoRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CreateRepoRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs_v2.Repo.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 2, ""),
    update: jspb.Message.getBooleanFieldWithDefault(msg, 3, false)
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
 * @return {!proto.pfs_v2.CreateRepoRequest}
 */
proto.pfs_v2.CreateRepoRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.CreateRepoRequest;
  return proto.pfs_v2.CreateRepoRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.CreateRepoRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.CreateRepoRequest}
 */
proto.pfs_v2.CreateRepoRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Repo;
      reader.readMessage(value,proto.pfs_v2.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 3:
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
proto.pfs_v2.CreateRepoRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.CreateRepoRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.CreateRepoRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CreateRepoRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Repo.serializeBinaryToWriter
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getUpdate();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs_v2.Repo}
 */
proto.pfs_v2.CreateRepoRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs_v2.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Repo, 1));
};


/**
 * @param {?proto.pfs_v2.Repo|undefined} value
 * @return {!proto.pfs_v2.CreateRepoRequest} returns this
*/
proto.pfs_v2.CreateRepoRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CreateRepoRequest} returns this
 */
proto.pfs_v2.CreateRepoRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CreateRepoRequest.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string description = 2;
 * @return {string}
 */
proto.pfs_v2.CreateRepoRequest.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.CreateRepoRequest} returns this
 */
proto.pfs_v2.CreateRepoRequest.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional bool update = 3;
 * @return {boolean}
 */
proto.pfs_v2.CreateRepoRequest.prototype.getUpdate = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.CreateRepoRequest} returns this
 */
proto.pfs_v2.CreateRepoRequest.prototype.setUpdate = function(value) {
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
proto.pfs_v2.InspectRepoRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.InspectRepoRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.InspectRepoRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.InspectRepoRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs_v2.Repo.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.InspectRepoRequest}
 */
proto.pfs_v2.InspectRepoRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.InspectRepoRequest;
  return proto.pfs_v2.InspectRepoRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.InspectRepoRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.InspectRepoRequest}
 */
proto.pfs_v2.InspectRepoRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Repo;
      reader.readMessage(value,proto.pfs_v2.Repo.deserializeBinaryFromReader);
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
proto.pfs_v2.InspectRepoRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.InspectRepoRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.InspectRepoRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.InspectRepoRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Repo.serializeBinaryToWriter
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs_v2.Repo}
 */
proto.pfs_v2.InspectRepoRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs_v2.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Repo, 1));
};


/**
 * @param {?proto.pfs_v2.Repo|undefined} value
 * @return {!proto.pfs_v2.InspectRepoRequest} returns this
*/
proto.pfs_v2.InspectRepoRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.InspectRepoRequest} returns this
 */
proto.pfs_v2.InspectRepoRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.InspectRepoRequest.prototype.hasRepo = function() {
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
proto.pfs_v2.ListRepoRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.ListRepoRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.ListRepoRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ListRepoRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    type: jspb.Message.getFieldWithDefault(msg, 1, "")
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
 * @return {!proto.pfs_v2.ListRepoRequest}
 */
proto.pfs_v2.ListRepoRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.ListRepoRequest;
  return proto.pfs_v2.ListRepoRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.ListRepoRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.ListRepoRequest}
 */
proto.pfs_v2.ListRepoRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setType(value);
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
proto.pfs_v2.ListRepoRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.ListRepoRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.ListRepoRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ListRepoRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getType();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string type = 1;
 * @return {string}
 */
proto.pfs_v2.ListRepoRequest.prototype.getType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.ListRepoRequest} returns this
 */
proto.pfs_v2.ListRepoRequest.prototype.setType = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.ListRepoResponse.repeatedFields_ = [1];



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
proto.pfs_v2.ListRepoResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.ListRepoResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.ListRepoResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ListRepoResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    repoInfoList: jspb.Message.toObjectList(msg.getRepoInfoList(),
    proto.pfs_v2.RepoInfo.toObject, includeInstance)
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
 * @return {!proto.pfs_v2.ListRepoResponse}
 */
proto.pfs_v2.ListRepoResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.ListRepoResponse;
  return proto.pfs_v2.ListRepoResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.ListRepoResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.ListRepoResponse}
 */
proto.pfs_v2.ListRepoResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.RepoInfo;
      reader.readMessage(value,proto.pfs_v2.RepoInfo.deserializeBinaryFromReader);
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
proto.pfs_v2.ListRepoResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.ListRepoResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.ListRepoResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ListRepoResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepoInfoList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs_v2.RepoInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated RepoInfo repo_info = 1;
 * @return {!Array<!proto.pfs_v2.RepoInfo>}
 */
proto.pfs_v2.ListRepoResponse.prototype.getRepoInfoList = function() {
  return /** @type{!Array<!proto.pfs_v2.RepoInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.RepoInfo, 1));
};


/**
 * @param {!Array<!proto.pfs_v2.RepoInfo>} value
 * @return {!proto.pfs_v2.ListRepoResponse} returns this
*/
proto.pfs_v2.ListRepoResponse.prototype.setRepoInfoList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs_v2.RepoInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.RepoInfo}
 */
proto.pfs_v2.ListRepoResponse.prototype.addRepoInfo = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs_v2.RepoInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.ListRepoResponse} returns this
 */
proto.pfs_v2.ListRepoResponse.prototype.clearRepoInfoList = function() {
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
proto.pfs_v2.DeleteRepoRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.DeleteRepoRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.DeleteRepoRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.DeleteRepoRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs_v2.Repo.toObject(includeInstance, f),
    force: jspb.Message.getBooleanFieldWithDefault(msg, 2, false),
    all: jspb.Message.getBooleanFieldWithDefault(msg, 3, false)
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
 * @return {!proto.pfs_v2.DeleteRepoRequest}
 */
proto.pfs_v2.DeleteRepoRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.DeleteRepoRequest;
  return proto.pfs_v2.DeleteRepoRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.DeleteRepoRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.DeleteRepoRequest}
 */
proto.pfs_v2.DeleteRepoRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Repo;
      reader.readMessage(value,proto.pfs_v2.Repo.deserializeBinaryFromReader);
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
proto.pfs_v2.DeleteRepoRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.DeleteRepoRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.DeleteRepoRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.DeleteRepoRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Repo.serializeBinaryToWriter
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
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs_v2.Repo}
 */
proto.pfs_v2.DeleteRepoRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs_v2.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Repo, 1));
};


/**
 * @param {?proto.pfs_v2.Repo|undefined} value
 * @return {!proto.pfs_v2.DeleteRepoRequest} returns this
*/
proto.pfs_v2.DeleteRepoRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.DeleteRepoRequest} returns this
 */
proto.pfs_v2.DeleteRepoRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.DeleteRepoRequest.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool force = 2;
 * @return {boolean}
 */
proto.pfs_v2.DeleteRepoRequest.prototype.getForce = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.DeleteRepoRequest} returns this
 */
proto.pfs_v2.DeleteRepoRequest.prototype.setForce = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};


/**
 * optional bool all = 3;
 * @return {boolean}
 */
proto.pfs_v2.DeleteRepoRequest.prototype.getAll = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.DeleteRepoRequest} returns this
 */
proto.pfs_v2.DeleteRepoRequest.prototype.setAll = function(value) {
  return jspb.Message.setProto3BooleanField(this, 3, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.StartCommitRequest.repeatedFields_ = [4];



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
proto.pfs_v2.StartCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.StartCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.StartCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.StartCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    parent: (f = msg.getParent()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 2, ""),
    branch: (f = msg.getBranch()) && proto.pfs_v2.Branch.toObject(includeInstance, f),
    provenanceList: jspb.Message.toObjectList(msg.getProvenanceList(),
    proto.pfs_v2.CommitProvenance.toObject, includeInstance)
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
 * @return {!proto.pfs_v2.StartCommitRequest}
 */
proto.pfs_v2.StartCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.StartCommitRequest;
  return proto.pfs_v2.StartCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.StartCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.StartCommitRequest}
 */
proto.pfs_v2.StartCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setParent(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 3:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
      msg.setBranch(value);
      break;
    case 4:
      var value = new proto.pfs_v2.CommitProvenance;
      reader.readMessage(value,proto.pfs_v2.CommitProvenance.deserializeBinaryFromReader);
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
proto.pfs_v2.StartCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.StartCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.StartCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.StartCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getParent();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getDescription();
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
      proto.pfs_v2.Branch.serializeBinaryToWriter
    );
  }
  f = message.getProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      4,
      f,
      proto.pfs_v2.CommitProvenance.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit parent = 1;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.StartCommitRequest.prototype.getParent = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.StartCommitRequest} returns this
*/
proto.pfs_v2.StartCommitRequest.prototype.setParent = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.StartCommitRequest} returns this
 */
proto.pfs_v2.StartCommitRequest.prototype.clearParent = function() {
  return this.setParent(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.StartCommitRequest.prototype.hasParent = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string description = 2;
 * @return {string}
 */
proto.pfs_v2.StartCommitRequest.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.StartCommitRequest} returns this
 */
proto.pfs_v2.StartCommitRequest.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional Branch branch = 3;
 * @return {?proto.pfs_v2.Branch}
 */
proto.pfs_v2.StartCommitRequest.prototype.getBranch = function() {
  return /** @type{?proto.pfs_v2.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Branch, 3));
};


/**
 * @param {?proto.pfs_v2.Branch|undefined} value
 * @return {!proto.pfs_v2.StartCommitRequest} returns this
*/
proto.pfs_v2.StartCommitRequest.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.StartCommitRequest} returns this
 */
proto.pfs_v2.StartCommitRequest.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.StartCommitRequest.prototype.hasBranch = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * repeated CommitProvenance provenance = 4;
 * @return {!Array<!proto.pfs_v2.CommitProvenance>}
 */
proto.pfs_v2.StartCommitRequest.prototype.getProvenanceList = function() {
  return /** @type{!Array<!proto.pfs_v2.CommitProvenance>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.CommitProvenance, 4));
};


/**
 * @param {!Array<!proto.pfs_v2.CommitProvenance>} value
 * @return {!proto.pfs_v2.StartCommitRequest} returns this
*/
proto.pfs_v2.StartCommitRequest.prototype.setProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 4, value);
};


/**
 * @param {!proto.pfs_v2.CommitProvenance=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.CommitProvenance}
 */
proto.pfs_v2.StartCommitRequest.prototype.addProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 4, opt_value, proto.pfs_v2.CommitProvenance, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.StartCommitRequest} returns this
 */
proto.pfs_v2.StartCommitRequest.prototype.clearProvenanceList = function() {
  return this.setProvenanceList([]);
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
proto.pfs_v2.FinishCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.FinishCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.FinishCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.FinishCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 2, ""),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 3, 0),
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
 * @return {!proto.pfs_v2.FinishCommitRequest}
 */
proto.pfs_v2.FinishCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.FinishCommitRequest;
  return proto.pfs_v2.FinishCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.FinishCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.FinishCommitRequest}
 */
proto.pfs_v2.FinishCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 3:
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
proto.pfs_v2.FinishCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.FinishCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.FinishCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.FinishCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
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
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.FinishCommitRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.FinishCommitRequest} returns this
*/
proto.pfs_v2.FinishCommitRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.FinishCommitRequest} returns this
 */
proto.pfs_v2.FinishCommitRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.FinishCommitRequest.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string description = 2;
 * @return {string}
 */
proto.pfs_v2.FinishCommitRequest.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.FinishCommitRequest} returns this
 */
proto.pfs_v2.FinishCommitRequest.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional uint64 size_bytes = 3;
 * @return {number}
 */
proto.pfs_v2.FinishCommitRequest.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.FinishCommitRequest} returns this
 */
proto.pfs_v2.FinishCommitRequest.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 3, value);
};


/**
 * optional bool empty = 4;
 * @return {boolean}
 */
proto.pfs_v2.FinishCommitRequest.prototype.getEmpty = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.FinishCommitRequest} returns this
 */
proto.pfs_v2.FinishCommitRequest.prototype.setEmpty = function(value) {
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
proto.pfs_v2.InspectCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.InspectCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.InspectCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.InspectCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
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
 * @return {!proto.pfs_v2.InspectCommitRequest}
 */
proto.pfs_v2.InspectCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.InspectCommitRequest;
  return proto.pfs_v2.InspectCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.InspectCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.InspectCommitRequest}
 */
proto.pfs_v2.InspectCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = /** @type {!proto.pfs_v2.CommitState} */ (reader.readEnum());
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
proto.pfs_v2.InspectCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.InspectCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.InspectCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.InspectCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
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
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.InspectCommitRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.InspectCommitRequest} returns this
*/
proto.pfs_v2.InspectCommitRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.InspectCommitRequest} returns this
 */
proto.pfs_v2.InspectCommitRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.InspectCommitRequest.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional CommitState block_state = 2;
 * @return {!proto.pfs_v2.CommitState}
 */
proto.pfs_v2.InspectCommitRequest.prototype.getBlockState = function() {
  return /** @type {!proto.pfs_v2.CommitState} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.pfs_v2.CommitState} value
 * @return {!proto.pfs_v2.InspectCommitRequest} returns this
 */
proto.pfs_v2.InspectCommitRequest.prototype.setBlockState = function(value) {
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
proto.pfs_v2.ListCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.ListCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.ListCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ListCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs_v2.Repo.toObject(includeInstance, f),
    from: (f = msg.getFrom()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    to: (f = msg.getTo()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
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
 * @return {!proto.pfs_v2.ListCommitRequest}
 */
proto.pfs_v2.ListCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.ListCommitRequest;
  return proto.pfs_v2.ListCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.ListCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.ListCommitRequest}
 */
proto.pfs_v2.ListCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Repo;
      reader.readMessage(value,proto.pfs_v2.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
      break;
    case 2:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setFrom(value);
      break;
    case 3:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
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
proto.pfs_v2.ListCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.ListCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.ListCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ListCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Repo.serializeBinaryToWriter
    );
  }
  f = message.getFrom();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getTo();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
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
 * @return {?proto.pfs_v2.Repo}
 */
proto.pfs_v2.ListCommitRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs_v2.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Repo, 1));
};


/**
 * @param {?proto.pfs_v2.Repo|undefined} value
 * @return {!proto.pfs_v2.ListCommitRequest} returns this
*/
proto.pfs_v2.ListCommitRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.ListCommitRequest} returns this
 */
proto.pfs_v2.ListCommitRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.ListCommitRequest.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Commit from = 2;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.ListCommitRequest.prototype.getFrom = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 2));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.ListCommitRequest} returns this
*/
proto.pfs_v2.ListCommitRequest.prototype.setFrom = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.ListCommitRequest} returns this
 */
proto.pfs_v2.ListCommitRequest.prototype.clearFrom = function() {
  return this.setFrom(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.ListCommitRequest.prototype.hasFrom = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional Commit to = 3;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.ListCommitRequest.prototype.getTo = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 3));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.ListCommitRequest} returns this
*/
proto.pfs_v2.ListCommitRequest.prototype.setTo = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.ListCommitRequest} returns this
 */
proto.pfs_v2.ListCommitRequest.prototype.clearTo = function() {
  return this.setTo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.ListCommitRequest.prototype.hasTo = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional uint64 number = 4;
 * @return {number}
 */
proto.pfs_v2.ListCommitRequest.prototype.getNumber = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.ListCommitRequest} returns this
 */
proto.pfs_v2.ListCommitRequest.prototype.setNumber = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional bool reverse = 5;
 * @return {boolean}
 */
proto.pfs_v2.ListCommitRequest.prototype.getReverse = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 5, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.ListCommitRequest} returns this
 */
proto.pfs_v2.ListCommitRequest.prototype.setReverse = function(value) {
  return jspb.Message.setProto3BooleanField(this, 5, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.CommitInfos.repeatedFields_ = [1];



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
proto.pfs_v2.CommitInfos.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.CommitInfos.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.CommitInfos} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CommitInfos.toObject = function(includeInstance, msg) {
  var f, obj = {
    commitInfoList: jspb.Message.toObjectList(msg.getCommitInfoList(),
    proto.pfs_v2.CommitInfo.toObject, includeInstance)
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
 * @return {!proto.pfs_v2.CommitInfos}
 */
proto.pfs_v2.CommitInfos.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.CommitInfos;
  return proto.pfs_v2.CommitInfos.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.CommitInfos} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.CommitInfos}
 */
proto.pfs_v2.CommitInfos.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.CommitInfo;
      reader.readMessage(value,proto.pfs_v2.CommitInfo.deserializeBinaryFromReader);
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
proto.pfs_v2.CommitInfos.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.CommitInfos.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.CommitInfos} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CommitInfos.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommitInfoList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs_v2.CommitInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated CommitInfo commit_info = 1;
 * @return {!Array<!proto.pfs_v2.CommitInfo>}
 */
proto.pfs_v2.CommitInfos.prototype.getCommitInfoList = function() {
  return /** @type{!Array<!proto.pfs_v2.CommitInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.CommitInfo, 1));
};


/**
 * @param {!Array<!proto.pfs_v2.CommitInfo>} value
 * @return {!proto.pfs_v2.CommitInfos} returns this
*/
proto.pfs_v2.CommitInfos.prototype.setCommitInfoList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs_v2.CommitInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.CommitInfo}
 */
proto.pfs_v2.CommitInfos.prototype.addCommitInfo = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs_v2.CommitInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.CommitInfos} returns this
 */
proto.pfs_v2.CommitInfos.prototype.clearCommitInfoList = function() {
  return this.setCommitInfoList([]);
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
proto.pfs_v2.SquashCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.SquashCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.SquashCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.SquashCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.SquashCommitRequest}
 */
proto.pfs_v2.SquashCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.SquashCommitRequest;
  return proto.pfs_v2.SquashCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.SquashCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.SquashCommitRequest}
 */
proto.pfs_v2.SquashCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
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
proto.pfs_v2.SquashCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.SquashCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.SquashCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.SquashCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.SquashCommitRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.SquashCommitRequest} returns this
*/
proto.pfs_v2.SquashCommitRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.SquashCommitRequest} returns this
 */
proto.pfs_v2.SquashCommitRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.SquashCommitRequest.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.FlushCommitRequest.repeatedFields_ = [1,2];



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
proto.pfs_v2.FlushCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.FlushCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.FlushCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.FlushCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commitsList: jspb.Message.toObjectList(msg.getCommitsList(),
    proto.pfs_v2.Commit.toObject, includeInstance),
    toReposList: jspb.Message.toObjectList(msg.getToReposList(),
    proto.pfs_v2.Repo.toObject, includeInstance)
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
 * @return {!proto.pfs_v2.FlushCommitRequest}
 */
proto.pfs_v2.FlushCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.FlushCommitRequest;
  return proto.pfs_v2.FlushCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.FlushCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.FlushCommitRequest}
 */
proto.pfs_v2.FlushCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.addCommits(value);
      break;
    case 2:
      var value = new proto.pfs_v2.Repo;
      reader.readMessage(value,proto.pfs_v2.Repo.deserializeBinaryFromReader);
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
proto.pfs_v2.FlushCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.FlushCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.FlushCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.FlushCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommitsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getToReposList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.pfs_v2.Repo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Commit commits = 1;
 * @return {!Array<!proto.pfs_v2.Commit>}
 */
proto.pfs_v2.FlushCommitRequest.prototype.getCommitsList = function() {
  return /** @type{!Array<!proto.pfs_v2.Commit>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {!Array<!proto.pfs_v2.Commit>} value
 * @return {!proto.pfs_v2.FlushCommitRequest} returns this
*/
proto.pfs_v2.FlushCommitRequest.prototype.setCommitsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs_v2.Commit=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.Commit}
 */
proto.pfs_v2.FlushCommitRequest.prototype.addCommits = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs_v2.Commit, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.FlushCommitRequest} returns this
 */
proto.pfs_v2.FlushCommitRequest.prototype.clearCommitsList = function() {
  return this.setCommitsList([]);
};


/**
 * repeated Repo to_repos = 2;
 * @return {!Array<!proto.pfs_v2.Repo>}
 */
proto.pfs_v2.FlushCommitRequest.prototype.getToReposList = function() {
  return /** @type{!Array<!proto.pfs_v2.Repo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.Repo, 2));
};


/**
 * @param {!Array<!proto.pfs_v2.Repo>} value
 * @return {!proto.pfs_v2.FlushCommitRequest} returns this
*/
proto.pfs_v2.FlushCommitRequest.prototype.setToReposList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pfs_v2.Repo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.Repo}
 */
proto.pfs_v2.FlushCommitRequest.prototype.addToRepos = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pfs_v2.Repo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.FlushCommitRequest} returns this
 */
proto.pfs_v2.FlushCommitRequest.prototype.clearToReposList = function() {
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
proto.pfs_v2.SubscribeCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.SubscribeCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.SubscribeCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.SubscribeCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs_v2.Repo.toObject(includeInstance, f),
    branch: jspb.Message.getFieldWithDefault(msg, 2, ""),
    prov: (f = msg.getProv()) && proto.pfs_v2.CommitProvenance.toObject(includeInstance, f),
    from: (f = msg.getFrom()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    state: jspb.Message.getFieldWithDefault(msg, 5, 0)
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
 * @return {!proto.pfs_v2.SubscribeCommitRequest}
 */
proto.pfs_v2.SubscribeCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.SubscribeCommitRequest;
  return proto.pfs_v2.SubscribeCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.SubscribeCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.SubscribeCommitRequest}
 */
proto.pfs_v2.SubscribeCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Repo;
      reader.readMessage(value,proto.pfs_v2.Repo.deserializeBinaryFromReader);
      msg.setRepo(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setBranch(value);
      break;
    case 3:
      var value = new proto.pfs_v2.CommitProvenance;
      reader.readMessage(value,proto.pfs_v2.CommitProvenance.deserializeBinaryFromReader);
      msg.setProv(value);
      break;
    case 4:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setFrom(value);
      break;
    case 5:
      var value = /** @type {!proto.pfs_v2.CommitState} */ (reader.readEnum());
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
proto.pfs_v2.SubscribeCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.SubscribeCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.SubscribeCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.SubscribeCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Repo.serializeBinaryToWriter
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
      3,
      f,
      proto.pfs_v2.CommitProvenance.serializeBinaryToWriter
    );
  }
  f = message.getFrom();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      5,
      f
    );
  }
};


/**
 * optional Repo repo = 1;
 * @return {?proto.pfs_v2.Repo}
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs_v2.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Repo, 1));
};


/**
 * @param {?proto.pfs_v2.Repo|undefined} value
 * @return {!proto.pfs_v2.SubscribeCommitRequest} returns this
*/
proto.pfs_v2.SubscribeCommitRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.SubscribeCommitRequest} returns this
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string branch = 2;
 * @return {string}
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.getBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.SubscribeCommitRequest} returns this
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.setBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional CommitProvenance prov = 3;
 * @return {?proto.pfs_v2.CommitProvenance}
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.getProv = function() {
  return /** @type{?proto.pfs_v2.CommitProvenance} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.CommitProvenance, 3));
};


/**
 * @param {?proto.pfs_v2.CommitProvenance|undefined} value
 * @return {!proto.pfs_v2.SubscribeCommitRequest} returns this
*/
proto.pfs_v2.SubscribeCommitRequest.prototype.setProv = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.SubscribeCommitRequest} returns this
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.clearProv = function() {
  return this.setProv(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.hasProv = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional Commit from = 4;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.getFrom = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 4));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.SubscribeCommitRequest} returns this
*/
proto.pfs_v2.SubscribeCommitRequest.prototype.setFrom = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.SubscribeCommitRequest} returns this
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.clearFrom = function() {
  return this.setFrom(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.hasFrom = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional CommitState state = 5;
 * @return {!proto.pfs_v2.CommitState}
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.getState = function() {
  return /** @type {!proto.pfs_v2.CommitState} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {!proto.pfs_v2.CommitState} value
 * @return {!proto.pfs_v2.SubscribeCommitRequest} returns this
 */
proto.pfs_v2.SubscribeCommitRequest.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 5, value);
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
proto.pfs_v2.ClearCommitRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.ClearCommitRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.ClearCommitRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ClearCommitRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.ClearCommitRequest}
 */
proto.pfs_v2.ClearCommitRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.ClearCommitRequest;
  return proto.pfs_v2.ClearCommitRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.ClearCommitRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.ClearCommitRequest}
 */
proto.pfs_v2.ClearCommitRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
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
proto.pfs_v2.ClearCommitRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.ClearCommitRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.ClearCommitRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ClearCommitRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.ClearCommitRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.ClearCommitRequest} returns this
*/
proto.pfs_v2.ClearCommitRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.ClearCommitRequest} returns this
 */
proto.pfs_v2.ClearCommitRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.ClearCommitRequest.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pfs_v2.CreateBranchRequest.repeatedFields_ = [3];



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
proto.pfs_v2.CreateBranchRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.CreateBranchRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.CreateBranchRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CreateBranchRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    head: (f = msg.getHead()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    branch: (f = msg.getBranch()) && proto.pfs_v2.Branch.toObject(includeInstance, f),
    provenanceList: jspb.Message.toObjectList(msg.getProvenanceList(),
    proto.pfs_v2.Branch.toObject, includeInstance),
    trigger: (f = msg.getTrigger()) && proto.pfs_v2.Trigger.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.CreateBranchRequest}
 */
proto.pfs_v2.CreateBranchRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.CreateBranchRequest;
  return proto.pfs_v2.CreateBranchRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.CreateBranchRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.CreateBranchRequest}
 */
proto.pfs_v2.CreateBranchRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setHead(value);
      break;
    case 2:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
      msg.setBranch(value);
      break;
    case 3:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
      msg.addProvenance(value);
      break;
    case 4:
      var value = new proto.pfs_v2.Trigger;
      reader.readMessage(value,proto.pfs_v2.Trigger.deserializeBinaryFromReader);
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
proto.pfs_v2.CreateBranchRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.CreateBranchRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.CreateBranchRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CreateBranchRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getHead();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs_v2.Branch.serializeBinaryToWriter
    );
  }
  f = message.getProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      3,
      f,
      proto.pfs_v2.Branch.serializeBinaryToWriter
    );
  }
  f = message.getTrigger();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pfs_v2.Trigger.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit head = 1;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.CreateBranchRequest.prototype.getHead = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.CreateBranchRequest} returns this
*/
proto.pfs_v2.CreateBranchRequest.prototype.setHead = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CreateBranchRequest} returns this
 */
proto.pfs_v2.CreateBranchRequest.prototype.clearHead = function() {
  return this.setHead(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CreateBranchRequest.prototype.hasHead = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Branch branch = 2;
 * @return {?proto.pfs_v2.Branch}
 */
proto.pfs_v2.CreateBranchRequest.prototype.getBranch = function() {
  return /** @type{?proto.pfs_v2.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Branch, 2));
};


/**
 * @param {?proto.pfs_v2.Branch|undefined} value
 * @return {!proto.pfs_v2.CreateBranchRequest} returns this
*/
proto.pfs_v2.CreateBranchRequest.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CreateBranchRequest} returns this
 */
proto.pfs_v2.CreateBranchRequest.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CreateBranchRequest.prototype.hasBranch = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * repeated Branch provenance = 3;
 * @return {!Array<!proto.pfs_v2.Branch>}
 */
proto.pfs_v2.CreateBranchRequest.prototype.getProvenanceList = function() {
  return /** @type{!Array<!proto.pfs_v2.Branch>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pfs_v2.Branch, 3));
};


/**
 * @param {!Array<!proto.pfs_v2.Branch>} value
 * @return {!proto.pfs_v2.CreateBranchRequest} returns this
*/
proto.pfs_v2.CreateBranchRequest.prototype.setProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 3, value);
};


/**
 * @param {!proto.pfs_v2.Branch=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.Branch}
 */
proto.pfs_v2.CreateBranchRequest.prototype.addProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 3, opt_value, proto.pfs_v2.Branch, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pfs_v2.CreateBranchRequest} returns this
 */
proto.pfs_v2.CreateBranchRequest.prototype.clearProvenanceList = function() {
  return this.setProvenanceList([]);
};


/**
 * optional Trigger trigger = 4;
 * @return {?proto.pfs_v2.Trigger}
 */
proto.pfs_v2.CreateBranchRequest.prototype.getTrigger = function() {
  return /** @type{?proto.pfs_v2.Trigger} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Trigger, 4));
};


/**
 * @param {?proto.pfs_v2.Trigger|undefined} value
 * @return {!proto.pfs_v2.CreateBranchRequest} returns this
*/
proto.pfs_v2.CreateBranchRequest.prototype.setTrigger = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CreateBranchRequest} returns this
 */
proto.pfs_v2.CreateBranchRequest.prototype.clearTrigger = function() {
  return this.setTrigger(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CreateBranchRequest.prototype.hasTrigger = function() {
  return jspb.Message.getField(this, 4) != null;
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
proto.pfs_v2.InspectBranchRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.InspectBranchRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.InspectBranchRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.InspectBranchRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    branch: (f = msg.getBranch()) && proto.pfs_v2.Branch.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.InspectBranchRequest}
 */
proto.pfs_v2.InspectBranchRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.InspectBranchRequest;
  return proto.pfs_v2.InspectBranchRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.InspectBranchRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.InspectBranchRequest}
 */
proto.pfs_v2.InspectBranchRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
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
proto.pfs_v2.InspectBranchRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.InspectBranchRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.InspectBranchRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.InspectBranchRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Branch.serializeBinaryToWriter
    );
  }
};


/**
 * optional Branch branch = 1;
 * @return {?proto.pfs_v2.Branch}
 */
proto.pfs_v2.InspectBranchRequest.prototype.getBranch = function() {
  return /** @type{?proto.pfs_v2.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Branch, 1));
};


/**
 * @param {?proto.pfs_v2.Branch|undefined} value
 * @return {!proto.pfs_v2.InspectBranchRequest} returns this
*/
proto.pfs_v2.InspectBranchRequest.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.InspectBranchRequest} returns this
 */
proto.pfs_v2.InspectBranchRequest.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.InspectBranchRequest.prototype.hasBranch = function() {
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
proto.pfs_v2.ListBranchRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.ListBranchRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.ListBranchRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ListBranchRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    repo: (f = msg.getRepo()) && proto.pfs_v2.Repo.toObject(includeInstance, f),
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
 * @return {!proto.pfs_v2.ListBranchRequest}
 */
proto.pfs_v2.ListBranchRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.ListBranchRequest;
  return proto.pfs_v2.ListBranchRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.ListBranchRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.ListBranchRequest}
 */
proto.pfs_v2.ListBranchRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Repo;
      reader.readMessage(value,proto.pfs_v2.Repo.deserializeBinaryFromReader);
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
proto.pfs_v2.ListBranchRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.ListBranchRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.ListBranchRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ListBranchRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRepo();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Repo.serializeBinaryToWriter
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
 * @return {?proto.pfs_v2.Repo}
 */
proto.pfs_v2.ListBranchRequest.prototype.getRepo = function() {
  return /** @type{?proto.pfs_v2.Repo} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Repo, 1));
};


/**
 * @param {?proto.pfs_v2.Repo|undefined} value
 * @return {!proto.pfs_v2.ListBranchRequest} returns this
*/
proto.pfs_v2.ListBranchRequest.prototype.setRepo = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.ListBranchRequest} returns this
 */
proto.pfs_v2.ListBranchRequest.prototype.clearRepo = function() {
  return this.setRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.ListBranchRequest.prototype.hasRepo = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool reverse = 2;
 * @return {boolean}
 */
proto.pfs_v2.ListBranchRequest.prototype.getReverse = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.ListBranchRequest} returns this
 */
proto.pfs_v2.ListBranchRequest.prototype.setReverse = function(value) {
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
proto.pfs_v2.DeleteBranchRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.DeleteBranchRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.DeleteBranchRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.DeleteBranchRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    branch: (f = msg.getBranch()) && proto.pfs_v2.Branch.toObject(includeInstance, f),
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
 * @return {!proto.pfs_v2.DeleteBranchRequest}
 */
proto.pfs_v2.DeleteBranchRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.DeleteBranchRequest;
  return proto.pfs_v2.DeleteBranchRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.DeleteBranchRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.DeleteBranchRequest}
 */
proto.pfs_v2.DeleteBranchRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
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
proto.pfs_v2.DeleteBranchRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.DeleteBranchRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.DeleteBranchRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.DeleteBranchRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Branch.serializeBinaryToWriter
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
 * @return {?proto.pfs_v2.Branch}
 */
proto.pfs_v2.DeleteBranchRequest.prototype.getBranch = function() {
  return /** @type{?proto.pfs_v2.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Branch, 1));
};


/**
 * @param {?proto.pfs_v2.Branch|undefined} value
 * @return {!proto.pfs_v2.DeleteBranchRequest} returns this
*/
proto.pfs_v2.DeleteBranchRequest.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.DeleteBranchRequest} returns this
 */
proto.pfs_v2.DeleteBranchRequest.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.DeleteBranchRequest.prototype.hasBranch = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool force = 2;
 * @return {boolean}
 */
proto.pfs_v2.DeleteBranchRequest.prototype.getForce = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.DeleteBranchRequest} returns this
 */
proto.pfs_v2.DeleteBranchRequest.prototype.setForce = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};



/**
 * Oneof group definitions for this message. Each group defines the field
 * numbers belonging to that group. When of these fields' value is set, all
 * other fields in the group are cleared. During deserialization, if multiple
 * fields are encountered for a group, only the last value seen will be kept.
 * @private {!Array<!Array<number>>}
 * @const
 */
proto.pfs_v2.PutFile.oneofGroups_ = [[3,4,5]];

/**
 * @enum {number}
 */
proto.pfs_v2.PutFile.SourceCase = {
  SOURCE_NOT_SET: 0,
  RAW_FILE_SOURCE: 3,
  TAR_FILE_SOURCE: 4,
  URL_FILE_SOURCE: 5
};

/**
 * @return {proto.pfs_v2.PutFile.SourceCase}
 */
proto.pfs_v2.PutFile.prototype.getSourceCase = function() {
  return /** @type {proto.pfs_v2.PutFile.SourceCase} */(jspb.Message.computeOneofCase(this, proto.pfs_v2.PutFile.oneofGroups_[0]));
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
proto.pfs_v2.PutFile.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.PutFile.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.PutFile} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.PutFile.toObject = function(includeInstance, msg) {
  var f, obj = {
    append: jspb.Message.getBooleanFieldWithDefault(msg, 1, false),
    tag: jspb.Message.getFieldWithDefault(msg, 2, ""),
    rawFileSource: (f = msg.getRawFileSource()) && proto.pfs_v2.RawFileSource.toObject(includeInstance, f),
    tarFileSource: (f = msg.getTarFileSource()) && proto.pfs_v2.TarFileSource.toObject(includeInstance, f),
    urlFileSource: (f = msg.getUrlFileSource()) && proto.pfs_v2.URLFileSource.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.PutFile}
 */
proto.pfs_v2.PutFile.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.PutFile;
  return proto.pfs_v2.PutFile.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.PutFile} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.PutFile}
 */
proto.pfs_v2.PutFile.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAppend(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setTag(value);
      break;
    case 3:
      var value = new proto.pfs_v2.RawFileSource;
      reader.readMessage(value,proto.pfs_v2.RawFileSource.deserializeBinaryFromReader);
      msg.setRawFileSource(value);
      break;
    case 4:
      var value = new proto.pfs_v2.TarFileSource;
      reader.readMessage(value,proto.pfs_v2.TarFileSource.deserializeBinaryFromReader);
      msg.setTarFileSource(value);
      break;
    case 5:
      var value = new proto.pfs_v2.URLFileSource;
      reader.readMessage(value,proto.pfs_v2.URLFileSource.deserializeBinaryFromReader);
      msg.setUrlFileSource(value);
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
proto.pfs_v2.PutFile.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.PutFile.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.PutFile} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.PutFile.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAppend();
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
  f = message.getRawFileSource();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pfs_v2.RawFileSource.serializeBinaryToWriter
    );
  }
  f = message.getTarFileSource();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pfs_v2.TarFileSource.serializeBinaryToWriter
    );
  }
  f = message.getUrlFileSource();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.pfs_v2.URLFileSource.serializeBinaryToWriter
    );
  }
};


/**
 * optional bool append = 1;
 * @return {boolean}
 */
proto.pfs_v2.PutFile.prototype.getAppend = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.PutFile} returns this
 */
proto.pfs_v2.PutFile.prototype.setAppend = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};


/**
 * optional string tag = 2;
 * @return {string}
 */
proto.pfs_v2.PutFile.prototype.getTag = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.PutFile} returns this
 */
proto.pfs_v2.PutFile.prototype.setTag = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional RawFileSource raw_file_source = 3;
 * @return {?proto.pfs_v2.RawFileSource}
 */
proto.pfs_v2.PutFile.prototype.getRawFileSource = function() {
  return /** @type{?proto.pfs_v2.RawFileSource} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.RawFileSource, 3));
};


/**
 * @param {?proto.pfs_v2.RawFileSource|undefined} value
 * @return {!proto.pfs_v2.PutFile} returns this
*/
proto.pfs_v2.PutFile.prototype.setRawFileSource = function(value) {
  return jspb.Message.setOneofWrapperField(this, 3, proto.pfs_v2.PutFile.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.PutFile} returns this
 */
proto.pfs_v2.PutFile.prototype.clearRawFileSource = function() {
  return this.setRawFileSource(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.PutFile.prototype.hasRawFileSource = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional TarFileSource tar_file_source = 4;
 * @return {?proto.pfs_v2.TarFileSource}
 */
proto.pfs_v2.PutFile.prototype.getTarFileSource = function() {
  return /** @type{?proto.pfs_v2.TarFileSource} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.TarFileSource, 4));
};


/**
 * @param {?proto.pfs_v2.TarFileSource|undefined} value
 * @return {!proto.pfs_v2.PutFile} returns this
*/
proto.pfs_v2.PutFile.prototype.setTarFileSource = function(value) {
  return jspb.Message.setOneofWrapperField(this, 4, proto.pfs_v2.PutFile.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.PutFile} returns this
 */
proto.pfs_v2.PutFile.prototype.clearTarFileSource = function() {
  return this.setTarFileSource(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.PutFile.prototype.hasTarFileSource = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional URLFileSource url_file_source = 5;
 * @return {?proto.pfs_v2.URLFileSource}
 */
proto.pfs_v2.PutFile.prototype.getUrlFileSource = function() {
  return /** @type{?proto.pfs_v2.URLFileSource} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.URLFileSource, 5));
};


/**
 * @param {?proto.pfs_v2.URLFileSource|undefined} value
 * @return {!proto.pfs_v2.PutFile} returns this
*/
proto.pfs_v2.PutFile.prototype.setUrlFileSource = function(value) {
  return jspb.Message.setOneofWrapperField(this, 5, proto.pfs_v2.PutFile.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.PutFile} returns this
 */
proto.pfs_v2.PutFile.prototype.clearUrlFileSource = function() {
  return this.setUrlFileSource(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.PutFile.prototype.hasUrlFileSource = function() {
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
proto.pfs_v2.RawFileSource.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.RawFileSource.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.RawFileSource} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RawFileSource.toObject = function(includeInstance, msg) {
  var f, obj = {
    path: jspb.Message.getFieldWithDefault(msg, 1, ""),
    data: msg.getData_asB64(),
    eof: jspb.Message.getBooleanFieldWithDefault(msg, 3, false)
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
 * @return {!proto.pfs_v2.RawFileSource}
 */
proto.pfs_v2.RawFileSource.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.RawFileSource;
  return proto.pfs_v2.RawFileSource.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.RawFileSource} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.RawFileSource}
 */
proto.pfs_v2.RawFileSource.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setPath(value);
      break;
    case 2:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setData(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEof(value);
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
proto.pfs_v2.RawFileSource.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.RawFileSource.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.RawFileSource} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RawFileSource.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getData_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      2,
      f
    );
  }
  f = message.getEof();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
};


/**
 * optional string path = 1;
 * @return {string}
 */
proto.pfs_v2.RawFileSource.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.RawFileSource} returns this
 */
proto.pfs_v2.RawFileSource.prototype.setPath = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional bytes data = 2;
 * @return {string}
 */
proto.pfs_v2.RawFileSource.prototype.getData = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * optional bytes data = 2;
 * This is a type-conversion wrapper around `getData()`
 * @return {string}
 */
proto.pfs_v2.RawFileSource.prototype.getData_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getData()));
};


/**
 * optional bytes data = 2;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getData()`
 * @return {!Uint8Array}
 */
proto.pfs_v2.RawFileSource.prototype.getData_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getData()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pfs_v2.RawFileSource} returns this
 */
proto.pfs_v2.RawFileSource.prototype.setData = function(value) {
  return jspb.Message.setProto3BytesField(this, 2, value);
};


/**
 * optional bool EOF = 3;
 * @return {boolean}
 */
proto.pfs_v2.RawFileSource.prototype.getEof = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.RawFileSource} returns this
 */
proto.pfs_v2.RawFileSource.prototype.setEof = function(value) {
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
proto.pfs_v2.TarFileSource.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.TarFileSource.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.TarFileSource} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.TarFileSource.toObject = function(includeInstance, msg) {
  var f, obj = {
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
 * @return {!proto.pfs_v2.TarFileSource}
 */
proto.pfs_v2.TarFileSource.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.TarFileSource;
  return proto.pfs_v2.TarFileSource.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.TarFileSource} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.TarFileSource}
 */
proto.pfs_v2.TarFileSource.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
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
proto.pfs_v2.TarFileSource.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.TarFileSource.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.TarFileSource} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.TarFileSource.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getData_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      1,
      f
    );
  }
};


/**
 * optional bytes data = 1;
 * @return {string}
 */
proto.pfs_v2.TarFileSource.prototype.getData = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * optional bytes data = 1;
 * This is a type-conversion wrapper around `getData()`
 * @return {string}
 */
proto.pfs_v2.TarFileSource.prototype.getData_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getData()));
};


/**
 * optional bytes data = 1;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getData()`
 * @return {!Uint8Array}
 */
proto.pfs_v2.TarFileSource.prototype.getData_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getData()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pfs_v2.TarFileSource} returns this
 */
proto.pfs_v2.TarFileSource.prototype.setData = function(value) {
  return jspb.Message.setProto3BytesField(this, 1, value);
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
proto.pfs_v2.URLFileSource.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.URLFileSource.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.URLFileSource} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.URLFileSource.toObject = function(includeInstance, msg) {
  var f, obj = {
    path: jspb.Message.getFieldWithDefault(msg, 1, ""),
    url: jspb.Message.getFieldWithDefault(msg, 2, ""),
    recursive: jspb.Message.getBooleanFieldWithDefault(msg, 3, false)
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
 * @return {!proto.pfs_v2.URLFileSource}
 */
proto.pfs_v2.URLFileSource.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.URLFileSource;
  return proto.pfs_v2.URLFileSource.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.URLFileSource} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.URLFileSource}
 */
proto.pfs_v2.URLFileSource.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setPath(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setUrl(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setRecursive(value);
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
proto.pfs_v2.URLFileSource.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.URLFileSource.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.URLFileSource} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.URLFileSource.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getUrl();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getRecursive();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
};


/**
 * optional string path = 1;
 * @return {string}
 */
proto.pfs_v2.URLFileSource.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.URLFileSource} returns this
 */
proto.pfs_v2.URLFileSource.prototype.setPath = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string URL = 2;
 * @return {string}
 */
proto.pfs_v2.URLFileSource.prototype.getUrl = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.URLFileSource} returns this
 */
proto.pfs_v2.URLFileSource.prototype.setUrl = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional bool recursive = 3;
 * @return {boolean}
 */
proto.pfs_v2.URLFileSource.prototype.getRecursive = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.URLFileSource} returns this
 */
proto.pfs_v2.URLFileSource.prototype.setRecursive = function(value) {
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
proto.pfs_v2.DeleteFile.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.DeleteFile.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.DeleteFile} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.DeleteFile.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: jspb.Message.getFieldWithDefault(msg, 1, ""),
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
 * @return {!proto.pfs_v2.DeleteFile}
 */
proto.pfs_v2.DeleteFile.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.DeleteFile;
  return proto.pfs_v2.DeleteFile.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.DeleteFile} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.DeleteFile}
 */
proto.pfs_v2.DeleteFile.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setFile(value);
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
proto.pfs_v2.DeleteFile.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.DeleteFile.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.DeleteFile} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.DeleteFile.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f.length > 0) {
    writer.writeString(
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
 * optional string file = 1;
 * @return {string}
 */
proto.pfs_v2.DeleteFile.prototype.getFile = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.DeleteFile} returns this
 */
proto.pfs_v2.DeleteFile.prototype.setFile = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string tag = 2;
 * @return {string}
 */
proto.pfs_v2.DeleteFile.prototype.getTag = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.DeleteFile} returns this
 */
proto.pfs_v2.DeleteFile.prototype.setTag = function(value) {
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
proto.pfs_v2.CopyFile.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.CopyFile.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.CopyFile} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CopyFile.toObject = function(includeInstance, msg) {
  var f, obj = {
    append: jspb.Message.getBooleanFieldWithDefault(msg, 1, false),
    tag: jspb.Message.getFieldWithDefault(msg, 2, ""),
    dst: jspb.Message.getFieldWithDefault(msg, 3, ""),
    src: (f = msg.getSrc()) && proto.pfs_v2.File.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.CopyFile}
 */
proto.pfs_v2.CopyFile.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.CopyFile;
  return proto.pfs_v2.CopyFile.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.CopyFile} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.CopyFile}
 */
proto.pfs_v2.CopyFile.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAppend(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setTag(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setDst(value);
      break;
    case 4:
      var value = new proto.pfs_v2.File;
      reader.readMessage(value,proto.pfs_v2.File.deserializeBinaryFromReader);
      msg.setSrc(value);
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
proto.pfs_v2.CopyFile.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.CopyFile.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.CopyFile} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CopyFile.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAppend();
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
  f = message.getDst();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getSrc();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pfs_v2.File.serializeBinaryToWriter
    );
  }
};


/**
 * optional bool append = 1;
 * @return {boolean}
 */
proto.pfs_v2.CopyFile.prototype.getAppend = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.CopyFile} returns this
 */
proto.pfs_v2.CopyFile.prototype.setAppend = function(value) {
  return jspb.Message.setProto3BooleanField(this, 1, value);
};


/**
 * optional string tag = 2;
 * @return {string}
 */
proto.pfs_v2.CopyFile.prototype.getTag = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.CopyFile} returns this
 */
proto.pfs_v2.CopyFile.prototype.setTag = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string dst = 3;
 * @return {string}
 */
proto.pfs_v2.CopyFile.prototype.getDst = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.CopyFile} returns this
 */
proto.pfs_v2.CopyFile.prototype.setDst = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional File src = 4;
 * @return {?proto.pfs_v2.File}
 */
proto.pfs_v2.CopyFile.prototype.getSrc = function() {
  return /** @type{?proto.pfs_v2.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.File, 4));
};


/**
 * @param {?proto.pfs_v2.File|undefined} value
 * @return {!proto.pfs_v2.CopyFile} returns this
*/
proto.pfs_v2.CopyFile.prototype.setSrc = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.CopyFile} returns this
 */
proto.pfs_v2.CopyFile.prototype.clearSrc = function() {
  return this.setSrc(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.CopyFile.prototype.hasSrc = function() {
  return jspb.Message.getField(this, 4) != null;
};



/**
 * Oneof group definitions for this message. Each group defines the field
 * numbers belonging to that group. When of these fields' value is set, all
 * other fields in the group are cleared. During deserialization, if multiple
 * fields are encountered for a group, only the last value seen will be kept.
 * @private {!Array<!Array<number>>}
 * @const
 */
proto.pfs_v2.ModifyFileRequest.oneofGroups_ = [[2,3,4]];

/**
 * @enum {number}
 */
proto.pfs_v2.ModifyFileRequest.ModificationCase = {
  MODIFICATION_NOT_SET: 0,
  PUT_FILE: 2,
  DELETE_FILE: 3,
  COPY_FILE: 4
};

/**
 * @return {proto.pfs_v2.ModifyFileRequest.ModificationCase}
 */
proto.pfs_v2.ModifyFileRequest.prototype.getModificationCase = function() {
  return /** @type {proto.pfs_v2.ModifyFileRequest.ModificationCase} */(jspb.Message.computeOneofCase(this, proto.pfs_v2.ModifyFileRequest.oneofGroups_[0]));
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
proto.pfs_v2.ModifyFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.ModifyFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.ModifyFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ModifyFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    putFile: (f = msg.getPutFile()) && proto.pfs_v2.PutFile.toObject(includeInstance, f),
    deleteFile: (f = msg.getDeleteFile()) && proto.pfs_v2.DeleteFile.toObject(includeInstance, f),
    copyFile: (f = msg.getCopyFile()) && proto.pfs_v2.CopyFile.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.ModifyFileRequest}
 */
proto.pfs_v2.ModifyFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.ModifyFileRequest;
  return proto.pfs_v2.ModifyFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.ModifyFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.ModifyFileRequest}
 */
proto.pfs_v2.ModifyFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = new proto.pfs_v2.PutFile;
      reader.readMessage(value,proto.pfs_v2.PutFile.deserializeBinaryFromReader);
      msg.setPutFile(value);
      break;
    case 3:
      var value = new proto.pfs_v2.DeleteFile;
      reader.readMessage(value,proto.pfs_v2.DeleteFile.deserializeBinaryFromReader);
      msg.setDeleteFile(value);
      break;
    case 4:
      var value = new proto.pfs_v2.CopyFile;
      reader.readMessage(value,proto.pfs_v2.CopyFile.deserializeBinaryFromReader);
      msg.setCopyFile(value);
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
proto.pfs_v2.ModifyFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.ModifyFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.ModifyFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ModifyFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getPutFile();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs_v2.PutFile.serializeBinaryToWriter
    );
  }
  f = message.getDeleteFile();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pfs_v2.DeleteFile.serializeBinaryToWriter
    );
  }
  f = message.getCopyFile();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pfs_v2.CopyFile.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.ModifyFileRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.ModifyFileRequest} returns this
*/
proto.pfs_v2.ModifyFileRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.ModifyFileRequest} returns this
 */
proto.pfs_v2.ModifyFileRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.ModifyFileRequest.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional PutFile put_file = 2;
 * @return {?proto.pfs_v2.PutFile}
 */
proto.pfs_v2.ModifyFileRequest.prototype.getPutFile = function() {
  return /** @type{?proto.pfs_v2.PutFile} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.PutFile, 2));
};


/**
 * @param {?proto.pfs_v2.PutFile|undefined} value
 * @return {!proto.pfs_v2.ModifyFileRequest} returns this
*/
proto.pfs_v2.ModifyFileRequest.prototype.setPutFile = function(value) {
  return jspb.Message.setOneofWrapperField(this, 2, proto.pfs_v2.ModifyFileRequest.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.ModifyFileRequest} returns this
 */
proto.pfs_v2.ModifyFileRequest.prototype.clearPutFile = function() {
  return this.setPutFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.ModifyFileRequest.prototype.hasPutFile = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional DeleteFile delete_file = 3;
 * @return {?proto.pfs_v2.DeleteFile}
 */
proto.pfs_v2.ModifyFileRequest.prototype.getDeleteFile = function() {
  return /** @type{?proto.pfs_v2.DeleteFile} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.DeleteFile, 3));
};


/**
 * @param {?proto.pfs_v2.DeleteFile|undefined} value
 * @return {!proto.pfs_v2.ModifyFileRequest} returns this
*/
proto.pfs_v2.ModifyFileRequest.prototype.setDeleteFile = function(value) {
  return jspb.Message.setOneofWrapperField(this, 3, proto.pfs_v2.ModifyFileRequest.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.ModifyFileRequest} returns this
 */
proto.pfs_v2.ModifyFileRequest.prototype.clearDeleteFile = function() {
  return this.setDeleteFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.ModifyFileRequest.prototype.hasDeleteFile = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional CopyFile copy_file = 4;
 * @return {?proto.pfs_v2.CopyFile}
 */
proto.pfs_v2.ModifyFileRequest.prototype.getCopyFile = function() {
  return /** @type{?proto.pfs_v2.CopyFile} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.CopyFile, 4));
};


/**
 * @param {?proto.pfs_v2.CopyFile|undefined} value
 * @return {!proto.pfs_v2.ModifyFileRequest} returns this
*/
proto.pfs_v2.ModifyFileRequest.prototype.setCopyFile = function(value) {
  return jspb.Message.setOneofWrapperField(this, 4, proto.pfs_v2.ModifyFileRequest.oneofGroups_[0], value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.ModifyFileRequest} returns this
 */
proto.pfs_v2.ModifyFileRequest.prototype.clearCopyFile = function() {
  return this.setCopyFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.ModifyFileRequest.prototype.hasCopyFile = function() {
  return jspb.Message.getField(this, 4) != null;
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
proto.pfs_v2.GetFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.GetFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.GetFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.GetFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs_v2.File.toObject(includeInstance, f),
    url: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pfs_v2.GetFileRequest}
 */
proto.pfs_v2.GetFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.GetFileRequest;
  return proto.pfs_v2.GetFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.GetFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.GetFileRequest}
 */
proto.pfs_v2.GetFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.File;
      reader.readMessage(value,proto.pfs_v2.File.deserializeBinaryFromReader);
      msg.setFile(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setUrl(value);
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
proto.pfs_v2.GetFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.GetFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.GetFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.GetFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.File.serializeBinaryToWriter
    );
  }
  f = message.getUrl();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs_v2.File}
 */
proto.pfs_v2.GetFileRequest.prototype.getFile = function() {
  return /** @type{?proto.pfs_v2.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.File, 1));
};


/**
 * @param {?proto.pfs_v2.File|undefined} value
 * @return {!proto.pfs_v2.GetFileRequest} returns this
*/
proto.pfs_v2.GetFileRequest.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.GetFileRequest} returns this
 */
proto.pfs_v2.GetFileRequest.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.GetFileRequest.prototype.hasFile = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string URL = 2;
 * @return {string}
 */
proto.pfs_v2.GetFileRequest.prototype.getUrl = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.GetFileRequest} returns this
 */
proto.pfs_v2.GetFileRequest.prototype.setUrl = function(value) {
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
proto.pfs_v2.InspectFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.InspectFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.InspectFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.InspectFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs_v2.File.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.InspectFileRequest}
 */
proto.pfs_v2.InspectFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.InspectFileRequest;
  return proto.pfs_v2.InspectFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.InspectFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.InspectFileRequest}
 */
proto.pfs_v2.InspectFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.File;
      reader.readMessage(value,proto.pfs_v2.File.deserializeBinaryFromReader);
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
proto.pfs_v2.InspectFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.InspectFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.InspectFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.InspectFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.File.serializeBinaryToWriter
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs_v2.File}
 */
proto.pfs_v2.InspectFileRequest.prototype.getFile = function() {
  return /** @type{?proto.pfs_v2.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.File, 1));
};


/**
 * @param {?proto.pfs_v2.File|undefined} value
 * @return {!proto.pfs_v2.InspectFileRequest} returns this
*/
proto.pfs_v2.InspectFileRequest.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.InspectFileRequest} returns this
 */
proto.pfs_v2.InspectFileRequest.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.InspectFileRequest.prototype.hasFile = function() {
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
proto.pfs_v2.ListFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.ListFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.ListFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ListFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs_v2.File.toObject(includeInstance, f),
    full: jspb.Message.getBooleanFieldWithDefault(msg, 2, false)
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
 * @return {!proto.pfs_v2.ListFileRequest}
 */
proto.pfs_v2.ListFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.ListFileRequest;
  return proto.pfs_v2.ListFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.ListFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.ListFileRequest}
 */
proto.pfs_v2.ListFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.File;
      reader.readMessage(value,proto.pfs_v2.File.deserializeBinaryFromReader);
      msg.setFile(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setFull(value);
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
proto.pfs_v2.ListFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.ListFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.ListFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ListFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.File.serializeBinaryToWriter
    );
  }
  f = message.getFull();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs_v2.File}
 */
proto.pfs_v2.ListFileRequest.prototype.getFile = function() {
  return /** @type{?proto.pfs_v2.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.File, 1));
};


/**
 * @param {?proto.pfs_v2.File|undefined} value
 * @return {!proto.pfs_v2.ListFileRequest} returns this
*/
proto.pfs_v2.ListFileRequest.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.ListFileRequest} returns this
 */
proto.pfs_v2.ListFileRequest.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.ListFileRequest.prototype.hasFile = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool full = 2;
 * @return {boolean}
 */
proto.pfs_v2.ListFileRequest.prototype.getFull = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.ListFileRequest} returns this
 */
proto.pfs_v2.ListFileRequest.prototype.setFull = function(value) {
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
proto.pfs_v2.WalkFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.WalkFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.WalkFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.WalkFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: (f = msg.getFile()) && proto.pfs_v2.File.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.WalkFileRequest}
 */
proto.pfs_v2.WalkFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.WalkFileRequest;
  return proto.pfs_v2.WalkFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.WalkFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.WalkFileRequest}
 */
proto.pfs_v2.WalkFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.File;
      reader.readMessage(value,proto.pfs_v2.File.deserializeBinaryFromReader);
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
proto.pfs_v2.WalkFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.WalkFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.WalkFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.WalkFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.File.serializeBinaryToWriter
    );
  }
};


/**
 * optional File file = 1;
 * @return {?proto.pfs_v2.File}
 */
proto.pfs_v2.WalkFileRequest.prototype.getFile = function() {
  return /** @type{?proto.pfs_v2.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.File, 1));
};


/**
 * @param {?proto.pfs_v2.File|undefined} value
 * @return {!proto.pfs_v2.WalkFileRequest} returns this
*/
proto.pfs_v2.WalkFileRequest.prototype.setFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.WalkFileRequest} returns this
 */
proto.pfs_v2.WalkFileRequest.prototype.clearFile = function() {
  return this.setFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.WalkFileRequest.prototype.hasFile = function() {
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
proto.pfs_v2.GlobFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.GlobFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.GlobFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.GlobFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
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
 * @return {!proto.pfs_v2.GlobFileRequest}
 */
proto.pfs_v2.GlobFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.GlobFileRequest;
  return proto.pfs_v2.GlobFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.GlobFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.GlobFileRequest}
 */
proto.pfs_v2.GlobFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
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
proto.pfs_v2.GlobFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.GlobFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.GlobFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.GlobFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
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
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.GlobFileRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.GlobFileRequest} returns this
*/
proto.pfs_v2.GlobFileRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.GlobFileRequest} returns this
 */
proto.pfs_v2.GlobFileRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.GlobFileRequest.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string pattern = 2;
 * @return {string}
 */
proto.pfs_v2.GlobFileRequest.prototype.getPattern = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.GlobFileRequest} returns this
 */
proto.pfs_v2.GlobFileRequest.prototype.setPattern = function(value) {
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
proto.pfs_v2.DiffFileRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.DiffFileRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.DiffFileRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.DiffFileRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    newFile: (f = msg.getNewFile()) && proto.pfs_v2.File.toObject(includeInstance, f),
    oldFile: (f = msg.getOldFile()) && proto.pfs_v2.File.toObject(includeInstance, f),
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
 * @return {!proto.pfs_v2.DiffFileRequest}
 */
proto.pfs_v2.DiffFileRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.DiffFileRequest;
  return proto.pfs_v2.DiffFileRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.DiffFileRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.DiffFileRequest}
 */
proto.pfs_v2.DiffFileRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.File;
      reader.readMessage(value,proto.pfs_v2.File.deserializeBinaryFromReader);
      msg.setNewFile(value);
      break;
    case 2:
      var value = new proto.pfs_v2.File;
      reader.readMessage(value,proto.pfs_v2.File.deserializeBinaryFromReader);
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
proto.pfs_v2.DiffFileRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.DiffFileRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.DiffFileRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.DiffFileRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getNewFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.File.serializeBinaryToWriter
    );
  }
  f = message.getOldFile();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs_v2.File.serializeBinaryToWriter
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
 * @return {?proto.pfs_v2.File}
 */
proto.pfs_v2.DiffFileRequest.prototype.getNewFile = function() {
  return /** @type{?proto.pfs_v2.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.File, 1));
};


/**
 * @param {?proto.pfs_v2.File|undefined} value
 * @return {!proto.pfs_v2.DiffFileRequest} returns this
*/
proto.pfs_v2.DiffFileRequest.prototype.setNewFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.DiffFileRequest} returns this
 */
proto.pfs_v2.DiffFileRequest.prototype.clearNewFile = function() {
  return this.setNewFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.DiffFileRequest.prototype.hasNewFile = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional File old_file = 2;
 * @return {?proto.pfs_v2.File}
 */
proto.pfs_v2.DiffFileRequest.prototype.getOldFile = function() {
  return /** @type{?proto.pfs_v2.File} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.File, 2));
};


/**
 * @param {?proto.pfs_v2.File|undefined} value
 * @return {!proto.pfs_v2.DiffFileRequest} returns this
*/
proto.pfs_v2.DiffFileRequest.prototype.setOldFile = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.DiffFileRequest} returns this
 */
proto.pfs_v2.DiffFileRequest.prototype.clearOldFile = function() {
  return this.setOldFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.DiffFileRequest.prototype.hasOldFile = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional bool shallow = 3;
 * @return {boolean}
 */
proto.pfs_v2.DiffFileRequest.prototype.getShallow = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.DiffFileRequest} returns this
 */
proto.pfs_v2.DiffFileRequest.prototype.setShallow = function(value) {
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
proto.pfs_v2.DiffFileResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.DiffFileResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.DiffFileResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.DiffFileResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    newFile: (f = msg.getNewFile()) && proto.pfs_v2.FileInfo.toObject(includeInstance, f),
    oldFile: (f = msg.getOldFile()) && proto.pfs_v2.FileInfo.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.DiffFileResponse}
 */
proto.pfs_v2.DiffFileResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.DiffFileResponse;
  return proto.pfs_v2.DiffFileResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.DiffFileResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.DiffFileResponse}
 */
proto.pfs_v2.DiffFileResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.FileInfo;
      reader.readMessage(value,proto.pfs_v2.FileInfo.deserializeBinaryFromReader);
      msg.setNewFile(value);
      break;
    case 2:
      var value = new proto.pfs_v2.FileInfo;
      reader.readMessage(value,proto.pfs_v2.FileInfo.deserializeBinaryFromReader);
      msg.setOldFile(value);
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
proto.pfs_v2.DiffFileResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.DiffFileResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.DiffFileResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.DiffFileResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getNewFile();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.FileInfo.serializeBinaryToWriter
    );
  }
  f = message.getOldFile();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pfs_v2.FileInfo.serializeBinaryToWriter
    );
  }
};


/**
 * optional FileInfo new_file = 1;
 * @return {?proto.pfs_v2.FileInfo}
 */
proto.pfs_v2.DiffFileResponse.prototype.getNewFile = function() {
  return /** @type{?proto.pfs_v2.FileInfo} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.FileInfo, 1));
};


/**
 * @param {?proto.pfs_v2.FileInfo|undefined} value
 * @return {!proto.pfs_v2.DiffFileResponse} returns this
*/
proto.pfs_v2.DiffFileResponse.prototype.setNewFile = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.DiffFileResponse} returns this
 */
proto.pfs_v2.DiffFileResponse.prototype.clearNewFile = function() {
  return this.setNewFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.DiffFileResponse.prototype.hasNewFile = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional FileInfo old_file = 2;
 * @return {?proto.pfs_v2.FileInfo}
 */
proto.pfs_v2.DiffFileResponse.prototype.getOldFile = function() {
  return /** @type{?proto.pfs_v2.FileInfo} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.FileInfo, 2));
};


/**
 * @param {?proto.pfs_v2.FileInfo|undefined} value
 * @return {!proto.pfs_v2.DiffFileResponse} returns this
*/
proto.pfs_v2.DiffFileResponse.prototype.setOldFile = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.DiffFileResponse} returns this
 */
proto.pfs_v2.DiffFileResponse.prototype.clearOldFile = function() {
  return this.setOldFile(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.DiffFileResponse.prototype.hasOldFile = function() {
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
proto.pfs_v2.FsckRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.FsckRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.FsckRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.FsckRequest.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pfs_v2.FsckRequest}
 */
proto.pfs_v2.FsckRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.FsckRequest;
  return proto.pfs_v2.FsckRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.FsckRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.FsckRequest}
 */
proto.pfs_v2.FsckRequest.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pfs_v2.FsckRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.FsckRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.FsckRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.FsckRequest.serializeBinaryToWriter = function(message, writer) {
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
proto.pfs_v2.FsckRequest.prototype.getFix = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 1, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pfs_v2.FsckRequest} returns this
 */
proto.pfs_v2.FsckRequest.prototype.setFix = function(value) {
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
proto.pfs_v2.FsckResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.FsckResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.FsckResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.FsckResponse.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pfs_v2.FsckResponse}
 */
proto.pfs_v2.FsckResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.FsckResponse;
  return proto.pfs_v2.FsckResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.FsckResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.FsckResponse}
 */
proto.pfs_v2.FsckResponse.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pfs_v2.FsckResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.FsckResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.FsckResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.FsckResponse.serializeBinaryToWriter = function(message, writer) {
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
proto.pfs_v2.FsckResponse.prototype.getFix = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.FsckResponse} returns this
 */
proto.pfs_v2.FsckResponse.prototype.setFix = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string error = 2;
 * @return {string}
 */
proto.pfs_v2.FsckResponse.prototype.getError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.FsckResponse} returns this
 */
proto.pfs_v2.FsckResponse.prototype.setError = function(value) {
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
proto.pfs_v2.CreateFilesetResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.CreateFilesetResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.CreateFilesetResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CreateFilesetResponse.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pfs_v2.CreateFilesetResponse}
 */
proto.pfs_v2.CreateFilesetResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.CreateFilesetResponse;
  return proto.pfs_v2.CreateFilesetResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.CreateFilesetResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.CreateFilesetResponse}
 */
proto.pfs_v2.CreateFilesetResponse.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pfs_v2.CreateFilesetResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.CreateFilesetResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.CreateFilesetResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.CreateFilesetResponse.serializeBinaryToWriter = function(message, writer) {
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
proto.pfs_v2.CreateFilesetResponse.prototype.getFilesetId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.CreateFilesetResponse} returns this
 */
proto.pfs_v2.CreateFilesetResponse.prototype.setFilesetId = function(value) {
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
proto.pfs_v2.GetFilesetRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.GetFilesetRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.GetFilesetRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.GetFilesetRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f)
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
 * @return {!proto.pfs_v2.GetFilesetRequest}
 */
proto.pfs_v2.GetFilesetRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.GetFilesetRequest;
  return proto.pfs_v2.GetFilesetRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.GetFilesetRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.GetFilesetRequest}
 */
proto.pfs_v2.GetFilesetRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
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
proto.pfs_v2.GetFilesetRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.GetFilesetRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.GetFilesetRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.GetFilesetRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.GetFilesetRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.GetFilesetRequest} returns this
*/
proto.pfs_v2.GetFilesetRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.GetFilesetRequest} returns this
 */
proto.pfs_v2.GetFilesetRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.GetFilesetRequest.prototype.hasCommit = function() {
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
proto.pfs_v2.AddFilesetRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.AddFilesetRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.AddFilesetRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.AddFilesetRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commit: (f = msg.getCommit()) && proto.pfs_v2.Commit.toObject(includeInstance, f),
    filesetId: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pfs_v2.AddFilesetRequest}
 */
proto.pfs_v2.AddFilesetRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.AddFilesetRequest;
  return proto.pfs_v2.AddFilesetRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.AddFilesetRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.AddFilesetRequest}
 */
proto.pfs_v2.AddFilesetRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Commit;
      reader.readMessage(value,proto.pfs_v2.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
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
proto.pfs_v2.AddFilesetRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.AddFilesetRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.AddFilesetRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.AddFilesetRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Commit.serializeBinaryToWriter
    );
  }
  f = message.getFilesetId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional Commit commit = 1;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pfs_v2.AddFilesetRequest.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Commit, 1));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pfs_v2.AddFilesetRequest} returns this
*/
proto.pfs_v2.AddFilesetRequest.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.AddFilesetRequest} returns this
 */
proto.pfs_v2.AddFilesetRequest.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.AddFilesetRequest.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string fileset_id = 2;
 * @return {string}
 */
proto.pfs_v2.AddFilesetRequest.prototype.getFilesetId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.AddFilesetRequest} returns this
 */
proto.pfs_v2.AddFilesetRequest.prototype.setFilesetId = function(value) {
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
proto.pfs_v2.RenewFilesetRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.RenewFilesetRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.RenewFilesetRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RenewFilesetRequest.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pfs_v2.RenewFilesetRequest}
 */
proto.pfs_v2.RenewFilesetRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.RenewFilesetRequest;
  return proto.pfs_v2.RenewFilesetRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.RenewFilesetRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.RenewFilesetRequest}
 */
proto.pfs_v2.RenewFilesetRequest.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pfs_v2.RenewFilesetRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.RenewFilesetRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.RenewFilesetRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RenewFilesetRequest.serializeBinaryToWriter = function(message, writer) {
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
proto.pfs_v2.RenewFilesetRequest.prototype.getFilesetId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.RenewFilesetRequest} returns this
 */
proto.pfs_v2.RenewFilesetRequest.prototype.setFilesetId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional int64 ttl_seconds = 2;
 * @return {number}
 */
proto.pfs_v2.RenewFilesetRequest.prototype.getTtlSeconds = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.RenewFilesetRequest} returns this
 */
proto.pfs_v2.RenewFilesetRequest.prototype.setTtlSeconds = function(value) {
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
proto.pfs_v2.ActivateAuthRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.ActivateAuthRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.ActivateAuthRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ActivateAuthRequest.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pfs_v2.ActivateAuthRequest}
 */
proto.pfs_v2.ActivateAuthRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.ActivateAuthRequest;
  return proto.pfs_v2.ActivateAuthRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.ActivateAuthRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.ActivateAuthRequest}
 */
proto.pfs_v2.ActivateAuthRequest.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pfs_v2.ActivateAuthRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.ActivateAuthRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.ActivateAuthRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ActivateAuthRequest.serializeBinaryToWriter = function(message, writer) {
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
proto.pfs_v2.ActivateAuthResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.ActivateAuthResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.ActivateAuthResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ActivateAuthResponse.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pfs_v2.ActivateAuthResponse}
 */
proto.pfs_v2.ActivateAuthResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.ActivateAuthResponse;
  return proto.pfs_v2.ActivateAuthResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.ActivateAuthResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.ActivateAuthResponse}
 */
proto.pfs_v2.ActivateAuthResponse.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pfs_v2.ActivateAuthResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.ActivateAuthResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.ActivateAuthResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.ActivateAuthResponse.serializeBinaryToWriter = function(message, writer) {
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
proto.pfs_v2.RunLoadTestRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.RunLoadTestRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.RunLoadTestRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RunLoadTestRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    spec: msg.getSpec_asB64(),
    seed: jspb.Message.getFieldWithDefault(msg, 2, 0)
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
 * @return {!proto.pfs_v2.RunLoadTestRequest}
 */
proto.pfs_v2.RunLoadTestRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.RunLoadTestRequest;
  return proto.pfs_v2.RunLoadTestRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.RunLoadTestRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.RunLoadTestRequest}
 */
proto.pfs_v2.RunLoadTestRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setSpec(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setSeed(value);
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
proto.pfs_v2.RunLoadTestRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.RunLoadTestRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.RunLoadTestRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RunLoadTestRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSpec_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      1,
      f
    );
  }
  f = message.getSeed();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
};


/**
 * optional bytes spec = 1;
 * @return {string}
 */
proto.pfs_v2.RunLoadTestRequest.prototype.getSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * optional bytes spec = 1;
 * This is a type-conversion wrapper around `getSpec()`
 * @return {string}
 */
proto.pfs_v2.RunLoadTestRequest.prototype.getSpec_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getSpec()));
};


/**
 * optional bytes spec = 1;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getSpec()`
 * @return {!Uint8Array}
 */
proto.pfs_v2.RunLoadTestRequest.prototype.getSpec_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getSpec()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pfs_v2.RunLoadTestRequest} returns this
 */
proto.pfs_v2.RunLoadTestRequest.prototype.setSpec = function(value) {
  return jspb.Message.setProto3BytesField(this, 1, value);
};


/**
 * optional int64 seed = 2;
 * @return {number}
 */
proto.pfs_v2.RunLoadTestRequest.prototype.getSeed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.RunLoadTestRequest} returns this
 */
proto.pfs_v2.RunLoadTestRequest.prototype.setSeed = function(value) {
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
proto.pfs_v2.RunLoadTestResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pfs_v2.RunLoadTestResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pfs_v2.RunLoadTestResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RunLoadTestResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    branch: (f = msg.getBranch()) && proto.pfs_v2.Branch.toObject(includeInstance, f),
    seed: jspb.Message.getFieldWithDefault(msg, 2, 0),
    error: jspb.Message.getFieldWithDefault(msg, 3, "")
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
 * @return {!proto.pfs_v2.RunLoadTestResponse}
 */
proto.pfs_v2.RunLoadTestResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pfs_v2.RunLoadTestResponse;
  return proto.pfs_v2.RunLoadTestResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pfs_v2.RunLoadTestResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pfs_v2.RunLoadTestResponse}
 */
proto.pfs_v2.RunLoadTestResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pfs_v2.Branch;
      reader.readMessage(value,proto.pfs_v2.Branch.deserializeBinaryFromReader);
      msg.setBranch(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setSeed(value);
      break;
    case 3:
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
proto.pfs_v2.RunLoadTestResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pfs_v2.RunLoadTestResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pfs_v2.RunLoadTestResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pfs_v2.RunLoadTestResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getBranch();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pfs_v2.Branch.serializeBinaryToWriter
    );
  }
  f = message.getSeed();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
  f = message.getError();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional Branch branch = 1;
 * @return {?proto.pfs_v2.Branch}
 */
proto.pfs_v2.RunLoadTestResponse.prototype.getBranch = function() {
  return /** @type{?proto.pfs_v2.Branch} */ (
    jspb.Message.getWrapperField(this, proto.pfs_v2.Branch, 1));
};


/**
 * @param {?proto.pfs_v2.Branch|undefined} value
 * @return {!proto.pfs_v2.RunLoadTestResponse} returns this
*/
proto.pfs_v2.RunLoadTestResponse.prototype.setBranch = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pfs_v2.RunLoadTestResponse} returns this
 */
proto.pfs_v2.RunLoadTestResponse.prototype.clearBranch = function() {
  return this.setBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pfs_v2.RunLoadTestResponse.prototype.hasBranch = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional int64 seed = 2;
 * @return {number}
 */
proto.pfs_v2.RunLoadTestResponse.prototype.getSeed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pfs_v2.RunLoadTestResponse} returns this
 */
proto.pfs_v2.RunLoadTestResponse.prototype.setSeed = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional string error = 3;
 * @return {string}
 */
proto.pfs_v2.RunLoadTestResponse.prototype.getError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pfs_v2.RunLoadTestResponse} returns this
 */
proto.pfs_v2.RunLoadTestResponse.prototype.setError = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * @enum {number}
 */
proto.pfs_v2.OriginKind = {
  USER: 0,
  AUTO: 1,
  FSCK: 2
};

/**
 * @enum {number}
 */
proto.pfs_v2.FileType = {
  RESERVED: 0,
  FILE: 1,
  DIR: 2
};

/**
 * @enum {number}
 */
proto.pfs_v2.CommitState = {
  STARTED: 0,
  READY: 1,
  FINISHED: 2
};

/**
 * @enum {number}
 */
proto.pfs_v2.Delimiter = {
  NONE: 0,
  JSON: 1,
  LINE: 2,
  SQL: 3,
  CSV: 4
};

goog.object.extend(exports, proto.pfs_v2);
