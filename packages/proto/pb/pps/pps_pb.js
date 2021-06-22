// source: pps/pps.proto
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
var google_protobuf_duration_pb = require('google-protobuf/google/protobuf/duration_pb.js');
goog.object.extend(proto, google_protobuf_duration_pb);
var gogoproto_gogo_pb = require('../gogoproto/gogo_pb.js');
goog.object.extend(proto, gogoproto_gogo_pb);
var pfs_pfs_pb = require('../pfs/pfs_pb.js');
goog.object.extend(proto, pfs_pfs_pb);
goog.exportSymbol('proto.pps_v2.ActivateAuthRequest', null, global);
goog.exportSymbol('proto.pps_v2.ActivateAuthResponse', null, global);
goog.exportSymbol('proto.pps_v2.Aggregate', null, global);
goog.exportSymbol('proto.pps_v2.AggregateProcessStats', null, global);
goog.exportSymbol('proto.pps_v2.CreatePipelineRequest', null, global);
goog.exportSymbol('proto.pps_v2.CreateSecretRequest', null, global);
goog.exportSymbol('proto.pps_v2.CronInput', null, global);
goog.exportSymbol('proto.pps_v2.Datum', null, global);
goog.exportSymbol('proto.pps_v2.DatumInfo', null, global);
goog.exportSymbol('proto.pps_v2.DatumSetSpec', null, global);
goog.exportSymbol('proto.pps_v2.DatumState', null, global);
goog.exportSymbol('proto.pps_v2.DatumStatus', null, global);
goog.exportSymbol('proto.pps_v2.DeleteJobRequest', null, global);
goog.exportSymbol('proto.pps_v2.DeletePipelineRequest', null, global);
goog.exportSymbol('proto.pps_v2.DeleteSecretRequest', null, global);
goog.exportSymbol('proto.pps_v2.Egress', null, global);
goog.exportSymbol('proto.pps_v2.GPUSpec', null, global);
goog.exportSymbol('proto.pps_v2.GetLogsRequest', null, global);
goog.exportSymbol('proto.pps_v2.Input', null, global);
goog.exportSymbol('proto.pps_v2.InputFile', null, global);
goog.exportSymbol('proto.pps_v2.InspectDatumRequest', null, global);
goog.exportSymbol('proto.pps_v2.InspectJobRequest', null, global);
goog.exportSymbol('proto.pps_v2.InspectJobSetRequest', null, global);
goog.exportSymbol('proto.pps_v2.InspectPipelineRequest', null, global);
goog.exportSymbol('proto.pps_v2.InspectSecretRequest', null, global);
goog.exportSymbol('proto.pps_v2.Job', null, global);
goog.exportSymbol('proto.pps_v2.JobInfo', null, global);
goog.exportSymbol('proto.pps_v2.JobInfo.Details', null, global);
goog.exportSymbol('proto.pps_v2.JobInput', null, global);
goog.exportSymbol('proto.pps_v2.JobSet', null, global);
goog.exportSymbol('proto.pps_v2.JobState', null, global);
goog.exportSymbol('proto.pps_v2.ListDatumRequest', null, global);
goog.exportSymbol('proto.pps_v2.ListJobRequest', null, global);
goog.exportSymbol('proto.pps_v2.ListPipelineRequest', null, global);
goog.exportSymbol('proto.pps_v2.LogMessage', null, global);
goog.exportSymbol('proto.pps_v2.Metadata', null, global);
goog.exportSymbol('proto.pps_v2.PFSInput', null, global);
goog.exportSymbol('proto.pps_v2.ParallelismSpec', null, global);
goog.exportSymbol('proto.pps_v2.Pipeline', null, global);
goog.exportSymbol('proto.pps_v2.PipelineInfo', null, global);
goog.exportSymbol('proto.pps_v2.PipelineInfo.Details', null, global);
goog.exportSymbol('proto.pps_v2.PipelineInfo.PipelineType', null, global);
goog.exportSymbol('proto.pps_v2.PipelineInfos', null, global);
goog.exportSymbol('proto.pps_v2.PipelineState', null, global);
goog.exportSymbol('proto.pps_v2.ProcessStats', null, global);
goog.exportSymbol('proto.pps_v2.ResourceSpec', null, global);
goog.exportSymbol('proto.pps_v2.RestartDatumRequest', null, global);
goog.exportSymbol('proto.pps_v2.RunCronRequest', null, global);
goog.exportSymbol('proto.pps_v2.RunPipelineRequest', null, global);
goog.exportSymbol('proto.pps_v2.SchedulingSpec', null, global);
goog.exportSymbol('proto.pps_v2.Secret', null, global);
goog.exportSymbol('proto.pps_v2.SecretInfo', null, global);
goog.exportSymbol('proto.pps_v2.SecretInfos', null, global);
goog.exportSymbol('proto.pps_v2.SecretMount', null, global);
goog.exportSymbol('proto.pps_v2.Service', null, global);
goog.exportSymbol('proto.pps_v2.Spout', null, global);
goog.exportSymbol('proto.pps_v2.StartPipelineRequest', null, global);
goog.exportSymbol('proto.pps_v2.StopJobRequest', null, global);
goog.exportSymbol('proto.pps_v2.StopPipelineRequest', null, global);
goog.exportSymbol('proto.pps_v2.SubscribeJobRequest', null, global);
goog.exportSymbol('proto.pps_v2.TFJob', null, global);
goog.exportSymbol('proto.pps_v2.Transform', null, global);
goog.exportSymbol('proto.pps_v2.UpdateJobStateRequest', null, global);
goog.exportSymbol('proto.pps_v2.Worker', null, global);
goog.exportSymbol('proto.pps_v2.WorkerState', null, global);
goog.exportSymbol('proto.pps_v2.WorkerStatus', null, global);
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
proto.pps_v2.SecretMount = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.SecretMount, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.SecretMount.displayName = 'proto.pps_v2.SecretMount';
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
proto.pps_v2.Transform = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.Transform.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.Transform, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Transform.displayName = 'proto.pps_v2.Transform';
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
proto.pps_v2.TFJob = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.TFJob, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.TFJob.displayName = 'proto.pps_v2.TFJob';
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
proto.pps_v2.Egress = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.Egress, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Egress.displayName = 'proto.pps_v2.Egress';
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
proto.pps_v2.Job = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.Job, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Job.displayName = 'proto.pps_v2.Job';
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
proto.pps_v2.Metadata = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.Metadata, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Metadata.displayName = 'proto.pps_v2.Metadata';
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
proto.pps_v2.Service = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.Service, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Service.displayName = 'proto.pps_v2.Service';
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
proto.pps_v2.Spout = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.Spout, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Spout.displayName = 'proto.pps_v2.Spout';
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
proto.pps_v2.PFSInput = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.PFSInput, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.PFSInput.displayName = 'proto.pps_v2.PFSInput';
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
proto.pps_v2.CronInput = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.CronInput, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.CronInput.displayName = 'proto.pps_v2.CronInput';
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
proto.pps_v2.Input = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.Input.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.Input, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Input.displayName = 'proto.pps_v2.Input';
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
proto.pps_v2.JobInput = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.JobInput, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.JobInput.displayName = 'proto.pps_v2.JobInput';
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
proto.pps_v2.ParallelismSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.ParallelismSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.ParallelismSpec.displayName = 'proto.pps_v2.ParallelismSpec';
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
proto.pps_v2.InputFile = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.InputFile, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.InputFile.displayName = 'proto.pps_v2.InputFile';
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
proto.pps_v2.Datum = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.Datum, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Datum.displayName = 'proto.pps_v2.Datum';
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
proto.pps_v2.DatumInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.DatumInfo.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.DatumInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.DatumInfo.displayName = 'proto.pps_v2.DatumInfo';
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
proto.pps_v2.Aggregate = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.Aggregate, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Aggregate.displayName = 'proto.pps_v2.Aggregate';
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
proto.pps_v2.ProcessStats = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.ProcessStats, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.ProcessStats.displayName = 'proto.pps_v2.ProcessStats';
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
proto.pps_v2.AggregateProcessStats = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.AggregateProcessStats, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.AggregateProcessStats.displayName = 'proto.pps_v2.AggregateProcessStats';
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
proto.pps_v2.WorkerStatus = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.WorkerStatus, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.WorkerStatus.displayName = 'proto.pps_v2.WorkerStatus';
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
proto.pps_v2.DatumStatus = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.DatumStatus.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.DatumStatus, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.DatumStatus.displayName = 'proto.pps_v2.DatumStatus';
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
proto.pps_v2.ResourceSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.ResourceSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.ResourceSpec.displayName = 'proto.pps_v2.ResourceSpec';
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
proto.pps_v2.GPUSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.GPUSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.GPUSpec.displayName = 'proto.pps_v2.GPUSpec';
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
proto.pps_v2.JobInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.JobInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.JobInfo.displayName = 'proto.pps_v2.JobInfo';
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
proto.pps_v2.JobInfo.Details = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.JobInfo.Details.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.JobInfo.Details, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.JobInfo.Details.displayName = 'proto.pps_v2.JobInfo.Details';
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
proto.pps_v2.Worker = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.Worker, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Worker.displayName = 'proto.pps_v2.Worker';
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
proto.pps_v2.Pipeline = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.Pipeline, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Pipeline.displayName = 'proto.pps_v2.Pipeline';
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
proto.pps_v2.PipelineInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.PipelineInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.PipelineInfo.displayName = 'proto.pps_v2.PipelineInfo';
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
proto.pps_v2.PipelineInfo.Details = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.PipelineInfo.Details, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.PipelineInfo.Details.displayName = 'proto.pps_v2.PipelineInfo.Details';
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
proto.pps_v2.PipelineInfos = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.PipelineInfos.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.PipelineInfos, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.PipelineInfos.displayName = 'proto.pps_v2.PipelineInfos';
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
proto.pps_v2.JobSet = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.JobSet, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.JobSet.displayName = 'proto.pps_v2.JobSet';
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
proto.pps_v2.InspectJobSetRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.InspectJobSetRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.InspectJobSetRequest.displayName = 'proto.pps_v2.InspectJobSetRequest';
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
proto.pps_v2.InspectJobRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.InspectJobRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.InspectJobRequest.displayName = 'proto.pps_v2.InspectJobRequest';
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
proto.pps_v2.ListJobRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.ListJobRequest.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.ListJobRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.ListJobRequest.displayName = 'proto.pps_v2.ListJobRequest';
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
proto.pps_v2.SubscribeJobRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.SubscribeJobRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.SubscribeJobRequest.displayName = 'proto.pps_v2.SubscribeJobRequest';
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
proto.pps_v2.DeleteJobRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.DeleteJobRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.DeleteJobRequest.displayName = 'proto.pps_v2.DeleteJobRequest';
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
proto.pps_v2.StopJobRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.StopJobRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.StopJobRequest.displayName = 'proto.pps_v2.StopJobRequest';
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
proto.pps_v2.UpdateJobStateRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.UpdateJobStateRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.UpdateJobStateRequest.displayName = 'proto.pps_v2.UpdateJobStateRequest';
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
proto.pps_v2.GetLogsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.GetLogsRequest.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.GetLogsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.GetLogsRequest.displayName = 'proto.pps_v2.GetLogsRequest';
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
proto.pps_v2.LogMessage = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.LogMessage.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.LogMessage, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.LogMessage.displayName = 'proto.pps_v2.LogMessage';
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
proto.pps_v2.RestartDatumRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.RestartDatumRequest.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.RestartDatumRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.RestartDatumRequest.displayName = 'proto.pps_v2.RestartDatumRequest';
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
proto.pps_v2.InspectDatumRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.InspectDatumRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.InspectDatumRequest.displayName = 'proto.pps_v2.InspectDatumRequest';
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
proto.pps_v2.ListDatumRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.ListDatumRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.ListDatumRequest.displayName = 'proto.pps_v2.ListDatumRequest';
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
proto.pps_v2.DatumSetSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.DatumSetSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.DatumSetSpec.displayName = 'proto.pps_v2.DatumSetSpec';
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
proto.pps_v2.SchedulingSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.SchedulingSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.SchedulingSpec.displayName = 'proto.pps_v2.SchedulingSpec';
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
proto.pps_v2.CreatePipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.CreatePipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.CreatePipelineRequest.displayName = 'proto.pps_v2.CreatePipelineRequest';
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
proto.pps_v2.InspectPipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.InspectPipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.InspectPipelineRequest.displayName = 'proto.pps_v2.InspectPipelineRequest';
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
proto.pps_v2.ListPipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.ListPipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.ListPipelineRequest.displayName = 'proto.pps_v2.ListPipelineRequest';
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
proto.pps_v2.DeletePipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.DeletePipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.DeletePipelineRequest.displayName = 'proto.pps_v2.DeletePipelineRequest';
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
proto.pps_v2.StartPipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.StartPipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.StartPipelineRequest.displayName = 'proto.pps_v2.StartPipelineRequest';
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
proto.pps_v2.StopPipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.StopPipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.StopPipelineRequest.displayName = 'proto.pps_v2.StopPipelineRequest';
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
proto.pps_v2.RunPipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.RunPipelineRequest.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.RunPipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.RunPipelineRequest.displayName = 'proto.pps_v2.RunPipelineRequest';
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
proto.pps_v2.RunCronRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.RunCronRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.RunCronRequest.displayName = 'proto.pps_v2.RunCronRequest';
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
proto.pps_v2.CreateSecretRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.CreateSecretRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.CreateSecretRequest.displayName = 'proto.pps_v2.CreateSecretRequest';
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
proto.pps_v2.DeleteSecretRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.DeleteSecretRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.DeleteSecretRequest.displayName = 'proto.pps_v2.DeleteSecretRequest';
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
proto.pps_v2.InspectSecretRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.InspectSecretRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.InspectSecretRequest.displayName = 'proto.pps_v2.InspectSecretRequest';
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
proto.pps_v2.Secret = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.Secret, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.Secret.displayName = 'proto.pps_v2.Secret';
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
proto.pps_v2.SecretInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.SecretInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.SecretInfo.displayName = 'proto.pps_v2.SecretInfo';
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
proto.pps_v2.SecretInfos = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps_v2.SecretInfos.repeatedFields_, null);
};
goog.inherits(proto.pps_v2.SecretInfos, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.SecretInfos.displayName = 'proto.pps_v2.SecretInfos';
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
proto.pps_v2.ActivateAuthRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.ActivateAuthRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.ActivateAuthRequest.displayName = 'proto.pps_v2.ActivateAuthRequest';
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
proto.pps_v2.ActivateAuthResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps_v2.ActivateAuthResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps_v2.ActivateAuthResponse.displayName = 'proto.pps_v2.ActivateAuthResponse';
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
proto.pps_v2.SecretMount.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.SecretMount.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.SecretMount} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.SecretMount.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    key: jspb.Message.getFieldWithDefault(msg, 2, ""),
    mountPath: jspb.Message.getFieldWithDefault(msg, 3, ""),
    envVar: jspb.Message.getFieldWithDefault(msg, 4, "")
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
 * @return {!proto.pps_v2.SecretMount}
 */
proto.pps_v2.SecretMount.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.SecretMount;
  return proto.pps_v2.SecretMount.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.SecretMount} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.SecretMount}
 */
proto.pps_v2.SecretMount.deserializeBinaryFromReader = function(msg, reader) {
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
      msg.setKey(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setMountPath(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setEnvVar(value);
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
proto.pps_v2.SecretMount.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.SecretMount.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.SecretMount} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.SecretMount.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getKey();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getMountPath();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getEnvVar();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pps_v2.SecretMount.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.SecretMount} returns this
 */
proto.pps_v2.SecretMount.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string key = 2;
 * @return {string}
 */
proto.pps_v2.SecretMount.prototype.getKey = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.SecretMount} returns this
 */
proto.pps_v2.SecretMount.prototype.setKey = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string mount_path = 3;
 * @return {string}
 */
proto.pps_v2.SecretMount.prototype.getMountPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.SecretMount} returns this
 */
proto.pps_v2.SecretMount.prototype.setMountPath = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string env_var = 4;
 * @return {string}
 */
proto.pps_v2.SecretMount.prototype.getEnvVar = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.SecretMount} returns this
 */
proto.pps_v2.SecretMount.prototype.setEnvVar = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.Transform.repeatedFields_ = [2,3,5,6,7,8,9];



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
proto.pps_v2.Transform.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Transform.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Transform} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Transform.toObject = function(includeInstance, msg) {
  var f, obj = {
    image: jspb.Message.getFieldWithDefault(msg, 1, ""),
    cmdList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f,
    errCmdList: (f = jspb.Message.getRepeatedField(msg, 3)) == null ? undefined : f,
    envMap: (f = msg.getEnvMap()) ? f.toObject(includeInstance, undefined) : [],
    secretsList: jspb.Message.toObjectList(msg.getSecretsList(),
    proto.pps_v2.SecretMount.toObject, includeInstance),
    imagePullSecretsList: (f = jspb.Message.getRepeatedField(msg, 6)) == null ? undefined : f,
    stdinList: (f = jspb.Message.getRepeatedField(msg, 7)) == null ? undefined : f,
    errStdinList: (f = jspb.Message.getRepeatedField(msg, 8)) == null ? undefined : f,
    acceptReturnCodeList: (f = jspb.Message.getRepeatedField(msg, 9)) == null ? undefined : f,
    debug: jspb.Message.getBooleanFieldWithDefault(msg, 10, false),
    user: jspb.Message.getFieldWithDefault(msg, 11, ""),
    workingDir: jspb.Message.getFieldWithDefault(msg, 12, ""),
    dockerfile: jspb.Message.getFieldWithDefault(msg, 13, "")
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
 * @return {!proto.pps_v2.Transform}
 */
proto.pps_v2.Transform.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Transform;
  return proto.pps_v2.Transform.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Transform} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Transform}
 */
proto.pps_v2.Transform.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setImage(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addCmd(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.addErrCmd(value);
      break;
    case 4:
      var value = msg.getEnvMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
         });
      break;
    case 5:
      var value = new proto.pps_v2.SecretMount;
      reader.readMessage(value,proto.pps_v2.SecretMount.deserializeBinaryFromReader);
      msg.addSecrets(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.addImagePullSecrets(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.addStdin(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.addErrStdin(value);
      break;
    case 9:
      var values = /** @type {!Array<number>} */ (reader.isDelimited() ? reader.readPackedInt64() : [reader.readInt64()]);
      for (var i = 0; i < values.length; i++) {
        msg.addAcceptReturnCode(values[i]);
      }
      break;
    case 10:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDebug(value);
      break;
    case 11:
      var value = /** @type {string} */ (reader.readString());
      msg.setUser(value);
      break;
    case 12:
      var value = /** @type {string} */ (reader.readString());
      msg.setWorkingDir(value);
      break;
    case 13:
      var value = /** @type {string} */ (reader.readString());
      msg.setDockerfile(value);
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
proto.pps_v2.Transform.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Transform.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Transform} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Transform.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getImage();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getCmdList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
  f = message.getErrCmdList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      3,
      f
    );
  }
  f = message.getEnvMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(4, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getSecretsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      5,
      f,
      proto.pps_v2.SecretMount.serializeBinaryToWriter
    );
  }
  f = message.getImagePullSecretsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      6,
      f
    );
  }
  f = message.getStdinList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      7,
      f
    );
  }
  f = message.getErrStdinList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      8,
      f
    );
  }
  f = message.getAcceptReturnCodeList();
  if (f.length > 0) {
    writer.writePackedInt64(
      9,
      f
    );
  }
  f = message.getDebug();
  if (f) {
    writer.writeBool(
      10,
      f
    );
  }
  f = message.getUser();
  if (f.length > 0) {
    writer.writeString(
      11,
      f
    );
  }
  f = message.getWorkingDir();
  if (f.length > 0) {
    writer.writeString(
      12,
      f
    );
  }
  f = message.getDockerfile();
  if (f.length > 0) {
    writer.writeString(
      13,
      f
    );
  }
};


/**
 * optional string image = 1;
 * @return {string}
 */
proto.pps_v2.Transform.prototype.getImage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.setImage = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * repeated string cmd = 2;
 * @return {!Array<string>}
 */
proto.pps_v2.Transform.prototype.getCmdList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.setCmdList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.addCmd = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.clearCmdList = function() {
  return this.setCmdList([]);
};


/**
 * repeated string err_cmd = 3;
 * @return {!Array<string>}
 */
proto.pps_v2.Transform.prototype.getErrCmdList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 3));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.setErrCmdList = function(value) {
  return jspb.Message.setField(this, 3, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.addErrCmd = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 3, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.clearErrCmdList = function() {
  return this.setErrCmdList([]);
};


/**
 * map<string, string> env = 4;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.pps_v2.Transform.prototype.getEnvMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 4, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.clearEnvMap = function() {
  this.getEnvMap().clear();
  return this;};


/**
 * repeated SecretMount secrets = 5;
 * @return {!Array<!proto.pps_v2.SecretMount>}
 */
proto.pps_v2.Transform.prototype.getSecretsList = function() {
  return /** @type{!Array<!proto.pps_v2.SecretMount>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps_v2.SecretMount, 5));
};


/**
 * @param {!Array<!proto.pps_v2.SecretMount>} value
 * @return {!proto.pps_v2.Transform} returns this
*/
proto.pps_v2.Transform.prototype.setSecretsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.pps_v2.SecretMount=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.SecretMount}
 */
proto.pps_v2.Transform.prototype.addSecrets = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.pps_v2.SecretMount, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.clearSecretsList = function() {
  return this.setSecretsList([]);
};


/**
 * repeated string image_pull_secrets = 6;
 * @return {!Array<string>}
 */
proto.pps_v2.Transform.prototype.getImagePullSecretsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 6));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.setImagePullSecretsList = function(value) {
  return jspb.Message.setField(this, 6, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.addImagePullSecrets = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 6, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.clearImagePullSecretsList = function() {
  return this.setImagePullSecretsList([]);
};


/**
 * repeated string stdin = 7;
 * @return {!Array<string>}
 */
proto.pps_v2.Transform.prototype.getStdinList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 7));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.setStdinList = function(value) {
  return jspb.Message.setField(this, 7, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.addStdin = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 7, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.clearStdinList = function() {
  return this.setStdinList([]);
};


/**
 * repeated string err_stdin = 8;
 * @return {!Array<string>}
 */
proto.pps_v2.Transform.prototype.getErrStdinList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 8));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.setErrStdinList = function(value) {
  return jspb.Message.setField(this, 8, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.addErrStdin = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 8, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.clearErrStdinList = function() {
  return this.setErrStdinList([]);
};


/**
 * repeated int64 accept_return_code = 9;
 * @return {!Array<number>}
 */
proto.pps_v2.Transform.prototype.getAcceptReturnCodeList = function() {
  return /** @type {!Array<number>} */ (jspb.Message.getRepeatedField(this, 9));
};


/**
 * @param {!Array<number>} value
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.setAcceptReturnCodeList = function(value) {
  return jspb.Message.setField(this, 9, value || []);
};


/**
 * @param {number} value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.addAcceptReturnCode = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 9, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.clearAcceptReturnCodeList = function() {
  return this.setAcceptReturnCodeList([]);
};


/**
 * optional bool debug = 10;
 * @return {boolean}
 */
proto.pps_v2.Transform.prototype.getDebug = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 10, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.setDebug = function(value) {
  return jspb.Message.setProto3BooleanField(this, 10, value);
};


/**
 * optional string user = 11;
 * @return {string}
 */
proto.pps_v2.Transform.prototype.getUser = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 11, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.setUser = function(value) {
  return jspb.Message.setProto3StringField(this, 11, value);
};


/**
 * optional string working_dir = 12;
 * @return {string}
 */
proto.pps_v2.Transform.prototype.getWorkingDir = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 12, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.setWorkingDir = function(value) {
  return jspb.Message.setProto3StringField(this, 12, value);
};


/**
 * optional string dockerfile = 13;
 * @return {string}
 */
proto.pps_v2.Transform.prototype.getDockerfile = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 13, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Transform} returns this
 */
proto.pps_v2.Transform.prototype.setDockerfile = function(value) {
  return jspb.Message.setProto3StringField(this, 13, value);
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
proto.pps_v2.TFJob.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.TFJob.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.TFJob} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.TFJob.toObject = function(includeInstance, msg) {
  var f, obj = {
    tfJob: jspb.Message.getFieldWithDefault(msg, 1, "")
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
 * @return {!proto.pps_v2.TFJob}
 */
proto.pps_v2.TFJob.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.TFJob;
  return proto.pps_v2.TFJob.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.TFJob} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.TFJob}
 */
proto.pps_v2.TFJob.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setTfJob(value);
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
proto.pps_v2.TFJob.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.TFJob.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.TFJob} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.TFJob.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTfJob();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string tf_job = 1;
 * @return {string}
 */
proto.pps_v2.TFJob.prototype.getTfJob = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.TFJob} returns this
 */
proto.pps_v2.TFJob.prototype.setTfJob = function(value) {
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
proto.pps_v2.Egress.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Egress.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Egress} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Egress.toObject = function(includeInstance, msg) {
  var f, obj = {
    url: jspb.Message.getFieldWithDefault(msg, 1, "")
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
 * @return {!proto.pps_v2.Egress}
 */
proto.pps_v2.Egress.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Egress;
  return proto.pps_v2.Egress.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Egress} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Egress}
 */
proto.pps_v2.Egress.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
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
proto.pps_v2.Egress.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Egress.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Egress} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Egress.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getUrl();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string URL = 1;
 * @return {string}
 */
proto.pps_v2.Egress.prototype.getUrl = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Egress} returns this
 */
proto.pps_v2.Egress.prototype.setUrl = function(value) {
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
proto.pps_v2.Job.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Job.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Job} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Job.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f),
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
 * @return {!proto.pps_v2.Job}
 */
proto.pps_v2.Job.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Job;
  return proto.pps_v2.Job.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Job} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Job}
 */
proto.pps_v2.Job.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
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
proto.pps_v2.Job.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Job.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Job} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Job.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
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
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.Job.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.Job} returns this
*/
proto.pps_v2.Job.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.Job} returns this
 */
proto.pps_v2.Job.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.Job.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string id = 2;
 * @return {string}
 */
proto.pps_v2.Job.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Job} returns this
 */
proto.pps_v2.Job.prototype.setId = function(value) {
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
proto.pps_v2.Metadata.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Metadata.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Metadata} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Metadata.toObject = function(includeInstance, msg) {
  var f, obj = {
    annotationsMap: (f = msg.getAnnotationsMap()) ? f.toObject(includeInstance, undefined) : [],
    labelsMap: (f = msg.getLabelsMap()) ? f.toObject(includeInstance, undefined) : []
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
 * @return {!proto.pps_v2.Metadata}
 */
proto.pps_v2.Metadata.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Metadata;
  return proto.pps_v2.Metadata.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Metadata} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Metadata}
 */
proto.pps_v2.Metadata.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getAnnotationsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
         });
      break;
    case 2:
      var value = msg.getLabelsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
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
proto.pps_v2.Metadata.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Metadata.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Metadata} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Metadata.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAnnotationsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getLabelsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(2, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
};


/**
 * map<string, string> annotations = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.pps_v2.Metadata.prototype.getAnnotationsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pps_v2.Metadata} returns this
 */
proto.pps_v2.Metadata.prototype.clearAnnotationsMap = function() {
  this.getAnnotationsMap().clear();
  return this;};


/**
 * map<string, string> labels = 2;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.pps_v2.Metadata.prototype.getLabelsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 2, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pps_v2.Metadata} returns this
 */
proto.pps_v2.Metadata.prototype.clearLabelsMap = function() {
  this.getLabelsMap().clear();
  return this;};





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
proto.pps_v2.Service.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Service.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Service} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Service.toObject = function(includeInstance, msg) {
  var f, obj = {
    internalPort: jspb.Message.getFieldWithDefault(msg, 1, 0),
    externalPort: jspb.Message.getFieldWithDefault(msg, 2, 0),
    ip: jspb.Message.getFieldWithDefault(msg, 3, ""),
    type: jspb.Message.getFieldWithDefault(msg, 4, "")
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
 * @return {!proto.pps_v2.Service}
 */
proto.pps_v2.Service.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Service;
  return proto.pps_v2.Service.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Service} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Service}
 */
proto.pps_v2.Service.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setInternalPort(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setExternalPort(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setIp(value);
      break;
    case 4:
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
proto.pps_v2.Service.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Service.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Service} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Service.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getInternalPort();
  if (f !== 0) {
    writer.writeInt32(
      1,
      f
    );
  }
  f = message.getExternalPort();
  if (f !== 0) {
    writer.writeInt32(
      2,
      f
    );
  }
  f = message.getIp();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getType();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * optional int32 internal_port = 1;
 * @return {number}
 */
proto.pps_v2.Service.prototype.getInternalPort = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.Service} returns this
 */
proto.pps_v2.Service.prototype.setInternalPort = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional int32 external_port = 2;
 * @return {number}
 */
proto.pps_v2.Service.prototype.getExternalPort = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.Service} returns this
 */
proto.pps_v2.Service.prototype.setExternalPort = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional string ip = 3;
 * @return {string}
 */
proto.pps_v2.Service.prototype.getIp = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Service} returns this
 */
proto.pps_v2.Service.prototype.setIp = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string type = 4;
 * @return {string}
 */
proto.pps_v2.Service.prototype.getType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Service} returns this
 */
proto.pps_v2.Service.prototype.setType = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
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
proto.pps_v2.Spout.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Spout.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Spout} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Spout.toObject = function(includeInstance, msg) {
  var f, obj = {
    service: (f = msg.getService()) && proto.pps_v2.Service.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.Spout}
 */
proto.pps_v2.Spout.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Spout;
  return proto.pps_v2.Spout.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Spout} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Spout}
 */
proto.pps_v2.Spout.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Service;
      reader.readMessage(value,proto.pps_v2.Service.deserializeBinaryFromReader);
      msg.setService(value);
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
proto.pps_v2.Spout.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Spout.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Spout} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Spout.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getService();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Service.serializeBinaryToWriter
    );
  }
};


/**
 * optional Service service = 1;
 * @return {?proto.pps_v2.Service}
 */
proto.pps_v2.Spout.prototype.getService = function() {
  return /** @type{?proto.pps_v2.Service} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Service, 1));
};


/**
 * @param {?proto.pps_v2.Service|undefined} value
 * @return {!proto.pps_v2.Spout} returns this
*/
proto.pps_v2.Spout.prototype.setService = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.Spout} returns this
 */
proto.pps_v2.Spout.prototype.clearService = function() {
  return this.setService(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.Spout.prototype.hasService = function() {
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
proto.pps_v2.PFSInput.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.PFSInput.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.PFSInput} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.PFSInput.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    repo: jspb.Message.getFieldWithDefault(msg, 2, ""),
    repoType: jspb.Message.getFieldWithDefault(msg, 13, ""),
    branch: jspb.Message.getFieldWithDefault(msg, 3, ""),
    commit: jspb.Message.getFieldWithDefault(msg, 4, ""),
    glob: jspb.Message.getFieldWithDefault(msg, 5, ""),
    joinOn: jspb.Message.getFieldWithDefault(msg, 6, ""),
    outerJoin: jspb.Message.getBooleanFieldWithDefault(msg, 7, false),
    groupBy: jspb.Message.getFieldWithDefault(msg, 8, ""),
    lazy: jspb.Message.getBooleanFieldWithDefault(msg, 9, false),
    emptyFiles: jspb.Message.getBooleanFieldWithDefault(msg, 10, false),
    s3: jspb.Message.getBooleanFieldWithDefault(msg, 11, false),
    trigger: (f = msg.getTrigger()) && pfs_pfs_pb.Trigger.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.PFSInput}
 */
proto.pps_v2.PFSInput.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.PFSInput;
  return proto.pps_v2.PFSInput.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.PFSInput} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.PFSInput}
 */
proto.pps_v2.PFSInput.deserializeBinaryFromReader = function(msg, reader) {
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
      msg.setRepo(value);
      break;
    case 13:
      var value = /** @type {string} */ (reader.readString());
      msg.setRepoType(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setBranch(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setCommit(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setGlob(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setJoinOn(value);
      break;
    case 7:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setOuterJoin(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setGroupBy(value);
      break;
    case 9:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setLazy(value);
      break;
    case 10:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEmptyFiles(value);
      break;
    case 11:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setS3(value);
      break;
    case 12:
      var value = new pfs_pfs_pb.Trigger;
      reader.readMessage(value,pfs_pfs_pb.Trigger.deserializeBinaryFromReader);
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
proto.pps_v2.PFSInput.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.PFSInput.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.PFSInput} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.PFSInput.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRepo();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getRepoType();
  if (f.length > 0) {
    writer.writeString(
      13,
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
  f = message.getCommit();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getGlob();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getJoinOn();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getOuterJoin();
  if (f) {
    writer.writeBool(
      7,
      f
    );
  }
  f = message.getGroupBy();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
  f = message.getLazy();
  if (f) {
    writer.writeBool(
      9,
      f
    );
  }
  f = message.getEmptyFiles();
  if (f) {
    writer.writeBool(
      10,
      f
    );
  }
  f = message.getS3();
  if (f) {
    writer.writeBool(
      11,
      f
    );
  }
  f = message.getTrigger();
  if (f != null) {
    writer.writeMessage(
      12,
      f,
      pfs_pfs_pb.Trigger.serializeBinaryToWriter
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pps_v2.PFSInput.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string repo = 2;
 * @return {string}
 */
proto.pps_v2.PFSInput.prototype.getRepo = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setRepo = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string repo_type = 13;
 * @return {string}
 */
proto.pps_v2.PFSInput.prototype.getRepoType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 13, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setRepoType = function(value) {
  return jspb.Message.setProto3StringField(this, 13, value);
};


/**
 * optional string branch = 3;
 * @return {string}
 */
proto.pps_v2.PFSInput.prototype.getBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string commit = 4;
 * @return {string}
 */
proto.pps_v2.PFSInput.prototype.getCommit = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setCommit = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string glob = 5;
 * @return {string}
 */
proto.pps_v2.PFSInput.prototype.getGlob = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setGlob = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string join_on = 6;
 * @return {string}
 */
proto.pps_v2.PFSInput.prototype.getJoinOn = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setJoinOn = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional bool outer_join = 7;
 * @return {boolean}
 */
proto.pps_v2.PFSInput.prototype.getOuterJoin = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 7, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setOuterJoin = function(value) {
  return jspb.Message.setProto3BooleanField(this, 7, value);
};


/**
 * optional string group_by = 8;
 * @return {string}
 */
proto.pps_v2.PFSInput.prototype.getGroupBy = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setGroupBy = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};


/**
 * optional bool lazy = 9;
 * @return {boolean}
 */
proto.pps_v2.PFSInput.prototype.getLazy = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 9, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setLazy = function(value) {
  return jspb.Message.setProto3BooleanField(this, 9, value);
};


/**
 * optional bool empty_files = 10;
 * @return {boolean}
 */
proto.pps_v2.PFSInput.prototype.getEmptyFiles = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 10, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setEmptyFiles = function(value) {
  return jspb.Message.setProto3BooleanField(this, 10, value);
};


/**
 * optional bool s3 = 11;
 * @return {boolean}
 */
proto.pps_v2.PFSInput.prototype.getS3 = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 11, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.setS3 = function(value) {
  return jspb.Message.setProto3BooleanField(this, 11, value);
};


/**
 * optional pfs_v2.Trigger trigger = 12;
 * @return {?proto.pfs_v2.Trigger}
 */
proto.pps_v2.PFSInput.prototype.getTrigger = function() {
  return /** @type{?proto.pfs_v2.Trigger} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Trigger, 12));
};


/**
 * @param {?proto.pfs_v2.Trigger|undefined} value
 * @return {!proto.pps_v2.PFSInput} returns this
*/
proto.pps_v2.PFSInput.prototype.setTrigger = function(value) {
  return jspb.Message.setWrapperField(this, 12, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PFSInput} returns this
 */
proto.pps_v2.PFSInput.prototype.clearTrigger = function() {
  return this.setTrigger(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PFSInput.prototype.hasTrigger = function() {
  return jspb.Message.getField(this, 12) != null;
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
proto.pps_v2.CronInput.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.CronInput.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.CronInput} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.CronInput.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    repo: jspb.Message.getFieldWithDefault(msg, 2, ""),
    repoType: jspb.Message.getFieldWithDefault(msg, 13, ""),
    commit: jspb.Message.getFieldWithDefault(msg, 3, ""),
    spec: jspb.Message.getFieldWithDefault(msg, 4, ""),
    overwrite: jspb.Message.getBooleanFieldWithDefault(msg, 5, false),
    start: (f = msg.getStart()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.CronInput}
 */
proto.pps_v2.CronInput.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.CronInput;
  return proto.pps_v2.CronInput.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.CronInput} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.CronInput}
 */
proto.pps_v2.CronInput.deserializeBinaryFromReader = function(msg, reader) {
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
      msg.setRepo(value);
      break;
    case 13:
      var value = /** @type {string} */ (reader.readString());
      msg.setRepoType(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setCommit(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setSpec(value);
      break;
    case 5:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setOverwrite(value);
      break;
    case 6:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setStart(value);
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
proto.pps_v2.CronInput.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.CronInput.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.CronInput} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.CronInput.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRepo();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getRepoType();
  if (f.length > 0) {
    writer.writeString(
      13,
      f
    );
  }
  f = message.getCommit();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getSpec();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getOverwrite();
  if (f) {
    writer.writeBool(
      5,
      f
    );
  }
  f = message.getStart();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pps_v2.CronInput.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CronInput} returns this
 */
proto.pps_v2.CronInput.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string repo = 2;
 * @return {string}
 */
proto.pps_v2.CronInput.prototype.getRepo = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CronInput} returns this
 */
proto.pps_v2.CronInput.prototype.setRepo = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string repo_type = 13;
 * @return {string}
 */
proto.pps_v2.CronInput.prototype.getRepoType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 13, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CronInput} returns this
 */
proto.pps_v2.CronInput.prototype.setRepoType = function(value) {
  return jspb.Message.setProto3StringField(this, 13, value);
};


/**
 * optional string commit = 3;
 * @return {string}
 */
proto.pps_v2.CronInput.prototype.getCommit = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CronInput} returns this
 */
proto.pps_v2.CronInput.prototype.setCommit = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string spec = 4;
 * @return {string}
 */
proto.pps_v2.CronInput.prototype.getSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CronInput} returns this
 */
proto.pps_v2.CronInput.prototype.setSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional bool overwrite = 5;
 * @return {boolean}
 */
proto.pps_v2.CronInput.prototype.getOverwrite = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 5, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.CronInput} returns this
 */
proto.pps_v2.CronInput.prototype.setOverwrite = function(value) {
  return jspb.Message.setProto3BooleanField(this, 5, value);
};


/**
 * optional google.protobuf.Timestamp start = 6;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps_v2.CronInput.prototype.getStart = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 6));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps_v2.CronInput} returns this
*/
proto.pps_v2.CronInput.prototype.setStart = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CronInput} returns this
 */
proto.pps_v2.CronInput.prototype.clearStart = function() {
  return this.setStart(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CronInput.prototype.hasStart = function() {
  return jspb.Message.getField(this, 6) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.Input.repeatedFields_ = [2,3,4,5];



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
proto.pps_v2.Input.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Input.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Input} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Input.toObject = function(includeInstance, msg) {
  var f, obj = {
    pfs: (f = msg.getPfs()) && proto.pps_v2.PFSInput.toObject(includeInstance, f),
    joinList: jspb.Message.toObjectList(msg.getJoinList(),
    proto.pps_v2.Input.toObject, includeInstance),
    groupList: jspb.Message.toObjectList(msg.getGroupList(),
    proto.pps_v2.Input.toObject, includeInstance),
    crossList: jspb.Message.toObjectList(msg.getCrossList(),
    proto.pps_v2.Input.toObject, includeInstance),
    unionList: jspb.Message.toObjectList(msg.getUnionList(),
    proto.pps_v2.Input.toObject, includeInstance),
    cron: (f = msg.getCron()) && proto.pps_v2.CronInput.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.Input}
 */
proto.pps_v2.Input.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Input;
  return proto.pps_v2.Input.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Input} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Input}
 */
proto.pps_v2.Input.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.PFSInput;
      reader.readMessage(value,proto.pps_v2.PFSInput.deserializeBinaryFromReader);
      msg.setPfs(value);
      break;
    case 2:
      var value = new proto.pps_v2.Input;
      reader.readMessage(value,proto.pps_v2.Input.deserializeBinaryFromReader);
      msg.addJoin(value);
      break;
    case 3:
      var value = new proto.pps_v2.Input;
      reader.readMessage(value,proto.pps_v2.Input.deserializeBinaryFromReader);
      msg.addGroup(value);
      break;
    case 4:
      var value = new proto.pps_v2.Input;
      reader.readMessage(value,proto.pps_v2.Input.deserializeBinaryFromReader);
      msg.addCross(value);
      break;
    case 5:
      var value = new proto.pps_v2.Input;
      reader.readMessage(value,proto.pps_v2.Input.deserializeBinaryFromReader);
      msg.addUnion(value);
      break;
    case 6:
      var value = new proto.pps_v2.CronInput;
      reader.readMessage(value,proto.pps_v2.CronInput.deserializeBinaryFromReader);
      msg.setCron(value);
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
proto.pps_v2.Input.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Input.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Input} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Input.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPfs();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.PFSInput.serializeBinaryToWriter
    );
  }
  f = message.getJoinList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.pps_v2.Input.serializeBinaryToWriter
    );
  }
  f = message.getGroupList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      3,
      f,
      proto.pps_v2.Input.serializeBinaryToWriter
    );
  }
  f = message.getCrossList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      4,
      f,
      proto.pps_v2.Input.serializeBinaryToWriter
    );
  }
  f = message.getUnionList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      5,
      f,
      proto.pps_v2.Input.serializeBinaryToWriter
    );
  }
  f = message.getCron();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      proto.pps_v2.CronInput.serializeBinaryToWriter
    );
  }
};


/**
 * optional PFSInput pfs = 1;
 * @return {?proto.pps_v2.PFSInput}
 */
proto.pps_v2.Input.prototype.getPfs = function() {
  return /** @type{?proto.pps_v2.PFSInput} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.PFSInput, 1));
};


/**
 * @param {?proto.pps_v2.PFSInput|undefined} value
 * @return {!proto.pps_v2.Input} returns this
*/
proto.pps_v2.Input.prototype.setPfs = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.Input} returns this
 */
proto.pps_v2.Input.prototype.clearPfs = function() {
  return this.setPfs(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.Input.prototype.hasPfs = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * repeated Input join = 2;
 * @return {!Array<!proto.pps_v2.Input>}
 */
proto.pps_v2.Input.prototype.getJoinList = function() {
  return /** @type{!Array<!proto.pps_v2.Input>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps_v2.Input, 2));
};


/**
 * @param {!Array<!proto.pps_v2.Input>} value
 * @return {!proto.pps_v2.Input} returns this
*/
proto.pps_v2.Input.prototype.setJoinList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pps_v2.Input=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.Input}
 */
proto.pps_v2.Input.prototype.addJoin = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pps_v2.Input, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.Input} returns this
 */
proto.pps_v2.Input.prototype.clearJoinList = function() {
  return this.setJoinList([]);
};


/**
 * repeated Input group = 3;
 * @return {!Array<!proto.pps_v2.Input>}
 */
proto.pps_v2.Input.prototype.getGroupList = function() {
  return /** @type{!Array<!proto.pps_v2.Input>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps_v2.Input, 3));
};


/**
 * @param {!Array<!proto.pps_v2.Input>} value
 * @return {!proto.pps_v2.Input} returns this
*/
proto.pps_v2.Input.prototype.setGroupList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 3, value);
};


/**
 * @param {!proto.pps_v2.Input=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.Input}
 */
proto.pps_v2.Input.prototype.addGroup = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 3, opt_value, proto.pps_v2.Input, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.Input} returns this
 */
proto.pps_v2.Input.prototype.clearGroupList = function() {
  return this.setGroupList([]);
};


/**
 * repeated Input cross = 4;
 * @return {!Array<!proto.pps_v2.Input>}
 */
proto.pps_v2.Input.prototype.getCrossList = function() {
  return /** @type{!Array<!proto.pps_v2.Input>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps_v2.Input, 4));
};


/**
 * @param {!Array<!proto.pps_v2.Input>} value
 * @return {!proto.pps_v2.Input} returns this
*/
proto.pps_v2.Input.prototype.setCrossList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 4, value);
};


/**
 * @param {!proto.pps_v2.Input=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.Input}
 */
proto.pps_v2.Input.prototype.addCross = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 4, opt_value, proto.pps_v2.Input, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.Input} returns this
 */
proto.pps_v2.Input.prototype.clearCrossList = function() {
  return this.setCrossList([]);
};


/**
 * repeated Input union = 5;
 * @return {!Array<!proto.pps_v2.Input>}
 */
proto.pps_v2.Input.prototype.getUnionList = function() {
  return /** @type{!Array<!proto.pps_v2.Input>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps_v2.Input, 5));
};


/**
 * @param {!Array<!proto.pps_v2.Input>} value
 * @return {!proto.pps_v2.Input} returns this
*/
proto.pps_v2.Input.prototype.setUnionList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.pps_v2.Input=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.Input}
 */
proto.pps_v2.Input.prototype.addUnion = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.pps_v2.Input, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.Input} returns this
 */
proto.pps_v2.Input.prototype.clearUnionList = function() {
  return this.setUnionList([]);
};


/**
 * optional CronInput cron = 6;
 * @return {?proto.pps_v2.CronInput}
 */
proto.pps_v2.Input.prototype.getCron = function() {
  return /** @type{?proto.pps_v2.CronInput} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.CronInput, 6));
};


/**
 * @param {?proto.pps_v2.CronInput|undefined} value
 * @return {!proto.pps_v2.Input} returns this
*/
proto.pps_v2.Input.prototype.setCron = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.Input} returns this
 */
proto.pps_v2.Input.prototype.clearCron = function() {
  return this.setCron(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.Input.prototype.hasCron = function() {
  return jspb.Message.getField(this, 6) != null;
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
proto.pps_v2.JobInput.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.JobInput.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.JobInput} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.JobInput.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    commit: (f = msg.getCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    glob: jspb.Message.getFieldWithDefault(msg, 3, ""),
    lazy: jspb.Message.getBooleanFieldWithDefault(msg, 4, false)
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
 * @return {!proto.pps_v2.JobInput}
 */
proto.pps_v2.JobInput.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.JobInput;
  return proto.pps_v2.JobInput.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.JobInput} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.JobInput}
 */
proto.pps_v2.JobInput.deserializeBinaryFromReader = function(msg, reader) {
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
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setGlob(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setLazy(value);
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
proto.pps_v2.JobInput.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.JobInput.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.JobInput} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.JobInput.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getGlob();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getLazy();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pps_v2.JobInput.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.JobInput} returns this
 */
proto.pps_v2.JobInput.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional pfs_v2.Commit commit = 2;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pps_v2.JobInput.prototype.getCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 2));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pps_v2.JobInput} returns this
*/
proto.pps_v2.JobInput.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInput} returns this
 */
proto.pps_v2.JobInput.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInput.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional string glob = 3;
 * @return {string}
 */
proto.pps_v2.JobInput.prototype.getGlob = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.JobInput} returns this
 */
proto.pps_v2.JobInput.prototype.setGlob = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional bool lazy = 4;
 * @return {boolean}
 */
proto.pps_v2.JobInput.prototype.getLazy = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.JobInput} returns this
 */
proto.pps_v2.JobInput.prototype.setLazy = function(value) {
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
proto.pps_v2.ParallelismSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.ParallelismSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.ParallelismSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ParallelismSpec.toObject = function(includeInstance, msg) {
  var f, obj = {
    constant: jspb.Message.getFieldWithDefault(msg, 1, 0)
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
 * @return {!proto.pps_v2.ParallelismSpec}
 */
proto.pps_v2.ParallelismSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.ParallelismSpec;
  return proto.pps_v2.ParallelismSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.ParallelismSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.ParallelismSpec}
 */
proto.pps_v2.ParallelismSpec.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setConstant(value);
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
proto.pps_v2.ParallelismSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.ParallelismSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.ParallelismSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ParallelismSpec.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getConstant();
  if (f !== 0) {
    writer.writeUint64(
      1,
      f
    );
  }
};


/**
 * optional uint64 constant = 1;
 * @return {number}
 */
proto.pps_v2.ParallelismSpec.prototype.getConstant = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.ParallelismSpec} returns this
 */
proto.pps_v2.ParallelismSpec.prototype.setConstant = function(value) {
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
proto.pps_v2.InputFile.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.InputFile.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.InputFile} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InputFile.toObject = function(includeInstance, msg) {
  var f, obj = {
    path: jspb.Message.getFieldWithDefault(msg, 1, ""),
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
 * @return {!proto.pps_v2.InputFile}
 */
proto.pps_v2.InputFile.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.InputFile;
  return proto.pps_v2.InputFile.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.InputFile} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.InputFile}
 */
proto.pps_v2.InputFile.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps_v2.InputFile.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.InputFile.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.InputFile} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InputFile.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getHash_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      2,
      f
    );
  }
};


/**
 * optional string path = 1;
 * @return {string}
 */
proto.pps_v2.InputFile.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.InputFile} returns this
 */
proto.pps_v2.InputFile.prototype.setPath = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional bytes hash = 2;
 * @return {string}
 */
proto.pps_v2.InputFile.prototype.getHash = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * optional bytes hash = 2;
 * This is a type-conversion wrapper around `getHash()`
 * @return {string}
 */
proto.pps_v2.InputFile.prototype.getHash_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getHash()));
};


/**
 * optional bytes hash = 2;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getHash()`
 * @return {!Uint8Array}
 */
proto.pps_v2.InputFile.prototype.getHash_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getHash()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pps_v2.InputFile} returns this
 */
proto.pps_v2.InputFile.prototype.setHash = function(value) {
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
proto.pps_v2.Datum.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Datum.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Datum} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Datum.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps_v2.Job.toObject(includeInstance, f),
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
 * @return {!proto.pps_v2.Datum}
 */
proto.pps_v2.Datum.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Datum;
  return proto.pps_v2.Datum.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Datum} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Datum}
 */
proto.pps_v2.Datum.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Job;
      reader.readMessage(value,proto.pps_v2.Job.deserializeBinaryFromReader);
      msg.setJob(value);
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
proto.pps_v2.Datum.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Datum.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Datum} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Datum.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Job.serializeBinaryToWriter
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
 * optional Job job = 1;
 * @return {?proto.pps_v2.Job}
 */
proto.pps_v2.Datum.prototype.getJob = function() {
  return /** @type{?proto.pps_v2.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Job, 1));
};


/**
 * @param {?proto.pps_v2.Job|undefined} value
 * @return {!proto.pps_v2.Datum} returns this
*/
proto.pps_v2.Datum.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.Datum} returns this
 */
proto.pps_v2.Datum.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.Datum.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string id = 2;
 * @return {string}
 */
proto.pps_v2.Datum.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Datum} returns this
 */
proto.pps_v2.Datum.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.DatumInfo.repeatedFields_ = [5];



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
proto.pps_v2.DatumInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.DatumInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.DatumInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DatumInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    datum: (f = msg.getDatum()) && proto.pps_v2.Datum.toObject(includeInstance, f),
    state: jspb.Message.getFieldWithDefault(msg, 2, 0),
    stats: (f = msg.getStats()) && proto.pps_v2.ProcessStats.toObject(includeInstance, f),
    pfsState: (f = msg.getPfsState()) && pfs_pfs_pb.File.toObject(includeInstance, f),
    dataList: jspb.Message.toObjectList(msg.getDataList(),
    pfs_pfs_pb.FileInfo.toObject, includeInstance)
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
 * @return {!proto.pps_v2.DatumInfo}
 */
proto.pps_v2.DatumInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.DatumInfo;
  return proto.pps_v2.DatumInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.DatumInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.DatumInfo}
 */
proto.pps_v2.DatumInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Datum;
      reader.readMessage(value,proto.pps_v2.Datum.deserializeBinaryFromReader);
      msg.setDatum(value);
      break;
    case 2:
      var value = /** @type {!proto.pps_v2.DatumState} */ (reader.readEnum());
      msg.setState(value);
      break;
    case 3:
      var value = new proto.pps_v2.ProcessStats;
      reader.readMessage(value,proto.pps_v2.ProcessStats.deserializeBinaryFromReader);
      msg.setStats(value);
      break;
    case 4:
      var value = new pfs_pfs_pb.File;
      reader.readMessage(value,pfs_pfs_pb.File.deserializeBinaryFromReader);
      msg.setPfsState(value);
      break;
    case 5:
      var value = new pfs_pfs_pb.FileInfo;
      reader.readMessage(value,pfs_pfs_pb.FileInfo.deserializeBinaryFromReader);
      msg.addData(value);
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
proto.pps_v2.DatumInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.DatumInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.DatumInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DatumInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getDatum();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Datum.serializeBinaryToWriter
    );
  }
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
  f = message.getStats();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pps_v2.ProcessStats.serializeBinaryToWriter
    );
  }
  f = message.getPfsState();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      pfs_pfs_pb.File.serializeBinaryToWriter
    );
  }
  f = message.getDataList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      5,
      f,
      pfs_pfs_pb.FileInfo.serializeBinaryToWriter
    );
  }
};


/**
 * optional Datum datum = 1;
 * @return {?proto.pps_v2.Datum}
 */
proto.pps_v2.DatumInfo.prototype.getDatum = function() {
  return /** @type{?proto.pps_v2.Datum} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Datum, 1));
};


/**
 * @param {?proto.pps_v2.Datum|undefined} value
 * @return {!proto.pps_v2.DatumInfo} returns this
*/
proto.pps_v2.DatumInfo.prototype.setDatum = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.DatumInfo} returns this
 */
proto.pps_v2.DatumInfo.prototype.clearDatum = function() {
  return this.setDatum(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.DatumInfo.prototype.hasDatum = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional DatumState state = 2;
 * @return {!proto.pps_v2.DatumState}
 */
proto.pps_v2.DatumInfo.prototype.getState = function() {
  return /** @type {!proto.pps_v2.DatumState} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.pps_v2.DatumState} value
 * @return {!proto.pps_v2.DatumInfo} returns this
 */
proto.pps_v2.DatumInfo.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 2, value);
};


/**
 * optional ProcessStats stats = 3;
 * @return {?proto.pps_v2.ProcessStats}
 */
proto.pps_v2.DatumInfo.prototype.getStats = function() {
  return /** @type{?proto.pps_v2.ProcessStats} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ProcessStats, 3));
};


/**
 * @param {?proto.pps_v2.ProcessStats|undefined} value
 * @return {!proto.pps_v2.DatumInfo} returns this
*/
proto.pps_v2.DatumInfo.prototype.setStats = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.DatumInfo} returns this
 */
proto.pps_v2.DatumInfo.prototype.clearStats = function() {
  return this.setStats(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.DatumInfo.prototype.hasStats = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional pfs_v2.File pfs_state = 4;
 * @return {?proto.pfs_v2.File}
 */
proto.pps_v2.DatumInfo.prototype.getPfsState = function() {
  return /** @type{?proto.pfs_v2.File} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.File, 4));
};


/**
 * @param {?proto.pfs_v2.File|undefined} value
 * @return {!proto.pps_v2.DatumInfo} returns this
*/
proto.pps_v2.DatumInfo.prototype.setPfsState = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.DatumInfo} returns this
 */
proto.pps_v2.DatumInfo.prototype.clearPfsState = function() {
  return this.setPfsState(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.DatumInfo.prototype.hasPfsState = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * repeated pfs_v2.FileInfo data = 5;
 * @return {!Array<!proto.pfs_v2.FileInfo>}
 */
proto.pps_v2.DatumInfo.prototype.getDataList = function() {
  return /** @type{!Array<!proto.pfs_v2.FileInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, pfs_pfs_pb.FileInfo, 5));
};


/**
 * @param {!Array<!proto.pfs_v2.FileInfo>} value
 * @return {!proto.pps_v2.DatumInfo} returns this
*/
proto.pps_v2.DatumInfo.prototype.setDataList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.pfs_v2.FileInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.FileInfo}
 */
proto.pps_v2.DatumInfo.prototype.addData = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.pfs_v2.FileInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.DatumInfo} returns this
 */
proto.pps_v2.DatumInfo.prototype.clearDataList = function() {
  return this.setDataList([]);
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
proto.pps_v2.Aggregate.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Aggregate.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Aggregate} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Aggregate.toObject = function(includeInstance, msg) {
  var f, obj = {
    count: jspb.Message.getFieldWithDefault(msg, 1, 0),
    mean: jspb.Message.getFloatingPointFieldWithDefault(msg, 2, 0.0),
    stddev: jspb.Message.getFloatingPointFieldWithDefault(msg, 3, 0.0),
    fifthPercentile: jspb.Message.getFloatingPointFieldWithDefault(msg, 4, 0.0),
    ninetyFifthPercentile: jspb.Message.getFloatingPointFieldWithDefault(msg, 5, 0.0)
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
 * @return {!proto.pps_v2.Aggregate}
 */
proto.pps_v2.Aggregate.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Aggregate;
  return proto.pps_v2.Aggregate.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Aggregate} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Aggregate}
 */
proto.pps_v2.Aggregate.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setCount(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setMean(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setStddev(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setFifthPercentile(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setNinetyFifthPercentile(value);
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
proto.pps_v2.Aggregate.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Aggregate.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Aggregate} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Aggregate.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCount();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getMean();
  if (f !== 0.0) {
    writer.writeDouble(
      2,
      f
    );
  }
  f = message.getStddev();
  if (f !== 0.0) {
    writer.writeDouble(
      3,
      f
    );
  }
  f = message.getFifthPercentile();
  if (f !== 0.0) {
    writer.writeDouble(
      4,
      f
    );
  }
  f = message.getNinetyFifthPercentile();
  if (f !== 0.0) {
    writer.writeDouble(
      5,
      f
    );
  }
};


/**
 * optional int64 count = 1;
 * @return {number}
 */
proto.pps_v2.Aggregate.prototype.getCount = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.Aggregate} returns this
 */
proto.pps_v2.Aggregate.prototype.setCount = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional double mean = 2;
 * @return {number}
 */
proto.pps_v2.Aggregate.prototype.getMean = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 2, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.Aggregate} returns this
 */
proto.pps_v2.Aggregate.prototype.setMean = function(value) {
  return jspb.Message.setProto3FloatField(this, 2, value);
};


/**
 * optional double stddev = 3;
 * @return {number}
 */
proto.pps_v2.Aggregate.prototype.getStddev = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 3, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.Aggregate} returns this
 */
proto.pps_v2.Aggregate.prototype.setStddev = function(value) {
  return jspb.Message.setProto3FloatField(this, 3, value);
};


/**
 * optional double fifth_percentile = 4;
 * @return {number}
 */
proto.pps_v2.Aggregate.prototype.getFifthPercentile = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 4, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.Aggregate} returns this
 */
proto.pps_v2.Aggregate.prototype.setFifthPercentile = function(value) {
  return jspb.Message.setProto3FloatField(this, 4, value);
};


/**
 * optional double ninety_fifth_percentile = 5;
 * @return {number}
 */
proto.pps_v2.Aggregate.prototype.getNinetyFifthPercentile = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 5, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.Aggregate} returns this
 */
proto.pps_v2.Aggregate.prototype.setNinetyFifthPercentile = function(value) {
  return jspb.Message.setProto3FloatField(this, 5, value);
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
proto.pps_v2.ProcessStats.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.ProcessStats.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.ProcessStats} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ProcessStats.toObject = function(includeInstance, msg) {
  var f, obj = {
    downloadTime: (f = msg.getDownloadTime()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    processTime: (f = msg.getProcessTime()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    uploadTime: (f = msg.getUploadTime()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    downloadBytes: jspb.Message.getFieldWithDefault(msg, 4, 0),
    uploadBytes: jspb.Message.getFieldWithDefault(msg, 5, 0)
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
 * @return {!proto.pps_v2.ProcessStats}
 */
proto.pps_v2.ProcessStats.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.ProcessStats;
  return proto.pps_v2.ProcessStats.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.ProcessStats} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.ProcessStats}
 */
proto.pps_v2.ProcessStats.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setDownloadTime(value);
      break;
    case 2:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setProcessTime(value);
      break;
    case 3:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setUploadTime(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setDownloadBytes(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setUploadBytes(value);
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
proto.pps_v2.ProcessStats.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.ProcessStats.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.ProcessStats} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ProcessStats.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getDownloadTime();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getProcessTime();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getUploadTime();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getDownloadBytes();
  if (f !== 0) {
    writer.writeUint64(
      4,
      f
    );
  }
  f = message.getUploadBytes();
  if (f !== 0) {
    writer.writeUint64(
      5,
      f
    );
  }
};


/**
 * optional google.protobuf.Duration download_time = 1;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps_v2.ProcessStats.prototype.getDownloadTime = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 1));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps_v2.ProcessStats} returns this
*/
proto.pps_v2.ProcessStats.prototype.setDownloadTime = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.ProcessStats} returns this
 */
proto.pps_v2.ProcessStats.prototype.clearDownloadTime = function() {
  return this.setDownloadTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.ProcessStats.prototype.hasDownloadTime = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional google.protobuf.Duration process_time = 2;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps_v2.ProcessStats.prototype.getProcessTime = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 2));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps_v2.ProcessStats} returns this
*/
proto.pps_v2.ProcessStats.prototype.setProcessTime = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.ProcessStats} returns this
 */
proto.pps_v2.ProcessStats.prototype.clearProcessTime = function() {
  return this.setProcessTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.ProcessStats.prototype.hasProcessTime = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional google.protobuf.Duration upload_time = 3;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps_v2.ProcessStats.prototype.getUploadTime = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 3));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps_v2.ProcessStats} returns this
*/
proto.pps_v2.ProcessStats.prototype.setUploadTime = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.ProcessStats} returns this
 */
proto.pps_v2.ProcessStats.prototype.clearUploadTime = function() {
  return this.setUploadTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.ProcessStats.prototype.hasUploadTime = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional uint64 download_bytes = 4;
 * @return {number}
 */
proto.pps_v2.ProcessStats.prototype.getDownloadBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.ProcessStats} returns this
 */
proto.pps_v2.ProcessStats.prototype.setDownloadBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional uint64 upload_bytes = 5;
 * @return {number}
 */
proto.pps_v2.ProcessStats.prototype.getUploadBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.ProcessStats} returns this
 */
proto.pps_v2.ProcessStats.prototype.setUploadBytes = function(value) {
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
proto.pps_v2.AggregateProcessStats.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.AggregateProcessStats.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.AggregateProcessStats} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.AggregateProcessStats.toObject = function(includeInstance, msg) {
  var f, obj = {
    downloadTime: (f = msg.getDownloadTime()) && proto.pps_v2.Aggregate.toObject(includeInstance, f),
    processTime: (f = msg.getProcessTime()) && proto.pps_v2.Aggregate.toObject(includeInstance, f),
    uploadTime: (f = msg.getUploadTime()) && proto.pps_v2.Aggregate.toObject(includeInstance, f),
    downloadBytes: (f = msg.getDownloadBytes()) && proto.pps_v2.Aggregate.toObject(includeInstance, f),
    uploadBytes: (f = msg.getUploadBytes()) && proto.pps_v2.Aggregate.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.AggregateProcessStats}
 */
proto.pps_v2.AggregateProcessStats.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.AggregateProcessStats;
  return proto.pps_v2.AggregateProcessStats.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.AggregateProcessStats} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.AggregateProcessStats}
 */
proto.pps_v2.AggregateProcessStats.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Aggregate;
      reader.readMessage(value,proto.pps_v2.Aggregate.deserializeBinaryFromReader);
      msg.setDownloadTime(value);
      break;
    case 2:
      var value = new proto.pps_v2.Aggregate;
      reader.readMessage(value,proto.pps_v2.Aggregate.deserializeBinaryFromReader);
      msg.setProcessTime(value);
      break;
    case 3:
      var value = new proto.pps_v2.Aggregate;
      reader.readMessage(value,proto.pps_v2.Aggregate.deserializeBinaryFromReader);
      msg.setUploadTime(value);
      break;
    case 4:
      var value = new proto.pps_v2.Aggregate;
      reader.readMessage(value,proto.pps_v2.Aggregate.deserializeBinaryFromReader);
      msg.setDownloadBytes(value);
      break;
    case 5:
      var value = new proto.pps_v2.Aggregate;
      reader.readMessage(value,proto.pps_v2.Aggregate.deserializeBinaryFromReader);
      msg.setUploadBytes(value);
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
proto.pps_v2.AggregateProcessStats.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.AggregateProcessStats.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.AggregateProcessStats} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.AggregateProcessStats.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getDownloadTime();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Aggregate.serializeBinaryToWriter
    );
  }
  f = message.getProcessTime();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps_v2.Aggregate.serializeBinaryToWriter
    );
  }
  f = message.getUploadTime();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pps_v2.Aggregate.serializeBinaryToWriter
    );
  }
  f = message.getDownloadBytes();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pps_v2.Aggregate.serializeBinaryToWriter
    );
  }
  f = message.getUploadBytes();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.pps_v2.Aggregate.serializeBinaryToWriter
    );
  }
};


/**
 * optional Aggregate download_time = 1;
 * @return {?proto.pps_v2.Aggregate}
 */
proto.pps_v2.AggregateProcessStats.prototype.getDownloadTime = function() {
  return /** @type{?proto.pps_v2.Aggregate} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Aggregate, 1));
};


/**
 * @param {?proto.pps_v2.Aggregate|undefined} value
 * @return {!proto.pps_v2.AggregateProcessStats} returns this
*/
proto.pps_v2.AggregateProcessStats.prototype.setDownloadTime = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.AggregateProcessStats} returns this
 */
proto.pps_v2.AggregateProcessStats.prototype.clearDownloadTime = function() {
  return this.setDownloadTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.AggregateProcessStats.prototype.hasDownloadTime = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Aggregate process_time = 2;
 * @return {?proto.pps_v2.Aggregate}
 */
proto.pps_v2.AggregateProcessStats.prototype.getProcessTime = function() {
  return /** @type{?proto.pps_v2.Aggregate} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Aggregate, 2));
};


/**
 * @param {?proto.pps_v2.Aggregate|undefined} value
 * @return {!proto.pps_v2.AggregateProcessStats} returns this
*/
proto.pps_v2.AggregateProcessStats.prototype.setProcessTime = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.AggregateProcessStats} returns this
 */
proto.pps_v2.AggregateProcessStats.prototype.clearProcessTime = function() {
  return this.setProcessTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.AggregateProcessStats.prototype.hasProcessTime = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional Aggregate upload_time = 3;
 * @return {?proto.pps_v2.Aggregate}
 */
proto.pps_v2.AggregateProcessStats.prototype.getUploadTime = function() {
  return /** @type{?proto.pps_v2.Aggregate} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Aggregate, 3));
};


/**
 * @param {?proto.pps_v2.Aggregate|undefined} value
 * @return {!proto.pps_v2.AggregateProcessStats} returns this
*/
proto.pps_v2.AggregateProcessStats.prototype.setUploadTime = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.AggregateProcessStats} returns this
 */
proto.pps_v2.AggregateProcessStats.prototype.clearUploadTime = function() {
  return this.setUploadTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.AggregateProcessStats.prototype.hasUploadTime = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional Aggregate download_bytes = 4;
 * @return {?proto.pps_v2.Aggregate}
 */
proto.pps_v2.AggregateProcessStats.prototype.getDownloadBytes = function() {
  return /** @type{?proto.pps_v2.Aggregate} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Aggregate, 4));
};


/**
 * @param {?proto.pps_v2.Aggregate|undefined} value
 * @return {!proto.pps_v2.AggregateProcessStats} returns this
*/
proto.pps_v2.AggregateProcessStats.prototype.setDownloadBytes = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.AggregateProcessStats} returns this
 */
proto.pps_v2.AggregateProcessStats.prototype.clearDownloadBytes = function() {
  return this.setDownloadBytes(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.AggregateProcessStats.prototype.hasDownloadBytes = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional Aggregate upload_bytes = 5;
 * @return {?proto.pps_v2.Aggregate}
 */
proto.pps_v2.AggregateProcessStats.prototype.getUploadBytes = function() {
  return /** @type{?proto.pps_v2.Aggregate} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Aggregate, 5));
};


/**
 * @param {?proto.pps_v2.Aggregate|undefined} value
 * @return {!proto.pps_v2.AggregateProcessStats} returns this
*/
proto.pps_v2.AggregateProcessStats.prototype.setUploadBytes = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.AggregateProcessStats} returns this
 */
proto.pps_v2.AggregateProcessStats.prototype.clearUploadBytes = function() {
  return this.setUploadBytes(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.AggregateProcessStats.prototype.hasUploadBytes = function() {
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
proto.pps_v2.WorkerStatus.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.WorkerStatus.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.WorkerStatus} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.WorkerStatus.toObject = function(includeInstance, msg) {
  var f, obj = {
    workerId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    jobId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    datumStatus: (f = msg.getDatumStatus()) && proto.pps_v2.DatumStatus.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.WorkerStatus}
 */
proto.pps_v2.WorkerStatus.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.WorkerStatus;
  return proto.pps_v2.WorkerStatus.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.WorkerStatus} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.WorkerStatus}
 */
proto.pps_v2.WorkerStatus.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setWorkerId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setJobId(value);
      break;
    case 3:
      var value = new proto.pps_v2.DatumStatus;
      reader.readMessage(value,proto.pps_v2.DatumStatus.deserializeBinaryFromReader);
      msg.setDatumStatus(value);
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
proto.pps_v2.WorkerStatus.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.WorkerStatus.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.WorkerStatus} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.WorkerStatus.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getWorkerId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getJobId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getDatumStatus();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pps_v2.DatumStatus.serializeBinaryToWriter
    );
  }
};


/**
 * optional string worker_id = 1;
 * @return {string}
 */
proto.pps_v2.WorkerStatus.prototype.getWorkerId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.WorkerStatus} returns this
 */
proto.pps_v2.WorkerStatus.prototype.setWorkerId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string job_id = 2;
 * @return {string}
 */
proto.pps_v2.WorkerStatus.prototype.getJobId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.WorkerStatus} returns this
 */
proto.pps_v2.WorkerStatus.prototype.setJobId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional DatumStatus datum_status = 3;
 * @return {?proto.pps_v2.DatumStatus}
 */
proto.pps_v2.WorkerStatus.prototype.getDatumStatus = function() {
  return /** @type{?proto.pps_v2.DatumStatus} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.DatumStatus, 3));
};


/**
 * @param {?proto.pps_v2.DatumStatus|undefined} value
 * @return {!proto.pps_v2.WorkerStatus} returns this
*/
proto.pps_v2.WorkerStatus.prototype.setDatumStatus = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.WorkerStatus} returns this
 */
proto.pps_v2.WorkerStatus.prototype.clearDatumStatus = function() {
  return this.setDatumStatus(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.WorkerStatus.prototype.hasDatumStatus = function() {
  return jspb.Message.getField(this, 3) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.DatumStatus.repeatedFields_ = [2];



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
proto.pps_v2.DatumStatus.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.DatumStatus.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.DatumStatus} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DatumStatus.toObject = function(includeInstance, msg) {
  var f, obj = {
    started: (f = msg.getStarted()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    dataList: jspb.Message.toObjectList(msg.getDataList(),
    proto.pps_v2.InputFile.toObject, includeInstance)
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
 * @return {!proto.pps_v2.DatumStatus}
 */
proto.pps_v2.DatumStatus.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.DatumStatus;
  return proto.pps_v2.DatumStatus.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.DatumStatus} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.DatumStatus}
 */
proto.pps_v2.DatumStatus.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setStarted(value);
      break;
    case 2:
      var value = new proto.pps_v2.InputFile;
      reader.readMessage(value,proto.pps_v2.InputFile.deserializeBinaryFromReader);
      msg.addData(value);
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
proto.pps_v2.DatumStatus.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.DatumStatus.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.DatumStatus} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DatumStatus.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getStarted();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getDataList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.pps_v2.InputFile.serializeBinaryToWriter
    );
  }
};


/**
 * optional google.protobuf.Timestamp started = 1;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps_v2.DatumStatus.prototype.getStarted = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 1));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps_v2.DatumStatus} returns this
*/
proto.pps_v2.DatumStatus.prototype.setStarted = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.DatumStatus} returns this
 */
proto.pps_v2.DatumStatus.prototype.clearStarted = function() {
  return this.setStarted(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.DatumStatus.prototype.hasStarted = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * repeated InputFile data = 2;
 * @return {!Array<!proto.pps_v2.InputFile>}
 */
proto.pps_v2.DatumStatus.prototype.getDataList = function() {
  return /** @type{!Array<!proto.pps_v2.InputFile>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps_v2.InputFile, 2));
};


/**
 * @param {!Array<!proto.pps_v2.InputFile>} value
 * @return {!proto.pps_v2.DatumStatus} returns this
*/
proto.pps_v2.DatumStatus.prototype.setDataList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pps_v2.InputFile=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.InputFile}
 */
proto.pps_v2.DatumStatus.prototype.addData = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pps_v2.InputFile, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.DatumStatus} returns this
 */
proto.pps_v2.DatumStatus.prototype.clearDataList = function() {
  return this.setDataList([]);
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
proto.pps_v2.ResourceSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.ResourceSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.ResourceSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ResourceSpec.toObject = function(includeInstance, msg) {
  var f, obj = {
    cpu: jspb.Message.getFloatingPointFieldWithDefault(msg, 1, 0.0),
    memory: jspb.Message.getFieldWithDefault(msg, 2, ""),
    gpu: (f = msg.getGpu()) && proto.pps_v2.GPUSpec.toObject(includeInstance, f),
    disk: jspb.Message.getFieldWithDefault(msg, 4, "")
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
 * @return {!proto.pps_v2.ResourceSpec}
 */
proto.pps_v2.ResourceSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.ResourceSpec;
  return proto.pps_v2.ResourceSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.ResourceSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.ResourceSpec}
 */
proto.pps_v2.ResourceSpec.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readFloat());
      msg.setCpu(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setMemory(value);
      break;
    case 3:
      var value = new proto.pps_v2.GPUSpec;
      reader.readMessage(value,proto.pps_v2.GPUSpec.deserializeBinaryFromReader);
      msg.setGpu(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setDisk(value);
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
proto.pps_v2.ResourceSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.ResourceSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.ResourceSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ResourceSpec.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCpu();
  if (f !== 0.0) {
    writer.writeFloat(
      1,
      f
    );
  }
  f = message.getMemory();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getGpu();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pps_v2.GPUSpec.serializeBinaryToWriter
    );
  }
  f = message.getDisk();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * optional float cpu = 1;
 * @return {number}
 */
proto.pps_v2.ResourceSpec.prototype.getCpu = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 1, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.ResourceSpec} returns this
 */
proto.pps_v2.ResourceSpec.prototype.setCpu = function(value) {
  return jspb.Message.setProto3FloatField(this, 1, value);
};


/**
 * optional string memory = 2;
 * @return {string}
 */
proto.pps_v2.ResourceSpec.prototype.getMemory = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.ResourceSpec} returns this
 */
proto.pps_v2.ResourceSpec.prototype.setMemory = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional GPUSpec gpu = 3;
 * @return {?proto.pps_v2.GPUSpec}
 */
proto.pps_v2.ResourceSpec.prototype.getGpu = function() {
  return /** @type{?proto.pps_v2.GPUSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.GPUSpec, 3));
};


/**
 * @param {?proto.pps_v2.GPUSpec|undefined} value
 * @return {!proto.pps_v2.ResourceSpec} returns this
*/
proto.pps_v2.ResourceSpec.prototype.setGpu = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.ResourceSpec} returns this
 */
proto.pps_v2.ResourceSpec.prototype.clearGpu = function() {
  return this.setGpu(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.ResourceSpec.prototype.hasGpu = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional string disk = 4;
 * @return {string}
 */
proto.pps_v2.ResourceSpec.prototype.getDisk = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.ResourceSpec} returns this
 */
proto.pps_v2.ResourceSpec.prototype.setDisk = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
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
proto.pps_v2.GPUSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.GPUSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.GPUSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.GPUSpec.toObject = function(includeInstance, msg) {
  var f, obj = {
    type: jspb.Message.getFieldWithDefault(msg, 1, ""),
    number: jspb.Message.getFieldWithDefault(msg, 2, 0)
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
 * @return {!proto.pps_v2.GPUSpec}
 */
proto.pps_v2.GPUSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.GPUSpec;
  return proto.pps_v2.GPUSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.GPUSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.GPUSpec}
 */
proto.pps_v2.GPUSpec.deserializeBinaryFromReader = function(msg, reader) {
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
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setNumber(value);
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
proto.pps_v2.GPUSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.GPUSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.GPUSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.GPUSpec.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getType();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getNumber();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
};


/**
 * optional string type = 1;
 * @return {string}
 */
proto.pps_v2.GPUSpec.prototype.getType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.GPUSpec} returns this
 */
proto.pps_v2.GPUSpec.prototype.setType = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional int64 number = 2;
 * @return {number}
 */
proto.pps_v2.GPUSpec.prototype.getNumber = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.GPUSpec} returns this
 */
proto.pps_v2.GPUSpec.prototype.setNumber = function(value) {
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
proto.pps_v2.JobInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.JobInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.JobInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.JobInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps_v2.Job.toObject(includeInstance, f),
    pipelineVersion: jspb.Message.getFieldWithDefault(msg, 2, 0),
    outputCommit: (f = msg.getOutputCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    restart: jspb.Message.getFieldWithDefault(msg, 4, 0),
    dataProcessed: jspb.Message.getFieldWithDefault(msg, 5, 0),
    dataSkipped: jspb.Message.getFieldWithDefault(msg, 6, 0),
    dataTotal: jspb.Message.getFieldWithDefault(msg, 7, 0),
    dataFailed: jspb.Message.getFieldWithDefault(msg, 8, 0),
    dataRecovered: jspb.Message.getFieldWithDefault(msg, 9, 0),
    stats: (f = msg.getStats()) && proto.pps_v2.ProcessStats.toObject(includeInstance, f),
    state: jspb.Message.getFieldWithDefault(msg, 11, 0),
    reason: jspb.Message.getFieldWithDefault(msg, 12, ""),
    started: (f = msg.getStarted()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    finished: (f = msg.getFinished()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    details: (f = msg.getDetails()) && proto.pps_v2.JobInfo.Details.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.JobInfo}
 */
proto.pps_v2.JobInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.JobInfo;
  return proto.pps_v2.JobInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.JobInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.JobInfo}
 */
proto.pps_v2.JobInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Job;
      reader.readMessage(value,proto.pps_v2.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setPipelineVersion(value);
      break;
    case 3:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setOutputCommit(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setRestart(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataProcessed(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataSkipped(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataTotal(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataFailed(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataRecovered(value);
      break;
    case 10:
      var value = new proto.pps_v2.ProcessStats;
      reader.readMessage(value,proto.pps_v2.ProcessStats.deserializeBinaryFromReader);
      msg.setStats(value);
      break;
    case 11:
      var value = /** @type {!proto.pps_v2.JobState} */ (reader.readEnum());
      msg.setState(value);
      break;
    case 12:
      var value = /** @type {string} */ (reader.readString());
      msg.setReason(value);
      break;
    case 13:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setStarted(value);
      break;
    case 14:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setFinished(value);
      break;
    case 15:
      var value = new proto.pps_v2.JobInfo.Details;
      reader.readMessage(value,proto.pps_v2.JobInfo.Details.deserializeBinaryFromReader);
      msg.setDetails(value);
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
proto.pps_v2.JobInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.JobInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.JobInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.JobInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Job.serializeBinaryToWriter
    );
  }
  f = message.getPipelineVersion();
  if (f !== 0) {
    writer.writeUint64(
      2,
      f
    );
  }
  f = message.getOutputCommit();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getRestart();
  if (f !== 0) {
    writer.writeUint64(
      4,
      f
    );
  }
  f = message.getDataProcessed();
  if (f !== 0) {
    writer.writeInt64(
      5,
      f
    );
  }
  f = message.getDataSkipped();
  if (f !== 0) {
    writer.writeInt64(
      6,
      f
    );
  }
  f = message.getDataTotal();
  if (f !== 0) {
    writer.writeInt64(
      7,
      f
    );
  }
  f = message.getDataFailed();
  if (f !== 0) {
    writer.writeInt64(
      8,
      f
    );
  }
  f = message.getDataRecovered();
  if (f !== 0) {
    writer.writeInt64(
      9,
      f
    );
  }
  f = message.getStats();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      proto.pps_v2.ProcessStats.serializeBinaryToWriter
    );
  }
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      11,
      f
    );
  }
  f = message.getReason();
  if (f.length > 0) {
    writer.writeString(
      12,
      f
    );
  }
  f = message.getStarted();
  if (f != null) {
    writer.writeMessage(
      13,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getFinished();
  if (f != null) {
    writer.writeMessage(
      14,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getDetails();
  if (f != null) {
    writer.writeMessage(
      15,
      f,
      proto.pps_v2.JobInfo.Details.serializeBinaryToWriter
    );
  }
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.JobInfo.Details.repeatedFields_ = [6];



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
proto.pps_v2.JobInfo.Details.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.JobInfo.Details.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.JobInfo.Details} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.JobInfo.Details.toObject = function(includeInstance, msg) {
  var f, obj = {
    transform: (f = msg.getTransform()) && proto.pps_v2.Transform.toObject(includeInstance, f),
    parallelismSpec: (f = msg.getParallelismSpec()) && proto.pps_v2.ParallelismSpec.toObject(includeInstance, f),
    egress: (f = msg.getEgress()) && proto.pps_v2.Egress.toObject(includeInstance, f),
    service: (f = msg.getService()) && proto.pps_v2.Service.toObject(includeInstance, f),
    spout: (f = msg.getSpout()) && proto.pps_v2.Spout.toObject(includeInstance, f),
    workerStatusList: jspb.Message.toObjectList(msg.getWorkerStatusList(),
    proto.pps_v2.WorkerStatus.toObject, includeInstance),
    resourceRequests: (f = msg.getResourceRequests()) && proto.pps_v2.ResourceSpec.toObject(includeInstance, f),
    resourceLimits: (f = msg.getResourceLimits()) && proto.pps_v2.ResourceSpec.toObject(includeInstance, f),
    sidecarResourceLimits: (f = msg.getSidecarResourceLimits()) && proto.pps_v2.ResourceSpec.toObject(includeInstance, f),
    input: (f = msg.getInput()) && proto.pps_v2.Input.toObject(includeInstance, f),
    salt: jspb.Message.getFieldWithDefault(msg, 11, ""),
    datumSetSpec: (f = msg.getDatumSetSpec()) && proto.pps_v2.DatumSetSpec.toObject(includeInstance, f),
    datumTimeout: (f = msg.getDatumTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    jobTimeout: (f = msg.getJobTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    datumTries: jspb.Message.getFieldWithDefault(msg, 15, 0),
    schedulingSpec: (f = msg.getSchedulingSpec()) && proto.pps_v2.SchedulingSpec.toObject(includeInstance, f),
    podSpec: jspb.Message.getFieldWithDefault(msg, 17, ""),
    podPatch: jspb.Message.getFieldWithDefault(msg, 18, "")
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
 * @return {!proto.pps_v2.JobInfo.Details}
 */
proto.pps_v2.JobInfo.Details.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.JobInfo.Details;
  return proto.pps_v2.JobInfo.Details.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.JobInfo.Details} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.JobInfo.Details}
 */
proto.pps_v2.JobInfo.Details.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Transform;
      reader.readMessage(value,proto.pps_v2.Transform.deserializeBinaryFromReader);
      msg.setTransform(value);
      break;
    case 2:
      var value = new proto.pps_v2.ParallelismSpec;
      reader.readMessage(value,proto.pps_v2.ParallelismSpec.deserializeBinaryFromReader);
      msg.setParallelismSpec(value);
      break;
    case 3:
      var value = new proto.pps_v2.Egress;
      reader.readMessage(value,proto.pps_v2.Egress.deserializeBinaryFromReader);
      msg.setEgress(value);
      break;
    case 4:
      var value = new proto.pps_v2.Service;
      reader.readMessage(value,proto.pps_v2.Service.deserializeBinaryFromReader);
      msg.setService(value);
      break;
    case 5:
      var value = new proto.pps_v2.Spout;
      reader.readMessage(value,proto.pps_v2.Spout.deserializeBinaryFromReader);
      msg.setSpout(value);
      break;
    case 6:
      var value = new proto.pps_v2.WorkerStatus;
      reader.readMessage(value,proto.pps_v2.WorkerStatus.deserializeBinaryFromReader);
      msg.addWorkerStatus(value);
      break;
    case 7:
      var value = new proto.pps_v2.ResourceSpec;
      reader.readMessage(value,proto.pps_v2.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceRequests(value);
      break;
    case 8:
      var value = new proto.pps_v2.ResourceSpec;
      reader.readMessage(value,proto.pps_v2.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceLimits(value);
      break;
    case 9:
      var value = new proto.pps_v2.ResourceSpec;
      reader.readMessage(value,proto.pps_v2.ResourceSpec.deserializeBinaryFromReader);
      msg.setSidecarResourceLimits(value);
      break;
    case 10:
      var value = new proto.pps_v2.Input;
      reader.readMessage(value,proto.pps_v2.Input.deserializeBinaryFromReader);
      msg.setInput(value);
      break;
    case 11:
      var value = /** @type {string} */ (reader.readString());
      msg.setSalt(value);
      break;
    case 12:
      var value = new proto.pps_v2.DatumSetSpec;
      reader.readMessage(value,proto.pps_v2.DatumSetSpec.deserializeBinaryFromReader);
      msg.setDatumSetSpec(value);
      break;
    case 13:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setDatumTimeout(value);
      break;
    case 14:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setJobTimeout(value);
      break;
    case 15:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDatumTries(value);
      break;
    case 16:
      var value = new proto.pps_v2.SchedulingSpec;
      reader.readMessage(value,proto.pps_v2.SchedulingSpec.deserializeBinaryFromReader);
      msg.setSchedulingSpec(value);
      break;
    case 17:
      var value = /** @type {string} */ (reader.readString());
      msg.setPodSpec(value);
      break;
    case 18:
      var value = /** @type {string} */ (reader.readString());
      msg.setPodPatch(value);
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
proto.pps_v2.JobInfo.Details.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.JobInfo.Details.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.JobInfo.Details} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.JobInfo.Details.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTransform();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Transform.serializeBinaryToWriter
    );
  }
  f = message.getParallelismSpec();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps_v2.ParallelismSpec.serializeBinaryToWriter
    );
  }
  f = message.getEgress();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pps_v2.Egress.serializeBinaryToWriter
    );
  }
  f = message.getService();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pps_v2.Service.serializeBinaryToWriter
    );
  }
  f = message.getSpout();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.pps_v2.Spout.serializeBinaryToWriter
    );
  }
  f = message.getWorkerStatusList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      6,
      f,
      proto.pps_v2.WorkerStatus.serializeBinaryToWriter
    );
  }
  f = message.getResourceRequests();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      proto.pps_v2.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getResourceLimits();
  if (f != null) {
    writer.writeMessage(
      8,
      f,
      proto.pps_v2.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getSidecarResourceLimits();
  if (f != null) {
    writer.writeMessage(
      9,
      f,
      proto.pps_v2.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getInput();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      proto.pps_v2.Input.serializeBinaryToWriter
    );
  }
  f = message.getSalt();
  if (f.length > 0) {
    writer.writeString(
      11,
      f
    );
  }
  f = message.getDatumSetSpec();
  if (f != null) {
    writer.writeMessage(
      12,
      f,
      proto.pps_v2.DatumSetSpec.serializeBinaryToWriter
    );
  }
  f = message.getDatumTimeout();
  if (f != null) {
    writer.writeMessage(
      13,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getJobTimeout();
  if (f != null) {
    writer.writeMessage(
      14,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getDatumTries();
  if (f !== 0) {
    writer.writeInt64(
      15,
      f
    );
  }
  f = message.getSchedulingSpec();
  if (f != null) {
    writer.writeMessage(
      16,
      f,
      proto.pps_v2.SchedulingSpec.serializeBinaryToWriter
    );
  }
  f = message.getPodSpec();
  if (f.length > 0) {
    writer.writeString(
      17,
      f
    );
  }
  f = message.getPodPatch();
  if (f.length > 0) {
    writer.writeString(
      18,
      f
    );
  }
};


/**
 * optional Transform transform = 1;
 * @return {?proto.pps_v2.Transform}
 */
proto.pps_v2.JobInfo.Details.prototype.getTransform = function() {
  return /** @type{?proto.pps_v2.Transform} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Transform, 1));
};


/**
 * @param {?proto.pps_v2.Transform|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setTransform = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearTransform = function() {
  return this.setTransform(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasTransform = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional ParallelismSpec parallelism_spec = 2;
 * @return {?proto.pps_v2.ParallelismSpec}
 */
proto.pps_v2.JobInfo.Details.prototype.getParallelismSpec = function() {
  return /** @type{?proto.pps_v2.ParallelismSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ParallelismSpec, 2));
};


/**
 * @param {?proto.pps_v2.ParallelismSpec|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setParallelismSpec = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearParallelismSpec = function() {
  return this.setParallelismSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasParallelismSpec = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional Egress egress = 3;
 * @return {?proto.pps_v2.Egress}
 */
proto.pps_v2.JobInfo.Details.prototype.getEgress = function() {
  return /** @type{?proto.pps_v2.Egress} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Egress, 3));
};


/**
 * @param {?proto.pps_v2.Egress|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setEgress = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearEgress = function() {
  return this.setEgress(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasEgress = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional Service service = 4;
 * @return {?proto.pps_v2.Service}
 */
proto.pps_v2.JobInfo.Details.prototype.getService = function() {
  return /** @type{?proto.pps_v2.Service} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Service, 4));
};


/**
 * @param {?proto.pps_v2.Service|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setService = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearService = function() {
  return this.setService(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasService = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional Spout spout = 5;
 * @return {?proto.pps_v2.Spout}
 */
proto.pps_v2.JobInfo.Details.prototype.getSpout = function() {
  return /** @type{?proto.pps_v2.Spout} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Spout, 5));
};


/**
 * @param {?proto.pps_v2.Spout|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setSpout = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearSpout = function() {
  return this.setSpout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasSpout = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * repeated WorkerStatus worker_status = 6;
 * @return {!Array<!proto.pps_v2.WorkerStatus>}
 */
proto.pps_v2.JobInfo.Details.prototype.getWorkerStatusList = function() {
  return /** @type{!Array<!proto.pps_v2.WorkerStatus>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps_v2.WorkerStatus, 6));
};


/**
 * @param {!Array<!proto.pps_v2.WorkerStatus>} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setWorkerStatusList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 6, value);
};


/**
 * @param {!proto.pps_v2.WorkerStatus=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.WorkerStatus}
 */
proto.pps_v2.JobInfo.Details.prototype.addWorkerStatus = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 6, opt_value, proto.pps_v2.WorkerStatus, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearWorkerStatusList = function() {
  return this.setWorkerStatusList([]);
};


/**
 * optional ResourceSpec resource_requests = 7;
 * @return {?proto.pps_v2.ResourceSpec}
 */
proto.pps_v2.JobInfo.Details.prototype.getResourceRequests = function() {
  return /** @type{?proto.pps_v2.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ResourceSpec, 7));
};


/**
 * @param {?proto.pps_v2.ResourceSpec|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setResourceRequests = function(value) {
  return jspb.Message.setWrapperField(this, 7, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearResourceRequests = function() {
  return this.setResourceRequests(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasResourceRequests = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * optional ResourceSpec resource_limits = 8;
 * @return {?proto.pps_v2.ResourceSpec}
 */
proto.pps_v2.JobInfo.Details.prototype.getResourceLimits = function() {
  return /** @type{?proto.pps_v2.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ResourceSpec, 8));
};


/**
 * @param {?proto.pps_v2.ResourceSpec|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 8, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearResourceLimits = function() {
  return this.setResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasResourceLimits = function() {
  return jspb.Message.getField(this, 8) != null;
};


/**
 * optional ResourceSpec sidecar_resource_limits = 9;
 * @return {?proto.pps_v2.ResourceSpec}
 */
proto.pps_v2.JobInfo.Details.prototype.getSidecarResourceLimits = function() {
  return /** @type{?proto.pps_v2.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ResourceSpec, 9));
};


/**
 * @param {?proto.pps_v2.ResourceSpec|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setSidecarResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 9, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearSidecarResourceLimits = function() {
  return this.setSidecarResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasSidecarResourceLimits = function() {
  return jspb.Message.getField(this, 9) != null;
};


/**
 * optional Input input = 10;
 * @return {?proto.pps_v2.Input}
 */
proto.pps_v2.JobInfo.Details.prototype.getInput = function() {
  return /** @type{?proto.pps_v2.Input} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Input, 10));
};


/**
 * @param {?proto.pps_v2.Input|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setInput = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearInput = function() {
  return this.setInput(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasInput = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * optional string salt = 11;
 * @return {string}
 */
proto.pps_v2.JobInfo.Details.prototype.getSalt = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 11, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.setSalt = function(value) {
  return jspb.Message.setProto3StringField(this, 11, value);
};


/**
 * optional DatumSetSpec datum_set_spec = 12;
 * @return {?proto.pps_v2.DatumSetSpec}
 */
proto.pps_v2.JobInfo.Details.prototype.getDatumSetSpec = function() {
  return /** @type{?proto.pps_v2.DatumSetSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.DatumSetSpec, 12));
};


/**
 * @param {?proto.pps_v2.DatumSetSpec|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setDatumSetSpec = function(value) {
  return jspb.Message.setWrapperField(this, 12, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearDatumSetSpec = function() {
  return this.setDatumSetSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasDatumSetSpec = function() {
  return jspb.Message.getField(this, 12) != null;
};


/**
 * optional google.protobuf.Duration datum_timeout = 13;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps_v2.JobInfo.Details.prototype.getDatumTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 13));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setDatumTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 13, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearDatumTimeout = function() {
  return this.setDatumTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasDatumTimeout = function() {
  return jspb.Message.getField(this, 13) != null;
};


/**
 * optional google.protobuf.Duration job_timeout = 14;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps_v2.JobInfo.Details.prototype.getJobTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 14));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setJobTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 14, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearJobTimeout = function() {
  return this.setJobTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasJobTimeout = function() {
  return jspb.Message.getField(this, 14) != null;
};


/**
 * optional int64 datum_tries = 15;
 * @return {number}
 */
proto.pps_v2.JobInfo.Details.prototype.getDatumTries = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 15, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.setDatumTries = function(value) {
  return jspb.Message.setProto3IntField(this, 15, value);
};


/**
 * optional SchedulingSpec scheduling_spec = 16;
 * @return {?proto.pps_v2.SchedulingSpec}
 */
proto.pps_v2.JobInfo.Details.prototype.getSchedulingSpec = function() {
  return /** @type{?proto.pps_v2.SchedulingSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.SchedulingSpec, 16));
};


/**
 * @param {?proto.pps_v2.SchedulingSpec|undefined} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
*/
proto.pps_v2.JobInfo.Details.prototype.setSchedulingSpec = function(value) {
  return jspb.Message.setWrapperField(this, 16, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.clearSchedulingSpec = function() {
  return this.setSchedulingSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.Details.prototype.hasSchedulingSpec = function() {
  return jspb.Message.getField(this, 16) != null;
};


/**
 * optional string pod_spec = 17;
 * @return {string}
 */
proto.pps_v2.JobInfo.Details.prototype.getPodSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 17, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.setPodSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 17, value);
};


/**
 * optional string pod_patch = 18;
 * @return {string}
 */
proto.pps_v2.JobInfo.Details.prototype.getPodPatch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 18, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.JobInfo.Details} returns this
 */
proto.pps_v2.JobInfo.Details.prototype.setPodPatch = function(value) {
  return jspb.Message.setProto3StringField(this, 18, value);
};


/**
 * optional Job job = 1;
 * @return {?proto.pps_v2.Job}
 */
proto.pps_v2.JobInfo.prototype.getJob = function() {
  return /** @type{?proto.pps_v2.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Job, 1));
};


/**
 * @param {?proto.pps_v2.Job|undefined} value
 * @return {!proto.pps_v2.JobInfo} returns this
*/
proto.pps_v2.JobInfo.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional uint64 pipeline_version = 2;
 * @return {number}
 */
proto.pps_v2.JobInfo.prototype.getPipelineVersion = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.setPipelineVersion = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional pfs_v2.Commit output_commit = 3;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pps_v2.JobInfo.prototype.getOutputCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 3));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pps_v2.JobInfo} returns this
*/
proto.pps_v2.JobInfo.prototype.setOutputCommit = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.clearOutputCommit = function() {
  return this.setOutputCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.prototype.hasOutputCommit = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional uint64 restart = 4;
 * @return {number}
 */
proto.pps_v2.JobInfo.prototype.getRestart = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.setRestart = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional int64 data_processed = 5;
 * @return {number}
 */
proto.pps_v2.JobInfo.prototype.getDataProcessed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.setDataProcessed = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};


/**
 * optional int64 data_skipped = 6;
 * @return {number}
 */
proto.pps_v2.JobInfo.prototype.getDataSkipped = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.setDataSkipped = function(value) {
  return jspb.Message.setProto3IntField(this, 6, value);
};


/**
 * optional int64 data_total = 7;
 * @return {number}
 */
proto.pps_v2.JobInfo.prototype.getDataTotal = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.setDataTotal = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
};


/**
 * optional int64 data_failed = 8;
 * @return {number}
 */
proto.pps_v2.JobInfo.prototype.getDataFailed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.setDataFailed = function(value) {
  return jspb.Message.setProto3IntField(this, 8, value);
};


/**
 * optional int64 data_recovered = 9;
 * @return {number}
 */
proto.pps_v2.JobInfo.prototype.getDataRecovered = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.setDataRecovered = function(value) {
  return jspb.Message.setProto3IntField(this, 9, value);
};


/**
 * optional ProcessStats stats = 10;
 * @return {?proto.pps_v2.ProcessStats}
 */
proto.pps_v2.JobInfo.prototype.getStats = function() {
  return /** @type{?proto.pps_v2.ProcessStats} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ProcessStats, 10));
};


/**
 * @param {?proto.pps_v2.ProcessStats|undefined} value
 * @return {!proto.pps_v2.JobInfo} returns this
*/
proto.pps_v2.JobInfo.prototype.setStats = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.clearStats = function() {
  return this.setStats(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.prototype.hasStats = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * optional JobState state = 11;
 * @return {!proto.pps_v2.JobState}
 */
proto.pps_v2.JobInfo.prototype.getState = function() {
  return /** @type {!proto.pps_v2.JobState} */ (jspb.Message.getFieldWithDefault(this, 11, 0));
};


/**
 * @param {!proto.pps_v2.JobState} value
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 11, value);
};


/**
 * optional string reason = 12;
 * @return {string}
 */
proto.pps_v2.JobInfo.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 12, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.setReason = function(value) {
  return jspb.Message.setProto3StringField(this, 12, value);
};


/**
 * optional google.protobuf.Timestamp started = 13;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps_v2.JobInfo.prototype.getStarted = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 13));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps_v2.JobInfo} returns this
*/
proto.pps_v2.JobInfo.prototype.setStarted = function(value) {
  return jspb.Message.setWrapperField(this, 13, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.clearStarted = function() {
  return this.setStarted(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.prototype.hasStarted = function() {
  return jspb.Message.getField(this, 13) != null;
};


/**
 * optional google.protobuf.Timestamp finished = 14;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps_v2.JobInfo.prototype.getFinished = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 14));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps_v2.JobInfo} returns this
*/
proto.pps_v2.JobInfo.prototype.setFinished = function(value) {
  return jspb.Message.setWrapperField(this, 14, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.clearFinished = function() {
  return this.setFinished(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.prototype.hasFinished = function() {
  return jspb.Message.getField(this, 14) != null;
};


/**
 * optional Details details = 15;
 * @return {?proto.pps_v2.JobInfo.Details}
 */
proto.pps_v2.JobInfo.prototype.getDetails = function() {
  return /** @type{?proto.pps_v2.JobInfo.Details} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.JobInfo.Details, 15));
};


/**
 * @param {?proto.pps_v2.JobInfo.Details|undefined} value
 * @return {!proto.pps_v2.JobInfo} returns this
*/
proto.pps_v2.JobInfo.prototype.setDetails = function(value) {
  return jspb.Message.setWrapperField(this, 15, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.JobInfo} returns this
 */
proto.pps_v2.JobInfo.prototype.clearDetails = function() {
  return this.setDetails(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.JobInfo.prototype.hasDetails = function() {
  return jspb.Message.getField(this, 15) != null;
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
proto.pps_v2.Worker.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Worker.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Worker} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Worker.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    state: jspb.Message.getFieldWithDefault(msg, 2, 0)
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
 * @return {!proto.pps_v2.Worker}
 */
proto.pps_v2.Worker.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Worker;
  return proto.pps_v2.Worker.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Worker} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Worker}
 */
proto.pps_v2.Worker.deserializeBinaryFromReader = function(msg, reader) {
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
      var value = /** @type {!proto.pps_v2.WorkerState} */ (reader.readEnum());
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
proto.pps_v2.Worker.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Worker.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Worker} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Worker.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pps_v2.Worker.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Worker} returns this
 */
proto.pps_v2.Worker.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional WorkerState state = 2;
 * @return {!proto.pps_v2.WorkerState}
 */
proto.pps_v2.Worker.prototype.getState = function() {
  return /** @type {!proto.pps_v2.WorkerState} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.pps_v2.WorkerState} value
 * @return {!proto.pps_v2.Worker} returns this
 */
proto.pps_v2.Worker.prototype.setState = function(value) {
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
proto.pps_v2.Pipeline.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Pipeline.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Pipeline} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Pipeline.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps_v2.Pipeline}
 */
proto.pps_v2.Pipeline.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Pipeline;
  return proto.pps_v2.Pipeline.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Pipeline} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Pipeline}
 */
proto.pps_v2.Pipeline.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps_v2.Pipeline.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Pipeline.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Pipeline} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Pipeline.serializeBinaryToWriter = function(message, writer) {
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
proto.pps_v2.Pipeline.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Pipeline} returns this
 */
proto.pps_v2.Pipeline.prototype.setName = function(value) {
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
proto.pps_v2.PipelineInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.PipelineInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.PipelineInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.PipelineInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f),
    version: jspb.Message.getFieldWithDefault(msg, 2, 0),
    specCommit: (f = msg.getSpecCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    stopped: jspb.Message.getBooleanFieldWithDefault(msg, 4, false),
    state: jspb.Message.getFieldWithDefault(msg, 5, 0),
    reason: jspb.Message.getFieldWithDefault(msg, 6, ""),
    jobCountsMap: (f = msg.getJobCountsMap()) ? f.toObject(includeInstance, undefined) : [],
    lastJobState: jspb.Message.getFieldWithDefault(msg, 8, 0),
    parallelism: jspb.Message.getFieldWithDefault(msg, 9, 0),
    type: jspb.Message.getFieldWithDefault(msg, 10, 0),
    authToken: jspb.Message.getFieldWithDefault(msg, 11, ""),
    details: (f = msg.getDetails()) && proto.pps_v2.PipelineInfo.Details.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.PipelineInfo}
 */
proto.pps_v2.PipelineInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.PipelineInfo;
  return proto.pps_v2.PipelineInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.PipelineInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.PipelineInfo}
 */
proto.pps_v2.PipelineInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setVersion(value);
      break;
    case 3:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setSpecCommit(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setStopped(value);
      break;
    case 5:
      var value = /** @type {!proto.pps_v2.PipelineState} */ (reader.readEnum());
      msg.setState(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setReason(value);
      break;
    case 7:
      var value = msg.getJobCountsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readInt32, jspb.BinaryReader.prototype.readInt32, null, 0, 0);
         });
      break;
    case 8:
      var value = /** @type {!proto.pps_v2.JobState} */ (reader.readEnum());
      msg.setLastJobState(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setParallelism(value);
      break;
    case 10:
      var value = /** @type {!proto.pps_v2.PipelineInfo.PipelineType} */ (reader.readEnum());
      msg.setType(value);
      break;
    case 11:
      var value = /** @type {string} */ (reader.readString());
      msg.setAuthToken(value);
      break;
    case 12:
      var value = new proto.pps_v2.PipelineInfo.Details;
      reader.readMessage(value,proto.pps_v2.PipelineInfo.Details.deserializeBinaryFromReader);
      msg.setDetails(value);
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
proto.pps_v2.PipelineInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.PipelineInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.PipelineInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.PipelineInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getVersion();
  if (f !== 0) {
    writer.writeUint64(
      2,
      f
    );
  }
  f = message.getSpecCommit();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getStopped();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      5,
      f
    );
  }
  f = message.getReason();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getJobCountsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(7, writer, jspb.BinaryWriter.prototype.writeInt32, jspb.BinaryWriter.prototype.writeInt32);
  }
  f = message.getLastJobState();
  if (f !== 0.0) {
    writer.writeEnum(
      8,
      f
    );
  }
  f = message.getParallelism();
  if (f !== 0) {
    writer.writeUint64(
      9,
      f
    );
  }
  f = message.getType();
  if (f !== 0.0) {
    writer.writeEnum(
      10,
      f
    );
  }
  f = message.getAuthToken();
  if (f.length > 0) {
    writer.writeString(
      11,
      f
    );
  }
  f = message.getDetails();
  if (f != null) {
    writer.writeMessage(
      12,
      f,
      proto.pps_v2.PipelineInfo.Details.serializeBinaryToWriter
    );
  }
};


/**
 * @enum {number}
 */
proto.pps_v2.PipelineInfo.PipelineType = {
  PIPELINT_TYPE_UNKNOWN: 0,
  PIPELINE_TYPE_TRANSFORM: 1,
  PIPELINE_TYPE_SPOUT: 2,
  PIPELINE_TYPE_SERVICE: 3
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
proto.pps_v2.PipelineInfo.Details.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.PipelineInfo.Details.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.PipelineInfo.Details} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.PipelineInfo.Details.toObject = function(includeInstance, msg) {
  var f, obj = {
    transform: (f = msg.getTransform()) && proto.pps_v2.Transform.toObject(includeInstance, f),
    tfJob: (f = msg.getTfJob()) && proto.pps_v2.TFJob.toObject(includeInstance, f),
    parallelismSpec: (f = msg.getParallelismSpec()) && proto.pps_v2.ParallelismSpec.toObject(includeInstance, f),
    egress: (f = msg.getEgress()) && proto.pps_v2.Egress.toObject(includeInstance, f),
    createdAt: (f = msg.getCreatedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    recentError: jspb.Message.getFieldWithDefault(msg, 6, ""),
    workersRequested: jspb.Message.getFieldWithDefault(msg, 7, 0),
    workersAvailable: jspb.Message.getFieldWithDefault(msg, 8, 0),
    outputBranch: jspb.Message.getFieldWithDefault(msg, 9, ""),
    resourceRequests: (f = msg.getResourceRequests()) && proto.pps_v2.ResourceSpec.toObject(includeInstance, f),
    resourceLimits: (f = msg.getResourceLimits()) && proto.pps_v2.ResourceSpec.toObject(includeInstance, f),
    sidecarResourceLimits: (f = msg.getSidecarResourceLimits()) && proto.pps_v2.ResourceSpec.toObject(includeInstance, f),
    input: (f = msg.getInput()) && proto.pps_v2.Input.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 14, ""),
    cacheSize: jspb.Message.getFieldWithDefault(msg, 15, ""),
    salt: jspb.Message.getFieldWithDefault(msg, 16, ""),
    reason: jspb.Message.getFieldWithDefault(msg, 17, ""),
    maxQueueSize: jspb.Message.getFieldWithDefault(msg, 18, 0),
    service: (f = msg.getService()) && proto.pps_v2.Service.toObject(includeInstance, f),
    spout: (f = msg.getSpout()) && proto.pps_v2.Spout.toObject(includeInstance, f),
    datumSetSpec: (f = msg.getDatumSetSpec()) && proto.pps_v2.DatumSetSpec.toObject(includeInstance, f),
    datumTimeout: (f = msg.getDatumTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    jobTimeout: (f = msg.getJobTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    standby: jspb.Message.getBooleanFieldWithDefault(msg, 24, false),
    datumTries: jspb.Message.getFieldWithDefault(msg, 25, 0),
    schedulingSpec: (f = msg.getSchedulingSpec()) && proto.pps_v2.SchedulingSpec.toObject(includeInstance, f),
    podSpec: jspb.Message.getFieldWithDefault(msg, 27, ""),
    podPatch: jspb.Message.getFieldWithDefault(msg, 28, ""),
    s3Out: jspb.Message.getBooleanFieldWithDefault(msg, 29, false),
    metadata: (f = msg.getMetadata()) && proto.pps_v2.Metadata.toObject(includeInstance, f),
    reprocessSpec: jspb.Message.getFieldWithDefault(msg, 31, ""),
    unclaimedTasks: jspb.Message.getFieldWithDefault(msg, 32, 0),
    workerRc: jspb.Message.getFieldWithDefault(msg, 33, ""),
    autoscaling: jspb.Message.getBooleanFieldWithDefault(msg, 34, false)
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
 * @return {!proto.pps_v2.PipelineInfo.Details}
 */
proto.pps_v2.PipelineInfo.Details.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.PipelineInfo.Details;
  return proto.pps_v2.PipelineInfo.Details.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.PipelineInfo.Details} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.PipelineInfo.Details}
 */
proto.pps_v2.PipelineInfo.Details.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Transform;
      reader.readMessage(value,proto.pps_v2.Transform.deserializeBinaryFromReader);
      msg.setTransform(value);
      break;
    case 2:
      var value = new proto.pps_v2.TFJob;
      reader.readMessage(value,proto.pps_v2.TFJob.deserializeBinaryFromReader);
      msg.setTfJob(value);
      break;
    case 3:
      var value = new proto.pps_v2.ParallelismSpec;
      reader.readMessage(value,proto.pps_v2.ParallelismSpec.deserializeBinaryFromReader);
      msg.setParallelismSpec(value);
      break;
    case 4:
      var value = new proto.pps_v2.Egress;
      reader.readMessage(value,proto.pps_v2.Egress.deserializeBinaryFromReader);
      msg.setEgress(value);
      break;
    case 5:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCreatedAt(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setRecentError(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setWorkersRequested(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setWorkersAvailable(value);
      break;
    case 9:
      var value = /** @type {string} */ (reader.readString());
      msg.setOutputBranch(value);
      break;
    case 10:
      var value = new proto.pps_v2.ResourceSpec;
      reader.readMessage(value,proto.pps_v2.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceRequests(value);
      break;
    case 11:
      var value = new proto.pps_v2.ResourceSpec;
      reader.readMessage(value,proto.pps_v2.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceLimits(value);
      break;
    case 12:
      var value = new proto.pps_v2.ResourceSpec;
      reader.readMessage(value,proto.pps_v2.ResourceSpec.deserializeBinaryFromReader);
      msg.setSidecarResourceLimits(value);
      break;
    case 13:
      var value = new proto.pps_v2.Input;
      reader.readMessage(value,proto.pps_v2.Input.deserializeBinaryFromReader);
      msg.setInput(value);
      break;
    case 14:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 15:
      var value = /** @type {string} */ (reader.readString());
      msg.setCacheSize(value);
      break;
    case 16:
      var value = /** @type {string} */ (reader.readString());
      msg.setSalt(value);
      break;
    case 17:
      var value = /** @type {string} */ (reader.readString());
      msg.setReason(value);
      break;
    case 18:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setMaxQueueSize(value);
      break;
    case 19:
      var value = new proto.pps_v2.Service;
      reader.readMessage(value,proto.pps_v2.Service.deserializeBinaryFromReader);
      msg.setService(value);
      break;
    case 20:
      var value = new proto.pps_v2.Spout;
      reader.readMessage(value,proto.pps_v2.Spout.deserializeBinaryFromReader);
      msg.setSpout(value);
      break;
    case 21:
      var value = new proto.pps_v2.DatumSetSpec;
      reader.readMessage(value,proto.pps_v2.DatumSetSpec.deserializeBinaryFromReader);
      msg.setDatumSetSpec(value);
      break;
    case 22:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setDatumTimeout(value);
      break;
    case 23:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setJobTimeout(value);
      break;
    case 24:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setStandby(value);
      break;
    case 25:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDatumTries(value);
      break;
    case 26:
      var value = new proto.pps_v2.SchedulingSpec;
      reader.readMessage(value,proto.pps_v2.SchedulingSpec.deserializeBinaryFromReader);
      msg.setSchedulingSpec(value);
      break;
    case 27:
      var value = /** @type {string} */ (reader.readString());
      msg.setPodSpec(value);
      break;
    case 28:
      var value = /** @type {string} */ (reader.readString());
      msg.setPodPatch(value);
      break;
    case 29:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setS3Out(value);
      break;
    case 30:
      var value = new proto.pps_v2.Metadata;
      reader.readMessage(value,proto.pps_v2.Metadata.deserializeBinaryFromReader);
      msg.setMetadata(value);
      break;
    case 31:
      var value = /** @type {string} */ (reader.readString());
      msg.setReprocessSpec(value);
      break;
    case 32:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setUnclaimedTasks(value);
      break;
    case 33:
      var value = /** @type {string} */ (reader.readString());
      msg.setWorkerRc(value);
      break;
    case 34:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAutoscaling(value);
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
proto.pps_v2.PipelineInfo.Details.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.PipelineInfo.Details.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.PipelineInfo.Details} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.PipelineInfo.Details.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTransform();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Transform.serializeBinaryToWriter
    );
  }
  f = message.getTfJob();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps_v2.TFJob.serializeBinaryToWriter
    );
  }
  f = message.getParallelismSpec();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pps_v2.ParallelismSpec.serializeBinaryToWriter
    );
  }
  f = message.getEgress();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pps_v2.Egress.serializeBinaryToWriter
    );
  }
  f = message.getCreatedAt();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getRecentError();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getWorkersRequested();
  if (f !== 0) {
    writer.writeInt64(
      7,
      f
    );
  }
  f = message.getWorkersAvailable();
  if (f !== 0) {
    writer.writeInt64(
      8,
      f
    );
  }
  f = message.getOutputBranch();
  if (f.length > 0) {
    writer.writeString(
      9,
      f
    );
  }
  f = message.getResourceRequests();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      proto.pps_v2.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getResourceLimits();
  if (f != null) {
    writer.writeMessage(
      11,
      f,
      proto.pps_v2.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getSidecarResourceLimits();
  if (f != null) {
    writer.writeMessage(
      12,
      f,
      proto.pps_v2.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getInput();
  if (f != null) {
    writer.writeMessage(
      13,
      f,
      proto.pps_v2.Input.serializeBinaryToWriter
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      14,
      f
    );
  }
  f = message.getCacheSize();
  if (f.length > 0) {
    writer.writeString(
      15,
      f
    );
  }
  f = message.getSalt();
  if (f.length > 0) {
    writer.writeString(
      16,
      f
    );
  }
  f = message.getReason();
  if (f.length > 0) {
    writer.writeString(
      17,
      f
    );
  }
  f = message.getMaxQueueSize();
  if (f !== 0) {
    writer.writeInt64(
      18,
      f
    );
  }
  f = message.getService();
  if (f != null) {
    writer.writeMessage(
      19,
      f,
      proto.pps_v2.Service.serializeBinaryToWriter
    );
  }
  f = message.getSpout();
  if (f != null) {
    writer.writeMessage(
      20,
      f,
      proto.pps_v2.Spout.serializeBinaryToWriter
    );
  }
  f = message.getDatumSetSpec();
  if (f != null) {
    writer.writeMessage(
      21,
      f,
      proto.pps_v2.DatumSetSpec.serializeBinaryToWriter
    );
  }
  f = message.getDatumTimeout();
  if (f != null) {
    writer.writeMessage(
      22,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getJobTimeout();
  if (f != null) {
    writer.writeMessage(
      23,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getStandby();
  if (f) {
    writer.writeBool(
      24,
      f
    );
  }
  f = message.getDatumTries();
  if (f !== 0) {
    writer.writeInt64(
      25,
      f
    );
  }
  f = message.getSchedulingSpec();
  if (f != null) {
    writer.writeMessage(
      26,
      f,
      proto.pps_v2.SchedulingSpec.serializeBinaryToWriter
    );
  }
  f = message.getPodSpec();
  if (f.length > 0) {
    writer.writeString(
      27,
      f
    );
  }
  f = message.getPodPatch();
  if (f.length > 0) {
    writer.writeString(
      28,
      f
    );
  }
  f = message.getS3Out();
  if (f) {
    writer.writeBool(
      29,
      f
    );
  }
  f = message.getMetadata();
  if (f != null) {
    writer.writeMessage(
      30,
      f,
      proto.pps_v2.Metadata.serializeBinaryToWriter
    );
  }
  f = message.getReprocessSpec();
  if (f.length > 0) {
    writer.writeString(
      31,
      f
    );
  }
  f = message.getUnclaimedTasks();
  if (f !== 0) {
    writer.writeInt64(
      32,
      f
    );
  }
  f = message.getWorkerRc();
  if (f.length > 0) {
    writer.writeString(
      33,
      f
    );
  }
  f = message.getAutoscaling();
  if (f) {
    writer.writeBool(
      34,
      f
    );
  }
};


/**
 * optional Transform transform = 1;
 * @return {?proto.pps_v2.Transform}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getTransform = function() {
  return /** @type{?proto.pps_v2.Transform} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Transform, 1));
};


/**
 * @param {?proto.pps_v2.Transform|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setTransform = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearTransform = function() {
  return this.setTransform(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasTransform = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional TFJob tf_job = 2;
 * @return {?proto.pps_v2.TFJob}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getTfJob = function() {
  return /** @type{?proto.pps_v2.TFJob} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.TFJob, 2));
};


/**
 * @param {?proto.pps_v2.TFJob|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setTfJob = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearTfJob = function() {
  return this.setTfJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasTfJob = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional ParallelismSpec parallelism_spec = 3;
 * @return {?proto.pps_v2.ParallelismSpec}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getParallelismSpec = function() {
  return /** @type{?proto.pps_v2.ParallelismSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ParallelismSpec, 3));
};


/**
 * @param {?proto.pps_v2.ParallelismSpec|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setParallelismSpec = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearParallelismSpec = function() {
  return this.setParallelismSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasParallelismSpec = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional Egress egress = 4;
 * @return {?proto.pps_v2.Egress}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getEgress = function() {
  return /** @type{?proto.pps_v2.Egress} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Egress, 4));
};


/**
 * @param {?proto.pps_v2.Egress|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setEgress = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearEgress = function() {
  return this.setEgress(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasEgress = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional google.protobuf.Timestamp created_at = 5;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getCreatedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 5));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setCreatedAt = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearCreatedAt = function() {
  return this.setCreatedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasCreatedAt = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional string recent_error = 6;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getRecentError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setRecentError = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional int64 workers_requested = 7;
 * @return {number}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getWorkersRequested = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setWorkersRequested = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
};


/**
 * optional int64 workers_available = 8;
 * @return {number}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getWorkersAvailable = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setWorkersAvailable = function(value) {
  return jspb.Message.setProto3IntField(this, 8, value);
};


/**
 * optional string output_branch = 9;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getOutputBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 9, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setOutputBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 9, value);
};


/**
 * optional ResourceSpec resource_requests = 10;
 * @return {?proto.pps_v2.ResourceSpec}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getResourceRequests = function() {
  return /** @type{?proto.pps_v2.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ResourceSpec, 10));
};


/**
 * @param {?proto.pps_v2.ResourceSpec|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setResourceRequests = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearResourceRequests = function() {
  return this.setResourceRequests(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasResourceRequests = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * optional ResourceSpec resource_limits = 11;
 * @return {?proto.pps_v2.ResourceSpec}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getResourceLimits = function() {
  return /** @type{?proto.pps_v2.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ResourceSpec, 11));
};


/**
 * @param {?proto.pps_v2.ResourceSpec|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 11, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearResourceLimits = function() {
  return this.setResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasResourceLimits = function() {
  return jspb.Message.getField(this, 11) != null;
};


/**
 * optional ResourceSpec sidecar_resource_limits = 12;
 * @return {?proto.pps_v2.ResourceSpec}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getSidecarResourceLimits = function() {
  return /** @type{?proto.pps_v2.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ResourceSpec, 12));
};


/**
 * @param {?proto.pps_v2.ResourceSpec|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setSidecarResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 12, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearSidecarResourceLimits = function() {
  return this.setSidecarResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasSidecarResourceLimits = function() {
  return jspb.Message.getField(this, 12) != null;
};


/**
 * optional Input input = 13;
 * @return {?proto.pps_v2.Input}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getInput = function() {
  return /** @type{?proto.pps_v2.Input} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Input, 13));
};


/**
 * @param {?proto.pps_v2.Input|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setInput = function(value) {
  return jspb.Message.setWrapperField(this, 13, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearInput = function() {
  return this.setInput(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasInput = function() {
  return jspb.Message.getField(this, 13) != null;
};


/**
 * optional string description = 14;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 14, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 14, value);
};


/**
 * optional string cache_size = 15;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getCacheSize = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 15, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setCacheSize = function(value) {
  return jspb.Message.setProto3StringField(this, 15, value);
};


/**
 * optional string salt = 16;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getSalt = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 16, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setSalt = function(value) {
  return jspb.Message.setProto3StringField(this, 16, value);
};


/**
 * optional string reason = 17;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 17, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setReason = function(value) {
  return jspb.Message.setProto3StringField(this, 17, value);
};


/**
 * optional int64 max_queue_size = 18;
 * @return {number}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getMaxQueueSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 18, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setMaxQueueSize = function(value) {
  return jspb.Message.setProto3IntField(this, 18, value);
};


/**
 * optional Service service = 19;
 * @return {?proto.pps_v2.Service}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getService = function() {
  return /** @type{?proto.pps_v2.Service} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Service, 19));
};


/**
 * @param {?proto.pps_v2.Service|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setService = function(value) {
  return jspb.Message.setWrapperField(this, 19, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearService = function() {
  return this.setService(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasService = function() {
  return jspb.Message.getField(this, 19) != null;
};


/**
 * optional Spout spout = 20;
 * @return {?proto.pps_v2.Spout}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getSpout = function() {
  return /** @type{?proto.pps_v2.Spout} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Spout, 20));
};


/**
 * @param {?proto.pps_v2.Spout|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setSpout = function(value) {
  return jspb.Message.setWrapperField(this, 20, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearSpout = function() {
  return this.setSpout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasSpout = function() {
  return jspb.Message.getField(this, 20) != null;
};


/**
 * optional DatumSetSpec datum_set_spec = 21;
 * @return {?proto.pps_v2.DatumSetSpec}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getDatumSetSpec = function() {
  return /** @type{?proto.pps_v2.DatumSetSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.DatumSetSpec, 21));
};


/**
 * @param {?proto.pps_v2.DatumSetSpec|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setDatumSetSpec = function(value) {
  return jspb.Message.setWrapperField(this, 21, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearDatumSetSpec = function() {
  return this.setDatumSetSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasDatumSetSpec = function() {
  return jspb.Message.getField(this, 21) != null;
};


/**
 * optional google.protobuf.Duration datum_timeout = 22;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getDatumTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 22));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setDatumTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 22, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearDatumTimeout = function() {
  return this.setDatumTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasDatumTimeout = function() {
  return jspb.Message.getField(this, 22) != null;
};


/**
 * optional google.protobuf.Duration job_timeout = 23;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getJobTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 23));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setJobTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 23, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearJobTimeout = function() {
  return this.setJobTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasJobTimeout = function() {
  return jspb.Message.getField(this, 23) != null;
};


/**
 * optional bool standby = 24;
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getStandby = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 24, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setStandby = function(value) {
  return jspb.Message.setProto3BooleanField(this, 24, value);
};


/**
 * optional int64 datum_tries = 25;
 * @return {number}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getDatumTries = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 25, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setDatumTries = function(value) {
  return jspb.Message.setProto3IntField(this, 25, value);
};


/**
 * optional SchedulingSpec scheduling_spec = 26;
 * @return {?proto.pps_v2.SchedulingSpec}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getSchedulingSpec = function() {
  return /** @type{?proto.pps_v2.SchedulingSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.SchedulingSpec, 26));
};


/**
 * @param {?proto.pps_v2.SchedulingSpec|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setSchedulingSpec = function(value) {
  return jspb.Message.setWrapperField(this, 26, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearSchedulingSpec = function() {
  return this.setSchedulingSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasSchedulingSpec = function() {
  return jspb.Message.getField(this, 26) != null;
};


/**
 * optional string pod_spec = 27;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getPodSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 27, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setPodSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 27, value);
};


/**
 * optional string pod_patch = 28;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getPodPatch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 28, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setPodPatch = function(value) {
  return jspb.Message.setProto3StringField(this, 28, value);
};


/**
 * optional bool s3_out = 29;
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getS3Out = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 29, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setS3Out = function(value) {
  return jspb.Message.setProto3BooleanField(this, 29, value);
};


/**
 * optional Metadata metadata = 30;
 * @return {?proto.pps_v2.Metadata}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getMetadata = function() {
  return /** @type{?proto.pps_v2.Metadata} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Metadata, 30));
};


/**
 * @param {?proto.pps_v2.Metadata|undefined} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
*/
proto.pps_v2.PipelineInfo.Details.prototype.setMetadata = function(value) {
  return jspb.Message.setWrapperField(this, 30, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.clearMetadata = function() {
  return this.setMetadata(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.hasMetadata = function() {
  return jspb.Message.getField(this, 30) != null;
};


/**
 * optional string reprocess_spec = 31;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getReprocessSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 31, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setReprocessSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 31, value);
};


/**
 * optional int64 unclaimed_tasks = 32;
 * @return {number}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getUnclaimedTasks = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 32, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setUnclaimedTasks = function(value) {
  return jspb.Message.setProto3IntField(this, 32, value);
};


/**
 * optional string worker_rc = 33;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getWorkerRc = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 33, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setWorkerRc = function(value) {
  return jspb.Message.setProto3StringField(this, 33, value);
};


/**
 * optional bool autoscaling = 34;
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.Details.prototype.getAutoscaling = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 34, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.PipelineInfo.Details} returns this
 */
proto.pps_v2.PipelineInfo.Details.prototype.setAutoscaling = function(value) {
  return jspb.Message.setProto3BooleanField(this, 34, value);
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.PipelineInfo.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.PipelineInfo} returns this
*/
proto.pps_v2.PipelineInfo.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional uint64 version = 2;
 * @return {number}
 */
proto.pps_v2.PipelineInfo.prototype.getVersion = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.setVersion = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional pfs_v2.Commit spec_commit = 3;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pps_v2.PipelineInfo.prototype.getSpecCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 3));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pps_v2.PipelineInfo} returns this
*/
proto.pps_v2.PipelineInfo.prototype.setSpecCommit = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.clearSpecCommit = function() {
  return this.setSpecCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.prototype.hasSpecCommit = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional bool stopped = 4;
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.prototype.getStopped = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.setStopped = function(value) {
  return jspb.Message.setProto3BooleanField(this, 4, value);
};


/**
 * optional PipelineState state = 5;
 * @return {!proto.pps_v2.PipelineState}
 */
proto.pps_v2.PipelineInfo.prototype.getState = function() {
  return /** @type {!proto.pps_v2.PipelineState} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {!proto.pps_v2.PipelineState} value
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 5, value);
};


/**
 * optional string reason = 6;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.setReason = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * map<int32, int32> job_counts = 7;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<number,number>}
 */
proto.pps_v2.PipelineInfo.prototype.getJobCountsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<number,number>} */ (
      jspb.Message.getMapField(this, 7, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.clearJobCountsMap = function() {
  this.getJobCountsMap().clear();
  return this;};


/**
 * optional JobState last_job_state = 8;
 * @return {!proto.pps_v2.JobState}
 */
proto.pps_v2.PipelineInfo.prototype.getLastJobState = function() {
  return /** @type {!proto.pps_v2.JobState} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {!proto.pps_v2.JobState} value
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.setLastJobState = function(value) {
  return jspb.Message.setProto3EnumField(this, 8, value);
};


/**
 * optional uint64 parallelism = 9;
 * @return {number}
 */
proto.pps_v2.PipelineInfo.prototype.getParallelism = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.setParallelism = function(value) {
  return jspb.Message.setProto3IntField(this, 9, value);
};


/**
 * optional PipelineType type = 10;
 * @return {!proto.pps_v2.PipelineInfo.PipelineType}
 */
proto.pps_v2.PipelineInfo.prototype.getType = function() {
  return /** @type {!proto.pps_v2.PipelineInfo.PipelineType} */ (jspb.Message.getFieldWithDefault(this, 10, 0));
};


/**
 * @param {!proto.pps_v2.PipelineInfo.PipelineType} value
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.setType = function(value) {
  return jspb.Message.setProto3EnumField(this, 10, value);
};


/**
 * optional string auth_token = 11;
 * @return {string}
 */
proto.pps_v2.PipelineInfo.prototype.getAuthToken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 11, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.setAuthToken = function(value) {
  return jspb.Message.setProto3StringField(this, 11, value);
};


/**
 * optional Details details = 12;
 * @return {?proto.pps_v2.PipelineInfo.Details}
 */
proto.pps_v2.PipelineInfo.prototype.getDetails = function() {
  return /** @type{?proto.pps_v2.PipelineInfo.Details} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.PipelineInfo.Details, 12));
};


/**
 * @param {?proto.pps_v2.PipelineInfo.Details|undefined} value
 * @return {!proto.pps_v2.PipelineInfo} returns this
*/
proto.pps_v2.PipelineInfo.prototype.setDetails = function(value) {
  return jspb.Message.setWrapperField(this, 12, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.PipelineInfo} returns this
 */
proto.pps_v2.PipelineInfo.prototype.clearDetails = function() {
  return this.setDetails(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.PipelineInfo.prototype.hasDetails = function() {
  return jspb.Message.getField(this, 12) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.PipelineInfos.repeatedFields_ = [1];



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
proto.pps_v2.PipelineInfos.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.PipelineInfos.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.PipelineInfos} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.PipelineInfos.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipelineInfoList: jspb.Message.toObjectList(msg.getPipelineInfoList(),
    proto.pps_v2.PipelineInfo.toObject, includeInstance)
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
 * @return {!proto.pps_v2.PipelineInfos}
 */
proto.pps_v2.PipelineInfos.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.PipelineInfos;
  return proto.pps_v2.PipelineInfos.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.PipelineInfos} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.PipelineInfos}
 */
proto.pps_v2.PipelineInfos.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.PipelineInfo;
      reader.readMessage(value,proto.pps_v2.PipelineInfo.deserializeBinaryFromReader);
      msg.addPipelineInfo(value);
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
proto.pps_v2.PipelineInfos.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.PipelineInfos.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.PipelineInfos} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.PipelineInfos.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipelineInfoList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pps_v2.PipelineInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated PipelineInfo pipeline_info = 1;
 * @return {!Array<!proto.pps_v2.PipelineInfo>}
 */
proto.pps_v2.PipelineInfos.prototype.getPipelineInfoList = function() {
  return /** @type{!Array<!proto.pps_v2.PipelineInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps_v2.PipelineInfo, 1));
};


/**
 * @param {!Array<!proto.pps_v2.PipelineInfo>} value
 * @return {!proto.pps_v2.PipelineInfos} returns this
*/
proto.pps_v2.PipelineInfos.prototype.setPipelineInfoList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pps_v2.PipelineInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.PipelineInfo}
 */
proto.pps_v2.PipelineInfos.prototype.addPipelineInfo = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pps_v2.PipelineInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.PipelineInfos} returns this
 */
proto.pps_v2.PipelineInfos.prototype.clearPipelineInfoList = function() {
  return this.setPipelineInfoList([]);
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
proto.pps_v2.JobSet.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.JobSet.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.JobSet} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.JobSet.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, "")
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
 * @return {!proto.pps_v2.JobSet}
 */
proto.pps_v2.JobSet.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.JobSet;
  return proto.pps_v2.JobSet.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.JobSet} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.JobSet}
 */
proto.pps_v2.JobSet.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps_v2.JobSet.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.JobSet.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.JobSet} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.JobSet.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.pps_v2.JobSet.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.JobSet} returns this
 */
proto.pps_v2.JobSet.prototype.setId = function(value) {
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
proto.pps_v2.InspectJobSetRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.InspectJobSetRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.InspectJobSetRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InspectJobSetRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    jobSet: (f = msg.getJobSet()) && proto.pps_v2.JobSet.toObject(includeInstance, f),
    wait: jspb.Message.getBooleanFieldWithDefault(msg, 2, false),
    details: jspb.Message.getBooleanFieldWithDefault(msg, 3, false)
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
 * @return {!proto.pps_v2.InspectJobSetRequest}
 */
proto.pps_v2.InspectJobSetRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.InspectJobSetRequest;
  return proto.pps_v2.InspectJobSetRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.InspectJobSetRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.InspectJobSetRequest}
 */
proto.pps_v2.InspectJobSetRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.JobSet;
      reader.readMessage(value,proto.pps_v2.JobSet.deserializeBinaryFromReader);
      msg.setJobSet(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setWait(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDetails(value);
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
proto.pps_v2.InspectJobSetRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.InspectJobSetRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.InspectJobSetRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InspectJobSetRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJobSet();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.JobSet.serializeBinaryToWriter
    );
  }
  f = message.getWait();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
  f = message.getDetails();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
};


/**
 * optional JobSet job_set = 1;
 * @return {?proto.pps_v2.JobSet}
 */
proto.pps_v2.InspectJobSetRequest.prototype.getJobSet = function() {
  return /** @type{?proto.pps_v2.JobSet} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.JobSet, 1));
};


/**
 * @param {?proto.pps_v2.JobSet|undefined} value
 * @return {!proto.pps_v2.InspectJobSetRequest} returns this
*/
proto.pps_v2.InspectJobSetRequest.prototype.setJobSet = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.InspectJobSetRequest} returns this
 */
proto.pps_v2.InspectJobSetRequest.prototype.clearJobSet = function() {
  return this.setJobSet(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.InspectJobSetRequest.prototype.hasJobSet = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool wait = 2;
 * @return {boolean}
 */
proto.pps_v2.InspectJobSetRequest.prototype.getWait = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.InspectJobSetRequest} returns this
 */
proto.pps_v2.InspectJobSetRequest.prototype.setWait = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};


/**
 * optional bool details = 3;
 * @return {boolean}
 */
proto.pps_v2.InspectJobSetRequest.prototype.getDetails = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.InspectJobSetRequest} returns this
 */
proto.pps_v2.InspectJobSetRequest.prototype.setDetails = function(value) {
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
proto.pps_v2.InspectJobRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.InspectJobRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.InspectJobRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InspectJobRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps_v2.Job.toObject(includeInstance, f),
    wait: jspb.Message.getBooleanFieldWithDefault(msg, 2, false),
    details: jspb.Message.getBooleanFieldWithDefault(msg, 3, false)
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
 * @return {!proto.pps_v2.InspectJobRequest}
 */
proto.pps_v2.InspectJobRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.InspectJobRequest;
  return proto.pps_v2.InspectJobRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.InspectJobRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.InspectJobRequest}
 */
proto.pps_v2.InspectJobRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Job;
      reader.readMessage(value,proto.pps_v2.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setWait(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDetails(value);
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
proto.pps_v2.InspectJobRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.InspectJobRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.InspectJobRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InspectJobRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Job.serializeBinaryToWriter
    );
  }
  f = message.getWait();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
  f = message.getDetails();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps_v2.Job}
 */
proto.pps_v2.InspectJobRequest.prototype.getJob = function() {
  return /** @type{?proto.pps_v2.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Job, 1));
};


/**
 * @param {?proto.pps_v2.Job|undefined} value
 * @return {!proto.pps_v2.InspectJobRequest} returns this
*/
proto.pps_v2.InspectJobRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.InspectJobRequest} returns this
 */
proto.pps_v2.InspectJobRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.InspectJobRequest.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool wait = 2;
 * @return {boolean}
 */
proto.pps_v2.InspectJobRequest.prototype.getWait = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.InspectJobRequest} returns this
 */
proto.pps_v2.InspectJobRequest.prototype.setWait = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};


/**
 * optional bool details = 3;
 * @return {boolean}
 */
proto.pps_v2.InspectJobRequest.prototype.getDetails = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.InspectJobRequest} returns this
 */
proto.pps_v2.InspectJobRequest.prototype.setDetails = function(value) {
  return jspb.Message.setProto3BooleanField(this, 3, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.ListJobRequest.repeatedFields_ = [2];



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
proto.pps_v2.ListJobRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.ListJobRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.ListJobRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ListJobRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f),
    inputCommitList: jspb.Message.toObjectList(msg.getInputCommitList(),
    pfs_pfs_pb.Commit.toObject, includeInstance),
    history: jspb.Message.getFieldWithDefault(msg, 4, 0),
    details: jspb.Message.getBooleanFieldWithDefault(msg, 5, false),
    jqfilter: jspb.Message.getFieldWithDefault(msg, 6, "")
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
 * @return {!proto.pps_v2.ListJobRequest}
 */
proto.pps_v2.ListJobRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.ListJobRequest;
  return proto.pps_v2.ListJobRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.ListJobRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.ListJobRequest}
 */
proto.pps_v2.ListJobRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.addInputCommit(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setHistory(value);
      break;
    case 5:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDetails(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setJqfilter(value);
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
proto.pps_v2.ListJobRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.ListJobRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.ListJobRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ListJobRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getInputCommitList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getHistory();
  if (f !== 0) {
    writer.writeInt64(
      4,
      f
    );
  }
  f = message.getDetails();
  if (f) {
    writer.writeBool(
      5,
      f
    );
  }
  f = message.getJqfilter();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.ListJobRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.ListJobRequest} returns this
*/
proto.pps_v2.ListJobRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.ListJobRequest} returns this
 */
proto.pps_v2.ListJobRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.ListJobRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * repeated pfs_v2.Commit input_commit = 2;
 * @return {!Array<!proto.pfs_v2.Commit>}
 */
proto.pps_v2.ListJobRequest.prototype.getInputCommitList = function() {
  return /** @type{!Array<!proto.pfs_v2.Commit>} */ (
    jspb.Message.getRepeatedWrapperField(this, pfs_pfs_pb.Commit, 2));
};


/**
 * @param {!Array<!proto.pfs_v2.Commit>} value
 * @return {!proto.pps_v2.ListJobRequest} returns this
*/
proto.pps_v2.ListJobRequest.prototype.setInputCommitList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pfs_v2.Commit=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.Commit}
 */
proto.pps_v2.ListJobRequest.prototype.addInputCommit = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pfs_v2.Commit, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.ListJobRequest} returns this
 */
proto.pps_v2.ListJobRequest.prototype.clearInputCommitList = function() {
  return this.setInputCommitList([]);
};


/**
 * optional int64 history = 4;
 * @return {number}
 */
proto.pps_v2.ListJobRequest.prototype.getHistory = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.ListJobRequest} returns this
 */
proto.pps_v2.ListJobRequest.prototype.setHistory = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional bool details = 5;
 * @return {boolean}
 */
proto.pps_v2.ListJobRequest.prototype.getDetails = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 5, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.ListJobRequest} returns this
 */
proto.pps_v2.ListJobRequest.prototype.setDetails = function(value) {
  return jspb.Message.setProto3BooleanField(this, 5, value);
};


/**
 * optional string jqFilter = 6;
 * @return {string}
 */
proto.pps_v2.ListJobRequest.prototype.getJqfilter = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.ListJobRequest} returns this
 */
proto.pps_v2.ListJobRequest.prototype.setJqfilter = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
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
proto.pps_v2.SubscribeJobRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.SubscribeJobRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.SubscribeJobRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.SubscribeJobRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f),
    details: jspb.Message.getBooleanFieldWithDefault(msg, 2, false)
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
 * @return {!proto.pps_v2.SubscribeJobRequest}
 */
proto.pps_v2.SubscribeJobRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.SubscribeJobRequest;
  return proto.pps_v2.SubscribeJobRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.SubscribeJobRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.SubscribeJobRequest}
 */
proto.pps_v2.SubscribeJobRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDetails(value);
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
proto.pps_v2.SubscribeJobRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.SubscribeJobRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.SubscribeJobRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.SubscribeJobRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getDetails();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.SubscribeJobRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.SubscribeJobRequest} returns this
*/
proto.pps_v2.SubscribeJobRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.SubscribeJobRequest} returns this
 */
proto.pps_v2.SubscribeJobRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.SubscribeJobRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool details = 2;
 * @return {boolean}
 */
proto.pps_v2.SubscribeJobRequest.prototype.getDetails = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.SubscribeJobRequest} returns this
 */
proto.pps_v2.SubscribeJobRequest.prototype.setDetails = function(value) {
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
proto.pps_v2.DeleteJobRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.DeleteJobRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.DeleteJobRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DeleteJobRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps_v2.Job.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.DeleteJobRequest}
 */
proto.pps_v2.DeleteJobRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.DeleteJobRequest;
  return proto.pps_v2.DeleteJobRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.DeleteJobRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.DeleteJobRequest}
 */
proto.pps_v2.DeleteJobRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Job;
      reader.readMessage(value,proto.pps_v2.Job.deserializeBinaryFromReader);
      msg.setJob(value);
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
proto.pps_v2.DeleteJobRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.DeleteJobRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.DeleteJobRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DeleteJobRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Job.serializeBinaryToWriter
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps_v2.Job}
 */
proto.pps_v2.DeleteJobRequest.prototype.getJob = function() {
  return /** @type{?proto.pps_v2.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Job, 1));
};


/**
 * @param {?proto.pps_v2.Job|undefined} value
 * @return {!proto.pps_v2.DeleteJobRequest} returns this
*/
proto.pps_v2.DeleteJobRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.DeleteJobRequest} returns this
 */
proto.pps_v2.DeleteJobRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.DeleteJobRequest.prototype.hasJob = function() {
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
proto.pps_v2.StopJobRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.StopJobRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.StopJobRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.StopJobRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps_v2.Job.toObject(includeInstance, f),
    reason: jspb.Message.getFieldWithDefault(msg, 3, "")
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
 * @return {!proto.pps_v2.StopJobRequest}
 */
proto.pps_v2.StopJobRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.StopJobRequest;
  return proto.pps_v2.StopJobRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.StopJobRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.StopJobRequest}
 */
proto.pps_v2.StopJobRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Job;
      reader.readMessage(value,proto.pps_v2.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setReason(value);
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
proto.pps_v2.StopJobRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.StopJobRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.StopJobRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.StopJobRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Job.serializeBinaryToWriter
    );
  }
  f = message.getReason();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps_v2.Job}
 */
proto.pps_v2.StopJobRequest.prototype.getJob = function() {
  return /** @type{?proto.pps_v2.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Job, 1));
};


/**
 * @param {?proto.pps_v2.Job|undefined} value
 * @return {!proto.pps_v2.StopJobRequest} returns this
*/
proto.pps_v2.StopJobRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.StopJobRequest} returns this
 */
proto.pps_v2.StopJobRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.StopJobRequest.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string reason = 3;
 * @return {string}
 */
proto.pps_v2.StopJobRequest.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.StopJobRequest} returns this
 */
proto.pps_v2.StopJobRequest.prototype.setReason = function(value) {
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
proto.pps_v2.UpdateJobStateRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.UpdateJobStateRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.UpdateJobStateRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.UpdateJobStateRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps_v2.Job.toObject(includeInstance, f),
    state: jspb.Message.getFieldWithDefault(msg, 2, 0),
    reason: jspb.Message.getFieldWithDefault(msg, 3, ""),
    restart: jspb.Message.getFieldWithDefault(msg, 5, 0),
    dataProcessed: jspb.Message.getFieldWithDefault(msg, 6, 0),
    dataSkipped: jspb.Message.getFieldWithDefault(msg, 7, 0),
    dataFailed: jspb.Message.getFieldWithDefault(msg, 8, 0),
    dataRecovered: jspb.Message.getFieldWithDefault(msg, 9, 0),
    dataTotal: jspb.Message.getFieldWithDefault(msg, 10, 0),
    stats: (f = msg.getStats()) && proto.pps_v2.ProcessStats.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.UpdateJobStateRequest}
 */
proto.pps_v2.UpdateJobStateRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.UpdateJobStateRequest;
  return proto.pps_v2.UpdateJobStateRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.UpdateJobStateRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.UpdateJobStateRequest}
 */
proto.pps_v2.UpdateJobStateRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Job;
      reader.readMessage(value,proto.pps_v2.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 2:
      var value = /** @type {!proto.pps_v2.JobState} */ (reader.readEnum());
      msg.setState(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setReason(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setRestart(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataProcessed(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataSkipped(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataFailed(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataRecovered(value);
      break;
    case 10:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataTotal(value);
      break;
    case 11:
      var value = new proto.pps_v2.ProcessStats;
      reader.readMessage(value,proto.pps_v2.ProcessStats.deserializeBinaryFromReader);
      msg.setStats(value);
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
proto.pps_v2.UpdateJobStateRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.UpdateJobStateRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.UpdateJobStateRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.UpdateJobStateRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Job.serializeBinaryToWriter
    );
  }
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
  f = message.getReason();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getRestart();
  if (f !== 0) {
    writer.writeUint64(
      5,
      f
    );
  }
  f = message.getDataProcessed();
  if (f !== 0) {
    writer.writeInt64(
      6,
      f
    );
  }
  f = message.getDataSkipped();
  if (f !== 0) {
    writer.writeInt64(
      7,
      f
    );
  }
  f = message.getDataFailed();
  if (f !== 0) {
    writer.writeInt64(
      8,
      f
    );
  }
  f = message.getDataRecovered();
  if (f !== 0) {
    writer.writeInt64(
      9,
      f
    );
  }
  f = message.getDataTotal();
  if (f !== 0) {
    writer.writeInt64(
      10,
      f
    );
  }
  f = message.getStats();
  if (f != null) {
    writer.writeMessage(
      11,
      f,
      proto.pps_v2.ProcessStats.serializeBinaryToWriter
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps_v2.Job}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.getJob = function() {
  return /** @type{?proto.pps_v2.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Job, 1));
};


/**
 * @param {?proto.pps_v2.Job|undefined} value
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
*/
proto.pps_v2.UpdateJobStateRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
 */
proto.pps_v2.UpdateJobStateRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional JobState state = 2;
 * @return {!proto.pps_v2.JobState}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.getState = function() {
  return /** @type {!proto.pps_v2.JobState} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.pps_v2.JobState} value
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
 */
proto.pps_v2.UpdateJobStateRequest.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 2, value);
};


/**
 * optional string reason = 3;
 * @return {string}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
 */
proto.pps_v2.UpdateJobStateRequest.prototype.setReason = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional uint64 restart = 5;
 * @return {number}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.getRestart = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
 */
proto.pps_v2.UpdateJobStateRequest.prototype.setRestart = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};


/**
 * optional int64 data_processed = 6;
 * @return {number}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.getDataProcessed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
 */
proto.pps_v2.UpdateJobStateRequest.prototype.setDataProcessed = function(value) {
  return jspb.Message.setProto3IntField(this, 6, value);
};


/**
 * optional int64 data_skipped = 7;
 * @return {number}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.getDataSkipped = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
 */
proto.pps_v2.UpdateJobStateRequest.prototype.setDataSkipped = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
};


/**
 * optional int64 data_failed = 8;
 * @return {number}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.getDataFailed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
 */
proto.pps_v2.UpdateJobStateRequest.prototype.setDataFailed = function(value) {
  return jspb.Message.setProto3IntField(this, 8, value);
};


/**
 * optional int64 data_recovered = 9;
 * @return {number}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.getDataRecovered = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
 */
proto.pps_v2.UpdateJobStateRequest.prototype.setDataRecovered = function(value) {
  return jspb.Message.setProto3IntField(this, 9, value);
};


/**
 * optional int64 data_total = 10;
 * @return {number}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.getDataTotal = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 10, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
 */
proto.pps_v2.UpdateJobStateRequest.prototype.setDataTotal = function(value) {
  return jspb.Message.setProto3IntField(this, 10, value);
};


/**
 * optional ProcessStats stats = 11;
 * @return {?proto.pps_v2.ProcessStats}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.getStats = function() {
  return /** @type{?proto.pps_v2.ProcessStats} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ProcessStats, 11));
};


/**
 * @param {?proto.pps_v2.ProcessStats|undefined} value
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
*/
proto.pps_v2.UpdateJobStateRequest.prototype.setStats = function(value) {
  return jspb.Message.setWrapperField(this, 11, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.UpdateJobStateRequest} returns this
 */
proto.pps_v2.UpdateJobStateRequest.prototype.clearStats = function() {
  return this.setStats(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.UpdateJobStateRequest.prototype.hasStats = function() {
  return jspb.Message.getField(this, 11) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.GetLogsRequest.repeatedFields_ = [3];



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
proto.pps_v2.GetLogsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.GetLogsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.GetLogsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.GetLogsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f),
    job: (f = msg.getJob()) && proto.pps_v2.Job.toObject(includeInstance, f),
    dataFiltersList: (f = jspb.Message.getRepeatedField(msg, 3)) == null ? undefined : f,
    datum: (f = msg.getDatum()) && proto.pps_v2.Datum.toObject(includeInstance, f),
    master: jspb.Message.getBooleanFieldWithDefault(msg, 5, false),
    follow: jspb.Message.getBooleanFieldWithDefault(msg, 6, false),
    tail: jspb.Message.getFieldWithDefault(msg, 7, 0),
    useLokiBackend: jspb.Message.getBooleanFieldWithDefault(msg, 8, false),
    since: (f = msg.getSince()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.GetLogsRequest}
 */
proto.pps_v2.GetLogsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.GetLogsRequest;
  return proto.pps_v2.GetLogsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.GetLogsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.GetLogsRequest}
 */
proto.pps_v2.GetLogsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = new proto.pps_v2.Job;
      reader.readMessage(value,proto.pps_v2.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.addDataFilters(value);
      break;
    case 4:
      var value = new proto.pps_v2.Datum;
      reader.readMessage(value,proto.pps_v2.Datum.deserializeBinaryFromReader);
      msg.setDatum(value);
      break;
    case 5:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setMaster(value);
      break;
    case 6:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setFollow(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTail(value);
      break;
    case 8:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setUseLokiBackend(value);
      break;
    case 9:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setSince(value);
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
proto.pps_v2.GetLogsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.GetLogsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.GetLogsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.GetLogsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps_v2.Job.serializeBinaryToWriter
    );
  }
  f = message.getDataFiltersList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      3,
      f
    );
  }
  f = message.getDatum();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pps_v2.Datum.serializeBinaryToWriter
    );
  }
  f = message.getMaster();
  if (f) {
    writer.writeBool(
      5,
      f
    );
  }
  f = message.getFollow();
  if (f) {
    writer.writeBool(
      6,
      f
    );
  }
  f = message.getTail();
  if (f !== 0) {
    writer.writeInt64(
      7,
      f
    );
  }
  f = message.getUseLokiBackend();
  if (f) {
    writer.writeBool(
      8,
      f
    );
  }
  f = message.getSince();
  if (f != null) {
    writer.writeMessage(
      9,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.GetLogsRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.GetLogsRequest} returns this
*/
proto.pps_v2.GetLogsRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.GetLogsRequest} returns this
 */
proto.pps_v2.GetLogsRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.GetLogsRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Job job = 2;
 * @return {?proto.pps_v2.Job}
 */
proto.pps_v2.GetLogsRequest.prototype.getJob = function() {
  return /** @type{?proto.pps_v2.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Job, 2));
};


/**
 * @param {?proto.pps_v2.Job|undefined} value
 * @return {!proto.pps_v2.GetLogsRequest} returns this
*/
proto.pps_v2.GetLogsRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.GetLogsRequest} returns this
 */
proto.pps_v2.GetLogsRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.GetLogsRequest.prototype.hasJob = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * repeated string data_filters = 3;
 * @return {!Array<string>}
 */
proto.pps_v2.GetLogsRequest.prototype.getDataFiltersList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 3));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps_v2.GetLogsRequest} returns this
 */
proto.pps_v2.GetLogsRequest.prototype.setDataFiltersList = function(value) {
  return jspb.Message.setField(this, 3, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.GetLogsRequest} returns this
 */
proto.pps_v2.GetLogsRequest.prototype.addDataFilters = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 3, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.GetLogsRequest} returns this
 */
proto.pps_v2.GetLogsRequest.prototype.clearDataFiltersList = function() {
  return this.setDataFiltersList([]);
};


/**
 * optional Datum datum = 4;
 * @return {?proto.pps_v2.Datum}
 */
proto.pps_v2.GetLogsRequest.prototype.getDatum = function() {
  return /** @type{?proto.pps_v2.Datum} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Datum, 4));
};


/**
 * @param {?proto.pps_v2.Datum|undefined} value
 * @return {!proto.pps_v2.GetLogsRequest} returns this
*/
proto.pps_v2.GetLogsRequest.prototype.setDatum = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.GetLogsRequest} returns this
 */
proto.pps_v2.GetLogsRequest.prototype.clearDatum = function() {
  return this.setDatum(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.GetLogsRequest.prototype.hasDatum = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional bool master = 5;
 * @return {boolean}
 */
proto.pps_v2.GetLogsRequest.prototype.getMaster = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 5, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.GetLogsRequest} returns this
 */
proto.pps_v2.GetLogsRequest.prototype.setMaster = function(value) {
  return jspb.Message.setProto3BooleanField(this, 5, value);
};


/**
 * optional bool follow = 6;
 * @return {boolean}
 */
proto.pps_v2.GetLogsRequest.prototype.getFollow = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 6, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.GetLogsRequest} returns this
 */
proto.pps_v2.GetLogsRequest.prototype.setFollow = function(value) {
  return jspb.Message.setProto3BooleanField(this, 6, value);
};


/**
 * optional int64 tail = 7;
 * @return {number}
 */
proto.pps_v2.GetLogsRequest.prototype.getTail = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.GetLogsRequest} returns this
 */
proto.pps_v2.GetLogsRequest.prototype.setTail = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
};


/**
 * optional bool use_loki_backend = 8;
 * @return {boolean}
 */
proto.pps_v2.GetLogsRequest.prototype.getUseLokiBackend = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 8, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.GetLogsRequest} returns this
 */
proto.pps_v2.GetLogsRequest.prototype.setUseLokiBackend = function(value) {
  return jspb.Message.setProto3BooleanField(this, 8, value);
};


/**
 * optional google.protobuf.Duration since = 9;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps_v2.GetLogsRequest.prototype.getSince = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 9));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps_v2.GetLogsRequest} returns this
*/
proto.pps_v2.GetLogsRequest.prototype.setSince = function(value) {
  return jspb.Message.setWrapperField(this, 9, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.GetLogsRequest} returns this
 */
proto.pps_v2.GetLogsRequest.prototype.clearSince = function() {
  return this.setSince(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.GetLogsRequest.prototype.hasSince = function() {
  return jspb.Message.getField(this, 9) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.LogMessage.repeatedFields_ = [6];



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
proto.pps_v2.LogMessage.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.LogMessage.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.LogMessage} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.LogMessage.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipelineName: jspb.Message.getFieldWithDefault(msg, 1, ""),
    jobId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    workerId: jspb.Message.getFieldWithDefault(msg, 3, ""),
    datumId: jspb.Message.getFieldWithDefault(msg, 4, ""),
    master: jspb.Message.getBooleanFieldWithDefault(msg, 5, false),
    dataList: jspb.Message.toObjectList(msg.getDataList(),
    proto.pps_v2.InputFile.toObject, includeInstance),
    user: jspb.Message.getBooleanFieldWithDefault(msg, 7, false),
    ts: (f = msg.getTs()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    message: jspb.Message.getFieldWithDefault(msg, 9, "")
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
 * @return {!proto.pps_v2.LogMessage}
 */
proto.pps_v2.LogMessage.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.LogMessage;
  return proto.pps_v2.LogMessage.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.LogMessage} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.LogMessage}
 */
proto.pps_v2.LogMessage.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setPipelineName(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setJobId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setWorkerId(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setDatumId(value);
      break;
    case 5:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setMaster(value);
      break;
    case 6:
      var value = new proto.pps_v2.InputFile;
      reader.readMessage(value,proto.pps_v2.InputFile.deserializeBinaryFromReader);
      msg.addData(value);
      break;
    case 7:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setUser(value);
      break;
    case 8:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setTs(value);
      break;
    case 9:
      var value = /** @type {string} */ (reader.readString());
      msg.setMessage(value);
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
proto.pps_v2.LogMessage.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.LogMessage.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.LogMessage} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.LogMessage.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipelineName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getJobId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getWorkerId();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getDatumId();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getMaster();
  if (f) {
    writer.writeBool(
      5,
      f
    );
  }
  f = message.getDataList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      6,
      f,
      proto.pps_v2.InputFile.serializeBinaryToWriter
    );
  }
  f = message.getUser();
  if (f) {
    writer.writeBool(
      7,
      f
    );
  }
  f = message.getTs();
  if (f != null) {
    writer.writeMessage(
      8,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getMessage();
  if (f.length > 0) {
    writer.writeString(
      9,
      f
    );
  }
};


/**
 * optional string pipeline_name = 1;
 * @return {string}
 */
proto.pps_v2.LogMessage.prototype.getPipelineName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.LogMessage} returns this
 */
proto.pps_v2.LogMessage.prototype.setPipelineName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string job_id = 2;
 * @return {string}
 */
proto.pps_v2.LogMessage.prototype.getJobId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.LogMessage} returns this
 */
proto.pps_v2.LogMessage.prototype.setJobId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string worker_id = 3;
 * @return {string}
 */
proto.pps_v2.LogMessage.prototype.getWorkerId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.LogMessage} returns this
 */
proto.pps_v2.LogMessage.prototype.setWorkerId = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string datum_id = 4;
 * @return {string}
 */
proto.pps_v2.LogMessage.prototype.getDatumId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.LogMessage} returns this
 */
proto.pps_v2.LogMessage.prototype.setDatumId = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional bool master = 5;
 * @return {boolean}
 */
proto.pps_v2.LogMessage.prototype.getMaster = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 5, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.LogMessage} returns this
 */
proto.pps_v2.LogMessage.prototype.setMaster = function(value) {
  return jspb.Message.setProto3BooleanField(this, 5, value);
};


/**
 * repeated InputFile data = 6;
 * @return {!Array<!proto.pps_v2.InputFile>}
 */
proto.pps_v2.LogMessage.prototype.getDataList = function() {
  return /** @type{!Array<!proto.pps_v2.InputFile>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps_v2.InputFile, 6));
};


/**
 * @param {!Array<!proto.pps_v2.InputFile>} value
 * @return {!proto.pps_v2.LogMessage} returns this
*/
proto.pps_v2.LogMessage.prototype.setDataList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 6, value);
};


/**
 * @param {!proto.pps_v2.InputFile=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.InputFile}
 */
proto.pps_v2.LogMessage.prototype.addData = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 6, opt_value, proto.pps_v2.InputFile, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.LogMessage} returns this
 */
proto.pps_v2.LogMessage.prototype.clearDataList = function() {
  return this.setDataList([]);
};


/**
 * optional bool user = 7;
 * @return {boolean}
 */
proto.pps_v2.LogMessage.prototype.getUser = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 7, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.LogMessage} returns this
 */
proto.pps_v2.LogMessage.prototype.setUser = function(value) {
  return jspb.Message.setProto3BooleanField(this, 7, value);
};


/**
 * optional google.protobuf.Timestamp ts = 8;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps_v2.LogMessage.prototype.getTs = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 8));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps_v2.LogMessage} returns this
*/
proto.pps_v2.LogMessage.prototype.setTs = function(value) {
  return jspb.Message.setWrapperField(this, 8, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.LogMessage} returns this
 */
proto.pps_v2.LogMessage.prototype.clearTs = function() {
  return this.setTs(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.LogMessage.prototype.hasTs = function() {
  return jspb.Message.getField(this, 8) != null;
};


/**
 * optional string message = 9;
 * @return {string}
 */
proto.pps_v2.LogMessage.prototype.getMessage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 9, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.LogMessage} returns this
 */
proto.pps_v2.LogMessage.prototype.setMessage = function(value) {
  return jspb.Message.setProto3StringField(this, 9, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.RestartDatumRequest.repeatedFields_ = [2];



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
proto.pps_v2.RestartDatumRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.RestartDatumRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.RestartDatumRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.RestartDatumRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps_v2.Job.toObject(includeInstance, f),
    dataFiltersList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f
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
 * @return {!proto.pps_v2.RestartDatumRequest}
 */
proto.pps_v2.RestartDatumRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.RestartDatumRequest;
  return proto.pps_v2.RestartDatumRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.RestartDatumRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.RestartDatumRequest}
 */
proto.pps_v2.RestartDatumRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Job;
      reader.readMessage(value,proto.pps_v2.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.addDataFilters(value);
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
proto.pps_v2.RestartDatumRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.RestartDatumRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.RestartDatumRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.RestartDatumRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Job.serializeBinaryToWriter
    );
  }
  f = message.getDataFiltersList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      2,
      f
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps_v2.Job}
 */
proto.pps_v2.RestartDatumRequest.prototype.getJob = function() {
  return /** @type{?proto.pps_v2.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Job, 1));
};


/**
 * @param {?proto.pps_v2.Job|undefined} value
 * @return {!proto.pps_v2.RestartDatumRequest} returns this
*/
proto.pps_v2.RestartDatumRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.RestartDatumRequest} returns this
 */
proto.pps_v2.RestartDatumRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.RestartDatumRequest.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * repeated string data_filters = 2;
 * @return {!Array<string>}
 */
proto.pps_v2.RestartDatumRequest.prototype.getDataFiltersList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps_v2.RestartDatumRequest} returns this
 */
proto.pps_v2.RestartDatumRequest.prototype.setDataFiltersList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.RestartDatumRequest} returns this
 */
proto.pps_v2.RestartDatumRequest.prototype.addDataFilters = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.RestartDatumRequest} returns this
 */
proto.pps_v2.RestartDatumRequest.prototype.clearDataFiltersList = function() {
  return this.setDataFiltersList([]);
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
proto.pps_v2.InspectDatumRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.InspectDatumRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.InspectDatumRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InspectDatumRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    datum: (f = msg.getDatum()) && proto.pps_v2.Datum.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.InspectDatumRequest}
 */
proto.pps_v2.InspectDatumRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.InspectDatumRequest;
  return proto.pps_v2.InspectDatumRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.InspectDatumRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.InspectDatumRequest}
 */
proto.pps_v2.InspectDatumRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Datum;
      reader.readMessage(value,proto.pps_v2.Datum.deserializeBinaryFromReader);
      msg.setDatum(value);
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
proto.pps_v2.InspectDatumRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.InspectDatumRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.InspectDatumRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InspectDatumRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getDatum();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Datum.serializeBinaryToWriter
    );
  }
};


/**
 * optional Datum datum = 1;
 * @return {?proto.pps_v2.Datum}
 */
proto.pps_v2.InspectDatumRequest.prototype.getDatum = function() {
  return /** @type{?proto.pps_v2.Datum} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Datum, 1));
};


/**
 * @param {?proto.pps_v2.Datum|undefined} value
 * @return {!proto.pps_v2.InspectDatumRequest} returns this
*/
proto.pps_v2.InspectDatumRequest.prototype.setDatum = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.InspectDatumRequest} returns this
 */
proto.pps_v2.InspectDatumRequest.prototype.clearDatum = function() {
  return this.setDatum(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.InspectDatumRequest.prototype.hasDatum = function() {
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
proto.pps_v2.ListDatumRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.ListDatumRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.ListDatumRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ListDatumRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps_v2.Job.toObject(includeInstance, f),
    input: (f = msg.getInput()) && proto.pps_v2.Input.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.ListDatumRequest}
 */
proto.pps_v2.ListDatumRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.ListDatumRequest;
  return proto.pps_v2.ListDatumRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.ListDatumRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.ListDatumRequest}
 */
proto.pps_v2.ListDatumRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Job;
      reader.readMessage(value,proto.pps_v2.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 2:
      var value = new proto.pps_v2.Input;
      reader.readMessage(value,proto.pps_v2.Input.deserializeBinaryFromReader);
      msg.setInput(value);
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
proto.pps_v2.ListDatumRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.ListDatumRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.ListDatumRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ListDatumRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Job.serializeBinaryToWriter
    );
  }
  f = message.getInput();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps_v2.Input.serializeBinaryToWriter
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps_v2.Job}
 */
proto.pps_v2.ListDatumRequest.prototype.getJob = function() {
  return /** @type{?proto.pps_v2.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Job, 1));
};


/**
 * @param {?proto.pps_v2.Job|undefined} value
 * @return {!proto.pps_v2.ListDatumRequest} returns this
*/
proto.pps_v2.ListDatumRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.ListDatumRequest} returns this
 */
proto.pps_v2.ListDatumRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.ListDatumRequest.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Input input = 2;
 * @return {?proto.pps_v2.Input}
 */
proto.pps_v2.ListDatumRequest.prototype.getInput = function() {
  return /** @type{?proto.pps_v2.Input} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Input, 2));
};


/**
 * @param {?proto.pps_v2.Input|undefined} value
 * @return {!proto.pps_v2.ListDatumRequest} returns this
*/
proto.pps_v2.ListDatumRequest.prototype.setInput = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.ListDatumRequest} returns this
 */
proto.pps_v2.ListDatumRequest.prototype.clearInput = function() {
  return this.setInput(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.ListDatumRequest.prototype.hasInput = function() {
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
proto.pps_v2.DatumSetSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.DatumSetSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.DatumSetSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DatumSetSpec.toObject = function(includeInstance, msg) {
  var f, obj = {
    number: jspb.Message.getFieldWithDefault(msg, 1, 0),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 2, 0),
    perWorker: jspb.Message.getFieldWithDefault(msg, 3, 0)
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
 * @return {!proto.pps_v2.DatumSetSpec}
 */
proto.pps_v2.DatumSetSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.DatumSetSpec;
  return proto.pps_v2.DatumSetSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.DatumSetSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.DatumSetSpec}
 */
proto.pps_v2.DatumSetSpec.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setNumber(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setSizeBytes(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setPerWorker(value);
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
proto.pps_v2.DatumSetSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.DatumSetSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.DatumSetSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DatumSetSpec.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getNumber();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getSizeBytes();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
  f = message.getPerWorker();
  if (f !== 0) {
    writer.writeInt64(
      3,
      f
    );
  }
};


/**
 * optional int64 number = 1;
 * @return {number}
 */
proto.pps_v2.DatumSetSpec.prototype.getNumber = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.DatumSetSpec} returns this
 */
proto.pps_v2.DatumSetSpec.prototype.setNumber = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional int64 size_bytes = 2;
 * @return {number}
 */
proto.pps_v2.DatumSetSpec.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.DatumSetSpec} returns this
 */
proto.pps_v2.DatumSetSpec.prototype.setSizeBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional int64 per_worker = 3;
 * @return {number}
 */
proto.pps_v2.DatumSetSpec.prototype.getPerWorker = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.DatumSetSpec} returns this
 */
proto.pps_v2.DatumSetSpec.prototype.setPerWorker = function(value) {
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
proto.pps_v2.SchedulingSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.SchedulingSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.SchedulingSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.SchedulingSpec.toObject = function(includeInstance, msg) {
  var f, obj = {
    nodeSelectorMap: (f = msg.getNodeSelectorMap()) ? f.toObject(includeInstance, undefined) : [],
    priorityClassName: jspb.Message.getFieldWithDefault(msg, 2, "")
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
 * @return {!proto.pps_v2.SchedulingSpec}
 */
proto.pps_v2.SchedulingSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.SchedulingSpec;
  return proto.pps_v2.SchedulingSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.SchedulingSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.SchedulingSpec}
 */
proto.pps_v2.SchedulingSpec.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getNodeSelectorMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
         });
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setPriorityClassName(value);
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
proto.pps_v2.SchedulingSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.SchedulingSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.SchedulingSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.SchedulingSpec.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getNodeSelectorMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getPriorityClassName();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * map<string, string> node_selector = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.pps_v2.SchedulingSpec.prototype.getNodeSelectorMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pps_v2.SchedulingSpec} returns this
 */
proto.pps_v2.SchedulingSpec.prototype.clearNodeSelectorMap = function() {
  this.getNodeSelectorMap().clear();
  return this;};


/**
 * optional string priority_class_name = 2;
 * @return {string}
 */
proto.pps_v2.SchedulingSpec.prototype.getPriorityClassName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.SchedulingSpec} returns this
 */
proto.pps_v2.SchedulingSpec.prototype.setPriorityClassName = function(value) {
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
proto.pps_v2.CreatePipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.CreatePipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.CreatePipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.CreatePipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f),
    tfJob: (f = msg.getTfJob()) && proto.pps_v2.TFJob.toObject(includeInstance, f),
    transform: (f = msg.getTransform()) && proto.pps_v2.Transform.toObject(includeInstance, f),
    parallelismSpec: (f = msg.getParallelismSpec()) && proto.pps_v2.ParallelismSpec.toObject(includeInstance, f),
    egress: (f = msg.getEgress()) && proto.pps_v2.Egress.toObject(includeInstance, f),
    update: jspb.Message.getBooleanFieldWithDefault(msg, 6, false),
    outputBranch: jspb.Message.getFieldWithDefault(msg, 7, ""),
    s3Out: jspb.Message.getBooleanFieldWithDefault(msg, 8, false),
    resourceRequests: (f = msg.getResourceRequests()) && proto.pps_v2.ResourceSpec.toObject(includeInstance, f),
    resourceLimits: (f = msg.getResourceLimits()) && proto.pps_v2.ResourceSpec.toObject(includeInstance, f),
    sidecarResourceLimits: (f = msg.getSidecarResourceLimits()) && proto.pps_v2.ResourceSpec.toObject(includeInstance, f),
    input: (f = msg.getInput()) && proto.pps_v2.Input.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 13, ""),
    cacheSize: jspb.Message.getFieldWithDefault(msg, 14, ""),
    reprocess: jspb.Message.getBooleanFieldWithDefault(msg, 16, false),
    maxQueueSize: jspb.Message.getFieldWithDefault(msg, 17, 0),
    service: (f = msg.getService()) && proto.pps_v2.Service.toObject(includeInstance, f),
    spout: (f = msg.getSpout()) && proto.pps_v2.Spout.toObject(includeInstance, f),
    datumSetSpec: (f = msg.getDatumSetSpec()) && proto.pps_v2.DatumSetSpec.toObject(includeInstance, f),
    datumTimeout: (f = msg.getDatumTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    jobTimeout: (f = msg.getJobTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    salt: jspb.Message.getFieldWithDefault(msg, 23, ""),
    standby: jspb.Message.getBooleanFieldWithDefault(msg, 24, false),
    datumTries: jspb.Message.getFieldWithDefault(msg, 25, 0),
    schedulingSpec: (f = msg.getSchedulingSpec()) && proto.pps_v2.SchedulingSpec.toObject(includeInstance, f),
    podSpec: jspb.Message.getFieldWithDefault(msg, 27, ""),
    podPatch: jspb.Message.getFieldWithDefault(msg, 28, ""),
    specCommit: (f = msg.getSpecCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    metadata: (f = msg.getMetadata()) && proto.pps_v2.Metadata.toObject(includeInstance, f),
    reprocessSpec: jspb.Message.getFieldWithDefault(msg, 31, ""),
    autoscaling: jspb.Message.getBooleanFieldWithDefault(msg, 32, false)
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
 * @return {!proto.pps_v2.CreatePipelineRequest}
 */
proto.pps_v2.CreatePipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.CreatePipelineRequest;
  return proto.pps_v2.CreatePipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.CreatePipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.CreatePipelineRequest}
 */
proto.pps_v2.CreatePipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = new proto.pps_v2.TFJob;
      reader.readMessage(value,proto.pps_v2.TFJob.deserializeBinaryFromReader);
      msg.setTfJob(value);
      break;
    case 3:
      var value = new proto.pps_v2.Transform;
      reader.readMessage(value,proto.pps_v2.Transform.deserializeBinaryFromReader);
      msg.setTransform(value);
      break;
    case 4:
      var value = new proto.pps_v2.ParallelismSpec;
      reader.readMessage(value,proto.pps_v2.ParallelismSpec.deserializeBinaryFromReader);
      msg.setParallelismSpec(value);
      break;
    case 5:
      var value = new proto.pps_v2.Egress;
      reader.readMessage(value,proto.pps_v2.Egress.deserializeBinaryFromReader);
      msg.setEgress(value);
      break;
    case 6:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setUpdate(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setOutputBranch(value);
      break;
    case 8:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setS3Out(value);
      break;
    case 9:
      var value = new proto.pps_v2.ResourceSpec;
      reader.readMessage(value,proto.pps_v2.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceRequests(value);
      break;
    case 10:
      var value = new proto.pps_v2.ResourceSpec;
      reader.readMessage(value,proto.pps_v2.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceLimits(value);
      break;
    case 11:
      var value = new proto.pps_v2.ResourceSpec;
      reader.readMessage(value,proto.pps_v2.ResourceSpec.deserializeBinaryFromReader);
      msg.setSidecarResourceLimits(value);
      break;
    case 12:
      var value = new proto.pps_v2.Input;
      reader.readMessage(value,proto.pps_v2.Input.deserializeBinaryFromReader);
      msg.setInput(value);
      break;
    case 13:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 14:
      var value = /** @type {string} */ (reader.readString());
      msg.setCacheSize(value);
      break;
    case 16:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setReprocess(value);
      break;
    case 17:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setMaxQueueSize(value);
      break;
    case 18:
      var value = new proto.pps_v2.Service;
      reader.readMessage(value,proto.pps_v2.Service.deserializeBinaryFromReader);
      msg.setService(value);
      break;
    case 19:
      var value = new proto.pps_v2.Spout;
      reader.readMessage(value,proto.pps_v2.Spout.deserializeBinaryFromReader);
      msg.setSpout(value);
      break;
    case 20:
      var value = new proto.pps_v2.DatumSetSpec;
      reader.readMessage(value,proto.pps_v2.DatumSetSpec.deserializeBinaryFromReader);
      msg.setDatumSetSpec(value);
      break;
    case 21:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setDatumTimeout(value);
      break;
    case 22:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setJobTimeout(value);
      break;
    case 23:
      var value = /** @type {string} */ (reader.readString());
      msg.setSalt(value);
      break;
    case 24:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setStandby(value);
      break;
    case 25:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDatumTries(value);
      break;
    case 26:
      var value = new proto.pps_v2.SchedulingSpec;
      reader.readMessage(value,proto.pps_v2.SchedulingSpec.deserializeBinaryFromReader);
      msg.setSchedulingSpec(value);
      break;
    case 27:
      var value = /** @type {string} */ (reader.readString());
      msg.setPodSpec(value);
      break;
    case 28:
      var value = /** @type {string} */ (reader.readString());
      msg.setPodPatch(value);
      break;
    case 29:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setSpecCommit(value);
      break;
    case 30:
      var value = new proto.pps_v2.Metadata;
      reader.readMessage(value,proto.pps_v2.Metadata.deserializeBinaryFromReader);
      msg.setMetadata(value);
      break;
    case 31:
      var value = /** @type {string} */ (reader.readString());
      msg.setReprocessSpec(value);
      break;
    case 32:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAutoscaling(value);
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
proto.pps_v2.CreatePipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.CreatePipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.CreatePipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.CreatePipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getTfJob();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps_v2.TFJob.serializeBinaryToWriter
    );
  }
  f = message.getTransform();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pps_v2.Transform.serializeBinaryToWriter
    );
  }
  f = message.getParallelismSpec();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pps_v2.ParallelismSpec.serializeBinaryToWriter
    );
  }
  f = message.getEgress();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.pps_v2.Egress.serializeBinaryToWriter
    );
  }
  f = message.getUpdate();
  if (f) {
    writer.writeBool(
      6,
      f
    );
  }
  f = message.getOutputBranch();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
  f = message.getS3Out();
  if (f) {
    writer.writeBool(
      8,
      f
    );
  }
  f = message.getResourceRequests();
  if (f != null) {
    writer.writeMessage(
      9,
      f,
      proto.pps_v2.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getResourceLimits();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      proto.pps_v2.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getSidecarResourceLimits();
  if (f != null) {
    writer.writeMessage(
      11,
      f,
      proto.pps_v2.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getInput();
  if (f != null) {
    writer.writeMessage(
      12,
      f,
      proto.pps_v2.Input.serializeBinaryToWriter
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      13,
      f
    );
  }
  f = message.getCacheSize();
  if (f.length > 0) {
    writer.writeString(
      14,
      f
    );
  }
  f = message.getReprocess();
  if (f) {
    writer.writeBool(
      16,
      f
    );
  }
  f = message.getMaxQueueSize();
  if (f !== 0) {
    writer.writeInt64(
      17,
      f
    );
  }
  f = message.getService();
  if (f != null) {
    writer.writeMessage(
      18,
      f,
      proto.pps_v2.Service.serializeBinaryToWriter
    );
  }
  f = message.getSpout();
  if (f != null) {
    writer.writeMessage(
      19,
      f,
      proto.pps_v2.Spout.serializeBinaryToWriter
    );
  }
  f = message.getDatumSetSpec();
  if (f != null) {
    writer.writeMessage(
      20,
      f,
      proto.pps_v2.DatumSetSpec.serializeBinaryToWriter
    );
  }
  f = message.getDatumTimeout();
  if (f != null) {
    writer.writeMessage(
      21,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getJobTimeout();
  if (f != null) {
    writer.writeMessage(
      22,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getSalt();
  if (f.length > 0) {
    writer.writeString(
      23,
      f
    );
  }
  f = message.getStandby();
  if (f) {
    writer.writeBool(
      24,
      f
    );
  }
  f = message.getDatumTries();
  if (f !== 0) {
    writer.writeInt64(
      25,
      f
    );
  }
  f = message.getSchedulingSpec();
  if (f != null) {
    writer.writeMessage(
      26,
      f,
      proto.pps_v2.SchedulingSpec.serializeBinaryToWriter
    );
  }
  f = message.getPodSpec();
  if (f.length > 0) {
    writer.writeString(
      27,
      f
    );
  }
  f = message.getPodPatch();
  if (f.length > 0) {
    writer.writeString(
      28,
      f
    );
  }
  f = message.getSpecCommit();
  if (f != null) {
    writer.writeMessage(
      29,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getMetadata();
  if (f != null) {
    writer.writeMessage(
      30,
      f,
      proto.pps_v2.Metadata.serializeBinaryToWriter
    );
  }
  f = message.getReprocessSpec();
  if (f.length > 0) {
    writer.writeString(
      31,
      f
    );
  }
  f = message.getAutoscaling();
  if (f) {
    writer.writeBool(
      32,
      f
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional TFJob tf_job = 2;
 * @return {?proto.pps_v2.TFJob}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getTfJob = function() {
  return /** @type{?proto.pps_v2.TFJob} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.TFJob, 2));
};


/**
 * @param {?proto.pps_v2.TFJob|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setTfJob = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearTfJob = function() {
  return this.setTfJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasTfJob = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional Transform transform = 3;
 * @return {?proto.pps_v2.Transform}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getTransform = function() {
  return /** @type{?proto.pps_v2.Transform} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Transform, 3));
};


/**
 * @param {?proto.pps_v2.Transform|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setTransform = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearTransform = function() {
  return this.setTransform(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasTransform = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional ParallelismSpec parallelism_spec = 4;
 * @return {?proto.pps_v2.ParallelismSpec}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getParallelismSpec = function() {
  return /** @type{?proto.pps_v2.ParallelismSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ParallelismSpec, 4));
};


/**
 * @param {?proto.pps_v2.ParallelismSpec|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setParallelismSpec = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearParallelismSpec = function() {
  return this.setParallelismSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasParallelismSpec = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional Egress egress = 5;
 * @return {?proto.pps_v2.Egress}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getEgress = function() {
  return /** @type{?proto.pps_v2.Egress} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Egress, 5));
};


/**
 * @param {?proto.pps_v2.Egress|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setEgress = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearEgress = function() {
  return this.setEgress(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasEgress = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional bool update = 6;
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getUpdate = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 6, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setUpdate = function(value) {
  return jspb.Message.setProto3BooleanField(this, 6, value);
};


/**
 * optional string output_branch = 7;
 * @return {string}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getOutputBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setOutputBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};


/**
 * optional bool s3_out = 8;
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getS3Out = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 8, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setS3Out = function(value) {
  return jspb.Message.setProto3BooleanField(this, 8, value);
};


/**
 * optional ResourceSpec resource_requests = 9;
 * @return {?proto.pps_v2.ResourceSpec}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getResourceRequests = function() {
  return /** @type{?proto.pps_v2.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ResourceSpec, 9));
};


/**
 * @param {?proto.pps_v2.ResourceSpec|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setResourceRequests = function(value) {
  return jspb.Message.setWrapperField(this, 9, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearResourceRequests = function() {
  return this.setResourceRequests(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasResourceRequests = function() {
  return jspb.Message.getField(this, 9) != null;
};


/**
 * optional ResourceSpec resource_limits = 10;
 * @return {?proto.pps_v2.ResourceSpec}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getResourceLimits = function() {
  return /** @type{?proto.pps_v2.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ResourceSpec, 10));
};


/**
 * @param {?proto.pps_v2.ResourceSpec|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearResourceLimits = function() {
  return this.setResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasResourceLimits = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * optional ResourceSpec sidecar_resource_limits = 11;
 * @return {?proto.pps_v2.ResourceSpec}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getSidecarResourceLimits = function() {
  return /** @type{?proto.pps_v2.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.ResourceSpec, 11));
};


/**
 * @param {?proto.pps_v2.ResourceSpec|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setSidecarResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 11, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearSidecarResourceLimits = function() {
  return this.setSidecarResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasSidecarResourceLimits = function() {
  return jspb.Message.getField(this, 11) != null;
};


/**
 * optional Input input = 12;
 * @return {?proto.pps_v2.Input}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getInput = function() {
  return /** @type{?proto.pps_v2.Input} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Input, 12));
};


/**
 * @param {?proto.pps_v2.Input|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setInput = function(value) {
  return jspb.Message.setWrapperField(this, 12, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearInput = function() {
  return this.setInput(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasInput = function() {
  return jspb.Message.getField(this, 12) != null;
};


/**
 * optional string description = 13;
 * @return {string}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 13, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 13, value);
};


/**
 * optional string cache_size = 14;
 * @return {string}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getCacheSize = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 14, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setCacheSize = function(value) {
  return jspb.Message.setProto3StringField(this, 14, value);
};


/**
 * optional bool reprocess = 16;
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getReprocess = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 16, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setReprocess = function(value) {
  return jspb.Message.setProto3BooleanField(this, 16, value);
};


/**
 * optional int64 max_queue_size = 17;
 * @return {number}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getMaxQueueSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 17, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setMaxQueueSize = function(value) {
  return jspb.Message.setProto3IntField(this, 17, value);
};


/**
 * optional Service service = 18;
 * @return {?proto.pps_v2.Service}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getService = function() {
  return /** @type{?proto.pps_v2.Service} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Service, 18));
};


/**
 * @param {?proto.pps_v2.Service|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setService = function(value) {
  return jspb.Message.setWrapperField(this, 18, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearService = function() {
  return this.setService(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasService = function() {
  return jspb.Message.getField(this, 18) != null;
};


/**
 * optional Spout spout = 19;
 * @return {?proto.pps_v2.Spout}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getSpout = function() {
  return /** @type{?proto.pps_v2.Spout} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Spout, 19));
};


/**
 * @param {?proto.pps_v2.Spout|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setSpout = function(value) {
  return jspb.Message.setWrapperField(this, 19, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearSpout = function() {
  return this.setSpout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasSpout = function() {
  return jspb.Message.getField(this, 19) != null;
};


/**
 * optional DatumSetSpec datum_set_spec = 20;
 * @return {?proto.pps_v2.DatumSetSpec}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getDatumSetSpec = function() {
  return /** @type{?proto.pps_v2.DatumSetSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.DatumSetSpec, 20));
};


/**
 * @param {?proto.pps_v2.DatumSetSpec|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setDatumSetSpec = function(value) {
  return jspb.Message.setWrapperField(this, 20, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearDatumSetSpec = function() {
  return this.setDatumSetSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasDatumSetSpec = function() {
  return jspb.Message.getField(this, 20) != null;
};


/**
 * optional google.protobuf.Duration datum_timeout = 21;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getDatumTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 21));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setDatumTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 21, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearDatumTimeout = function() {
  return this.setDatumTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasDatumTimeout = function() {
  return jspb.Message.getField(this, 21) != null;
};


/**
 * optional google.protobuf.Duration job_timeout = 22;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getJobTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 22));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setJobTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 22, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearJobTimeout = function() {
  return this.setJobTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasJobTimeout = function() {
  return jspb.Message.getField(this, 22) != null;
};


/**
 * optional string salt = 23;
 * @return {string}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getSalt = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 23, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setSalt = function(value) {
  return jspb.Message.setProto3StringField(this, 23, value);
};


/**
 * optional bool standby = 24;
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getStandby = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 24, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setStandby = function(value) {
  return jspb.Message.setProto3BooleanField(this, 24, value);
};


/**
 * optional int64 datum_tries = 25;
 * @return {number}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getDatumTries = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 25, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setDatumTries = function(value) {
  return jspb.Message.setProto3IntField(this, 25, value);
};


/**
 * optional SchedulingSpec scheduling_spec = 26;
 * @return {?proto.pps_v2.SchedulingSpec}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getSchedulingSpec = function() {
  return /** @type{?proto.pps_v2.SchedulingSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.SchedulingSpec, 26));
};


/**
 * @param {?proto.pps_v2.SchedulingSpec|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setSchedulingSpec = function(value) {
  return jspb.Message.setWrapperField(this, 26, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearSchedulingSpec = function() {
  return this.setSchedulingSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasSchedulingSpec = function() {
  return jspb.Message.getField(this, 26) != null;
};


/**
 * optional string pod_spec = 27;
 * @return {string}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getPodSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 27, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setPodSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 27, value);
};


/**
 * optional string pod_patch = 28;
 * @return {string}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getPodPatch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 28, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setPodPatch = function(value) {
  return jspb.Message.setProto3StringField(this, 28, value);
};


/**
 * optional pfs_v2.Commit spec_commit = 29;
 * @return {?proto.pfs_v2.Commit}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getSpecCommit = function() {
  return /** @type{?proto.pfs_v2.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 29));
};


/**
 * @param {?proto.pfs_v2.Commit|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setSpecCommit = function(value) {
  return jspb.Message.setWrapperField(this, 29, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearSpecCommit = function() {
  return this.setSpecCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasSpecCommit = function() {
  return jspb.Message.getField(this, 29) != null;
};


/**
 * optional Metadata metadata = 30;
 * @return {?proto.pps_v2.Metadata}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getMetadata = function() {
  return /** @type{?proto.pps_v2.Metadata} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Metadata, 30));
};


/**
 * @param {?proto.pps_v2.Metadata|undefined} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
*/
proto.pps_v2.CreatePipelineRequest.prototype.setMetadata = function(value) {
  return jspb.Message.setWrapperField(this, 30, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.clearMetadata = function() {
  return this.setMetadata(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.hasMetadata = function() {
  return jspb.Message.getField(this, 30) != null;
};


/**
 * optional string reprocess_spec = 31;
 * @return {string}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getReprocessSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 31, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setReprocessSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 31, value);
};


/**
 * optional bool autoscaling = 32;
 * @return {boolean}
 */
proto.pps_v2.CreatePipelineRequest.prototype.getAutoscaling = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 32, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.CreatePipelineRequest} returns this
 */
proto.pps_v2.CreatePipelineRequest.prototype.setAutoscaling = function(value) {
  return jspb.Message.setProto3BooleanField(this, 32, value);
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
proto.pps_v2.InspectPipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.InspectPipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.InspectPipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InspectPipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f),
    details: jspb.Message.getBooleanFieldWithDefault(msg, 2, false)
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
 * @return {!proto.pps_v2.InspectPipelineRequest}
 */
proto.pps_v2.InspectPipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.InspectPipelineRequest;
  return proto.pps_v2.InspectPipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.InspectPipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.InspectPipelineRequest}
 */
proto.pps_v2.InspectPipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDetails(value);
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
proto.pps_v2.InspectPipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.InspectPipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.InspectPipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InspectPipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getDetails();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.InspectPipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.InspectPipelineRequest} returns this
*/
proto.pps_v2.InspectPipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.InspectPipelineRequest} returns this
 */
proto.pps_v2.InspectPipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.InspectPipelineRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool details = 2;
 * @return {boolean}
 */
proto.pps_v2.InspectPipelineRequest.prototype.getDetails = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.InspectPipelineRequest} returns this
 */
proto.pps_v2.InspectPipelineRequest.prototype.setDetails = function(value) {
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
proto.pps_v2.ListPipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.ListPipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.ListPipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ListPipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f),
    history: jspb.Message.getFieldWithDefault(msg, 2, 0),
    details: jspb.Message.getBooleanFieldWithDefault(msg, 3, false),
    jqfilter: jspb.Message.getFieldWithDefault(msg, 4, "")
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
 * @return {!proto.pps_v2.ListPipelineRequest}
 */
proto.pps_v2.ListPipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.ListPipelineRequest;
  return proto.pps_v2.ListPipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.ListPipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.ListPipelineRequest}
 */
proto.pps_v2.ListPipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setHistory(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDetails(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setJqfilter(value);
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
proto.pps_v2.ListPipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.ListPipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.ListPipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ListPipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getHistory();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
  f = message.getDetails();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
  f = message.getJqfilter();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.ListPipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.ListPipelineRequest} returns this
*/
proto.pps_v2.ListPipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.ListPipelineRequest} returns this
 */
proto.pps_v2.ListPipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.ListPipelineRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional int64 history = 2;
 * @return {number}
 */
proto.pps_v2.ListPipelineRequest.prototype.getHistory = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps_v2.ListPipelineRequest} returns this
 */
proto.pps_v2.ListPipelineRequest.prototype.setHistory = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional bool details = 3;
 * @return {boolean}
 */
proto.pps_v2.ListPipelineRequest.prototype.getDetails = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.ListPipelineRequest} returns this
 */
proto.pps_v2.ListPipelineRequest.prototype.setDetails = function(value) {
  return jspb.Message.setProto3BooleanField(this, 3, value);
};


/**
 * optional string jqFilter = 4;
 * @return {string}
 */
proto.pps_v2.ListPipelineRequest.prototype.getJqfilter = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.ListPipelineRequest} returns this
 */
proto.pps_v2.ListPipelineRequest.prototype.setJqfilter = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
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
proto.pps_v2.DeletePipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.DeletePipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.DeletePipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DeletePipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f),
    all: jspb.Message.getBooleanFieldWithDefault(msg, 2, false),
    force: jspb.Message.getBooleanFieldWithDefault(msg, 3, false),
    keepRepo: jspb.Message.getBooleanFieldWithDefault(msg, 4, false)
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
 * @return {!proto.pps_v2.DeletePipelineRequest}
 */
proto.pps_v2.DeletePipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.DeletePipelineRequest;
  return proto.pps_v2.DeletePipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.DeletePipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.DeletePipelineRequest}
 */
proto.pps_v2.DeletePipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAll(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setForce(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setKeepRepo(value);
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
proto.pps_v2.DeletePipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.DeletePipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.DeletePipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DeletePipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getAll();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
  f = message.getForce();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
  f = message.getKeepRepo();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.DeletePipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.DeletePipelineRequest} returns this
*/
proto.pps_v2.DeletePipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.DeletePipelineRequest} returns this
 */
proto.pps_v2.DeletePipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.DeletePipelineRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool all = 2;
 * @return {boolean}
 */
proto.pps_v2.DeletePipelineRequest.prototype.getAll = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.DeletePipelineRequest} returns this
 */
proto.pps_v2.DeletePipelineRequest.prototype.setAll = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};


/**
 * optional bool force = 3;
 * @return {boolean}
 */
proto.pps_v2.DeletePipelineRequest.prototype.getForce = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.DeletePipelineRequest} returns this
 */
proto.pps_v2.DeletePipelineRequest.prototype.setForce = function(value) {
  return jspb.Message.setProto3BooleanField(this, 3, value);
};


/**
 * optional bool keep_repo = 4;
 * @return {boolean}
 */
proto.pps_v2.DeletePipelineRequest.prototype.getKeepRepo = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps_v2.DeletePipelineRequest} returns this
 */
proto.pps_v2.DeletePipelineRequest.prototype.setKeepRepo = function(value) {
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
proto.pps_v2.StartPipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.StartPipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.StartPipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.StartPipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.StartPipelineRequest}
 */
proto.pps_v2.StartPipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.StartPipelineRequest;
  return proto.pps_v2.StartPipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.StartPipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.StartPipelineRequest}
 */
proto.pps_v2.StartPipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
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
proto.pps_v2.StartPipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.StartPipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.StartPipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.StartPipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.StartPipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.StartPipelineRequest} returns this
*/
proto.pps_v2.StartPipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.StartPipelineRequest} returns this
 */
proto.pps_v2.StartPipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.StartPipelineRequest.prototype.hasPipeline = function() {
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
proto.pps_v2.StopPipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.StopPipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.StopPipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.StopPipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.StopPipelineRequest}
 */
proto.pps_v2.StopPipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.StopPipelineRequest;
  return proto.pps_v2.StopPipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.StopPipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.StopPipelineRequest}
 */
proto.pps_v2.StopPipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
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
proto.pps_v2.StopPipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.StopPipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.StopPipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.StopPipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.StopPipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.StopPipelineRequest} returns this
*/
proto.pps_v2.StopPipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.StopPipelineRequest} returns this
 */
proto.pps_v2.StopPipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.StopPipelineRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.RunPipelineRequest.repeatedFields_ = [2];



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
proto.pps_v2.RunPipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.RunPipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.RunPipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.RunPipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f),
    provenanceList: jspb.Message.toObjectList(msg.getProvenanceList(),
    pfs_pfs_pb.Commit.toObject, includeInstance),
    jobId: jspb.Message.getFieldWithDefault(msg, 3, "")
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
 * @return {!proto.pps_v2.RunPipelineRequest}
 */
proto.pps_v2.RunPipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.RunPipelineRequest;
  return proto.pps_v2.RunPipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.RunPipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.RunPipelineRequest}
 */
proto.pps_v2.RunPipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.addProvenance(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setJobId(value);
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
proto.pps_v2.RunPipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.RunPipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.RunPipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.RunPipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getJobId();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.RunPipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.RunPipelineRequest} returns this
*/
proto.pps_v2.RunPipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.RunPipelineRequest} returns this
 */
proto.pps_v2.RunPipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.RunPipelineRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * repeated pfs_v2.Commit provenance = 2;
 * @return {!Array<!proto.pfs_v2.Commit>}
 */
proto.pps_v2.RunPipelineRequest.prototype.getProvenanceList = function() {
  return /** @type{!Array<!proto.pfs_v2.Commit>} */ (
    jspb.Message.getRepeatedWrapperField(this, pfs_pfs_pb.Commit, 2));
};


/**
 * @param {!Array<!proto.pfs_v2.Commit>} value
 * @return {!proto.pps_v2.RunPipelineRequest} returns this
*/
proto.pps_v2.RunPipelineRequest.prototype.setProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pfs_v2.Commit=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs_v2.Commit}
 */
proto.pps_v2.RunPipelineRequest.prototype.addProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pfs_v2.Commit, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.RunPipelineRequest} returns this
 */
proto.pps_v2.RunPipelineRequest.prototype.clearProvenanceList = function() {
  return this.setProvenanceList([]);
};


/**
 * optional string job_id = 3;
 * @return {string}
 */
proto.pps_v2.RunPipelineRequest.prototype.getJobId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.RunPipelineRequest} returns this
 */
proto.pps_v2.RunPipelineRequest.prototype.setJobId = function(value) {
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
proto.pps_v2.RunCronRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.RunCronRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.RunCronRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.RunCronRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps_v2.Pipeline.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.RunCronRequest}
 */
proto.pps_v2.RunCronRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.RunCronRequest;
  return proto.pps_v2.RunCronRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.RunCronRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.RunCronRequest}
 */
proto.pps_v2.RunCronRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Pipeline;
      reader.readMessage(value,proto.pps_v2.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
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
proto.pps_v2.RunCronRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.RunCronRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.RunCronRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.RunCronRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Pipeline.serializeBinaryToWriter
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps_v2.Pipeline}
 */
proto.pps_v2.RunCronRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps_v2.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Pipeline, 1));
};


/**
 * @param {?proto.pps_v2.Pipeline|undefined} value
 * @return {!proto.pps_v2.RunCronRequest} returns this
*/
proto.pps_v2.RunCronRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.RunCronRequest} returns this
 */
proto.pps_v2.RunCronRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.RunCronRequest.prototype.hasPipeline = function() {
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
proto.pps_v2.CreateSecretRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.CreateSecretRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.CreateSecretRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.CreateSecretRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    file: msg.getFile_asB64()
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
 * @return {!proto.pps_v2.CreateSecretRequest}
 */
proto.pps_v2.CreateSecretRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.CreateSecretRequest;
  return proto.pps_v2.CreateSecretRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.CreateSecretRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.CreateSecretRequest}
 */
proto.pps_v2.CreateSecretRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
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
proto.pps_v2.CreateSecretRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.CreateSecretRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.CreateSecretRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.CreateSecretRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getFile_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      1,
      f
    );
  }
};


/**
 * optional bytes file = 1;
 * @return {string}
 */
proto.pps_v2.CreateSecretRequest.prototype.getFile = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * optional bytes file = 1;
 * This is a type-conversion wrapper around `getFile()`
 * @return {string}
 */
proto.pps_v2.CreateSecretRequest.prototype.getFile_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getFile()));
};


/**
 * optional bytes file = 1;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getFile()`
 * @return {!Uint8Array}
 */
proto.pps_v2.CreateSecretRequest.prototype.getFile_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getFile()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pps_v2.CreateSecretRequest} returns this
 */
proto.pps_v2.CreateSecretRequest.prototype.setFile = function(value) {
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
proto.pps_v2.DeleteSecretRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.DeleteSecretRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.DeleteSecretRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DeleteSecretRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    secret: (f = msg.getSecret()) && proto.pps_v2.Secret.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.DeleteSecretRequest}
 */
proto.pps_v2.DeleteSecretRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.DeleteSecretRequest;
  return proto.pps_v2.DeleteSecretRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.DeleteSecretRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.DeleteSecretRequest}
 */
proto.pps_v2.DeleteSecretRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Secret;
      reader.readMessage(value,proto.pps_v2.Secret.deserializeBinaryFromReader);
      msg.setSecret(value);
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
proto.pps_v2.DeleteSecretRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.DeleteSecretRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.DeleteSecretRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.DeleteSecretRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSecret();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Secret.serializeBinaryToWriter
    );
  }
};


/**
 * optional Secret secret = 1;
 * @return {?proto.pps_v2.Secret}
 */
proto.pps_v2.DeleteSecretRequest.prototype.getSecret = function() {
  return /** @type{?proto.pps_v2.Secret} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Secret, 1));
};


/**
 * @param {?proto.pps_v2.Secret|undefined} value
 * @return {!proto.pps_v2.DeleteSecretRequest} returns this
*/
proto.pps_v2.DeleteSecretRequest.prototype.setSecret = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.DeleteSecretRequest} returns this
 */
proto.pps_v2.DeleteSecretRequest.prototype.clearSecret = function() {
  return this.setSecret(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.DeleteSecretRequest.prototype.hasSecret = function() {
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
proto.pps_v2.InspectSecretRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.InspectSecretRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.InspectSecretRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InspectSecretRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    secret: (f = msg.getSecret()) && proto.pps_v2.Secret.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.InspectSecretRequest}
 */
proto.pps_v2.InspectSecretRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.InspectSecretRequest;
  return proto.pps_v2.InspectSecretRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.InspectSecretRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.InspectSecretRequest}
 */
proto.pps_v2.InspectSecretRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Secret;
      reader.readMessage(value,proto.pps_v2.Secret.deserializeBinaryFromReader);
      msg.setSecret(value);
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
proto.pps_v2.InspectSecretRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.InspectSecretRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.InspectSecretRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.InspectSecretRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSecret();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Secret.serializeBinaryToWriter
    );
  }
};


/**
 * optional Secret secret = 1;
 * @return {?proto.pps_v2.Secret}
 */
proto.pps_v2.InspectSecretRequest.prototype.getSecret = function() {
  return /** @type{?proto.pps_v2.Secret} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Secret, 1));
};


/**
 * @param {?proto.pps_v2.Secret|undefined} value
 * @return {!proto.pps_v2.InspectSecretRequest} returns this
*/
proto.pps_v2.InspectSecretRequest.prototype.setSecret = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.InspectSecretRequest} returns this
 */
proto.pps_v2.InspectSecretRequest.prototype.clearSecret = function() {
  return this.setSecret(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.InspectSecretRequest.prototype.hasSecret = function() {
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
proto.pps_v2.Secret.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.Secret.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.Secret} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Secret.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps_v2.Secret}
 */
proto.pps_v2.Secret.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.Secret;
  return proto.pps_v2.Secret.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.Secret} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.Secret}
 */
proto.pps_v2.Secret.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps_v2.Secret.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.Secret.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.Secret} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.Secret.serializeBinaryToWriter = function(message, writer) {
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
proto.pps_v2.Secret.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.Secret} returns this
 */
proto.pps_v2.Secret.prototype.setName = function(value) {
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
proto.pps_v2.SecretInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.SecretInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.SecretInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.SecretInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    secret: (f = msg.getSecret()) && proto.pps_v2.Secret.toObject(includeInstance, f),
    type: jspb.Message.getFieldWithDefault(msg, 2, ""),
    creationTimestamp: (f = msg.getCreationTimestamp()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f)
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
 * @return {!proto.pps_v2.SecretInfo}
 */
proto.pps_v2.SecretInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.SecretInfo;
  return proto.pps_v2.SecretInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.SecretInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.SecretInfo}
 */
proto.pps_v2.SecretInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.Secret;
      reader.readMessage(value,proto.pps_v2.Secret.deserializeBinaryFromReader);
      msg.setSecret(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setType(value);
      break;
    case 3:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCreationTimestamp(value);
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
proto.pps_v2.SecretInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.SecretInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.SecretInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.SecretInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSecret();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps_v2.Secret.serializeBinaryToWriter
    );
  }
  f = message.getType();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getCreationTimestamp();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
};


/**
 * optional Secret secret = 1;
 * @return {?proto.pps_v2.Secret}
 */
proto.pps_v2.SecretInfo.prototype.getSecret = function() {
  return /** @type{?proto.pps_v2.Secret} */ (
    jspb.Message.getWrapperField(this, proto.pps_v2.Secret, 1));
};


/**
 * @param {?proto.pps_v2.Secret|undefined} value
 * @return {!proto.pps_v2.SecretInfo} returns this
*/
proto.pps_v2.SecretInfo.prototype.setSecret = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.SecretInfo} returns this
 */
proto.pps_v2.SecretInfo.prototype.clearSecret = function() {
  return this.setSecret(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.SecretInfo.prototype.hasSecret = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string type = 2;
 * @return {string}
 */
proto.pps_v2.SecretInfo.prototype.getType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps_v2.SecretInfo} returns this
 */
proto.pps_v2.SecretInfo.prototype.setType = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional google.protobuf.Timestamp creation_timestamp = 3;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps_v2.SecretInfo.prototype.getCreationTimestamp = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 3));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps_v2.SecretInfo} returns this
*/
proto.pps_v2.SecretInfo.prototype.setCreationTimestamp = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps_v2.SecretInfo} returns this
 */
proto.pps_v2.SecretInfo.prototype.clearCreationTimestamp = function() {
  return this.setCreationTimestamp(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps_v2.SecretInfo.prototype.hasCreationTimestamp = function() {
  return jspb.Message.getField(this, 3) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps_v2.SecretInfos.repeatedFields_ = [1];



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
proto.pps_v2.SecretInfos.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.SecretInfos.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.SecretInfos} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.SecretInfos.toObject = function(includeInstance, msg) {
  var f, obj = {
    secretInfoList: jspb.Message.toObjectList(msg.getSecretInfoList(),
    proto.pps_v2.SecretInfo.toObject, includeInstance)
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
 * @return {!proto.pps_v2.SecretInfos}
 */
proto.pps_v2.SecretInfos.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.SecretInfos;
  return proto.pps_v2.SecretInfos.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.SecretInfos} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.SecretInfos}
 */
proto.pps_v2.SecretInfos.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps_v2.SecretInfo;
      reader.readMessage(value,proto.pps_v2.SecretInfo.deserializeBinaryFromReader);
      msg.addSecretInfo(value);
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
proto.pps_v2.SecretInfos.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.SecretInfos.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.SecretInfos} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.SecretInfos.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSecretInfoList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pps_v2.SecretInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated SecretInfo secret_info = 1;
 * @return {!Array<!proto.pps_v2.SecretInfo>}
 */
proto.pps_v2.SecretInfos.prototype.getSecretInfoList = function() {
  return /** @type{!Array<!proto.pps_v2.SecretInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps_v2.SecretInfo, 1));
};


/**
 * @param {!Array<!proto.pps_v2.SecretInfo>} value
 * @return {!proto.pps_v2.SecretInfos} returns this
*/
proto.pps_v2.SecretInfos.prototype.setSecretInfoList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pps_v2.SecretInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps_v2.SecretInfo}
 */
proto.pps_v2.SecretInfos.prototype.addSecretInfo = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pps_v2.SecretInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps_v2.SecretInfos} returns this
 */
proto.pps_v2.SecretInfos.prototype.clearSecretInfoList = function() {
  return this.setSecretInfoList([]);
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
proto.pps_v2.ActivateAuthRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.ActivateAuthRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.ActivateAuthRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ActivateAuthRequest.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps_v2.ActivateAuthRequest}
 */
proto.pps_v2.ActivateAuthRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.ActivateAuthRequest;
  return proto.pps_v2.ActivateAuthRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.ActivateAuthRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.ActivateAuthRequest}
 */
proto.pps_v2.ActivateAuthRequest.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps_v2.ActivateAuthRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.ActivateAuthRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.ActivateAuthRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ActivateAuthRequest.serializeBinaryToWriter = function(message, writer) {
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
proto.pps_v2.ActivateAuthResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pps_v2.ActivateAuthResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps_v2.ActivateAuthResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ActivateAuthResponse.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps_v2.ActivateAuthResponse}
 */
proto.pps_v2.ActivateAuthResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps_v2.ActivateAuthResponse;
  return proto.pps_v2.ActivateAuthResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps_v2.ActivateAuthResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps_v2.ActivateAuthResponse}
 */
proto.pps_v2.ActivateAuthResponse.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps_v2.ActivateAuthResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps_v2.ActivateAuthResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps_v2.ActivateAuthResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps_v2.ActivateAuthResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};


/**
 * @enum {number}
 */
proto.pps_v2.JobState = {
  JOB_STATE_UNKNOWN: 0,
  JOB_CREATED: 1,
  JOB_STARTING: 2,
  JOB_RUNNING: 3,
  JOB_FAILURE: 4,
  JOB_SUCCESS: 5,
  JOB_KILLED: 6,
  JOB_EGRESSING: 7
};

/**
 * @enum {number}
 */
proto.pps_v2.DatumState = {
  DATUM_STATE_UNKNOWN: 0,
  FAILED: 1,
  SUCCESS: 2,
  SKIPPED: 3,
  STARTING: 4,
  RECOVERED: 5,
  UNPROCESSED: 6
};

/**
 * @enum {number}
 */
proto.pps_v2.WorkerState = {
  WORKER_STATE_UNKNOWN: 0,
  POD_RUNNING: 1,
  POD_SUCCESS: 2,
  POD_FAILED: 3
};

/**
 * @enum {number}
 */
proto.pps_v2.PipelineState = {
  PIPELINE_STATE_UNKNOWN: 0,
  PIPELINE_STARTING: 1,
  PIPELINE_RUNNING: 2,
  PIPELINE_RESTARTING: 3,
  PIPELINE_FAILURE: 4,
  PIPELINE_PAUSED: 5,
  PIPELINE_STANDBY: 6,
  PIPELINE_CRASHING: 7
};

goog.object.extend(exports, proto.pps_v2);
