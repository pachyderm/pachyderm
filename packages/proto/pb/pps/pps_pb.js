// source: pps/pps.proto
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
var google_protobuf_duration_pb = require('google-protobuf/google/protobuf/duration_pb.js');
goog.object.extend(proto, google_protobuf_duration_pb);
var gogoproto_gogo_pb = require('../gogoproto/gogo_pb.js');
goog.object.extend(proto, gogoproto_gogo_pb);
var pfs_pfs_pb = require('../pfs/pfs_pb.js');
goog.object.extend(proto, pfs_pfs_pb);
goog.exportSymbol('proto.pps.ActivateAuthRequest', null, global);
goog.exportSymbol('proto.pps.ActivateAuthResponse', null, global);
goog.exportSymbol('proto.pps.Aggregate', null, global);
goog.exportSymbol('proto.pps.AggregateProcessStats', null, global);
goog.exportSymbol('proto.pps.BuildSpec', null, global);
goog.exportSymbol('proto.pps.ChunkSpec', null, global);
goog.exportSymbol('proto.pps.CreateJobRequest', null, global);
goog.exportSymbol('proto.pps.CreatePipelineRequest', null, global);
goog.exportSymbol('proto.pps.CreateSecretRequest', null, global);
goog.exportSymbol('proto.pps.CronInput', null, global);
goog.exportSymbol('proto.pps.Datum', null, global);
goog.exportSymbol('proto.pps.DatumInfo', null, global);
goog.exportSymbol('proto.pps.DatumState', null, global);
goog.exportSymbol('proto.pps.DeleteJobRequest', null, global);
goog.exportSymbol('proto.pps.DeletePipelineRequest', null, global);
goog.exportSymbol('proto.pps.DeleteSecretRequest', null, global);
goog.exportSymbol('proto.pps.Egress', null, global);
goog.exportSymbol('proto.pps.EtcdJobInfo', null, global);
goog.exportSymbol('proto.pps.EtcdPipelineInfo', null, global);
goog.exportSymbol('proto.pps.FlushJobRequest', null, global);
goog.exportSymbol('proto.pps.GPUSpec', null, global);
goog.exportSymbol('proto.pps.GarbageCollectRequest', null, global);
goog.exportSymbol('proto.pps.GarbageCollectResponse', null, global);
goog.exportSymbol('proto.pps.GetLogsRequest', null, global);
goog.exportSymbol('proto.pps.GitInput', null, global);
goog.exportSymbol('proto.pps.Input', null, global);
goog.exportSymbol('proto.pps.InputFile', null, global);
goog.exportSymbol('proto.pps.InspectDatumRequest', null, global);
goog.exportSymbol('proto.pps.InspectJobRequest', null, global);
goog.exportSymbol('proto.pps.InspectPipelineRequest', null, global);
goog.exportSymbol('proto.pps.InspectSecretRequest', null, global);
goog.exportSymbol('proto.pps.Job', null, global);
goog.exportSymbol('proto.pps.JobInfo', null, global);
goog.exportSymbol('proto.pps.JobInput', null, global);
goog.exportSymbol('proto.pps.JobState', null, global);
goog.exportSymbol('proto.pps.ListDatumRequest', null, global);
goog.exportSymbol('proto.pps.ListJobRequest', null, global);
goog.exportSymbol('proto.pps.ListPipelineRequest', null, global);
goog.exportSymbol('proto.pps.LogMessage', null, global);
goog.exportSymbol('proto.pps.Metadata', null, global);
goog.exportSymbol('proto.pps.PFSInput', null, global);
goog.exportSymbol('proto.pps.ParallelismSpec', null, global);
goog.exportSymbol('proto.pps.Pipeline', null, global);
goog.exportSymbol('proto.pps.PipelineInfo', null, global);
goog.exportSymbol('proto.pps.PipelineInfos', null, global);
goog.exportSymbol('proto.pps.PipelineState', null, global);
goog.exportSymbol('proto.pps.ProcessStats', null, global);
goog.exportSymbol('proto.pps.ResourceSpec', null, global);
goog.exportSymbol('proto.pps.RestartDatumRequest', null, global);
goog.exportSymbol('proto.pps.RunCronRequest', null, global);
goog.exportSymbol('proto.pps.RunPipelineRequest', null, global);
goog.exportSymbol('proto.pps.SchedulingSpec', null, global);
goog.exportSymbol('proto.pps.Secret', null, global);
goog.exportSymbol('proto.pps.SecretInfo', null, global);
goog.exportSymbol('proto.pps.SecretInfos', null, global);
goog.exportSymbol('proto.pps.SecretMount', null, global);
goog.exportSymbol('proto.pps.Service', null, global);
goog.exportSymbol('proto.pps.Spout', null, global);
goog.exportSymbol('proto.pps.StartPipelineRequest', null, global);
goog.exportSymbol('proto.pps.StopJobRequest', null, global);
goog.exportSymbol('proto.pps.StopPipelineRequest', null, global);
goog.exportSymbol('proto.pps.TFJob', null, global);
goog.exportSymbol('proto.pps.Transform', null, global);
goog.exportSymbol('proto.pps.UpdateJobStateRequest', null, global);
goog.exportSymbol('proto.pps.Worker', null, global);
goog.exportSymbol('proto.pps.WorkerState', null, global);
goog.exportSymbol('proto.pps.WorkerStatus', null, global);
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
proto.pps.SecretMount = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.SecretMount, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.SecretMount.displayName = 'proto.pps.SecretMount';
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
proto.pps.Transform = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.Transform.repeatedFields_, null);
};
goog.inherits(proto.pps.Transform, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Transform.displayName = 'proto.pps.Transform';
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
proto.pps.BuildSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.BuildSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.BuildSpec.displayName = 'proto.pps.BuildSpec';
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
proto.pps.TFJob = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.TFJob, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.TFJob.displayName = 'proto.pps.TFJob';
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
proto.pps.Egress = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.Egress, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Egress.displayName = 'proto.pps.Egress';
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
proto.pps.Job = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.Job, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Job.displayName = 'proto.pps.Job';
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
proto.pps.Metadata = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.Metadata, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Metadata.displayName = 'proto.pps.Metadata';
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
proto.pps.Service = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.Service, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Service.displayName = 'proto.pps.Service';
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
proto.pps.Spout = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.Spout, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Spout.displayName = 'proto.pps.Spout';
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
proto.pps.PFSInput = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.PFSInput, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.PFSInput.displayName = 'proto.pps.PFSInput';
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
proto.pps.CronInput = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.CronInput, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.CronInput.displayName = 'proto.pps.CronInput';
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
proto.pps.GitInput = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.GitInput, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.GitInput.displayName = 'proto.pps.GitInput';
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
proto.pps.Input = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.Input.repeatedFields_, null);
};
goog.inherits(proto.pps.Input, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Input.displayName = 'proto.pps.Input';
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
proto.pps.JobInput = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.JobInput, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.JobInput.displayName = 'proto.pps.JobInput';
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
proto.pps.ParallelismSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.ParallelismSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.ParallelismSpec.displayName = 'proto.pps.ParallelismSpec';
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
proto.pps.InputFile = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.InputFile, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.InputFile.displayName = 'proto.pps.InputFile';
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
proto.pps.Datum = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.Datum, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Datum.displayName = 'proto.pps.Datum';
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
proto.pps.DatumInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.DatumInfo.repeatedFields_, null);
};
goog.inherits(proto.pps.DatumInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.DatumInfo.displayName = 'proto.pps.DatumInfo';
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
proto.pps.Aggregate = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.Aggregate, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Aggregate.displayName = 'proto.pps.Aggregate';
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
proto.pps.ProcessStats = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.ProcessStats, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.ProcessStats.displayName = 'proto.pps.ProcessStats';
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
proto.pps.AggregateProcessStats = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.AggregateProcessStats, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.AggregateProcessStats.displayName = 'proto.pps.AggregateProcessStats';
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
proto.pps.WorkerStatus = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.WorkerStatus.repeatedFields_, null);
};
goog.inherits(proto.pps.WorkerStatus, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.WorkerStatus.displayName = 'proto.pps.WorkerStatus';
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
proto.pps.ResourceSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.ResourceSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.ResourceSpec.displayName = 'proto.pps.ResourceSpec';
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
proto.pps.GPUSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.GPUSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.GPUSpec.displayName = 'proto.pps.GPUSpec';
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
proto.pps.EtcdJobInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.EtcdJobInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.EtcdJobInfo.displayName = 'proto.pps.EtcdJobInfo';
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
proto.pps.JobInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.JobInfo.repeatedFields_, null);
};
goog.inherits(proto.pps.JobInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.JobInfo.displayName = 'proto.pps.JobInfo';
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
proto.pps.Worker = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.Worker, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Worker.displayName = 'proto.pps.Worker';
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
proto.pps.Pipeline = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.Pipeline, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Pipeline.displayName = 'proto.pps.Pipeline';
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
proto.pps.EtcdPipelineInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.EtcdPipelineInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.EtcdPipelineInfo.displayName = 'proto.pps.EtcdPipelineInfo';
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
proto.pps.PipelineInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.PipelineInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.PipelineInfo.displayName = 'proto.pps.PipelineInfo';
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
proto.pps.PipelineInfos = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.PipelineInfos.repeatedFields_, null);
};
goog.inherits(proto.pps.PipelineInfos, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.PipelineInfos.displayName = 'proto.pps.PipelineInfos';
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
proto.pps.CreateJobRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.CreateJobRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.CreateJobRequest.displayName = 'proto.pps.CreateJobRequest';
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
proto.pps.InspectJobRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.InspectJobRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.InspectJobRequest.displayName = 'proto.pps.InspectJobRequest';
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
proto.pps.ListJobRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.ListJobRequest.repeatedFields_, null);
};
goog.inherits(proto.pps.ListJobRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.ListJobRequest.displayName = 'proto.pps.ListJobRequest';
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
proto.pps.FlushJobRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.FlushJobRequest.repeatedFields_, null);
};
goog.inherits(proto.pps.FlushJobRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.FlushJobRequest.displayName = 'proto.pps.FlushJobRequest';
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
proto.pps.DeleteJobRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.DeleteJobRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.DeleteJobRequest.displayName = 'proto.pps.DeleteJobRequest';
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
proto.pps.StopJobRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.StopJobRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.StopJobRequest.displayName = 'proto.pps.StopJobRequest';
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
proto.pps.UpdateJobStateRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.UpdateJobStateRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.UpdateJobStateRequest.displayName = 'proto.pps.UpdateJobStateRequest';
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
proto.pps.GetLogsRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.GetLogsRequest.repeatedFields_, null);
};
goog.inherits(proto.pps.GetLogsRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.GetLogsRequest.displayName = 'proto.pps.GetLogsRequest';
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
proto.pps.LogMessage = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.LogMessage.repeatedFields_, null);
};
goog.inherits(proto.pps.LogMessage, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.LogMessage.displayName = 'proto.pps.LogMessage';
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
proto.pps.RestartDatumRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.RestartDatumRequest.repeatedFields_, null);
};
goog.inherits(proto.pps.RestartDatumRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.RestartDatumRequest.displayName = 'proto.pps.RestartDatumRequest';
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
proto.pps.InspectDatumRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.InspectDatumRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.InspectDatumRequest.displayName = 'proto.pps.InspectDatumRequest';
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
proto.pps.ListDatumRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.ListDatumRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.ListDatumRequest.displayName = 'proto.pps.ListDatumRequest';
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
proto.pps.ChunkSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.ChunkSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.ChunkSpec.displayName = 'proto.pps.ChunkSpec';
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
proto.pps.SchedulingSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.SchedulingSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.SchedulingSpec.displayName = 'proto.pps.SchedulingSpec';
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
proto.pps.CreatePipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.CreatePipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.CreatePipelineRequest.displayName = 'proto.pps.CreatePipelineRequest';
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
proto.pps.InspectPipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.InspectPipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.InspectPipelineRequest.displayName = 'proto.pps.InspectPipelineRequest';
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
proto.pps.ListPipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.ListPipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.ListPipelineRequest.displayName = 'proto.pps.ListPipelineRequest';
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
proto.pps.DeletePipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.DeletePipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.DeletePipelineRequest.displayName = 'proto.pps.DeletePipelineRequest';
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
proto.pps.StartPipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.StartPipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.StartPipelineRequest.displayName = 'proto.pps.StartPipelineRequest';
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
proto.pps.StopPipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.StopPipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.StopPipelineRequest.displayName = 'proto.pps.StopPipelineRequest';
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
proto.pps.RunPipelineRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.RunPipelineRequest.repeatedFields_, null);
};
goog.inherits(proto.pps.RunPipelineRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.RunPipelineRequest.displayName = 'proto.pps.RunPipelineRequest';
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
proto.pps.RunCronRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.RunCronRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.RunCronRequest.displayName = 'proto.pps.RunCronRequest';
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
proto.pps.CreateSecretRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.CreateSecretRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.CreateSecretRequest.displayName = 'proto.pps.CreateSecretRequest';
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
proto.pps.DeleteSecretRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.DeleteSecretRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.DeleteSecretRequest.displayName = 'proto.pps.DeleteSecretRequest';
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
proto.pps.InspectSecretRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.InspectSecretRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.InspectSecretRequest.displayName = 'proto.pps.InspectSecretRequest';
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
proto.pps.Secret = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.Secret, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.Secret.displayName = 'proto.pps.Secret';
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
proto.pps.SecretInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.SecretInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.SecretInfo.displayName = 'proto.pps.SecretInfo';
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
proto.pps.SecretInfos = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.pps.SecretInfos.repeatedFields_, null);
};
goog.inherits(proto.pps.SecretInfos, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.SecretInfos.displayName = 'proto.pps.SecretInfos';
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
proto.pps.GarbageCollectRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.GarbageCollectRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.GarbageCollectRequest.displayName = 'proto.pps.GarbageCollectRequest';
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
proto.pps.GarbageCollectResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.GarbageCollectResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.GarbageCollectResponse.displayName = 'proto.pps.GarbageCollectResponse';
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
proto.pps.ActivateAuthRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.ActivateAuthRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.ActivateAuthRequest.displayName = 'proto.pps.ActivateAuthRequest';
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
proto.pps.ActivateAuthResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pps.ActivateAuthResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pps.ActivateAuthResponse.displayName = 'proto.pps.ActivateAuthResponse';
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
proto.pps.SecretMount.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.SecretMount.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.SecretMount} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.SecretMount.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    key: jspb.Message.getFieldWithDefault(msg, 4, ""),
    mountPath: jspb.Message.getFieldWithDefault(msg, 2, ""),
    envVar: jspb.Message.getFieldWithDefault(msg, 3, "")
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
 * @return {!proto.pps.SecretMount}
 */
proto.pps.SecretMount.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.SecretMount;
  return proto.pps.SecretMount.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.SecretMount} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.SecretMount}
 */
proto.pps.SecretMount.deserializeBinaryFromReader = function(msg, reader) {
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
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setKey(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setMountPath(value);
      break;
    case 3:
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
proto.pps.SecretMount.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.SecretMount.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.SecretMount} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.SecretMount.serializeBinaryToWriter = function(message, writer) {
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
      4,
      f
    );
  }
  f = message.getMountPath();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getEnvVar();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pps.SecretMount.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.SecretMount} returns this
 */
proto.pps.SecretMount.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string key = 4;
 * @return {string}
 */
proto.pps.SecretMount.prototype.getKey = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.SecretMount} returns this
 */
proto.pps.SecretMount.prototype.setKey = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string mount_path = 2;
 * @return {string}
 */
proto.pps.SecretMount.prototype.getMountPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.SecretMount} returns this
 */
proto.pps.SecretMount.prototype.setMountPath = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string env_var = 3;
 * @return {string}
 */
proto.pps.SecretMount.prototype.getEnvVar = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.SecretMount} returns this
 */
proto.pps.SecretMount.prototype.setEnvVar = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.Transform.repeatedFields_ = [2,13,4,9,5,14,6];



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
proto.pps.Transform.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Transform.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Transform} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Transform.toObject = function(includeInstance, msg) {
  var f, obj = {
    image: jspb.Message.getFieldWithDefault(msg, 1, ""),
    cmdList: (f = jspb.Message.getRepeatedField(msg, 2)) == null ? undefined : f,
    errCmdList: (f = jspb.Message.getRepeatedField(msg, 13)) == null ? undefined : f,
    envMap: (f = msg.getEnvMap()) ? f.toObject(includeInstance, undefined) : [],
    secretsList: jspb.Message.toObjectList(msg.getSecretsList(),
    proto.pps.SecretMount.toObject, includeInstance),
    imagePullSecretsList: (f = jspb.Message.getRepeatedField(msg, 9)) == null ? undefined : f,
    stdinList: (f = jspb.Message.getRepeatedField(msg, 5)) == null ? undefined : f,
    errStdinList: (f = jspb.Message.getRepeatedField(msg, 14)) == null ? undefined : f,
    acceptReturnCodeList: (f = jspb.Message.getRepeatedField(msg, 6)) == null ? undefined : f,
    debug: jspb.Message.getBooleanFieldWithDefault(msg, 7, false),
    user: jspb.Message.getFieldWithDefault(msg, 10, ""),
    workingDir: jspb.Message.getFieldWithDefault(msg, 11, ""),
    dockerfile: jspb.Message.getFieldWithDefault(msg, 12, ""),
    build: (f = msg.getBuild()) && proto.pps.BuildSpec.toObject(includeInstance, f)
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
 * @return {!proto.pps.Transform}
 */
proto.pps.Transform.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Transform;
  return proto.pps.Transform.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Transform} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Transform}
 */
proto.pps.Transform.deserializeBinaryFromReader = function(msg, reader) {
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
    case 13:
      var value = /** @type {string} */ (reader.readString());
      msg.addErrCmd(value);
      break;
    case 3:
      var value = msg.getEnvMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString, null, "", "");
         });
      break;
    case 4:
      var value = new proto.pps.SecretMount;
      reader.readMessage(value,proto.pps.SecretMount.deserializeBinaryFromReader);
      msg.addSecrets(value);
      break;
    case 9:
      var value = /** @type {string} */ (reader.readString());
      msg.addImagePullSecrets(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.addStdin(value);
      break;
    case 14:
      var value = /** @type {string} */ (reader.readString());
      msg.addErrStdin(value);
      break;
    case 6:
      var values = /** @type {!Array<number>} */ (reader.isDelimited() ? reader.readPackedInt64() : [reader.readInt64()]);
      for (var i = 0; i < values.length; i++) {
        msg.addAcceptReturnCode(values[i]);
      }
      break;
    case 7:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDebug(value);
      break;
    case 10:
      var value = /** @type {string} */ (reader.readString());
      msg.setUser(value);
      break;
    case 11:
      var value = /** @type {string} */ (reader.readString());
      msg.setWorkingDir(value);
      break;
    case 12:
      var value = /** @type {string} */ (reader.readString());
      msg.setDockerfile(value);
      break;
    case 15:
      var value = new proto.pps.BuildSpec;
      reader.readMessage(value,proto.pps.BuildSpec.deserializeBinaryFromReader);
      msg.setBuild(value);
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
proto.pps.Transform.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Transform.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Transform} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Transform.serializeBinaryToWriter = function(message, writer) {
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
      13,
      f
    );
  }
  f = message.getEnvMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(3, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getSecretsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      4,
      f,
      proto.pps.SecretMount.serializeBinaryToWriter
    );
  }
  f = message.getImagePullSecretsList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      9,
      f
    );
  }
  f = message.getStdinList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      5,
      f
    );
  }
  f = message.getErrStdinList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      14,
      f
    );
  }
  f = message.getAcceptReturnCodeList();
  if (f.length > 0) {
    writer.writePackedInt64(
      6,
      f
    );
  }
  f = message.getDebug();
  if (f) {
    writer.writeBool(
      7,
      f
    );
  }
  f = message.getUser();
  if (f.length > 0) {
    writer.writeString(
      10,
      f
    );
  }
  f = message.getWorkingDir();
  if (f.length > 0) {
    writer.writeString(
      11,
      f
    );
  }
  f = message.getDockerfile();
  if (f.length > 0) {
    writer.writeString(
      12,
      f
    );
  }
  f = message.getBuild();
  if (f != null) {
    writer.writeMessage(
      15,
      f,
      proto.pps.BuildSpec.serializeBinaryToWriter
    );
  }
};


/**
 * optional string image = 1;
 * @return {string}
 */
proto.pps.Transform.prototype.getImage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.setImage = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * repeated string cmd = 2;
 * @return {!Array<string>}
 */
proto.pps.Transform.prototype.getCmdList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.setCmdList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.addCmd = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.clearCmdList = function() {
  return this.setCmdList([]);
};


/**
 * repeated string err_cmd = 13;
 * @return {!Array<string>}
 */
proto.pps.Transform.prototype.getErrCmdList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 13));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.setErrCmdList = function(value) {
  return jspb.Message.setField(this, 13, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.addErrCmd = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 13, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.clearErrCmdList = function() {
  return this.setErrCmdList([]);
};


/**
 * map<string, string> env = 3;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.pps.Transform.prototype.getEnvMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 3, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.clearEnvMap = function() {
  this.getEnvMap().clear();
  return this;};


/**
 * repeated SecretMount secrets = 4;
 * @return {!Array<!proto.pps.SecretMount>}
 */
proto.pps.Transform.prototype.getSecretsList = function() {
  return /** @type{!Array<!proto.pps.SecretMount>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps.SecretMount, 4));
};


/**
 * @param {!Array<!proto.pps.SecretMount>} value
 * @return {!proto.pps.Transform} returns this
*/
proto.pps.Transform.prototype.setSecretsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 4, value);
};


/**
 * @param {!proto.pps.SecretMount=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps.SecretMount}
 */
proto.pps.Transform.prototype.addSecrets = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 4, opt_value, proto.pps.SecretMount, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.clearSecretsList = function() {
  return this.setSecretsList([]);
};


/**
 * repeated string image_pull_secrets = 9;
 * @return {!Array<string>}
 */
proto.pps.Transform.prototype.getImagePullSecretsList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 9));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.setImagePullSecretsList = function(value) {
  return jspb.Message.setField(this, 9, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.addImagePullSecrets = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 9, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.clearImagePullSecretsList = function() {
  return this.setImagePullSecretsList([]);
};


/**
 * repeated string stdin = 5;
 * @return {!Array<string>}
 */
proto.pps.Transform.prototype.getStdinList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 5));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.setStdinList = function(value) {
  return jspb.Message.setField(this, 5, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.addStdin = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 5, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.clearStdinList = function() {
  return this.setStdinList([]);
};


/**
 * repeated string err_stdin = 14;
 * @return {!Array<string>}
 */
proto.pps.Transform.prototype.getErrStdinList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 14));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.setErrStdinList = function(value) {
  return jspb.Message.setField(this, 14, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.addErrStdin = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 14, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.clearErrStdinList = function() {
  return this.setErrStdinList([]);
};


/**
 * repeated int64 accept_return_code = 6;
 * @return {!Array<number>}
 */
proto.pps.Transform.prototype.getAcceptReturnCodeList = function() {
  return /** @type {!Array<number>} */ (jspb.Message.getRepeatedField(this, 6));
};


/**
 * @param {!Array<number>} value
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.setAcceptReturnCodeList = function(value) {
  return jspb.Message.setField(this, 6, value || []);
};


/**
 * @param {number} value
 * @param {number=} opt_index
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.addAcceptReturnCode = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 6, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.clearAcceptReturnCodeList = function() {
  return this.setAcceptReturnCodeList([]);
};


/**
 * optional bool debug = 7;
 * @return {boolean}
 */
proto.pps.Transform.prototype.getDebug = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 7, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.setDebug = function(value) {
  return jspb.Message.setProto3BooleanField(this, 7, value);
};


/**
 * optional string user = 10;
 * @return {string}
 */
proto.pps.Transform.prototype.getUser = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 10, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.setUser = function(value) {
  return jspb.Message.setProto3StringField(this, 10, value);
};


/**
 * optional string working_dir = 11;
 * @return {string}
 */
proto.pps.Transform.prototype.getWorkingDir = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 11, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.setWorkingDir = function(value) {
  return jspb.Message.setProto3StringField(this, 11, value);
};


/**
 * optional string dockerfile = 12;
 * @return {string}
 */
proto.pps.Transform.prototype.getDockerfile = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 12, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.setDockerfile = function(value) {
  return jspb.Message.setProto3StringField(this, 12, value);
};


/**
 * optional BuildSpec build = 15;
 * @return {?proto.pps.BuildSpec}
 */
proto.pps.Transform.prototype.getBuild = function() {
  return /** @type{?proto.pps.BuildSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.BuildSpec, 15));
};


/**
 * @param {?proto.pps.BuildSpec|undefined} value
 * @return {!proto.pps.Transform} returns this
*/
proto.pps.Transform.prototype.setBuild = function(value) {
  return jspb.Message.setWrapperField(this, 15, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.Transform} returns this
 */
proto.pps.Transform.prototype.clearBuild = function() {
  return this.setBuild(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.Transform.prototype.hasBuild = function() {
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
proto.pps.BuildSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.BuildSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.BuildSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.BuildSpec.toObject = function(includeInstance, msg) {
  var f, obj = {
    path: jspb.Message.getFieldWithDefault(msg, 1, ""),
    language: jspb.Message.getFieldWithDefault(msg, 2, ""),
    image: jspb.Message.getFieldWithDefault(msg, 3, "")
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
 * @return {!proto.pps.BuildSpec}
 */
proto.pps.BuildSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.BuildSpec;
  return proto.pps.BuildSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.BuildSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.BuildSpec}
 */
proto.pps.BuildSpec.deserializeBinaryFromReader = function(msg, reader) {
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
      msg.setLanguage(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setImage(value);
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
proto.pps.BuildSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.BuildSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.BuildSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.BuildSpec.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getLanguage();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getImage();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
};


/**
 * optional string path = 1;
 * @return {string}
 */
proto.pps.BuildSpec.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.BuildSpec} returns this
 */
proto.pps.BuildSpec.prototype.setPath = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string language = 2;
 * @return {string}
 */
proto.pps.BuildSpec.prototype.getLanguage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.BuildSpec} returns this
 */
proto.pps.BuildSpec.prototype.setLanguage = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string image = 3;
 * @return {string}
 */
proto.pps.BuildSpec.prototype.getImage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.BuildSpec} returns this
 */
proto.pps.BuildSpec.prototype.setImage = function(value) {
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
proto.pps.TFJob.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.TFJob.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.TFJob} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.TFJob.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.TFJob}
 */
proto.pps.TFJob.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.TFJob;
  return proto.pps.TFJob.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.TFJob} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.TFJob}
 */
proto.pps.TFJob.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.TFJob.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.TFJob.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.TFJob} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.TFJob.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.TFJob.prototype.getTfJob = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.TFJob} returns this
 */
proto.pps.TFJob.prototype.setTfJob = function(value) {
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
proto.pps.Egress.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Egress.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Egress} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Egress.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.Egress}
 */
proto.pps.Egress.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Egress;
  return proto.pps.Egress.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Egress} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Egress}
 */
proto.pps.Egress.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.Egress.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Egress.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Egress} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Egress.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.Egress.prototype.getUrl = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Egress} returns this
 */
proto.pps.Egress.prototype.setUrl = function(value) {
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
proto.pps.Job.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Job.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Job} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Job.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.Job}
 */
proto.pps.Job.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Job;
  return proto.pps.Job.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Job} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Job}
 */
proto.pps.Job.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.Job.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Job.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Job} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Job.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.Job.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Job} returns this
 */
proto.pps.Job.prototype.setId = function(value) {
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
proto.pps.Metadata.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Metadata.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Metadata} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Metadata.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.Metadata}
 */
proto.pps.Metadata.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Metadata;
  return proto.pps.Metadata.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Metadata} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Metadata}
 */
proto.pps.Metadata.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.Metadata.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Metadata.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Metadata} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Metadata.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.Metadata.prototype.getAnnotationsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pps.Metadata} returns this
 */
proto.pps.Metadata.prototype.clearAnnotationsMap = function() {
  this.getAnnotationsMap().clear();
  return this;};


/**
 * map<string, string> labels = 2;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.pps.Metadata.prototype.getLabelsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 2, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pps.Metadata} returns this
 */
proto.pps.Metadata.prototype.clearLabelsMap = function() {
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
proto.pps.Service.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Service.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Service} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Service.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.Service}
 */
proto.pps.Service.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Service;
  return proto.pps.Service.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Service} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Service}
 */
proto.pps.Service.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.Service.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Service.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Service} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Service.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.Service.prototype.getInternalPort = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.Service} returns this
 */
proto.pps.Service.prototype.setInternalPort = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional int32 external_port = 2;
 * @return {number}
 */
proto.pps.Service.prototype.getExternalPort = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.Service} returns this
 */
proto.pps.Service.prototype.setExternalPort = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional string ip = 3;
 * @return {string}
 */
proto.pps.Service.prototype.getIp = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Service} returns this
 */
proto.pps.Service.prototype.setIp = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string type = 4;
 * @return {string}
 */
proto.pps.Service.prototype.getType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Service} returns this
 */
proto.pps.Service.prototype.setType = function(value) {
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
proto.pps.Spout.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Spout.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Spout} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Spout.toObject = function(includeInstance, msg) {
  var f, obj = {
    service: (f = msg.getService()) && proto.pps.Service.toObject(includeInstance, f)
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
 * @return {!proto.pps.Spout}
 */
proto.pps.Spout.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Spout;
  return proto.pps.Spout.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Spout} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Spout}
 */
proto.pps.Spout.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Service;
      reader.readMessage(value,proto.pps.Service.deserializeBinaryFromReader);
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
proto.pps.Spout.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Spout.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Spout} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Spout.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getService();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Service.serializeBinaryToWriter
    );
  }
};


/**
 * optional Service service = 1;
 * @return {?proto.pps.Service}
 */
proto.pps.Spout.prototype.getService = function() {
  return /** @type{?proto.pps.Service} */ (
    jspb.Message.getWrapperField(this, proto.pps.Service, 1));
};


/**
 * @param {?proto.pps.Service|undefined} value
 * @return {!proto.pps.Spout} returns this
*/
proto.pps.Spout.prototype.setService = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.Spout} returns this
 */
proto.pps.Spout.prototype.clearService = function() {
  return this.setService(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.Spout.prototype.hasService = function() {
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
proto.pps.PFSInput.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.PFSInput.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.PFSInput} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.PFSInput.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    repo: jspb.Message.getFieldWithDefault(msg, 2, ""),
    branch: jspb.Message.getFieldWithDefault(msg, 3, ""),
    commit: jspb.Message.getFieldWithDefault(msg, 4, ""),
    glob: jspb.Message.getFieldWithDefault(msg, 5, ""),
    joinOn: jspb.Message.getFieldWithDefault(msg, 8, ""),
    outerJoin: jspb.Message.getBooleanFieldWithDefault(msg, 12, false),
    groupBy: jspb.Message.getFieldWithDefault(msg, 11, ""),
    lazy: jspb.Message.getBooleanFieldWithDefault(msg, 6, false),
    emptyFiles: jspb.Message.getBooleanFieldWithDefault(msg, 7, false),
    s3: jspb.Message.getBooleanFieldWithDefault(msg, 9, false),
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
 * @return {!proto.pps.PFSInput}
 */
proto.pps.PFSInput.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.PFSInput;
  return proto.pps.PFSInput.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.PFSInput} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.PFSInput}
 */
proto.pps.PFSInput.deserializeBinaryFromReader = function(msg, reader) {
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
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setJoinOn(value);
      break;
    case 12:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setOuterJoin(value);
      break;
    case 11:
      var value = /** @type {string} */ (reader.readString());
      msg.setGroupBy(value);
      break;
    case 6:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setLazy(value);
      break;
    case 7:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEmptyFiles(value);
      break;
    case 9:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setS3(value);
      break;
    case 10:
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
proto.pps.PFSInput.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.PFSInput.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.PFSInput} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.PFSInput.serializeBinaryToWriter = function(message, writer) {
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
      8,
      f
    );
  }
  f = message.getOuterJoin();
  if (f) {
    writer.writeBool(
      12,
      f
    );
  }
  f = message.getGroupBy();
  if (f.length > 0) {
    writer.writeString(
      11,
      f
    );
  }
  f = message.getLazy();
  if (f) {
    writer.writeBool(
      6,
      f
    );
  }
  f = message.getEmptyFiles();
  if (f) {
    writer.writeBool(
      7,
      f
    );
  }
  f = message.getS3();
  if (f) {
    writer.writeBool(
      9,
      f
    );
  }
  f = message.getTrigger();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      pfs_pfs_pb.Trigger.serializeBinaryToWriter
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pps.PFSInput.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string repo = 2;
 * @return {string}
 */
proto.pps.PFSInput.prototype.getRepo = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.setRepo = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string branch = 3;
 * @return {string}
 */
proto.pps.PFSInput.prototype.getBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.setBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string commit = 4;
 * @return {string}
 */
proto.pps.PFSInput.prototype.getCommit = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.setCommit = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional string glob = 5;
 * @return {string}
 */
proto.pps.PFSInput.prototype.getGlob = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.setGlob = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string join_on = 8;
 * @return {string}
 */
proto.pps.PFSInput.prototype.getJoinOn = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.setJoinOn = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};


/**
 * optional bool outer_join = 12;
 * @return {boolean}
 */
proto.pps.PFSInput.prototype.getOuterJoin = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 12, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.setOuterJoin = function(value) {
  return jspb.Message.setProto3BooleanField(this, 12, value);
};


/**
 * optional string group_by = 11;
 * @return {string}
 */
proto.pps.PFSInput.prototype.getGroupBy = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 11, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.setGroupBy = function(value) {
  return jspb.Message.setProto3StringField(this, 11, value);
};


/**
 * optional bool lazy = 6;
 * @return {boolean}
 */
proto.pps.PFSInput.prototype.getLazy = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 6, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.setLazy = function(value) {
  return jspb.Message.setProto3BooleanField(this, 6, value);
};


/**
 * optional bool empty_files = 7;
 * @return {boolean}
 */
proto.pps.PFSInput.prototype.getEmptyFiles = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 7, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.setEmptyFiles = function(value) {
  return jspb.Message.setProto3BooleanField(this, 7, value);
};


/**
 * optional bool s3 = 9;
 * @return {boolean}
 */
proto.pps.PFSInput.prototype.getS3 = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 9, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.setS3 = function(value) {
  return jspb.Message.setProto3BooleanField(this, 9, value);
};


/**
 * optional pfs.Trigger trigger = 10;
 * @return {?proto.pfs.Trigger}
 */
proto.pps.PFSInput.prototype.getTrigger = function() {
  return /** @type{?proto.pfs.Trigger} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Trigger, 10));
};


/**
 * @param {?proto.pfs.Trigger|undefined} value
 * @return {!proto.pps.PFSInput} returns this
*/
proto.pps.PFSInput.prototype.setTrigger = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PFSInput} returns this
 */
proto.pps.PFSInput.prototype.clearTrigger = function() {
  return this.setTrigger(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PFSInput.prototype.hasTrigger = function() {
  return jspb.Message.getField(this, 10) != null;
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
proto.pps.CronInput.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.CronInput.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.CronInput} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.CronInput.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    repo: jspb.Message.getFieldWithDefault(msg, 2, ""),
    commit: jspb.Message.getFieldWithDefault(msg, 3, ""),
    spec: jspb.Message.getFieldWithDefault(msg, 4, ""),
    overwrite: jspb.Message.getBooleanFieldWithDefault(msg, 6, false),
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
 * @return {!proto.pps.CronInput}
 */
proto.pps.CronInput.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.CronInput;
  return proto.pps.CronInput.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.CronInput} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.CronInput}
 */
proto.pps.CronInput.deserializeBinaryFromReader = function(msg, reader) {
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
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setCommit(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setSpec(value);
      break;
    case 6:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setOverwrite(value);
      break;
    case 5:
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
proto.pps.CronInput.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.CronInput.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.CronInput} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.CronInput.serializeBinaryToWriter = function(message, writer) {
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
      6,
      f
    );
  }
  f = message.getStart();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pps.CronInput.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.CronInput} returns this
 */
proto.pps.CronInput.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string repo = 2;
 * @return {string}
 */
proto.pps.CronInput.prototype.getRepo = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.CronInput} returns this
 */
proto.pps.CronInput.prototype.setRepo = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string commit = 3;
 * @return {string}
 */
proto.pps.CronInput.prototype.getCommit = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.CronInput} returns this
 */
proto.pps.CronInput.prototype.setCommit = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string spec = 4;
 * @return {string}
 */
proto.pps.CronInput.prototype.getSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.CronInput} returns this
 */
proto.pps.CronInput.prototype.setSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional bool overwrite = 6;
 * @return {boolean}
 */
proto.pps.CronInput.prototype.getOverwrite = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 6, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.CronInput} returns this
 */
proto.pps.CronInput.prototype.setOverwrite = function(value) {
  return jspb.Message.setProto3BooleanField(this, 6, value);
};


/**
 * optional google.protobuf.Timestamp start = 5;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps.CronInput.prototype.getStart = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 5));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps.CronInput} returns this
*/
proto.pps.CronInput.prototype.setStart = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CronInput} returns this
 */
proto.pps.CronInput.prototype.clearStart = function() {
  return this.setStart(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CronInput.prototype.hasStart = function() {
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
proto.pps.GitInput.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.GitInput.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.GitInput} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.GitInput.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    url: jspb.Message.getFieldWithDefault(msg, 2, ""),
    branch: jspb.Message.getFieldWithDefault(msg, 3, ""),
    commit: jspb.Message.getFieldWithDefault(msg, 4, "")
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
 * @return {!proto.pps.GitInput}
 */
proto.pps.GitInput.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.GitInput;
  return proto.pps.GitInput.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.GitInput} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.GitInput}
 */
proto.pps.GitInput.deserializeBinaryFromReader = function(msg, reader) {
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
      msg.setUrl(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setBranch(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
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
proto.pps.GitInput.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.GitInput.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.GitInput} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.GitInput.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
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
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.pps.GitInput.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.GitInput} returns this
 */
proto.pps.GitInput.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string url = 2;
 * @return {string}
 */
proto.pps.GitInput.prototype.getUrl = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.GitInput} returns this
 */
proto.pps.GitInput.prototype.setUrl = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string branch = 3;
 * @return {string}
 */
proto.pps.GitInput.prototype.getBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.GitInput} returns this
 */
proto.pps.GitInput.prototype.setBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string commit = 4;
 * @return {string}
 */
proto.pps.GitInput.prototype.getCommit = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.GitInput} returns this
 */
proto.pps.GitInput.prototype.setCommit = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.Input.repeatedFields_ = [7,8,2,3];



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
proto.pps.Input.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Input.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Input} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Input.toObject = function(includeInstance, msg) {
  var f, obj = {
    pfs: (f = msg.getPfs()) && proto.pps.PFSInput.toObject(includeInstance, f),
    joinList: jspb.Message.toObjectList(msg.getJoinList(),
    proto.pps.Input.toObject, includeInstance),
    groupList: jspb.Message.toObjectList(msg.getGroupList(),
    proto.pps.Input.toObject, includeInstance),
    crossList: jspb.Message.toObjectList(msg.getCrossList(),
    proto.pps.Input.toObject, includeInstance),
    unionList: jspb.Message.toObjectList(msg.getUnionList(),
    proto.pps.Input.toObject, includeInstance),
    cron: (f = msg.getCron()) && proto.pps.CronInput.toObject(includeInstance, f),
    git: (f = msg.getGit()) && proto.pps.GitInput.toObject(includeInstance, f)
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
 * @return {!proto.pps.Input}
 */
proto.pps.Input.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Input;
  return proto.pps.Input.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Input} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Input}
 */
proto.pps.Input.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 6:
      var value = new proto.pps.PFSInput;
      reader.readMessage(value,proto.pps.PFSInput.deserializeBinaryFromReader);
      msg.setPfs(value);
      break;
    case 7:
      var value = new proto.pps.Input;
      reader.readMessage(value,proto.pps.Input.deserializeBinaryFromReader);
      msg.addJoin(value);
      break;
    case 8:
      var value = new proto.pps.Input;
      reader.readMessage(value,proto.pps.Input.deserializeBinaryFromReader);
      msg.addGroup(value);
      break;
    case 2:
      var value = new proto.pps.Input;
      reader.readMessage(value,proto.pps.Input.deserializeBinaryFromReader);
      msg.addCross(value);
      break;
    case 3:
      var value = new proto.pps.Input;
      reader.readMessage(value,proto.pps.Input.deserializeBinaryFromReader);
      msg.addUnion(value);
      break;
    case 4:
      var value = new proto.pps.CronInput;
      reader.readMessage(value,proto.pps.CronInput.deserializeBinaryFromReader);
      msg.setCron(value);
      break;
    case 5:
      var value = new proto.pps.GitInput;
      reader.readMessage(value,proto.pps.GitInput.deserializeBinaryFromReader);
      msg.setGit(value);
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
proto.pps.Input.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Input.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Input} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Input.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPfs();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      proto.pps.PFSInput.serializeBinaryToWriter
    );
  }
  f = message.getJoinList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      7,
      f,
      proto.pps.Input.serializeBinaryToWriter
    );
  }
  f = message.getGroupList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      8,
      f,
      proto.pps.Input.serializeBinaryToWriter
    );
  }
  f = message.getCrossList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.pps.Input.serializeBinaryToWriter
    );
  }
  f = message.getUnionList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      3,
      f,
      proto.pps.Input.serializeBinaryToWriter
    );
  }
  f = message.getCron();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pps.CronInput.serializeBinaryToWriter
    );
  }
  f = message.getGit();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.pps.GitInput.serializeBinaryToWriter
    );
  }
};


/**
 * optional PFSInput pfs = 6;
 * @return {?proto.pps.PFSInput}
 */
proto.pps.Input.prototype.getPfs = function() {
  return /** @type{?proto.pps.PFSInput} */ (
    jspb.Message.getWrapperField(this, proto.pps.PFSInput, 6));
};


/**
 * @param {?proto.pps.PFSInput|undefined} value
 * @return {!proto.pps.Input} returns this
*/
proto.pps.Input.prototype.setPfs = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.Input} returns this
 */
proto.pps.Input.prototype.clearPfs = function() {
  return this.setPfs(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.Input.prototype.hasPfs = function() {
  return jspb.Message.getField(this, 6) != null;
};


/**
 * repeated Input join = 7;
 * @return {!Array<!proto.pps.Input>}
 */
proto.pps.Input.prototype.getJoinList = function() {
  return /** @type{!Array<!proto.pps.Input>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps.Input, 7));
};


/**
 * @param {!Array<!proto.pps.Input>} value
 * @return {!proto.pps.Input} returns this
*/
proto.pps.Input.prototype.setJoinList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 7, value);
};


/**
 * @param {!proto.pps.Input=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps.Input}
 */
proto.pps.Input.prototype.addJoin = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 7, opt_value, proto.pps.Input, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.Input} returns this
 */
proto.pps.Input.prototype.clearJoinList = function() {
  return this.setJoinList([]);
};


/**
 * repeated Input group = 8;
 * @return {!Array<!proto.pps.Input>}
 */
proto.pps.Input.prototype.getGroupList = function() {
  return /** @type{!Array<!proto.pps.Input>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps.Input, 8));
};


/**
 * @param {!Array<!proto.pps.Input>} value
 * @return {!proto.pps.Input} returns this
*/
proto.pps.Input.prototype.setGroupList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 8, value);
};


/**
 * @param {!proto.pps.Input=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps.Input}
 */
proto.pps.Input.prototype.addGroup = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 8, opt_value, proto.pps.Input, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.Input} returns this
 */
proto.pps.Input.prototype.clearGroupList = function() {
  return this.setGroupList([]);
};


/**
 * repeated Input cross = 2;
 * @return {!Array<!proto.pps.Input>}
 */
proto.pps.Input.prototype.getCrossList = function() {
  return /** @type{!Array<!proto.pps.Input>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps.Input, 2));
};


/**
 * @param {!Array<!proto.pps.Input>} value
 * @return {!proto.pps.Input} returns this
*/
proto.pps.Input.prototype.setCrossList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pps.Input=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps.Input}
 */
proto.pps.Input.prototype.addCross = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pps.Input, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.Input} returns this
 */
proto.pps.Input.prototype.clearCrossList = function() {
  return this.setCrossList([]);
};


/**
 * repeated Input union = 3;
 * @return {!Array<!proto.pps.Input>}
 */
proto.pps.Input.prototype.getUnionList = function() {
  return /** @type{!Array<!proto.pps.Input>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps.Input, 3));
};


/**
 * @param {!Array<!proto.pps.Input>} value
 * @return {!proto.pps.Input} returns this
*/
proto.pps.Input.prototype.setUnionList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 3, value);
};


/**
 * @param {!proto.pps.Input=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps.Input}
 */
proto.pps.Input.prototype.addUnion = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 3, opt_value, proto.pps.Input, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.Input} returns this
 */
proto.pps.Input.prototype.clearUnionList = function() {
  return this.setUnionList([]);
};


/**
 * optional CronInput cron = 4;
 * @return {?proto.pps.CronInput}
 */
proto.pps.Input.prototype.getCron = function() {
  return /** @type{?proto.pps.CronInput} */ (
    jspb.Message.getWrapperField(this, proto.pps.CronInput, 4));
};


/**
 * @param {?proto.pps.CronInput|undefined} value
 * @return {!proto.pps.Input} returns this
*/
proto.pps.Input.prototype.setCron = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.Input} returns this
 */
proto.pps.Input.prototype.clearCron = function() {
  return this.setCron(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.Input.prototype.hasCron = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional GitInput git = 5;
 * @return {?proto.pps.GitInput}
 */
proto.pps.Input.prototype.getGit = function() {
  return /** @type{?proto.pps.GitInput} */ (
    jspb.Message.getWrapperField(this, proto.pps.GitInput, 5));
};


/**
 * @param {?proto.pps.GitInput|undefined} value
 * @return {!proto.pps.Input} returns this
*/
proto.pps.Input.prototype.setGit = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.Input} returns this
 */
proto.pps.Input.prototype.clearGit = function() {
  return this.setGit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.Input.prototype.hasGit = function() {
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
proto.pps.JobInput.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.JobInput.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.JobInput} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.JobInput.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 4, ""),
    commit: (f = msg.getCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    glob: jspb.Message.getFieldWithDefault(msg, 2, ""),
    lazy: jspb.Message.getBooleanFieldWithDefault(msg, 3, false)
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
 * @return {!proto.pps.JobInput}
 */
proto.pps.JobInput.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.JobInput;
  return proto.pps.JobInput.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.JobInput} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.JobInput}
 */
proto.pps.JobInput.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 1:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setCommit(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setGlob(value);
      break;
    case 3:
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
proto.pps.JobInput.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.JobInput.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.JobInput} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.JobInput.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getCommit();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getGlob();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getLazy();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
};


/**
 * optional string name = 4;
 * @return {string}
 */
proto.pps.JobInput.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.JobInput} returns this
 */
proto.pps.JobInput.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional pfs.Commit commit = 1;
 * @return {?proto.pfs.Commit}
 */
proto.pps.JobInput.prototype.getCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 1));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.JobInput} returns this
*/
proto.pps.JobInput.prototype.setCommit = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInput} returns this
 */
proto.pps.JobInput.prototype.clearCommit = function() {
  return this.setCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInput.prototype.hasCommit = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string glob = 2;
 * @return {string}
 */
proto.pps.JobInput.prototype.getGlob = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.JobInput} returns this
 */
proto.pps.JobInput.prototype.setGlob = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional bool lazy = 3;
 * @return {boolean}
 */
proto.pps.JobInput.prototype.getLazy = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.JobInput} returns this
 */
proto.pps.JobInput.prototype.setLazy = function(value) {
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
proto.pps.ParallelismSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.ParallelismSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.ParallelismSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ParallelismSpec.toObject = function(includeInstance, msg) {
  var f, obj = {
    constant: jspb.Message.getFieldWithDefault(msg, 2, 0),
    coefficient: jspb.Message.getFloatingPointFieldWithDefault(msg, 3, 0.0)
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
 * @return {!proto.pps.ParallelismSpec}
 */
proto.pps.ParallelismSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.ParallelismSpec;
  return proto.pps.ParallelismSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.ParallelismSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.ParallelismSpec}
 */
proto.pps.ParallelismSpec.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 2:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setConstant(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setCoefficient(value);
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
proto.pps.ParallelismSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.ParallelismSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.ParallelismSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ParallelismSpec.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getConstant();
  if (f !== 0) {
    writer.writeUint64(
      2,
      f
    );
  }
  f = message.getCoefficient();
  if (f !== 0.0) {
    writer.writeDouble(
      3,
      f
    );
  }
};


/**
 * optional uint64 constant = 2;
 * @return {number}
 */
proto.pps.ParallelismSpec.prototype.getConstant = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.ParallelismSpec} returns this
 */
proto.pps.ParallelismSpec.prototype.setConstant = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional double coefficient = 3;
 * @return {number}
 */
proto.pps.ParallelismSpec.prototype.getCoefficient = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 3, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.pps.ParallelismSpec} returns this
 */
proto.pps.ParallelismSpec.prototype.setCoefficient = function(value) {
  return jspb.Message.setProto3FloatField(this, 3, value);
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
proto.pps.InputFile.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.InputFile.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.InputFile} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.InputFile.toObject = function(includeInstance, msg) {
  var f, obj = {
    path: jspb.Message.getFieldWithDefault(msg, 4, ""),
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
 * @return {!proto.pps.InputFile}
 */
proto.pps.InputFile.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.InputFile;
  return proto.pps.InputFile.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.InputFile} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.InputFile}
 */
proto.pps.InputFile.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setPath(value);
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
proto.pps.InputFile.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.InputFile.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.InputFile} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.InputFile.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
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
 * optional string path = 4;
 * @return {string}
 */
proto.pps.InputFile.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.InputFile} returns this
 */
proto.pps.InputFile.prototype.setPath = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional bytes hash = 5;
 * @return {string}
 */
proto.pps.InputFile.prototype.getHash = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * optional bytes hash = 5;
 * This is a type-conversion wrapper around `getHash()`
 * @return {string}
 */
proto.pps.InputFile.prototype.getHash_asB64 = function() {
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
proto.pps.InputFile.prototype.getHash_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getHash()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pps.InputFile} returns this
 */
proto.pps.InputFile.prototype.setHash = function(value) {
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
proto.pps.Datum.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Datum.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Datum} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Datum.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    job: (f = msg.getJob()) && proto.pps.Job.toObject(includeInstance, f)
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
 * @return {!proto.pps.Datum}
 */
proto.pps.Datum.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Datum;
  return proto.pps.Datum.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Datum} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Datum}
 */
proto.pps.Datum.deserializeBinaryFromReader = function(msg, reader) {
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
      var value = new proto.pps.Job;
      reader.readMessage(value,proto.pps.Job.deserializeBinaryFromReader);
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
proto.pps.Datum.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Datum.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Datum} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Datum.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps.Job.serializeBinaryToWriter
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.pps.Datum.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Datum} returns this
 */
proto.pps.Datum.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional Job job = 2;
 * @return {?proto.pps.Job}
 */
proto.pps.Datum.prototype.getJob = function() {
  return /** @type{?proto.pps.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps.Job, 2));
};


/**
 * @param {?proto.pps.Job|undefined} value
 * @return {!proto.pps.Datum} returns this
*/
proto.pps.Datum.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.Datum} returns this
 */
proto.pps.Datum.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.Datum.prototype.hasJob = function() {
  return jspb.Message.getField(this, 2) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.DatumInfo.repeatedFields_ = [5];



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
proto.pps.DatumInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.DatumInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.DatumInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.DatumInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    datum: (f = msg.getDatum()) && proto.pps.Datum.toObject(includeInstance, f),
    state: jspb.Message.getFieldWithDefault(msg, 2, 0),
    stats: (f = msg.getStats()) && proto.pps.ProcessStats.toObject(includeInstance, f),
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
 * @return {!proto.pps.DatumInfo}
 */
proto.pps.DatumInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.DatumInfo;
  return proto.pps.DatumInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.DatumInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.DatumInfo}
 */
proto.pps.DatumInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Datum;
      reader.readMessage(value,proto.pps.Datum.deserializeBinaryFromReader);
      msg.setDatum(value);
      break;
    case 2:
      var value = /** @type {!proto.pps.DatumState} */ (reader.readEnum());
      msg.setState(value);
      break;
    case 3:
      var value = new proto.pps.ProcessStats;
      reader.readMessage(value,proto.pps.ProcessStats.deserializeBinaryFromReader);
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
proto.pps.DatumInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.DatumInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.DatumInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.DatumInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getDatum();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Datum.serializeBinaryToWriter
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
      proto.pps.ProcessStats.serializeBinaryToWriter
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
 * @return {?proto.pps.Datum}
 */
proto.pps.DatumInfo.prototype.getDatum = function() {
  return /** @type{?proto.pps.Datum} */ (
    jspb.Message.getWrapperField(this, proto.pps.Datum, 1));
};


/**
 * @param {?proto.pps.Datum|undefined} value
 * @return {!proto.pps.DatumInfo} returns this
*/
proto.pps.DatumInfo.prototype.setDatum = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.DatumInfo} returns this
 */
proto.pps.DatumInfo.prototype.clearDatum = function() {
  return this.setDatum(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.DatumInfo.prototype.hasDatum = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional DatumState state = 2;
 * @return {!proto.pps.DatumState}
 */
proto.pps.DatumInfo.prototype.getState = function() {
  return /** @type {!proto.pps.DatumState} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.pps.DatumState} value
 * @return {!proto.pps.DatumInfo} returns this
 */
proto.pps.DatumInfo.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 2, value);
};


/**
 * optional ProcessStats stats = 3;
 * @return {?proto.pps.ProcessStats}
 */
proto.pps.DatumInfo.prototype.getStats = function() {
  return /** @type{?proto.pps.ProcessStats} */ (
    jspb.Message.getWrapperField(this, proto.pps.ProcessStats, 3));
};


/**
 * @param {?proto.pps.ProcessStats|undefined} value
 * @return {!proto.pps.DatumInfo} returns this
*/
proto.pps.DatumInfo.prototype.setStats = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.DatumInfo} returns this
 */
proto.pps.DatumInfo.prototype.clearStats = function() {
  return this.setStats(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.DatumInfo.prototype.hasStats = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional pfs.File pfs_state = 4;
 * @return {?proto.pfs.File}
 */
proto.pps.DatumInfo.prototype.getPfsState = function() {
  return /** @type{?proto.pfs.File} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.File, 4));
};


/**
 * @param {?proto.pfs.File|undefined} value
 * @return {!proto.pps.DatumInfo} returns this
*/
proto.pps.DatumInfo.prototype.setPfsState = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.DatumInfo} returns this
 */
proto.pps.DatumInfo.prototype.clearPfsState = function() {
  return this.setPfsState(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.DatumInfo.prototype.hasPfsState = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * repeated pfs.FileInfo data = 5;
 * @return {!Array<!proto.pfs.FileInfo>}
 */
proto.pps.DatumInfo.prototype.getDataList = function() {
  return /** @type{!Array<!proto.pfs.FileInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, pfs_pfs_pb.FileInfo, 5));
};


/**
 * @param {!Array<!proto.pfs.FileInfo>} value
 * @return {!proto.pps.DatumInfo} returns this
*/
proto.pps.DatumInfo.prototype.setDataList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 5, value);
};


/**
 * @param {!proto.pfs.FileInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.FileInfo}
 */
proto.pps.DatumInfo.prototype.addData = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 5, opt_value, proto.pfs.FileInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.DatumInfo} returns this
 */
proto.pps.DatumInfo.prototype.clearDataList = function() {
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
proto.pps.Aggregate.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Aggregate.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Aggregate} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Aggregate.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.Aggregate}
 */
proto.pps.Aggregate.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Aggregate;
  return proto.pps.Aggregate.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Aggregate} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Aggregate}
 */
proto.pps.Aggregate.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.Aggregate.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Aggregate.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Aggregate} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Aggregate.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.Aggregate.prototype.getCount = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.Aggregate} returns this
 */
proto.pps.Aggregate.prototype.setCount = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional double mean = 2;
 * @return {number}
 */
proto.pps.Aggregate.prototype.getMean = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 2, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.pps.Aggregate} returns this
 */
proto.pps.Aggregate.prototype.setMean = function(value) {
  return jspb.Message.setProto3FloatField(this, 2, value);
};


/**
 * optional double stddev = 3;
 * @return {number}
 */
proto.pps.Aggregate.prototype.getStddev = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 3, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.pps.Aggregate} returns this
 */
proto.pps.Aggregate.prototype.setStddev = function(value) {
  return jspb.Message.setProto3FloatField(this, 3, value);
};


/**
 * optional double fifth_percentile = 4;
 * @return {number}
 */
proto.pps.Aggregate.prototype.getFifthPercentile = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 4, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.pps.Aggregate} returns this
 */
proto.pps.Aggregate.prototype.setFifthPercentile = function(value) {
  return jspb.Message.setProto3FloatField(this, 4, value);
};


/**
 * optional double ninety_fifth_percentile = 5;
 * @return {number}
 */
proto.pps.Aggregate.prototype.getNinetyFifthPercentile = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 5, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.pps.Aggregate} returns this
 */
proto.pps.Aggregate.prototype.setNinetyFifthPercentile = function(value) {
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
proto.pps.ProcessStats.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.ProcessStats.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.ProcessStats} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ProcessStats.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.ProcessStats}
 */
proto.pps.ProcessStats.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.ProcessStats;
  return proto.pps.ProcessStats.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.ProcessStats} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.ProcessStats}
 */
proto.pps.ProcessStats.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.ProcessStats.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.ProcessStats.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.ProcessStats} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ProcessStats.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.ProcessStats.prototype.getDownloadTime = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 1));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps.ProcessStats} returns this
*/
proto.pps.ProcessStats.prototype.setDownloadTime = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.ProcessStats} returns this
 */
proto.pps.ProcessStats.prototype.clearDownloadTime = function() {
  return this.setDownloadTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.ProcessStats.prototype.hasDownloadTime = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional google.protobuf.Duration process_time = 2;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps.ProcessStats.prototype.getProcessTime = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 2));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps.ProcessStats} returns this
*/
proto.pps.ProcessStats.prototype.setProcessTime = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.ProcessStats} returns this
 */
proto.pps.ProcessStats.prototype.clearProcessTime = function() {
  return this.setProcessTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.ProcessStats.prototype.hasProcessTime = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional google.protobuf.Duration upload_time = 3;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps.ProcessStats.prototype.getUploadTime = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 3));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps.ProcessStats} returns this
*/
proto.pps.ProcessStats.prototype.setUploadTime = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.ProcessStats} returns this
 */
proto.pps.ProcessStats.prototype.clearUploadTime = function() {
  return this.setUploadTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.ProcessStats.prototype.hasUploadTime = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional uint64 download_bytes = 4;
 * @return {number}
 */
proto.pps.ProcessStats.prototype.getDownloadBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.ProcessStats} returns this
 */
proto.pps.ProcessStats.prototype.setDownloadBytes = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional uint64 upload_bytes = 5;
 * @return {number}
 */
proto.pps.ProcessStats.prototype.getUploadBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.ProcessStats} returns this
 */
proto.pps.ProcessStats.prototype.setUploadBytes = function(value) {
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
proto.pps.AggregateProcessStats.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.AggregateProcessStats.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.AggregateProcessStats} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.AggregateProcessStats.toObject = function(includeInstance, msg) {
  var f, obj = {
    downloadTime: (f = msg.getDownloadTime()) && proto.pps.Aggregate.toObject(includeInstance, f),
    processTime: (f = msg.getProcessTime()) && proto.pps.Aggregate.toObject(includeInstance, f),
    uploadTime: (f = msg.getUploadTime()) && proto.pps.Aggregate.toObject(includeInstance, f),
    downloadBytes: (f = msg.getDownloadBytes()) && proto.pps.Aggregate.toObject(includeInstance, f),
    uploadBytes: (f = msg.getUploadBytes()) && proto.pps.Aggregate.toObject(includeInstance, f)
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
 * @return {!proto.pps.AggregateProcessStats}
 */
proto.pps.AggregateProcessStats.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.AggregateProcessStats;
  return proto.pps.AggregateProcessStats.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.AggregateProcessStats} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.AggregateProcessStats}
 */
proto.pps.AggregateProcessStats.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Aggregate;
      reader.readMessage(value,proto.pps.Aggregate.deserializeBinaryFromReader);
      msg.setDownloadTime(value);
      break;
    case 2:
      var value = new proto.pps.Aggregate;
      reader.readMessage(value,proto.pps.Aggregate.deserializeBinaryFromReader);
      msg.setProcessTime(value);
      break;
    case 3:
      var value = new proto.pps.Aggregate;
      reader.readMessage(value,proto.pps.Aggregate.deserializeBinaryFromReader);
      msg.setUploadTime(value);
      break;
    case 4:
      var value = new proto.pps.Aggregate;
      reader.readMessage(value,proto.pps.Aggregate.deserializeBinaryFromReader);
      msg.setDownloadBytes(value);
      break;
    case 5:
      var value = new proto.pps.Aggregate;
      reader.readMessage(value,proto.pps.Aggregate.deserializeBinaryFromReader);
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
proto.pps.AggregateProcessStats.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.AggregateProcessStats.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.AggregateProcessStats} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.AggregateProcessStats.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getDownloadTime();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Aggregate.serializeBinaryToWriter
    );
  }
  f = message.getProcessTime();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps.Aggregate.serializeBinaryToWriter
    );
  }
  f = message.getUploadTime();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pps.Aggregate.serializeBinaryToWriter
    );
  }
  f = message.getDownloadBytes();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      proto.pps.Aggregate.serializeBinaryToWriter
    );
  }
  f = message.getUploadBytes();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.pps.Aggregate.serializeBinaryToWriter
    );
  }
};


/**
 * optional Aggregate download_time = 1;
 * @return {?proto.pps.Aggregate}
 */
proto.pps.AggregateProcessStats.prototype.getDownloadTime = function() {
  return /** @type{?proto.pps.Aggregate} */ (
    jspb.Message.getWrapperField(this, proto.pps.Aggregate, 1));
};


/**
 * @param {?proto.pps.Aggregate|undefined} value
 * @return {!proto.pps.AggregateProcessStats} returns this
*/
proto.pps.AggregateProcessStats.prototype.setDownloadTime = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.AggregateProcessStats} returns this
 */
proto.pps.AggregateProcessStats.prototype.clearDownloadTime = function() {
  return this.setDownloadTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.AggregateProcessStats.prototype.hasDownloadTime = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Aggregate process_time = 2;
 * @return {?proto.pps.Aggregate}
 */
proto.pps.AggregateProcessStats.prototype.getProcessTime = function() {
  return /** @type{?proto.pps.Aggregate} */ (
    jspb.Message.getWrapperField(this, proto.pps.Aggregate, 2));
};


/**
 * @param {?proto.pps.Aggregate|undefined} value
 * @return {!proto.pps.AggregateProcessStats} returns this
*/
proto.pps.AggregateProcessStats.prototype.setProcessTime = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.AggregateProcessStats} returns this
 */
proto.pps.AggregateProcessStats.prototype.clearProcessTime = function() {
  return this.setProcessTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.AggregateProcessStats.prototype.hasProcessTime = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional Aggregate upload_time = 3;
 * @return {?proto.pps.Aggregate}
 */
proto.pps.AggregateProcessStats.prototype.getUploadTime = function() {
  return /** @type{?proto.pps.Aggregate} */ (
    jspb.Message.getWrapperField(this, proto.pps.Aggregate, 3));
};


/**
 * @param {?proto.pps.Aggregate|undefined} value
 * @return {!proto.pps.AggregateProcessStats} returns this
*/
proto.pps.AggregateProcessStats.prototype.setUploadTime = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.AggregateProcessStats} returns this
 */
proto.pps.AggregateProcessStats.prototype.clearUploadTime = function() {
  return this.setUploadTime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.AggregateProcessStats.prototype.hasUploadTime = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional Aggregate download_bytes = 4;
 * @return {?proto.pps.Aggregate}
 */
proto.pps.AggregateProcessStats.prototype.getDownloadBytes = function() {
  return /** @type{?proto.pps.Aggregate} */ (
    jspb.Message.getWrapperField(this, proto.pps.Aggregate, 4));
};


/**
 * @param {?proto.pps.Aggregate|undefined} value
 * @return {!proto.pps.AggregateProcessStats} returns this
*/
proto.pps.AggregateProcessStats.prototype.setDownloadBytes = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.AggregateProcessStats} returns this
 */
proto.pps.AggregateProcessStats.prototype.clearDownloadBytes = function() {
  return this.setDownloadBytes(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.AggregateProcessStats.prototype.hasDownloadBytes = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional Aggregate upload_bytes = 5;
 * @return {?proto.pps.Aggregate}
 */
proto.pps.AggregateProcessStats.prototype.getUploadBytes = function() {
  return /** @type{?proto.pps.Aggregate} */ (
    jspb.Message.getWrapperField(this, proto.pps.Aggregate, 5));
};


/**
 * @param {?proto.pps.Aggregate|undefined} value
 * @return {!proto.pps.AggregateProcessStats} returns this
*/
proto.pps.AggregateProcessStats.prototype.setUploadBytes = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.AggregateProcessStats} returns this
 */
proto.pps.AggregateProcessStats.prototype.clearUploadBytes = function() {
  return this.setUploadBytes(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.AggregateProcessStats.prototype.hasUploadBytes = function() {
  return jspb.Message.getField(this, 5) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.WorkerStatus.repeatedFields_ = [3];



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
proto.pps.WorkerStatus.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.WorkerStatus.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.WorkerStatus} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.WorkerStatus.toObject = function(includeInstance, msg) {
  var f, obj = {
    workerId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    jobId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    dataList: jspb.Message.toObjectList(msg.getDataList(),
    proto.pps.InputFile.toObject, includeInstance),
    started: (f = msg.getStarted()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    stats: (f = msg.getStats()) && proto.pps.ProcessStats.toObject(includeInstance, f),
    queueSize: jspb.Message.getFieldWithDefault(msg, 6, 0),
    dataProcessed: jspb.Message.getFieldWithDefault(msg, 7, 0),
    dataRecovered: jspb.Message.getFieldWithDefault(msg, 8, 0)
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
 * @return {!proto.pps.WorkerStatus}
 */
proto.pps.WorkerStatus.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.WorkerStatus;
  return proto.pps.WorkerStatus.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.WorkerStatus} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.WorkerStatus}
 */
proto.pps.WorkerStatus.deserializeBinaryFromReader = function(msg, reader) {
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
      var value = new proto.pps.InputFile;
      reader.readMessage(value,proto.pps.InputFile.deserializeBinaryFromReader);
      msg.addData(value);
      break;
    case 4:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setStarted(value);
      break;
    case 5:
      var value = new proto.pps.ProcessStats;
      reader.readMessage(value,proto.pps.ProcessStats.deserializeBinaryFromReader);
      msg.setStats(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setQueueSize(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataProcessed(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataRecovered(value);
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
proto.pps.WorkerStatus.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.WorkerStatus.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.WorkerStatus} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.WorkerStatus.serializeBinaryToWriter = function(message, writer) {
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
  f = message.getDataList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      3,
      f,
      proto.pps.InputFile.serializeBinaryToWriter
    );
  }
  f = message.getStarted();
  if (f != null) {
    writer.writeMessage(
      4,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getStats();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.pps.ProcessStats.serializeBinaryToWriter
    );
  }
  f = message.getQueueSize();
  if (f !== 0) {
    writer.writeInt64(
      6,
      f
    );
  }
  f = message.getDataProcessed();
  if (f !== 0) {
    writer.writeInt64(
      7,
      f
    );
  }
  f = message.getDataRecovered();
  if (f !== 0) {
    writer.writeInt64(
      8,
      f
    );
  }
};


/**
 * optional string worker_id = 1;
 * @return {string}
 */
proto.pps.WorkerStatus.prototype.getWorkerId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.WorkerStatus} returns this
 */
proto.pps.WorkerStatus.prototype.setWorkerId = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string job_id = 2;
 * @return {string}
 */
proto.pps.WorkerStatus.prototype.getJobId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.WorkerStatus} returns this
 */
proto.pps.WorkerStatus.prototype.setJobId = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * repeated InputFile data = 3;
 * @return {!Array<!proto.pps.InputFile>}
 */
proto.pps.WorkerStatus.prototype.getDataList = function() {
  return /** @type{!Array<!proto.pps.InputFile>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps.InputFile, 3));
};


/**
 * @param {!Array<!proto.pps.InputFile>} value
 * @return {!proto.pps.WorkerStatus} returns this
*/
proto.pps.WorkerStatus.prototype.setDataList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 3, value);
};


/**
 * @param {!proto.pps.InputFile=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps.InputFile}
 */
proto.pps.WorkerStatus.prototype.addData = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 3, opt_value, proto.pps.InputFile, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.WorkerStatus} returns this
 */
proto.pps.WorkerStatus.prototype.clearDataList = function() {
  return this.setDataList([]);
};


/**
 * optional google.protobuf.Timestamp started = 4;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps.WorkerStatus.prototype.getStarted = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 4));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps.WorkerStatus} returns this
*/
proto.pps.WorkerStatus.prototype.setStarted = function(value) {
  return jspb.Message.setWrapperField(this, 4, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.WorkerStatus} returns this
 */
proto.pps.WorkerStatus.prototype.clearStarted = function() {
  return this.setStarted(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.WorkerStatus.prototype.hasStarted = function() {
  return jspb.Message.getField(this, 4) != null;
};


/**
 * optional ProcessStats stats = 5;
 * @return {?proto.pps.ProcessStats}
 */
proto.pps.WorkerStatus.prototype.getStats = function() {
  return /** @type{?proto.pps.ProcessStats} */ (
    jspb.Message.getWrapperField(this, proto.pps.ProcessStats, 5));
};


/**
 * @param {?proto.pps.ProcessStats|undefined} value
 * @return {!proto.pps.WorkerStatus} returns this
*/
proto.pps.WorkerStatus.prototype.setStats = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.WorkerStatus} returns this
 */
proto.pps.WorkerStatus.prototype.clearStats = function() {
  return this.setStats(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.WorkerStatus.prototype.hasStats = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional int64 queue_size = 6;
 * @return {number}
 */
proto.pps.WorkerStatus.prototype.getQueueSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.WorkerStatus} returns this
 */
proto.pps.WorkerStatus.prototype.setQueueSize = function(value) {
  return jspb.Message.setProto3IntField(this, 6, value);
};


/**
 * optional int64 data_processed = 7;
 * @return {number}
 */
proto.pps.WorkerStatus.prototype.getDataProcessed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.WorkerStatus} returns this
 */
proto.pps.WorkerStatus.prototype.setDataProcessed = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
};


/**
 * optional int64 data_recovered = 8;
 * @return {number}
 */
proto.pps.WorkerStatus.prototype.getDataRecovered = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.WorkerStatus} returns this
 */
proto.pps.WorkerStatus.prototype.setDataRecovered = function(value) {
  return jspb.Message.setProto3IntField(this, 8, value);
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
proto.pps.ResourceSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.ResourceSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.ResourceSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ResourceSpec.toObject = function(includeInstance, msg) {
  var f, obj = {
    cpu: jspb.Message.getFloatingPointFieldWithDefault(msg, 1, 0.0),
    memory: jspb.Message.getFieldWithDefault(msg, 2, ""),
    gpu: (f = msg.getGpu()) && proto.pps.GPUSpec.toObject(includeInstance, f),
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
 * @return {!proto.pps.ResourceSpec}
 */
proto.pps.ResourceSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.ResourceSpec;
  return proto.pps.ResourceSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.ResourceSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.ResourceSpec}
 */
proto.pps.ResourceSpec.deserializeBinaryFromReader = function(msg, reader) {
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
    case 5:
      var value = new proto.pps.GPUSpec;
      reader.readMessage(value,proto.pps.GPUSpec.deserializeBinaryFromReader);
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
proto.pps.ResourceSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.ResourceSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.ResourceSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ResourceSpec.serializeBinaryToWriter = function(message, writer) {
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
      5,
      f,
      proto.pps.GPUSpec.serializeBinaryToWriter
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
proto.pps.ResourceSpec.prototype.getCpu = function() {
  return /** @type {number} */ (jspb.Message.getFloatingPointFieldWithDefault(this, 1, 0.0));
};


/**
 * @param {number} value
 * @return {!proto.pps.ResourceSpec} returns this
 */
proto.pps.ResourceSpec.prototype.setCpu = function(value) {
  return jspb.Message.setProto3FloatField(this, 1, value);
};


/**
 * optional string memory = 2;
 * @return {string}
 */
proto.pps.ResourceSpec.prototype.getMemory = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.ResourceSpec} returns this
 */
proto.pps.ResourceSpec.prototype.setMemory = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional GPUSpec gpu = 5;
 * @return {?proto.pps.GPUSpec}
 */
proto.pps.ResourceSpec.prototype.getGpu = function() {
  return /** @type{?proto.pps.GPUSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.GPUSpec, 5));
};


/**
 * @param {?proto.pps.GPUSpec|undefined} value
 * @return {!proto.pps.ResourceSpec} returns this
*/
proto.pps.ResourceSpec.prototype.setGpu = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.ResourceSpec} returns this
 */
proto.pps.ResourceSpec.prototype.clearGpu = function() {
  return this.setGpu(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.ResourceSpec.prototype.hasGpu = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional string disk = 4;
 * @return {string}
 */
proto.pps.ResourceSpec.prototype.getDisk = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.ResourceSpec} returns this
 */
proto.pps.ResourceSpec.prototype.setDisk = function(value) {
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
proto.pps.GPUSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.GPUSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.GPUSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.GPUSpec.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.GPUSpec}
 */
proto.pps.GPUSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.GPUSpec;
  return proto.pps.GPUSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.GPUSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.GPUSpec}
 */
proto.pps.GPUSpec.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.GPUSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.GPUSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.GPUSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.GPUSpec.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.GPUSpec.prototype.getType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.GPUSpec} returns this
 */
proto.pps.GPUSpec.prototype.setType = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional int64 number = 2;
 * @return {number}
 */
proto.pps.GPUSpec.prototype.getNumber = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.GPUSpec} returns this
 */
proto.pps.GPUSpec.prototype.setNumber = function(value) {
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
proto.pps.EtcdJobInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.EtcdJobInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.EtcdJobInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.EtcdJobInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps.Job.toObject(includeInstance, f),
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f),
    outputCommit: (f = msg.getOutputCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    restart: jspb.Message.getFieldWithDefault(msg, 4, 0),
    dataProcessed: jspb.Message.getFieldWithDefault(msg, 5, 0),
    dataSkipped: jspb.Message.getFieldWithDefault(msg, 6, 0),
    dataTotal: jspb.Message.getFieldWithDefault(msg, 7, 0),
    dataFailed: jspb.Message.getFieldWithDefault(msg, 8, 0),
    dataRecovered: jspb.Message.getFieldWithDefault(msg, 15, 0),
    stats: (f = msg.getStats()) && proto.pps.ProcessStats.toObject(includeInstance, f),
    statsCommit: (f = msg.getStatsCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    state: jspb.Message.getFieldWithDefault(msg, 11, 0),
    reason: jspb.Message.getFieldWithDefault(msg, 12, ""),
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
 * @return {!proto.pps.EtcdJobInfo}
 */
proto.pps.EtcdJobInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.EtcdJobInfo;
  return proto.pps.EtcdJobInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.EtcdJobInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.EtcdJobInfo}
 */
proto.pps.EtcdJobInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Job;
      reader.readMessage(value,proto.pps.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 2:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
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
    case 15:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataRecovered(value);
      break;
    case 9:
      var value = new proto.pps.ProcessStats;
      reader.readMessage(value,proto.pps.ProcessStats.deserializeBinaryFromReader);
      msg.setStats(value);
      break;
    case 10:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setStatsCommit(value);
      break;
    case 11:
      var value = /** @type {!proto.pps.JobState} */ (reader.readEnum());
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
proto.pps.EtcdJobInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.EtcdJobInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.EtcdJobInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.EtcdJobInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Job.serializeBinaryToWriter
    );
  }
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
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
      15,
      f
    );
  }
  f = message.getStats();
  if (f != null) {
    writer.writeMessage(
      9,
      f,
      proto.pps.ProcessStats.serializeBinaryToWriter
    );
  }
  f = message.getStatsCommit();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
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
};


/**
 * optional Job job = 1;
 * @return {?proto.pps.Job}
 */
proto.pps.EtcdJobInfo.prototype.getJob = function() {
  return /** @type{?proto.pps.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps.Job, 1));
};


/**
 * @param {?proto.pps.Job|undefined} value
 * @return {!proto.pps.EtcdJobInfo} returns this
*/
proto.pps.EtcdJobInfo.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.EtcdJobInfo.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Pipeline pipeline = 2;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.EtcdJobInfo.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 2));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.EtcdJobInfo} returns this
*/
proto.pps.EtcdJobInfo.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.EtcdJobInfo.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional pfs.Commit output_commit = 3;
 * @return {?proto.pfs.Commit}
 */
proto.pps.EtcdJobInfo.prototype.getOutputCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 3));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.EtcdJobInfo} returns this
*/
proto.pps.EtcdJobInfo.prototype.setOutputCommit = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.clearOutputCommit = function() {
  return this.setOutputCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.EtcdJobInfo.prototype.hasOutputCommit = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional uint64 restart = 4;
 * @return {number}
 */
proto.pps.EtcdJobInfo.prototype.getRestart = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.setRestart = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional int64 data_processed = 5;
 * @return {number}
 */
proto.pps.EtcdJobInfo.prototype.getDataProcessed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.setDataProcessed = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};


/**
 * optional int64 data_skipped = 6;
 * @return {number}
 */
proto.pps.EtcdJobInfo.prototype.getDataSkipped = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.setDataSkipped = function(value) {
  return jspb.Message.setProto3IntField(this, 6, value);
};


/**
 * optional int64 data_total = 7;
 * @return {number}
 */
proto.pps.EtcdJobInfo.prototype.getDataTotal = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.setDataTotal = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
};


/**
 * optional int64 data_failed = 8;
 * @return {number}
 */
proto.pps.EtcdJobInfo.prototype.getDataFailed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.setDataFailed = function(value) {
  return jspb.Message.setProto3IntField(this, 8, value);
};


/**
 * optional int64 data_recovered = 15;
 * @return {number}
 */
proto.pps.EtcdJobInfo.prototype.getDataRecovered = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 15, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.setDataRecovered = function(value) {
  return jspb.Message.setProto3IntField(this, 15, value);
};


/**
 * optional ProcessStats stats = 9;
 * @return {?proto.pps.ProcessStats}
 */
proto.pps.EtcdJobInfo.prototype.getStats = function() {
  return /** @type{?proto.pps.ProcessStats} */ (
    jspb.Message.getWrapperField(this, proto.pps.ProcessStats, 9));
};


/**
 * @param {?proto.pps.ProcessStats|undefined} value
 * @return {!proto.pps.EtcdJobInfo} returns this
*/
proto.pps.EtcdJobInfo.prototype.setStats = function(value) {
  return jspb.Message.setWrapperField(this, 9, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.clearStats = function() {
  return this.setStats(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.EtcdJobInfo.prototype.hasStats = function() {
  return jspb.Message.getField(this, 9) != null;
};


/**
 * optional pfs.Commit stats_commit = 10;
 * @return {?proto.pfs.Commit}
 */
proto.pps.EtcdJobInfo.prototype.getStatsCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 10));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.EtcdJobInfo} returns this
*/
proto.pps.EtcdJobInfo.prototype.setStatsCommit = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.clearStatsCommit = function() {
  return this.setStatsCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.EtcdJobInfo.prototype.hasStatsCommit = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * optional JobState state = 11;
 * @return {!proto.pps.JobState}
 */
proto.pps.EtcdJobInfo.prototype.getState = function() {
  return /** @type {!proto.pps.JobState} */ (jspb.Message.getFieldWithDefault(this, 11, 0));
};


/**
 * @param {!proto.pps.JobState} value
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 11, value);
};


/**
 * optional string reason = 12;
 * @return {string}
 */
proto.pps.EtcdJobInfo.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 12, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.setReason = function(value) {
  return jspb.Message.setProto3StringField(this, 12, value);
};


/**
 * optional google.protobuf.Timestamp started = 13;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps.EtcdJobInfo.prototype.getStarted = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 13));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps.EtcdJobInfo} returns this
*/
proto.pps.EtcdJobInfo.prototype.setStarted = function(value) {
  return jspb.Message.setWrapperField(this, 13, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.clearStarted = function() {
  return this.setStarted(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.EtcdJobInfo.prototype.hasStarted = function() {
  return jspb.Message.getField(this, 13) != null;
};


/**
 * optional google.protobuf.Timestamp finished = 14;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps.EtcdJobInfo.prototype.getFinished = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 14));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps.EtcdJobInfo} returns this
*/
proto.pps.EtcdJobInfo.prototype.setFinished = function(value) {
  return jspb.Message.setWrapperField(this, 14, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.EtcdJobInfo} returns this
 */
proto.pps.EtcdJobInfo.prototype.clearFinished = function() {
  return this.setFinished(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.EtcdJobInfo.prototype.hasFinished = function() {
  return jspb.Message.getField(this, 14) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.JobInfo.repeatedFields_ = [24];



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
proto.pps.JobInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.JobInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.JobInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.JobInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps.Job.toObject(includeInstance, f),
    transform: (f = msg.getTransform()) && proto.pps.Transform.toObject(includeInstance, f),
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f),
    pipelineVersion: jspb.Message.getFieldWithDefault(msg, 13, 0),
    specCommit: (f = msg.getSpecCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    parallelismSpec: (f = msg.getParallelismSpec()) && proto.pps.ParallelismSpec.toObject(includeInstance, f),
    egress: (f = msg.getEgress()) && proto.pps.Egress.toObject(includeInstance, f),
    parentJob: (f = msg.getParentJob()) && proto.pps.Job.toObject(includeInstance, f),
    started: (f = msg.getStarted()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    finished: (f = msg.getFinished()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    outputCommit: (f = msg.getOutputCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    state: jspb.Message.getFieldWithDefault(msg, 10, 0),
    reason: jspb.Message.getFieldWithDefault(msg, 35, ""),
    service: (f = msg.getService()) && proto.pps.Service.toObject(includeInstance, f),
    spout: (f = msg.getSpout()) && proto.pps.Spout.toObject(includeInstance, f),
    outputRepo: (f = msg.getOutputRepo()) && pfs_pfs_pb.Repo.toObject(includeInstance, f),
    outputBranch: jspb.Message.getFieldWithDefault(msg, 17, ""),
    restart: jspb.Message.getFieldWithDefault(msg, 20, 0),
    dataProcessed: jspb.Message.getFieldWithDefault(msg, 22, 0),
    dataSkipped: jspb.Message.getFieldWithDefault(msg, 30, 0),
    dataFailed: jspb.Message.getFieldWithDefault(msg, 40, 0),
    dataRecovered: jspb.Message.getFieldWithDefault(msg, 46, 0),
    dataTotal: jspb.Message.getFieldWithDefault(msg, 23, 0),
    stats: (f = msg.getStats()) && proto.pps.ProcessStats.toObject(includeInstance, f),
    workerStatusList: jspb.Message.toObjectList(msg.getWorkerStatusList(),
    proto.pps.WorkerStatus.toObject, includeInstance),
    resourceRequests: (f = msg.getResourceRequests()) && proto.pps.ResourceSpec.toObject(includeInstance, f),
    resourceLimits: (f = msg.getResourceLimits()) && proto.pps.ResourceSpec.toObject(includeInstance, f),
    sidecarResourceLimits: (f = msg.getSidecarResourceLimits()) && proto.pps.ResourceSpec.toObject(includeInstance, f),
    input: (f = msg.getInput()) && proto.pps.Input.toObject(includeInstance, f),
    newBranch: (f = msg.getNewBranch()) && pfs_pfs_pb.BranchInfo.toObject(includeInstance, f),
    statsCommit: (f = msg.getStatsCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    enableStats: jspb.Message.getBooleanFieldWithDefault(msg, 32, false),
    salt: jspb.Message.getFieldWithDefault(msg, 33, ""),
    chunkSpec: (f = msg.getChunkSpec()) && proto.pps.ChunkSpec.toObject(includeInstance, f),
    datumTimeout: (f = msg.getDatumTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    jobTimeout: (f = msg.getJobTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    datumTries: jspb.Message.getFieldWithDefault(msg, 41, 0),
    schedulingSpec: (f = msg.getSchedulingSpec()) && proto.pps.SchedulingSpec.toObject(includeInstance, f),
    podSpec: jspb.Message.getFieldWithDefault(msg, 43, ""),
    podPatch: jspb.Message.getFieldWithDefault(msg, 44, "")
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
 * @return {!proto.pps.JobInfo}
 */
proto.pps.JobInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.JobInfo;
  return proto.pps.JobInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.JobInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.JobInfo}
 */
proto.pps.JobInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Job;
      reader.readMessage(value,proto.pps.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 2:
      var value = new proto.pps.Transform;
      reader.readMessage(value,proto.pps.Transform.deserializeBinaryFromReader);
      msg.setTransform(value);
      break;
    case 3:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 13:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setPipelineVersion(value);
      break;
    case 47:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setSpecCommit(value);
      break;
    case 12:
      var value = new proto.pps.ParallelismSpec;
      reader.readMessage(value,proto.pps.ParallelismSpec.deserializeBinaryFromReader);
      msg.setParallelismSpec(value);
      break;
    case 15:
      var value = new proto.pps.Egress;
      reader.readMessage(value,proto.pps.Egress.deserializeBinaryFromReader);
      msg.setEgress(value);
      break;
    case 6:
      var value = new proto.pps.Job;
      reader.readMessage(value,proto.pps.Job.deserializeBinaryFromReader);
      msg.setParentJob(value);
      break;
    case 7:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setStarted(value);
      break;
    case 8:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setFinished(value);
      break;
    case 9:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setOutputCommit(value);
      break;
    case 10:
      var value = /** @type {!proto.pps.JobState} */ (reader.readEnum());
      msg.setState(value);
      break;
    case 35:
      var value = /** @type {string} */ (reader.readString());
      msg.setReason(value);
      break;
    case 14:
      var value = new proto.pps.Service;
      reader.readMessage(value,proto.pps.Service.deserializeBinaryFromReader);
      msg.setService(value);
      break;
    case 45:
      var value = new proto.pps.Spout;
      reader.readMessage(value,proto.pps.Spout.deserializeBinaryFromReader);
      msg.setSpout(value);
      break;
    case 18:
      var value = new pfs_pfs_pb.Repo;
      reader.readMessage(value,pfs_pfs_pb.Repo.deserializeBinaryFromReader);
      msg.setOutputRepo(value);
      break;
    case 17:
      var value = /** @type {string} */ (reader.readString());
      msg.setOutputBranch(value);
      break;
    case 20:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setRestart(value);
      break;
    case 22:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataProcessed(value);
      break;
    case 30:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataSkipped(value);
      break;
    case 40:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataFailed(value);
      break;
    case 46:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataRecovered(value);
      break;
    case 23:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataTotal(value);
      break;
    case 31:
      var value = new proto.pps.ProcessStats;
      reader.readMessage(value,proto.pps.ProcessStats.deserializeBinaryFromReader);
      msg.setStats(value);
      break;
    case 24:
      var value = new proto.pps.WorkerStatus;
      reader.readMessage(value,proto.pps.WorkerStatus.deserializeBinaryFromReader);
      msg.addWorkerStatus(value);
      break;
    case 25:
      var value = new proto.pps.ResourceSpec;
      reader.readMessage(value,proto.pps.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceRequests(value);
      break;
    case 36:
      var value = new proto.pps.ResourceSpec;
      reader.readMessage(value,proto.pps.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceLimits(value);
      break;
    case 48:
      var value = new proto.pps.ResourceSpec;
      reader.readMessage(value,proto.pps.ResourceSpec.deserializeBinaryFromReader);
      msg.setSidecarResourceLimits(value);
      break;
    case 26:
      var value = new proto.pps.Input;
      reader.readMessage(value,proto.pps.Input.deserializeBinaryFromReader);
      msg.setInput(value);
      break;
    case 27:
      var value = new pfs_pfs_pb.BranchInfo;
      reader.readMessage(value,pfs_pfs_pb.BranchInfo.deserializeBinaryFromReader);
      msg.setNewBranch(value);
      break;
    case 29:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setStatsCommit(value);
      break;
    case 32:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEnableStats(value);
      break;
    case 33:
      var value = /** @type {string} */ (reader.readString());
      msg.setSalt(value);
      break;
    case 37:
      var value = new proto.pps.ChunkSpec;
      reader.readMessage(value,proto.pps.ChunkSpec.deserializeBinaryFromReader);
      msg.setChunkSpec(value);
      break;
    case 38:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setDatumTimeout(value);
      break;
    case 39:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setJobTimeout(value);
      break;
    case 41:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDatumTries(value);
      break;
    case 42:
      var value = new proto.pps.SchedulingSpec;
      reader.readMessage(value,proto.pps.SchedulingSpec.deserializeBinaryFromReader);
      msg.setSchedulingSpec(value);
      break;
    case 43:
      var value = /** @type {string} */ (reader.readString());
      msg.setPodSpec(value);
      break;
    case 44:
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
proto.pps.JobInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.JobInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.JobInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.JobInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Job.serializeBinaryToWriter
    );
  }
  f = message.getTransform();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps.Transform.serializeBinaryToWriter
    );
  }
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getPipelineVersion();
  if (f !== 0) {
    writer.writeUint64(
      13,
      f
    );
  }
  f = message.getSpecCommit();
  if (f != null) {
    writer.writeMessage(
      47,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getParallelismSpec();
  if (f != null) {
    writer.writeMessage(
      12,
      f,
      proto.pps.ParallelismSpec.serializeBinaryToWriter
    );
  }
  f = message.getEgress();
  if (f != null) {
    writer.writeMessage(
      15,
      f,
      proto.pps.Egress.serializeBinaryToWriter
    );
  }
  f = message.getParentJob();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      proto.pps.Job.serializeBinaryToWriter
    );
  }
  f = message.getStarted();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getFinished();
  if (f != null) {
    writer.writeMessage(
      8,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getOutputCommit();
  if (f != null) {
    writer.writeMessage(
      9,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      10,
      f
    );
  }
  f = message.getReason();
  if (f.length > 0) {
    writer.writeString(
      35,
      f
    );
  }
  f = message.getService();
  if (f != null) {
    writer.writeMessage(
      14,
      f,
      proto.pps.Service.serializeBinaryToWriter
    );
  }
  f = message.getSpout();
  if (f != null) {
    writer.writeMessage(
      45,
      f,
      proto.pps.Spout.serializeBinaryToWriter
    );
  }
  f = message.getOutputRepo();
  if (f != null) {
    writer.writeMessage(
      18,
      f,
      pfs_pfs_pb.Repo.serializeBinaryToWriter
    );
  }
  f = message.getOutputBranch();
  if (f.length > 0) {
    writer.writeString(
      17,
      f
    );
  }
  f = message.getRestart();
  if (f !== 0) {
    writer.writeUint64(
      20,
      f
    );
  }
  f = message.getDataProcessed();
  if (f !== 0) {
    writer.writeInt64(
      22,
      f
    );
  }
  f = message.getDataSkipped();
  if (f !== 0) {
    writer.writeInt64(
      30,
      f
    );
  }
  f = message.getDataFailed();
  if (f !== 0) {
    writer.writeInt64(
      40,
      f
    );
  }
  f = message.getDataRecovered();
  if (f !== 0) {
    writer.writeInt64(
      46,
      f
    );
  }
  f = message.getDataTotal();
  if (f !== 0) {
    writer.writeInt64(
      23,
      f
    );
  }
  f = message.getStats();
  if (f != null) {
    writer.writeMessage(
      31,
      f,
      proto.pps.ProcessStats.serializeBinaryToWriter
    );
  }
  f = message.getWorkerStatusList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      24,
      f,
      proto.pps.WorkerStatus.serializeBinaryToWriter
    );
  }
  f = message.getResourceRequests();
  if (f != null) {
    writer.writeMessage(
      25,
      f,
      proto.pps.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getResourceLimits();
  if (f != null) {
    writer.writeMessage(
      36,
      f,
      proto.pps.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getSidecarResourceLimits();
  if (f != null) {
    writer.writeMessage(
      48,
      f,
      proto.pps.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getInput();
  if (f != null) {
    writer.writeMessage(
      26,
      f,
      proto.pps.Input.serializeBinaryToWriter
    );
  }
  f = message.getNewBranch();
  if (f != null) {
    writer.writeMessage(
      27,
      f,
      pfs_pfs_pb.BranchInfo.serializeBinaryToWriter
    );
  }
  f = message.getStatsCommit();
  if (f != null) {
    writer.writeMessage(
      29,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getEnableStats();
  if (f) {
    writer.writeBool(
      32,
      f
    );
  }
  f = message.getSalt();
  if (f.length > 0) {
    writer.writeString(
      33,
      f
    );
  }
  f = message.getChunkSpec();
  if (f != null) {
    writer.writeMessage(
      37,
      f,
      proto.pps.ChunkSpec.serializeBinaryToWriter
    );
  }
  f = message.getDatumTimeout();
  if (f != null) {
    writer.writeMessage(
      38,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getJobTimeout();
  if (f != null) {
    writer.writeMessage(
      39,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getDatumTries();
  if (f !== 0) {
    writer.writeInt64(
      41,
      f
    );
  }
  f = message.getSchedulingSpec();
  if (f != null) {
    writer.writeMessage(
      42,
      f,
      proto.pps.SchedulingSpec.serializeBinaryToWriter
    );
  }
  f = message.getPodSpec();
  if (f.length > 0) {
    writer.writeString(
      43,
      f
    );
  }
  f = message.getPodPatch();
  if (f.length > 0) {
    writer.writeString(
      44,
      f
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps.Job}
 */
proto.pps.JobInfo.prototype.getJob = function() {
  return /** @type{?proto.pps.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps.Job, 1));
};


/**
 * @param {?proto.pps.Job|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Transform transform = 2;
 * @return {?proto.pps.Transform}
 */
proto.pps.JobInfo.prototype.getTransform = function() {
  return /** @type{?proto.pps.Transform} */ (
    jspb.Message.getWrapperField(this, proto.pps.Transform, 2));
};


/**
 * @param {?proto.pps.Transform|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setTransform = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearTransform = function() {
  return this.setTransform(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasTransform = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional Pipeline pipeline = 3;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.JobInfo.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 3));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional uint64 pipeline_version = 13;
 * @return {number}
 */
proto.pps.JobInfo.prototype.getPipelineVersion = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 13, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setPipelineVersion = function(value) {
  return jspb.Message.setProto3IntField(this, 13, value);
};


/**
 * optional pfs.Commit spec_commit = 47;
 * @return {?proto.pfs.Commit}
 */
proto.pps.JobInfo.prototype.getSpecCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 47));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setSpecCommit = function(value) {
  return jspb.Message.setWrapperField(this, 47, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearSpecCommit = function() {
  return this.setSpecCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasSpecCommit = function() {
  return jspb.Message.getField(this, 47) != null;
};


/**
 * optional ParallelismSpec parallelism_spec = 12;
 * @return {?proto.pps.ParallelismSpec}
 */
proto.pps.JobInfo.prototype.getParallelismSpec = function() {
  return /** @type{?proto.pps.ParallelismSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ParallelismSpec, 12));
};


/**
 * @param {?proto.pps.ParallelismSpec|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setParallelismSpec = function(value) {
  return jspb.Message.setWrapperField(this, 12, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearParallelismSpec = function() {
  return this.setParallelismSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasParallelismSpec = function() {
  return jspb.Message.getField(this, 12) != null;
};


/**
 * optional Egress egress = 15;
 * @return {?proto.pps.Egress}
 */
proto.pps.JobInfo.prototype.getEgress = function() {
  return /** @type{?proto.pps.Egress} */ (
    jspb.Message.getWrapperField(this, proto.pps.Egress, 15));
};


/**
 * @param {?proto.pps.Egress|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setEgress = function(value) {
  return jspb.Message.setWrapperField(this, 15, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearEgress = function() {
  return this.setEgress(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasEgress = function() {
  return jspb.Message.getField(this, 15) != null;
};


/**
 * optional Job parent_job = 6;
 * @return {?proto.pps.Job}
 */
proto.pps.JobInfo.prototype.getParentJob = function() {
  return /** @type{?proto.pps.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps.Job, 6));
};


/**
 * @param {?proto.pps.Job|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setParentJob = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearParentJob = function() {
  return this.setParentJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasParentJob = function() {
  return jspb.Message.getField(this, 6) != null;
};


/**
 * optional google.protobuf.Timestamp started = 7;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps.JobInfo.prototype.getStarted = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 7));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setStarted = function(value) {
  return jspb.Message.setWrapperField(this, 7, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearStarted = function() {
  return this.setStarted(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasStarted = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * optional google.protobuf.Timestamp finished = 8;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps.JobInfo.prototype.getFinished = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 8));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setFinished = function(value) {
  return jspb.Message.setWrapperField(this, 8, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearFinished = function() {
  return this.setFinished(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasFinished = function() {
  return jspb.Message.getField(this, 8) != null;
};


/**
 * optional pfs.Commit output_commit = 9;
 * @return {?proto.pfs.Commit}
 */
proto.pps.JobInfo.prototype.getOutputCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 9));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setOutputCommit = function(value) {
  return jspb.Message.setWrapperField(this, 9, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearOutputCommit = function() {
  return this.setOutputCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasOutputCommit = function() {
  return jspb.Message.getField(this, 9) != null;
};


/**
 * optional JobState state = 10;
 * @return {!proto.pps.JobState}
 */
proto.pps.JobInfo.prototype.getState = function() {
  return /** @type {!proto.pps.JobState} */ (jspb.Message.getFieldWithDefault(this, 10, 0));
};


/**
 * @param {!proto.pps.JobState} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 10, value);
};


/**
 * optional string reason = 35;
 * @return {string}
 */
proto.pps.JobInfo.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 35, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setReason = function(value) {
  return jspb.Message.setProto3StringField(this, 35, value);
};


/**
 * optional Service service = 14;
 * @return {?proto.pps.Service}
 */
proto.pps.JobInfo.prototype.getService = function() {
  return /** @type{?proto.pps.Service} */ (
    jspb.Message.getWrapperField(this, proto.pps.Service, 14));
};


/**
 * @param {?proto.pps.Service|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setService = function(value) {
  return jspb.Message.setWrapperField(this, 14, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearService = function() {
  return this.setService(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasService = function() {
  return jspb.Message.getField(this, 14) != null;
};


/**
 * optional Spout spout = 45;
 * @return {?proto.pps.Spout}
 */
proto.pps.JobInfo.prototype.getSpout = function() {
  return /** @type{?proto.pps.Spout} */ (
    jspb.Message.getWrapperField(this, proto.pps.Spout, 45));
};


/**
 * @param {?proto.pps.Spout|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setSpout = function(value) {
  return jspb.Message.setWrapperField(this, 45, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearSpout = function() {
  return this.setSpout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasSpout = function() {
  return jspb.Message.getField(this, 45) != null;
};


/**
 * optional pfs.Repo output_repo = 18;
 * @return {?proto.pfs.Repo}
 */
proto.pps.JobInfo.prototype.getOutputRepo = function() {
  return /** @type{?proto.pfs.Repo} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Repo, 18));
};


/**
 * @param {?proto.pfs.Repo|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setOutputRepo = function(value) {
  return jspb.Message.setWrapperField(this, 18, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearOutputRepo = function() {
  return this.setOutputRepo(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasOutputRepo = function() {
  return jspb.Message.getField(this, 18) != null;
};


/**
 * optional string output_branch = 17;
 * @return {string}
 */
proto.pps.JobInfo.prototype.getOutputBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 17, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setOutputBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 17, value);
};


/**
 * optional uint64 restart = 20;
 * @return {number}
 */
proto.pps.JobInfo.prototype.getRestart = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 20, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setRestart = function(value) {
  return jspb.Message.setProto3IntField(this, 20, value);
};


/**
 * optional int64 data_processed = 22;
 * @return {number}
 */
proto.pps.JobInfo.prototype.getDataProcessed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 22, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setDataProcessed = function(value) {
  return jspb.Message.setProto3IntField(this, 22, value);
};


/**
 * optional int64 data_skipped = 30;
 * @return {number}
 */
proto.pps.JobInfo.prototype.getDataSkipped = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 30, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setDataSkipped = function(value) {
  return jspb.Message.setProto3IntField(this, 30, value);
};


/**
 * optional int64 data_failed = 40;
 * @return {number}
 */
proto.pps.JobInfo.prototype.getDataFailed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 40, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setDataFailed = function(value) {
  return jspb.Message.setProto3IntField(this, 40, value);
};


/**
 * optional int64 data_recovered = 46;
 * @return {number}
 */
proto.pps.JobInfo.prototype.getDataRecovered = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 46, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setDataRecovered = function(value) {
  return jspb.Message.setProto3IntField(this, 46, value);
};


/**
 * optional int64 data_total = 23;
 * @return {number}
 */
proto.pps.JobInfo.prototype.getDataTotal = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 23, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setDataTotal = function(value) {
  return jspb.Message.setProto3IntField(this, 23, value);
};


/**
 * optional ProcessStats stats = 31;
 * @return {?proto.pps.ProcessStats}
 */
proto.pps.JobInfo.prototype.getStats = function() {
  return /** @type{?proto.pps.ProcessStats} */ (
    jspb.Message.getWrapperField(this, proto.pps.ProcessStats, 31));
};


/**
 * @param {?proto.pps.ProcessStats|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setStats = function(value) {
  return jspb.Message.setWrapperField(this, 31, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearStats = function() {
  return this.setStats(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasStats = function() {
  return jspb.Message.getField(this, 31) != null;
};


/**
 * repeated WorkerStatus worker_status = 24;
 * @return {!Array<!proto.pps.WorkerStatus>}
 */
proto.pps.JobInfo.prototype.getWorkerStatusList = function() {
  return /** @type{!Array<!proto.pps.WorkerStatus>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps.WorkerStatus, 24));
};


/**
 * @param {!Array<!proto.pps.WorkerStatus>} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setWorkerStatusList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 24, value);
};


/**
 * @param {!proto.pps.WorkerStatus=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps.WorkerStatus}
 */
proto.pps.JobInfo.prototype.addWorkerStatus = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 24, opt_value, proto.pps.WorkerStatus, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearWorkerStatusList = function() {
  return this.setWorkerStatusList([]);
};


/**
 * optional ResourceSpec resource_requests = 25;
 * @return {?proto.pps.ResourceSpec}
 */
proto.pps.JobInfo.prototype.getResourceRequests = function() {
  return /** @type{?proto.pps.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ResourceSpec, 25));
};


/**
 * @param {?proto.pps.ResourceSpec|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setResourceRequests = function(value) {
  return jspb.Message.setWrapperField(this, 25, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearResourceRequests = function() {
  return this.setResourceRequests(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasResourceRequests = function() {
  return jspb.Message.getField(this, 25) != null;
};


/**
 * optional ResourceSpec resource_limits = 36;
 * @return {?proto.pps.ResourceSpec}
 */
proto.pps.JobInfo.prototype.getResourceLimits = function() {
  return /** @type{?proto.pps.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ResourceSpec, 36));
};


/**
 * @param {?proto.pps.ResourceSpec|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 36, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearResourceLimits = function() {
  return this.setResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasResourceLimits = function() {
  return jspb.Message.getField(this, 36) != null;
};


/**
 * optional ResourceSpec sidecar_resource_limits = 48;
 * @return {?proto.pps.ResourceSpec}
 */
proto.pps.JobInfo.prototype.getSidecarResourceLimits = function() {
  return /** @type{?proto.pps.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ResourceSpec, 48));
};


/**
 * @param {?proto.pps.ResourceSpec|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setSidecarResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 48, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearSidecarResourceLimits = function() {
  return this.setSidecarResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasSidecarResourceLimits = function() {
  return jspb.Message.getField(this, 48) != null;
};


/**
 * optional Input input = 26;
 * @return {?proto.pps.Input}
 */
proto.pps.JobInfo.prototype.getInput = function() {
  return /** @type{?proto.pps.Input} */ (
    jspb.Message.getWrapperField(this, proto.pps.Input, 26));
};


/**
 * @param {?proto.pps.Input|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setInput = function(value) {
  return jspb.Message.setWrapperField(this, 26, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearInput = function() {
  return this.setInput(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasInput = function() {
  return jspb.Message.getField(this, 26) != null;
};


/**
 * optional pfs.BranchInfo new_branch = 27;
 * @return {?proto.pfs.BranchInfo}
 */
proto.pps.JobInfo.prototype.getNewBranch = function() {
  return /** @type{?proto.pfs.BranchInfo} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.BranchInfo, 27));
};


/**
 * @param {?proto.pfs.BranchInfo|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setNewBranch = function(value) {
  return jspb.Message.setWrapperField(this, 27, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearNewBranch = function() {
  return this.setNewBranch(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasNewBranch = function() {
  return jspb.Message.getField(this, 27) != null;
};


/**
 * optional pfs.Commit stats_commit = 29;
 * @return {?proto.pfs.Commit}
 */
proto.pps.JobInfo.prototype.getStatsCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 29));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setStatsCommit = function(value) {
  return jspb.Message.setWrapperField(this, 29, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearStatsCommit = function() {
  return this.setStatsCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasStatsCommit = function() {
  return jspb.Message.getField(this, 29) != null;
};


/**
 * optional bool enable_stats = 32;
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.getEnableStats = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 32, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setEnableStats = function(value) {
  return jspb.Message.setProto3BooleanField(this, 32, value);
};


/**
 * optional string salt = 33;
 * @return {string}
 */
proto.pps.JobInfo.prototype.getSalt = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 33, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setSalt = function(value) {
  return jspb.Message.setProto3StringField(this, 33, value);
};


/**
 * optional ChunkSpec chunk_spec = 37;
 * @return {?proto.pps.ChunkSpec}
 */
proto.pps.JobInfo.prototype.getChunkSpec = function() {
  return /** @type{?proto.pps.ChunkSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ChunkSpec, 37));
};


/**
 * @param {?proto.pps.ChunkSpec|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setChunkSpec = function(value) {
  return jspb.Message.setWrapperField(this, 37, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearChunkSpec = function() {
  return this.setChunkSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasChunkSpec = function() {
  return jspb.Message.getField(this, 37) != null;
};


/**
 * optional google.protobuf.Duration datum_timeout = 38;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps.JobInfo.prototype.getDatumTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 38));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setDatumTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 38, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearDatumTimeout = function() {
  return this.setDatumTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasDatumTimeout = function() {
  return jspb.Message.getField(this, 38) != null;
};


/**
 * optional google.protobuf.Duration job_timeout = 39;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps.JobInfo.prototype.getJobTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 39));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setJobTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 39, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearJobTimeout = function() {
  return this.setJobTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasJobTimeout = function() {
  return jspb.Message.getField(this, 39) != null;
};


/**
 * optional int64 datum_tries = 41;
 * @return {number}
 */
proto.pps.JobInfo.prototype.getDatumTries = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 41, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setDatumTries = function(value) {
  return jspb.Message.setProto3IntField(this, 41, value);
};


/**
 * optional SchedulingSpec scheduling_spec = 42;
 * @return {?proto.pps.SchedulingSpec}
 */
proto.pps.JobInfo.prototype.getSchedulingSpec = function() {
  return /** @type{?proto.pps.SchedulingSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.SchedulingSpec, 42));
};


/**
 * @param {?proto.pps.SchedulingSpec|undefined} value
 * @return {!proto.pps.JobInfo} returns this
*/
proto.pps.JobInfo.prototype.setSchedulingSpec = function(value) {
  return jspb.Message.setWrapperField(this, 42, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.clearSchedulingSpec = function() {
  return this.setSchedulingSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.JobInfo.prototype.hasSchedulingSpec = function() {
  return jspb.Message.getField(this, 42) != null;
};


/**
 * optional string pod_spec = 43;
 * @return {string}
 */
proto.pps.JobInfo.prototype.getPodSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 43, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setPodSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 43, value);
};


/**
 * optional string pod_patch = 44;
 * @return {string}
 */
proto.pps.JobInfo.prototype.getPodPatch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 44, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.JobInfo} returns this
 */
proto.pps.JobInfo.prototype.setPodPatch = function(value) {
  return jspb.Message.setProto3StringField(this, 44, value);
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
proto.pps.Worker.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Worker.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Worker} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Worker.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.Worker}
 */
proto.pps.Worker.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Worker;
  return proto.pps.Worker.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Worker} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Worker}
 */
proto.pps.Worker.deserializeBinaryFromReader = function(msg, reader) {
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
      var value = /** @type {!proto.pps.WorkerState} */ (reader.readEnum());
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
proto.pps.Worker.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Worker.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Worker} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Worker.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.Worker.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Worker} returns this
 */
proto.pps.Worker.prototype.setName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional WorkerState state = 2;
 * @return {!proto.pps.WorkerState}
 */
proto.pps.Worker.prototype.getState = function() {
  return /** @type {!proto.pps.WorkerState} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.pps.WorkerState} value
 * @return {!proto.pps.Worker} returns this
 */
proto.pps.Worker.prototype.setState = function(value) {
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
proto.pps.Pipeline.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Pipeline.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Pipeline} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Pipeline.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.Pipeline}
 */
proto.pps.Pipeline.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Pipeline;
  return proto.pps.Pipeline.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Pipeline} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Pipeline}
 */
proto.pps.Pipeline.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.Pipeline.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Pipeline.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Pipeline} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Pipeline.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.Pipeline.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Pipeline} returns this
 */
proto.pps.Pipeline.prototype.setName = function(value) {
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
proto.pps.EtcdPipelineInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.EtcdPipelineInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.EtcdPipelineInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.EtcdPipelineInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    state: jspb.Message.getFieldWithDefault(msg, 1, 0),
    reason: jspb.Message.getFieldWithDefault(msg, 4, ""),
    specCommit: (f = msg.getSpecCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    jobCountsMap: (f = msg.getJobCountsMap()) ? f.toObject(includeInstance, undefined) : [],
    authToken: jspb.Message.getFieldWithDefault(msg, 5, ""),
    lastJobState: jspb.Message.getFieldWithDefault(msg, 6, 0),
    parallelism: jspb.Message.getFieldWithDefault(msg, 7, 0)
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
 * @return {!proto.pps.EtcdPipelineInfo}
 */
proto.pps.EtcdPipelineInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.EtcdPipelineInfo;
  return proto.pps.EtcdPipelineInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.EtcdPipelineInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.EtcdPipelineInfo}
 */
proto.pps.EtcdPipelineInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!proto.pps.PipelineState} */ (reader.readEnum());
      msg.setState(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setReason(value);
      break;
    case 2:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setSpecCommit(value);
      break;
    case 3:
      var value = msg.getJobCountsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readInt32, jspb.BinaryReader.prototype.readInt32, null, 0, 0);
         });
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setAuthToken(value);
      break;
    case 6:
      var value = /** @type {!proto.pps.JobState} */ (reader.readEnum());
      msg.setLastJobState(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setParallelism(value);
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
proto.pps.EtcdPipelineInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.EtcdPipelineInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.EtcdPipelineInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.EtcdPipelineInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      1,
      f
    );
  }
  f = message.getReason();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getSpecCommit();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getJobCountsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(3, writer, jspb.BinaryWriter.prototype.writeInt32, jspb.BinaryWriter.prototype.writeInt32);
  }
  f = message.getAuthToken();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getLastJobState();
  if (f !== 0.0) {
    writer.writeEnum(
      6,
      f
    );
  }
  f = message.getParallelism();
  if (f !== 0) {
    writer.writeUint64(
      7,
      f
    );
  }
};


/**
 * optional PipelineState state = 1;
 * @return {!proto.pps.PipelineState}
 */
proto.pps.EtcdPipelineInfo.prototype.getState = function() {
  return /** @type {!proto.pps.PipelineState} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {!proto.pps.PipelineState} value
 * @return {!proto.pps.EtcdPipelineInfo} returns this
 */
proto.pps.EtcdPipelineInfo.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 1, value);
};


/**
 * optional string reason = 4;
 * @return {string}
 */
proto.pps.EtcdPipelineInfo.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.EtcdPipelineInfo} returns this
 */
proto.pps.EtcdPipelineInfo.prototype.setReason = function(value) {
  return jspb.Message.setProto3StringField(this, 4, value);
};


/**
 * optional pfs.Commit spec_commit = 2;
 * @return {?proto.pfs.Commit}
 */
proto.pps.EtcdPipelineInfo.prototype.getSpecCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 2));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.EtcdPipelineInfo} returns this
*/
proto.pps.EtcdPipelineInfo.prototype.setSpecCommit = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.EtcdPipelineInfo} returns this
 */
proto.pps.EtcdPipelineInfo.prototype.clearSpecCommit = function() {
  return this.setSpecCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.EtcdPipelineInfo.prototype.hasSpecCommit = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * map<int32, int32> job_counts = 3;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<number,number>}
 */
proto.pps.EtcdPipelineInfo.prototype.getJobCountsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<number,number>} */ (
      jspb.Message.getMapField(this, 3, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pps.EtcdPipelineInfo} returns this
 */
proto.pps.EtcdPipelineInfo.prototype.clearJobCountsMap = function() {
  this.getJobCountsMap().clear();
  return this;};


/**
 * optional string auth_token = 5;
 * @return {string}
 */
proto.pps.EtcdPipelineInfo.prototype.getAuthToken = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.EtcdPipelineInfo} returns this
 */
proto.pps.EtcdPipelineInfo.prototype.setAuthToken = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional JobState last_job_state = 6;
 * @return {!proto.pps.JobState}
 */
proto.pps.EtcdPipelineInfo.prototype.getLastJobState = function() {
  return /** @type {!proto.pps.JobState} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {!proto.pps.JobState} value
 * @return {!proto.pps.EtcdPipelineInfo} returns this
 */
proto.pps.EtcdPipelineInfo.prototype.setLastJobState = function(value) {
  return jspb.Message.setProto3EnumField(this, 6, value);
};


/**
 * optional uint64 parallelism = 7;
 * @return {number}
 */
proto.pps.EtcdPipelineInfo.prototype.getParallelism = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.EtcdPipelineInfo} returns this
 */
proto.pps.EtcdPipelineInfo.prototype.setParallelism = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
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
proto.pps.PipelineInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.PipelineInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.PipelineInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.PipelineInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 17, ""),
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f),
    version: jspb.Message.getFieldWithDefault(msg, 11, 0),
    transform: (f = msg.getTransform()) && proto.pps.Transform.toObject(includeInstance, f),
    tfJob: (f = msg.getTfJob()) && proto.pps.TFJob.toObject(includeInstance, f),
    parallelismSpec: (f = msg.getParallelismSpec()) && proto.pps.ParallelismSpec.toObject(includeInstance, f),
    egress: (f = msg.getEgress()) && proto.pps.Egress.toObject(includeInstance, f),
    createdAt: (f = msg.getCreatedAt()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    state: jspb.Message.getFieldWithDefault(msg, 7, 0),
    stopped: jspb.Message.getBooleanFieldWithDefault(msg, 38, false),
    recentError: jspb.Message.getFieldWithDefault(msg, 8, ""),
    workersRequested: jspb.Message.getFieldWithDefault(msg, 49, 0),
    workersAvailable: jspb.Message.getFieldWithDefault(msg, 50, 0),
    jobCountsMap: (f = msg.getJobCountsMap()) ? f.toObject(includeInstance, undefined) : [],
    lastJobState: jspb.Message.getFieldWithDefault(msg, 43, 0),
    outputBranch: jspb.Message.getFieldWithDefault(msg, 16, ""),
    resourceRequests: (f = msg.getResourceRequests()) && proto.pps.ResourceSpec.toObject(includeInstance, f),
    resourceLimits: (f = msg.getResourceLimits()) && proto.pps.ResourceSpec.toObject(includeInstance, f),
    sidecarResourceLimits: (f = msg.getSidecarResourceLimits()) && proto.pps.ResourceSpec.toObject(includeInstance, f),
    input: (f = msg.getInput()) && proto.pps.Input.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 21, ""),
    cacheSize: jspb.Message.getFieldWithDefault(msg, 23, ""),
    enableStats: jspb.Message.getBooleanFieldWithDefault(msg, 24, false),
    salt: jspb.Message.getFieldWithDefault(msg, 25, ""),
    reason: jspb.Message.getFieldWithDefault(msg, 28, ""),
    maxQueueSize: jspb.Message.getFieldWithDefault(msg, 29, 0),
    service: (f = msg.getService()) && proto.pps.Service.toObject(includeInstance, f),
    spout: (f = msg.getSpout()) && proto.pps.Spout.toObject(includeInstance, f),
    chunkSpec: (f = msg.getChunkSpec()) && proto.pps.ChunkSpec.toObject(includeInstance, f),
    datumTimeout: (f = msg.getDatumTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    jobTimeout: (f = msg.getJobTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    githookUrl: jspb.Message.getFieldWithDefault(msg, 35, ""),
    specCommit: (f = msg.getSpecCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    standby: jspb.Message.getBooleanFieldWithDefault(msg, 37, false),
    datumTries: jspb.Message.getFieldWithDefault(msg, 39, 0),
    schedulingSpec: (f = msg.getSchedulingSpec()) && proto.pps.SchedulingSpec.toObject(includeInstance, f),
    podSpec: jspb.Message.getFieldWithDefault(msg, 41, ""),
    podPatch: jspb.Message.getFieldWithDefault(msg, 44, ""),
    s3Out: jspb.Message.getBooleanFieldWithDefault(msg, 47, false),
    metadata: (f = msg.getMetadata()) && proto.pps.Metadata.toObject(includeInstance, f),
    noSkip: jspb.Message.getBooleanFieldWithDefault(msg, 52, false)
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
 * @return {!proto.pps.PipelineInfo}
 */
proto.pps.PipelineInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.PipelineInfo;
  return proto.pps.PipelineInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.PipelineInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.PipelineInfo}
 */
proto.pps.PipelineInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 17:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 1:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 11:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setVersion(value);
      break;
    case 2:
      var value = new proto.pps.Transform;
      reader.readMessage(value,proto.pps.Transform.deserializeBinaryFromReader);
      msg.setTransform(value);
      break;
    case 46:
      var value = new proto.pps.TFJob;
      reader.readMessage(value,proto.pps.TFJob.deserializeBinaryFromReader);
      msg.setTfJob(value);
      break;
    case 10:
      var value = new proto.pps.ParallelismSpec;
      reader.readMessage(value,proto.pps.ParallelismSpec.deserializeBinaryFromReader);
      msg.setParallelismSpec(value);
      break;
    case 15:
      var value = new proto.pps.Egress;
      reader.readMessage(value,proto.pps.Egress.deserializeBinaryFromReader);
      msg.setEgress(value);
      break;
    case 6:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCreatedAt(value);
      break;
    case 7:
      var value = /** @type {!proto.pps.PipelineState} */ (reader.readEnum());
      msg.setState(value);
      break;
    case 38:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setStopped(value);
      break;
    case 8:
      var value = /** @type {string} */ (reader.readString());
      msg.setRecentError(value);
      break;
    case 49:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setWorkersRequested(value);
      break;
    case 50:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setWorkersAvailable(value);
      break;
    case 9:
      var value = msg.getJobCountsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readInt32, jspb.BinaryReader.prototype.readInt32, null, 0, 0);
         });
      break;
    case 43:
      var value = /** @type {!proto.pps.JobState} */ (reader.readEnum());
      msg.setLastJobState(value);
      break;
    case 16:
      var value = /** @type {string} */ (reader.readString());
      msg.setOutputBranch(value);
      break;
    case 19:
      var value = new proto.pps.ResourceSpec;
      reader.readMessage(value,proto.pps.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceRequests(value);
      break;
    case 31:
      var value = new proto.pps.ResourceSpec;
      reader.readMessage(value,proto.pps.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceLimits(value);
      break;
    case 51:
      var value = new proto.pps.ResourceSpec;
      reader.readMessage(value,proto.pps.ResourceSpec.deserializeBinaryFromReader);
      msg.setSidecarResourceLimits(value);
      break;
    case 20:
      var value = new proto.pps.Input;
      reader.readMessage(value,proto.pps.Input.deserializeBinaryFromReader);
      msg.setInput(value);
      break;
    case 21:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 23:
      var value = /** @type {string} */ (reader.readString());
      msg.setCacheSize(value);
      break;
    case 24:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEnableStats(value);
      break;
    case 25:
      var value = /** @type {string} */ (reader.readString());
      msg.setSalt(value);
      break;
    case 28:
      var value = /** @type {string} */ (reader.readString());
      msg.setReason(value);
      break;
    case 29:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setMaxQueueSize(value);
      break;
    case 30:
      var value = new proto.pps.Service;
      reader.readMessage(value,proto.pps.Service.deserializeBinaryFromReader);
      msg.setService(value);
      break;
    case 45:
      var value = new proto.pps.Spout;
      reader.readMessage(value,proto.pps.Spout.deserializeBinaryFromReader);
      msg.setSpout(value);
      break;
    case 32:
      var value = new proto.pps.ChunkSpec;
      reader.readMessage(value,proto.pps.ChunkSpec.deserializeBinaryFromReader);
      msg.setChunkSpec(value);
      break;
    case 33:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setDatumTimeout(value);
      break;
    case 34:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setJobTimeout(value);
      break;
    case 35:
      var value = /** @type {string} */ (reader.readString());
      msg.setGithookUrl(value);
      break;
    case 36:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setSpecCommit(value);
      break;
    case 37:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setStandby(value);
      break;
    case 39:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDatumTries(value);
      break;
    case 40:
      var value = new proto.pps.SchedulingSpec;
      reader.readMessage(value,proto.pps.SchedulingSpec.deserializeBinaryFromReader);
      msg.setSchedulingSpec(value);
      break;
    case 41:
      var value = /** @type {string} */ (reader.readString());
      msg.setPodSpec(value);
      break;
    case 44:
      var value = /** @type {string} */ (reader.readString());
      msg.setPodPatch(value);
      break;
    case 47:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setS3Out(value);
      break;
    case 48:
      var value = new proto.pps.Metadata;
      reader.readMessage(value,proto.pps.Metadata.deserializeBinaryFromReader);
      msg.setMetadata(value);
      break;
    case 52:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setNoSkip(value);
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
proto.pps.PipelineInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.PipelineInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.PipelineInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.PipelineInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      17,
      f
    );
  }
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getVersion();
  if (f !== 0) {
    writer.writeUint64(
      11,
      f
    );
  }
  f = message.getTransform();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps.Transform.serializeBinaryToWriter
    );
  }
  f = message.getTfJob();
  if (f != null) {
    writer.writeMessage(
      46,
      f,
      proto.pps.TFJob.serializeBinaryToWriter
    );
  }
  f = message.getParallelismSpec();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      proto.pps.ParallelismSpec.serializeBinaryToWriter
    );
  }
  f = message.getEgress();
  if (f != null) {
    writer.writeMessage(
      15,
      f,
      proto.pps.Egress.serializeBinaryToWriter
    );
  }
  f = message.getCreatedAt();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      7,
      f
    );
  }
  f = message.getStopped();
  if (f) {
    writer.writeBool(
      38,
      f
    );
  }
  f = message.getRecentError();
  if (f.length > 0) {
    writer.writeString(
      8,
      f
    );
  }
  f = message.getWorkersRequested();
  if (f !== 0) {
    writer.writeInt64(
      49,
      f
    );
  }
  f = message.getWorkersAvailable();
  if (f !== 0) {
    writer.writeInt64(
      50,
      f
    );
  }
  f = message.getJobCountsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(9, writer, jspb.BinaryWriter.prototype.writeInt32, jspb.BinaryWriter.prototype.writeInt32);
  }
  f = message.getLastJobState();
  if (f !== 0.0) {
    writer.writeEnum(
      43,
      f
    );
  }
  f = message.getOutputBranch();
  if (f.length > 0) {
    writer.writeString(
      16,
      f
    );
  }
  f = message.getResourceRequests();
  if (f != null) {
    writer.writeMessage(
      19,
      f,
      proto.pps.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getResourceLimits();
  if (f != null) {
    writer.writeMessage(
      31,
      f,
      proto.pps.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getSidecarResourceLimits();
  if (f != null) {
    writer.writeMessage(
      51,
      f,
      proto.pps.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getInput();
  if (f != null) {
    writer.writeMessage(
      20,
      f,
      proto.pps.Input.serializeBinaryToWriter
    );
  }
  f = message.getDescription();
  if (f.length > 0) {
    writer.writeString(
      21,
      f
    );
  }
  f = message.getCacheSize();
  if (f.length > 0) {
    writer.writeString(
      23,
      f
    );
  }
  f = message.getEnableStats();
  if (f) {
    writer.writeBool(
      24,
      f
    );
  }
  f = message.getSalt();
  if (f.length > 0) {
    writer.writeString(
      25,
      f
    );
  }
  f = message.getReason();
  if (f.length > 0) {
    writer.writeString(
      28,
      f
    );
  }
  f = message.getMaxQueueSize();
  if (f !== 0) {
    writer.writeInt64(
      29,
      f
    );
  }
  f = message.getService();
  if (f != null) {
    writer.writeMessage(
      30,
      f,
      proto.pps.Service.serializeBinaryToWriter
    );
  }
  f = message.getSpout();
  if (f != null) {
    writer.writeMessage(
      45,
      f,
      proto.pps.Spout.serializeBinaryToWriter
    );
  }
  f = message.getChunkSpec();
  if (f != null) {
    writer.writeMessage(
      32,
      f,
      proto.pps.ChunkSpec.serializeBinaryToWriter
    );
  }
  f = message.getDatumTimeout();
  if (f != null) {
    writer.writeMessage(
      33,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getJobTimeout();
  if (f != null) {
    writer.writeMessage(
      34,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getGithookUrl();
  if (f.length > 0) {
    writer.writeString(
      35,
      f
    );
  }
  f = message.getSpecCommit();
  if (f != null) {
    writer.writeMessage(
      36,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getStandby();
  if (f) {
    writer.writeBool(
      37,
      f
    );
  }
  f = message.getDatumTries();
  if (f !== 0) {
    writer.writeInt64(
      39,
      f
    );
  }
  f = message.getSchedulingSpec();
  if (f != null) {
    writer.writeMessage(
      40,
      f,
      proto.pps.SchedulingSpec.serializeBinaryToWriter
    );
  }
  f = message.getPodSpec();
  if (f.length > 0) {
    writer.writeString(
      41,
      f
    );
  }
  f = message.getPodPatch();
  if (f.length > 0) {
    writer.writeString(
      44,
      f
    );
  }
  f = message.getS3Out();
  if (f) {
    writer.writeBool(
      47,
      f
    );
  }
  f = message.getMetadata();
  if (f != null) {
    writer.writeMessage(
      48,
      f,
      proto.pps.Metadata.serializeBinaryToWriter
    );
  }
  f = message.getNoSkip();
  if (f) {
    writer.writeBool(
      52,
      f
    );
  }
};


/**
 * optional string id = 17;
 * @return {string}
 */
proto.pps.PipelineInfo.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 17, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setId = function(value) {
  return jspb.Message.setProto3StringField(this, 17, value);
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.PipelineInfo.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 1));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional uint64 version = 11;
 * @return {number}
 */
proto.pps.PipelineInfo.prototype.getVersion = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 11, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setVersion = function(value) {
  return jspb.Message.setProto3IntField(this, 11, value);
};


/**
 * optional Transform transform = 2;
 * @return {?proto.pps.Transform}
 */
proto.pps.PipelineInfo.prototype.getTransform = function() {
  return /** @type{?proto.pps.Transform} */ (
    jspb.Message.getWrapperField(this, proto.pps.Transform, 2));
};


/**
 * @param {?proto.pps.Transform|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setTransform = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearTransform = function() {
  return this.setTransform(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasTransform = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional TFJob tf_job = 46;
 * @return {?proto.pps.TFJob}
 */
proto.pps.PipelineInfo.prototype.getTfJob = function() {
  return /** @type{?proto.pps.TFJob} */ (
    jspb.Message.getWrapperField(this, proto.pps.TFJob, 46));
};


/**
 * @param {?proto.pps.TFJob|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setTfJob = function(value) {
  return jspb.Message.setWrapperField(this, 46, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearTfJob = function() {
  return this.setTfJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasTfJob = function() {
  return jspb.Message.getField(this, 46) != null;
};


/**
 * optional ParallelismSpec parallelism_spec = 10;
 * @return {?proto.pps.ParallelismSpec}
 */
proto.pps.PipelineInfo.prototype.getParallelismSpec = function() {
  return /** @type{?proto.pps.ParallelismSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ParallelismSpec, 10));
};


/**
 * @param {?proto.pps.ParallelismSpec|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setParallelismSpec = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearParallelismSpec = function() {
  return this.setParallelismSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasParallelismSpec = function() {
  return jspb.Message.getField(this, 10) != null;
};


/**
 * optional Egress egress = 15;
 * @return {?proto.pps.Egress}
 */
proto.pps.PipelineInfo.prototype.getEgress = function() {
  return /** @type{?proto.pps.Egress} */ (
    jspb.Message.getWrapperField(this, proto.pps.Egress, 15));
};


/**
 * @param {?proto.pps.Egress|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setEgress = function(value) {
  return jspb.Message.setWrapperField(this, 15, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearEgress = function() {
  return this.setEgress(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasEgress = function() {
  return jspb.Message.getField(this, 15) != null;
};


/**
 * optional google.protobuf.Timestamp created_at = 6;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps.PipelineInfo.prototype.getCreatedAt = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 6));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setCreatedAt = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearCreatedAt = function() {
  return this.setCreatedAt(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasCreatedAt = function() {
  return jspb.Message.getField(this, 6) != null;
};


/**
 * optional PipelineState state = 7;
 * @return {!proto.pps.PipelineState}
 */
proto.pps.PipelineInfo.prototype.getState = function() {
  return /** @type {!proto.pps.PipelineState} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {!proto.pps.PipelineState} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 7, value);
};


/**
 * optional bool stopped = 38;
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.getStopped = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 38, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setStopped = function(value) {
  return jspb.Message.setProto3BooleanField(this, 38, value);
};


/**
 * optional string recent_error = 8;
 * @return {string}
 */
proto.pps.PipelineInfo.prototype.getRecentError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 8, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setRecentError = function(value) {
  return jspb.Message.setProto3StringField(this, 8, value);
};


/**
 * optional int64 workers_requested = 49;
 * @return {number}
 */
proto.pps.PipelineInfo.prototype.getWorkersRequested = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 49, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setWorkersRequested = function(value) {
  return jspb.Message.setProto3IntField(this, 49, value);
};


/**
 * optional int64 workers_available = 50;
 * @return {number}
 */
proto.pps.PipelineInfo.prototype.getWorkersAvailable = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 50, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setWorkersAvailable = function(value) {
  return jspb.Message.setProto3IntField(this, 50, value);
};


/**
 * map<int32, int32> job_counts = 9;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<number,number>}
 */
proto.pps.PipelineInfo.prototype.getJobCountsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<number,number>} */ (
      jspb.Message.getMapField(this, 9, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearJobCountsMap = function() {
  this.getJobCountsMap().clear();
  return this;};


/**
 * optional JobState last_job_state = 43;
 * @return {!proto.pps.JobState}
 */
proto.pps.PipelineInfo.prototype.getLastJobState = function() {
  return /** @type {!proto.pps.JobState} */ (jspb.Message.getFieldWithDefault(this, 43, 0));
};


/**
 * @param {!proto.pps.JobState} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setLastJobState = function(value) {
  return jspb.Message.setProto3EnumField(this, 43, value);
};


/**
 * optional string output_branch = 16;
 * @return {string}
 */
proto.pps.PipelineInfo.prototype.getOutputBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 16, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setOutputBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 16, value);
};


/**
 * optional ResourceSpec resource_requests = 19;
 * @return {?proto.pps.ResourceSpec}
 */
proto.pps.PipelineInfo.prototype.getResourceRequests = function() {
  return /** @type{?proto.pps.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ResourceSpec, 19));
};


/**
 * @param {?proto.pps.ResourceSpec|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setResourceRequests = function(value) {
  return jspb.Message.setWrapperField(this, 19, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearResourceRequests = function() {
  return this.setResourceRequests(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasResourceRequests = function() {
  return jspb.Message.getField(this, 19) != null;
};


/**
 * optional ResourceSpec resource_limits = 31;
 * @return {?proto.pps.ResourceSpec}
 */
proto.pps.PipelineInfo.prototype.getResourceLimits = function() {
  return /** @type{?proto.pps.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ResourceSpec, 31));
};


/**
 * @param {?proto.pps.ResourceSpec|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 31, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearResourceLimits = function() {
  return this.setResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasResourceLimits = function() {
  return jspb.Message.getField(this, 31) != null;
};


/**
 * optional ResourceSpec sidecar_resource_limits = 51;
 * @return {?proto.pps.ResourceSpec}
 */
proto.pps.PipelineInfo.prototype.getSidecarResourceLimits = function() {
  return /** @type{?proto.pps.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ResourceSpec, 51));
};


/**
 * @param {?proto.pps.ResourceSpec|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setSidecarResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 51, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearSidecarResourceLimits = function() {
  return this.setSidecarResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasSidecarResourceLimits = function() {
  return jspb.Message.getField(this, 51) != null;
};


/**
 * optional Input input = 20;
 * @return {?proto.pps.Input}
 */
proto.pps.PipelineInfo.prototype.getInput = function() {
  return /** @type{?proto.pps.Input} */ (
    jspb.Message.getWrapperField(this, proto.pps.Input, 20));
};


/**
 * @param {?proto.pps.Input|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setInput = function(value) {
  return jspb.Message.setWrapperField(this, 20, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearInput = function() {
  return this.setInput(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasInput = function() {
  return jspb.Message.getField(this, 20) != null;
};


/**
 * optional string description = 21;
 * @return {string}
 */
proto.pps.PipelineInfo.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 21, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 21, value);
};


/**
 * optional string cache_size = 23;
 * @return {string}
 */
proto.pps.PipelineInfo.prototype.getCacheSize = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 23, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setCacheSize = function(value) {
  return jspb.Message.setProto3StringField(this, 23, value);
};


/**
 * optional bool enable_stats = 24;
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.getEnableStats = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 24, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setEnableStats = function(value) {
  return jspb.Message.setProto3BooleanField(this, 24, value);
};


/**
 * optional string salt = 25;
 * @return {string}
 */
proto.pps.PipelineInfo.prototype.getSalt = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 25, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setSalt = function(value) {
  return jspb.Message.setProto3StringField(this, 25, value);
};


/**
 * optional string reason = 28;
 * @return {string}
 */
proto.pps.PipelineInfo.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 28, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setReason = function(value) {
  return jspb.Message.setProto3StringField(this, 28, value);
};


/**
 * optional int64 max_queue_size = 29;
 * @return {number}
 */
proto.pps.PipelineInfo.prototype.getMaxQueueSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 29, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setMaxQueueSize = function(value) {
  return jspb.Message.setProto3IntField(this, 29, value);
};


/**
 * optional Service service = 30;
 * @return {?proto.pps.Service}
 */
proto.pps.PipelineInfo.prototype.getService = function() {
  return /** @type{?proto.pps.Service} */ (
    jspb.Message.getWrapperField(this, proto.pps.Service, 30));
};


/**
 * @param {?proto.pps.Service|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setService = function(value) {
  return jspb.Message.setWrapperField(this, 30, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearService = function() {
  return this.setService(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasService = function() {
  return jspb.Message.getField(this, 30) != null;
};


/**
 * optional Spout spout = 45;
 * @return {?proto.pps.Spout}
 */
proto.pps.PipelineInfo.prototype.getSpout = function() {
  return /** @type{?proto.pps.Spout} */ (
    jspb.Message.getWrapperField(this, proto.pps.Spout, 45));
};


/**
 * @param {?proto.pps.Spout|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setSpout = function(value) {
  return jspb.Message.setWrapperField(this, 45, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearSpout = function() {
  return this.setSpout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasSpout = function() {
  return jspb.Message.getField(this, 45) != null;
};


/**
 * optional ChunkSpec chunk_spec = 32;
 * @return {?proto.pps.ChunkSpec}
 */
proto.pps.PipelineInfo.prototype.getChunkSpec = function() {
  return /** @type{?proto.pps.ChunkSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ChunkSpec, 32));
};


/**
 * @param {?proto.pps.ChunkSpec|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setChunkSpec = function(value) {
  return jspb.Message.setWrapperField(this, 32, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearChunkSpec = function() {
  return this.setChunkSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasChunkSpec = function() {
  return jspb.Message.getField(this, 32) != null;
};


/**
 * optional google.protobuf.Duration datum_timeout = 33;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps.PipelineInfo.prototype.getDatumTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 33));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setDatumTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 33, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearDatumTimeout = function() {
  return this.setDatumTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasDatumTimeout = function() {
  return jspb.Message.getField(this, 33) != null;
};


/**
 * optional google.protobuf.Duration job_timeout = 34;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps.PipelineInfo.prototype.getJobTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 34));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setJobTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 34, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearJobTimeout = function() {
  return this.setJobTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasJobTimeout = function() {
  return jspb.Message.getField(this, 34) != null;
};


/**
 * optional string githook_url = 35;
 * @return {string}
 */
proto.pps.PipelineInfo.prototype.getGithookUrl = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 35, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setGithookUrl = function(value) {
  return jspb.Message.setProto3StringField(this, 35, value);
};


/**
 * optional pfs.Commit spec_commit = 36;
 * @return {?proto.pfs.Commit}
 */
proto.pps.PipelineInfo.prototype.getSpecCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 36));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setSpecCommit = function(value) {
  return jspb.Message.setWrapperField(this, 36, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearSpecCommit = function() {
  return this.setSpecCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasSpecCommit = function() {
  return jspb.Message.getField(this, 36) != null;
};


/**
 * optional bool standby = 37;
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.getStandby = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 37, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setStandby = function(value) {
  return jspb.Message.setProto3BooleanField(this, 37, value);
};


/**
 * optional int64 datum_tries = 39;
 * @return {number}
 */
proto.pps.PipelineInfo.prototype.getDatumTries = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 39, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setDatumTries = function(value) {
  return jspb.Message.setProto3IntField(this, 39, value);
};


/**
 * optional SchedulingSpec scheduling_spec = 40;
 * @return {?proto.pps.SchedulingSpec}
 */
proto.pps.PipelineInfo.prototype.getSchedulingSpec = function() {
  return /** @type{?proto.pps.SchedulingSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.SchedulingSpec, 40));
};


/**
 * @param {?proto.pps.SchedulingSpec|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setSchedulingSpec = function(value) {
  return jspb.Message.setWrapperField(this, 40, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearSchedulingSpec = function() {
  return this.setSchedulingSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasSchedulingSpec = function() {
  return jspb.Message.getField(this, 40) != null;
};


/**
 * optional string pod_spec = 41;
 * @return {string}
 */
proto.pps.PipelineInfo.prototype.getPodSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 41, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setPodSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 41, value);
};


/**
 * optional string pod_patch = 44;
 * @return {string}
 */
proto.pps.PipelineInfo.prototype.getPodPatch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 44, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setPodPatch = function(value) {
  return jspb.Message.setProto3StringField(this, 44, value);
};


/**
 * optional bool s3_out = 47;
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.getS3Out = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 47, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setS3Out = function(value) {
  return jspb.Message.setProto3BooleanField(this, 47, value);
};


/**
 * optional Metadata metadata = 48;
 * @return {?proto.pps.Metadata}
 */
proto.pps.PipelineInfo.prototype.getMetadata = function() {
  return /** @type{?proto.pps.Metadata} */ (
    jspb.Message.getWrapperField(this, proto.pps.Metadata, 48));
};


/**
 * @param {?proto.pps.Metadata|undefined} value
 * @return {!proto.pps.PipelineInfo} returns this
*/
proto.pps.PipelineInfo.prototype.setMetadata = function(value) {
  return jspb.Message.setWrapperField(this, 48, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.clearMetadata = function() {
  return this.setMetadata(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.hasMetadata = function() {
  return jspb.Message.getField(this, 48) != null;
};


/**
 * optional bool no_skip = 52;
 * @return {boolean}
 */
proto.pps.PipelineInfo.prototype.getNoSkip = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 52, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.PipelineInfo} returns this
 */
proto.pps.PipelineInfo.prototype.setNoSkip = function(value) {
  return jspb.Message.setProto3BooleanField(this, 52, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.PipelineInfos.repeatedFields_ = [1];



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
proto.pps.PipelineInfos.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.PipelineInfos.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.PipelineInfos} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.PipelineInfos.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipelineInfoList: jspb.Message.toObjectList(msg.getPipelineInfoList(),
    proto.pps.PipelineInfo.toObject, includeInstance)
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
 * @return {!proto.pps.PipelineInfos}
 */
proto.pps.PipelineInfos.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.PipelineInfos;
  return proto.pps.PipelineInfos.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.PipelineInfos} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.PipelineInfos}
 */
proto.pps.PipelineInfos.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.PipelineInfo;
      reader.readMessage(value,proto.pps.PipelineInfo.deserializeBinaryFromReader);
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
proto.pps.PipelineInfos.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.PipelineInfos.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.PipelineInfos} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.PipelineInfos.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipelineInfoList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pps.PipelineInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated PipelineInfo pipeline_info = 1;
 * @return {!Array<!proto.pps.PipelineInfo>}
 */
proto.pps.PipelineInfos.prototype.getPipelineInfoList = function() {
  return /** @type{!Array<!proto.pps.PipelineInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps.PipelineInfo, 1));
};


/**
 * @param {!Array<!proto.pps.PipelineInfo>} value
 * @return {!proto.pps.PipelineInfos} returns this
*/
proto.pps.PipelineInfos.prototype.setPipelineInfoList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pps.PipelineInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps.PipelineInfo}
 */
proto.pps.PipelineInfos.prototype.addPipelineInfo = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pps.PipelineInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.PipelineInfos} returns this
 */
proto.pps.PipelineInfos.prototype.clearPipelineInfoList = function() {
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
proto.pps.CreateJobRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.CreateJobRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.CreateJobRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.CreateJobRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f),
    outputCommit: (f = msg.getOutputCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    restart: jspb.Message.getFieldWithDefault(msg, 26, 0),
    dataProcessed: jspb.Message.getFieldWithDefault(msg, 27, 0),
    dataSkipped: jspb.Message.getFieldWithDefault(msg, 28, 0),
    dataTotal: jspb.Message.getFieldWithDefault(msg, 29, 0),
    dataFailed: jspb.Message.getFieldWithDefault(msg, 30, 0),
    dataRecovered: jspb.Message.getFieldWithDefault(msg, 31, 0),
    stats: (f = msg.getStats()) && proto.pps.ProcessStats.toObject(includeInstance, f),
    statsCommit: (f = msg.getStatsCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    state: jspb.Message.getFieldWithDefault(msg, 34, 0),
    reason: jspb.Message.getFieldWithDefault(msg, 35, ""),
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
 * @return {!proto.pps.CreateJobRequest}
 */
proto.pps.CreateJobRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.CreateJobRequest;
  return proto.pps.CreateJobRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.CreateJobRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.CreateJobRequest}
 */
proto.pps.CreateJobRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 2:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 25:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setOutputCommit(value);
      break;
    case 26:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setRestart(value);
      break;
    case 27:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataProcessed(value);
      break;
    case 28:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataSkipped(value);
      break;
    case 29:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataTotal(value);
      break;
    case 30:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataFailed(value);
      break;
    case 31:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataRecovered(value);
      break;
    case 32:
      var value = new proto.pps.ProcessStats;
      reader.readMessage(value,proto.pps.ProcessStats.deserializeBinaryFromReader);
      msg.setStats(value);
      break;
    case 33:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setStatsCommit(value);
      break;
    case 34:
      var value = /** @type {!proto.pps.JobState} */ (reader.readEnum());
      msg.setState(value);
      break;
    case 35:
      var value = /** @type {string} */ (reader.readString());
      msg.setReason(value);
      break;
    case 36:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setStarted(value);
      break;
    case 37:
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
proto.pps.CreateJobRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.CreateJobRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.CreateJobRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.CreateJobRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getOutputCommit();
  if (f != null) {
    writer.writeMessage(
      25,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getRestart();
  if (f !== 0) {
    writer.writeUint64(
      26,
      f
    );
  }
  f = message.getDataProcessed();
  if (f !== 0) {
    writer.writeInt64(
      27,
      f
    );
  }
  f = message.getDataSkipped();
  if (f !== 0) {
    writer.writeInt64(
      28,
      f
    );
  }
  f = message.getDataTotal();
  if (f !== 0) {
    writer.writeInt64(
      29,
      f
    );
  }
  f = message.getDataFailed();
  if (f !== 0) {
    writer.writeInt64(
      30,
      f
    );
  }
  f = message.getDataRecovered();
  if (f !== 0) {
    writer.writeInt64(
      31,
      f
    );
  }
  f = message.getStats();
  if (f != null) {
    writer.writeMessage(
      32,
      f,
      proto.pps.ProcessStats.serializeBinaryToWriter
    );
  }
  f = message.getStatsCommit();
  if (f != null) {
    writer.writeMessage(
      33,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      34,
      f
    );
  }
  f = message.getReason();
  if (f.length > 0) {
    writer.writeString(
      35,
      f
    );
  }
  f = message.getStarted();
  if (f != null) {
    writer.writeMessage(
      36,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getFinished();
  if (f != null) {
    writer.writeMessage(
      37,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
};


/**
 * optional Pipeline pipeline = 2;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.CreateJobRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 2));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.CreateJobRequest} returns this
*/
proto.pps.CreateJobRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreateJobRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional pfs.Commit output_commit = 25;
 * @return {?proto.pfs.Commit}
 */
proto.pps.CreateJobRequest.prototype.getOutputCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 25));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.CreateJobRequest} returns this
*/
proto.pps.CreateJobRequest.prototype.setOutputCommit = function(value) {
  return jspb.Message.setWrapperField(this, 25, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.clearOutputCommit = function() {
  return this.setOutputCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreateJobRequest.prototype.hasOutputCommit = function() {
  return jspb.Message.getField(this, 25) != null;
};


/**
 * optional uint64 restart = 26;
 * @return {number}
 */
proto.pps.CreateJobRequest.prototype.getRestart = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 26, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.setRestart = function(value) {
  return jspb.Message.setProto3IntField(this, 26, value);
};


/**
 * optional int64 data_processed = 27;
 * @return {number}
 */
proto.pps.CreateJobRequest.prototype.getDataProcessed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 27, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.setDataProcessed = function(value) {
  return jspb.Message.setProto3IntField(this, 27, value);
};


/**
 * optional int64 data_skipped = 28;
 * @return {number}
 */
proto.pps.CreateJobRequest.prototype.getDataSkipped = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 28, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.setDataSkipped = function(value) {
  return jspb.Message.setProto3IntField(this, 28, value);
};


/**
 * optional int64 data_total = 29;
 * @return {number}
 */
proto.pps.CreateJobRequest.prototype.getDataTotal = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 29, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.setDataTotal = function(value) {
  return jspb.Message.setProto3IntField(this, 29, value);
};


/**
 * optional int64 data_failed = 30;
 * @return {number}
 */
proto.pps.CreateJobRequest.prototype.getDataFailed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 30, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.setDataFailed = function(value) {
  return jspb.Message.setProto3IntField(this, 30, value);
};


/**
 * optional int64 data_recovered = 31;
 * @return {number}
 */
proto.pps.CreateJobRequest.prototype.getDataRecovered = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 31, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.setDataRecovered = function(value) {
  return jspb.Message.setProto3IntField(this, 31, value);
};


/**
 * optional ProcessStats stats = 32;
 * @return {?proto.pps.ProcessStats}
 */
proto.pps.CreateJobRequest.prototype.getStats = function() {
  return /** @type{?proto.pps.ProcessStats} */ (
    jspb.Message.getWrapperField(this, proto.pps.ProcessStats, 32));
};


/**
 * @param {?proto.pps.ProcessStats|undefined} value
 * @return {!proto.pps.CreateJobRequest} returns this
*/
proto.pps.CreateJobRequest.prototype.setStats = function(value) {
  return jspb.Message.setWrapperField(this, 32, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.clearStats = function() {
  return this.setStats(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreateJobRequest.prototype.hasStats = function() {
  return jspb.Message.getField(this, 32) != null;
};


/**
 * optional pfs.Commit stats_commit = 33;
 * @return {?proto.pfs.Commit}
 */
proto.pps.CreateJobRequest.prototype.getStatsCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 33));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.CreateJobRequest} returns this
*/
proto.pps.CreateJobRequest.prototype.setStatsCommit = function(value) {
  return jspb.Message.setWrapperField(this, 33, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.clearStatsCommit = function() {
  return this.setStatsCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreateJobRequest.prototype.hasStatsCommit = function() {
  return jspb.Message.getField(this, 33) != null;
};


/**
 * optional JobState state = 34;
 * @return {!proto.pps.JobState}
 */
proto.pps.CreateJobRequest.prototype.getState = function() {
  return /** @type {!proto.pps.JobState} */ (jspb.Message.getFieldWithDefault(this, 34, 0));
};


/**
 * @param {!proto.pps.JobState} value
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 34, value);
};


/**
 * optional string reason = 35;
 * @return {string}
 */
proto.pps.CreateJobRequest.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 35, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.setReason = function(value) {
  return jspb.Message.setProto3StringField(this, 35, value);
};


/**
 * optional google.protobuf.Timestamp started = 36;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps.CreateJobRequest.prototype.getStarted = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 36));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps.CreateJobRequest} returns this
*/
proto.pps.CreateJobRequest.prototype.setStarted = function(value) {
  return jspb.Message.setWrapperField(this, 36, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.clearStarted = function() {
  return this.setStarted(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreateJobRequest.prototype.hasStarted = function() {
  return jspb.Message.getField(this, 36) != null;
};


/**
 * optional google.protobuf.Timestamp finished = 37;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps.CreateJobRequest.prototype.getFinished = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 37));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps.CreateJobRequest} returns this
*/
proto.pps.CreateJobRequest.prototype.setFinished = function(value) {
  return jspb.Message.setWrapperField(this, 37, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreateJobRequest} returns this
 */
proto.pps.CreateJobRequest.prototype.clearFinished = function() {
  return this.setFinished(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreateJobRequest.prototype.hasFinished = function() {
  return jspb.Message.getField(this, 37) != null;
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
proto.pps.InspectJobRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.InspectJobRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.InspectJobRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.InspectJobRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps.Job.toObject(includeInstance, f),
    outputCommit: (f = msg.getOutputCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    blockState: jspb.Message.getBooleanFieldWithDefault(msg, 2, false),
    full: jspb.Message.getBooleanFieldWithDefault(msg, 4, false)
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
 * @return {!proto.pps.InspectJobRequest}
 */
proto.pps.InspectJobRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.InspectJobRequest;
  return proto.pps.InspectJobRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.InspectJobRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.InspectJobRequest}
 */
proto.pps.InspectJobRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Job;
      reader.readMessage(value,proto.pps.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 3:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setOutputCommit(value);
      break;
    case 2:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setBlockState(value);
      break;
    case 4:
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
proto.pps.InspectJobRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.InspectJobRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.InspectJobRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.InspectJobRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Job.serializeBinaryToWriter
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
  f = message.getBlockState();
  if (f) {
    writer.writeBool(
      2,
      f
    );
  }
  f = message.getFull();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps.Job}
 */
proto.pps.InspectJobRequest.prototype.getJob = function() {
  return /** @type{?proto.pps.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps.Job, 1));
};


/**
 * @param {?proto.pps.Job|undefined} value
 * @return {!proto.pps.InspectJobRequest} returns this
*/
proto.pps.InspectJobRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.InspectJobRequest} returns this
 */
proto.pps.InspectJobRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.InspectJobRequest.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional pfs.Commit output_commit = 3;
 * @return {?proto.pfs.Commit}
 */
proto.pps.InspectJobRequest.prototype.getOutputCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 3));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.InspectJobRequest} returns this
*/
proto.pps.InspectJobRequest.prototype.setOutputCommit = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.InspectJobRequest} returns this
 */
proto.pps.InspectJobRequest.prototype.clearOutputCommit = function() {
  return this.setOutputCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.InspectJobRequest.prototype.hasOutputCommit = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional bool block_state = 2;
 * @return {boolean}
 */
proto.pps.InspectJobRequest.prototype.getBlockState = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 2, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.InspectJobRequest} returns this
 */
proto.pps.InspectJobRequest.prototype.setBlockState = function(value) {
  return jspb.Message.setProto3BooleanField(this, 2, value);
};


/**
 * optional bool full = 4;
 * @return {boolean}
 */
proto.pps.InspectJobRequest.prototype.getFull = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.InspectJobRequest} returns this
 */
proto.pps.InspectJobRequest.prototype.setFull = function(value) {
  return jspb.Message.setProto3BooleanField(this, 4, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.ListJobRequest.repeatedFields_ = [2];



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
proto.pps.ListJobRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.ListJobRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.ListJobRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ListJobRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f),
    inputCommitList: jspb.Message.toObjectList(msg.getInputCommitList(),
    pfs_pfs_pb.Commit.toObject, includeInstance),
    outputCommit: (f = msg.getOutputCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    history: jspb.Message.getFieldWithDefault(msg, 4, 0),
    full: jspb.Message.getBooleanFieldWithDefault(msg, 5, false),
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
 * @return {!proto.pps.ListJobRequest}
 */
proto.pps.ListJobRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.ListJobRequest;
  return proto.pps.ListJobRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.ListJobRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.ListJobRequest}
 */
proto.pps.ListJobRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.addInputCommit(value);
      break;
    case 3:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setOutputCommit(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setHistory(value);
      break;
    case 5:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setFull(value);
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
proto.pps.ListJobRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.ListJobRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.ListJobRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ListJobRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
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
  f = message.getOutputCommit();
  if (f != null) {
    writer.writeMessage(
      3,
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
  f = message.getFull();
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
 * @return {?proto.pps.Pipeline}
 */
proto.pps.ListJobRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 1));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.ListJobRequest} returns this
*/
proto.pps.ListJobRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.ListJobRequest} returns this
 */
proto.pps.ListJobRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.ListJobRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * repeated pfs.Commit input_commit = 2;
 * @return {!Array<!proto.pfs.Commit>}
 */
proto.pps.ListJobRequest.prototype.getInputCommitList = function() {
  return /** @type{!Array<!proto.pfs.Commit>} */ (
    jspb.Message.getRepeatedWrapperField(this, pfs_pfs_pb.Commit, 2));
};


/**
 * @param {!Array<!proto.pfs.Commit>} value
 * @return {!proto.pps.ListJobRequest} returns this
*/
proto.pps.ListJobRequest.prototype.setInputCommitList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pfs.Commit=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Commit}
 */
proto.pps.ListJobRequest.prototype.addInputCommit = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pfs.Commit, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.ListJobRequest} returns this
 */
proto.pps.ListJobRequest.prototype.clearInputCommitList = function() {
  return this.setInputCommitList([]);
};


/**
 * optional pfs.Commit output_commit = 3;
 * @return {?proto.pfs.Commit}
 */
proto.pps.ListJobRequest.prototype.getOutputCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 3));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.ListJobRequest} returns this
*/
proto.pps.ListJobRequest.prototype.setOutputCommit = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.ListJobRequest} returns this
 */
proto.pps.ListJobRequest.prototype.clearOutputCommit = function() {
  return this.setOutputCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.ListJobRequest.prototype.hasOutputCommit = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional int64 history = 4;
 * @return {number}
 */
proto.pps.ListJobRequest.prototype.getHistory = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.ListJobRequest} returns this
 */
proto.pps.ListJobRequest.prototype.setHistory = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional bool full = 5;
 * @return {boolean}
 */
proto.pps.ListJobRequest.prototype.getFull = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 5, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.ListJobRequest} returns this
 */
proto.pps.ListJobRequest.prototype.setFull = function(value) {
  return jspb.Message.setProto3BooleanField(this, 5, value);
};


/**
 * optional string jqFilter = 6;
 * @return {string}
 */
proto.pps.ListJobRequest.prototype.getJqfilter = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.ListJobRequest} returns this
 */
proto.pps.ListJobRequest.prototype.setJqfilter = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.FlushJobRequest.repeatedFields_ = [1,2];



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
proto.pps.FlushJobRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.FlushJobRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.FlushJobRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.FlushJobRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    commitsList: jspb.Message.toObjectList(msg.getCommitsList(),
    pfs_pfs_pb.Commit.toObject, includeInstance),
    toPipelinesList: jspb.Message.toObjectList(msg.getToPipelinesList(),
    proto.pps.Pipeline.toObject, includeInstance)
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
 * @return {!proto.pps.FlushJobRequest}
 */
proto.pps.FlushJobRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.FlushJobRequest;
  return proto.pps.FlushJobRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.FlushJobRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.FlushJobRequest}
 */
proto.pps.FlushJobRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.addCommits(value);
      break;
    case 2:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
      msg.addToPipelines(value);
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
proto.pps.FlushJobRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.FlushJobRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.FlushJobRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.FlushJobRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCommitsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getToPipelinesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
};


/**
 * repeated pfs.Commit commits = 1;
 * @return {!Array<!proto.pfs.Commit>}
 */
proto.pps.FlushJobRequest.prototype.getCommitsList = function() {
  return /** @type{!Array<!proto.pfs.Commit>} */ (
    jspb.Message.getRepeatedWrapperField(this, pfs_pfs_pb.Commit, 1));
};


/**
 * @param {!Array<!proto.pfs.Commit>} value
 * @return {!proto.pps.FlushJobRequest} returns this
*/
proto.pps.FlushJobRequest.prototype.setCommitsList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pfs.Commit=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.Commit}
 */
proto.pps.FlushJobRequest.prototype.addCommits = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pfs.Commit, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.FlushJobRequest} returns this
 */
proto.pps.FlushJobRequest.prototype.clearCommitsList = function() {
  return this.setCommitsList([]);
};


/**
 * repeated Pipeline to_pipelines = 2;
 * @return {!Array<!proto.pps.Pipeline>}
 */
proto.pps.FlushJobRequest.prototype.getToPipelinesList = function() {
  return /** @type{!Array<!proto.pps.Pipeline>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps.Pipeline, 2));
};


/**
 * @param {!Array<!proto.pps.Pipeline>} value
 * @return {!proto.pps.FlushJobRequest} returns this
*/
proto.pps.FlushJobRequest.prototype.setToPipelinesList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pps.Pipeline=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps.Pipeline}
 */
proto.pps.FlushJobRequest.prototype.addToPipelines = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pps.Pipeline, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.FlushJobRequest} returns this
 */
proto.pps.FlushJobRequest.prototype.clearToPipelinesList = function() {
  return this.setToPipelinesList([]);
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
proto.pps.DeleteJobRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.DeleteJobRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.DeleteJobRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.DeleteJobRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps.Job.toObject(includeInstance, f)
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
 * @return {!proto.pps.DeleteJobRequest}
 */
proto.pps.DeleteJobRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.DeleteJobRequest;
  return proto.pps.DeleteJobRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.DeleteJobRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.DeleteJobRequest}
 */
proto.pps.DeleteJobRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Job;
      reader.readMessage(value,proto.pps.Job.deserializeBinaryFromReader);
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
proto.pps.DeleteJobRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.DeleteJobRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.DeleteJobRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.DeleteJobRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Job.serializeBinaryToWriter
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps.Job}
 */
proto.pps.DeleteJobRequest.prototype.getJob = function() {
  return /** @type{?proto.pps.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps.Job, 1));
};


/**
 * @param {?proto.pps.Job|undefined} value
 * @return {!proto.pps.DeleteJobRequest} returns this
*/
proto.pps.DeleteJobRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.DeleteJobRequest} returns this
 */
proto.pps.DeleteJobRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.DeleteJobRequest.prototype.hasJob = function() {
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
proto.pps.StopJobRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.StopJobRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.StopJobRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.StopJobRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps.Job.toObject(includeInstance, f),
    outputCommit: (f = msg.getOutputCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f)
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
 * @return {!proto.pps.StopJobRequest}
 */
proto.pps.StopJobRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.StopJobRequest;
  return proto.pps.StopJobRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.StopJobRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.StopJobRequest}
 */
proto.pps.StopJobRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Job;
      reader.readMessage(value,proto.pps.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 2:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setOutputCommit(value);
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
proto.pps.StopJobRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.StopJobRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.StopJobRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.StopJobRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Job.serializeBinaryToWriter
    );
  }
  f = message.getOutputCommit();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps.Job}
 */
proto.pps.StopJobRequest.prototype.getJob = function() {
  return /** @type{?proto.pps.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps.Job, 1));
};


/**
 * @param {?proto.pps.Job|undefined} value
 * @return {!proto.pps.StopJobRequest} returns this
*/
proto.pps.StopJobRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.StopJobRequest} returns this
 */
proto.pps.StopJobRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.StopJobRequest.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional pfs.Commit output_commit = 2;
 * @return {?proto.pfs.Commit}
 */
proto.pps.StopJobRequest.prototype.getOutputCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 2));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.StopJobRequest} returns this
*/
proto.pps.StopJobRequest.prototype.setOutputCommit = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.StopJobRequest} returns this
 */
proto.pps.StopJobRequest.prototype.clearOutputCommit = function() {
  return this.setOutputCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.StopJobRequest.prototype.hasOutputCommit = function() {
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
proto.pps.UpdateJobStateRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.UpdateJobStateRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.UpdateJobStateRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.UpdateJobStateRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps.Job.toObject(includeInstance, f),
    state: jspb.Message.getFieldWithDefault(msg, 2, 0),
    reason: jspb.Message.getFieldWithDefault(msg, 3, ""),
    restart: jspb.Message.getFieldWithDefault(msg, 4, 0),
    dataProcessed: jspb.Message.getFieldWithDefault(msg, 5, 0),
    dataSkipped: jspb.Message.getFieldWithDefault(msg, 6, 0),
    dataFailed: jspb.Message.getFieldWithDefault(msg, 7, 0),
    dataRecovered: jspb.Message.getFieldWithDefault(msg, 8, 0),
    dataTotal: jspb.Message.getFieldWithDefault(msg, 9, 0),
    stats: (f = msg.getStats()) && proto.pps.ProcessStats.toObject(includeInstance, f)
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
 * @return {!proto.pps.UpdateJobStateRequest}
 */
proto.pps.UpdateJobStateRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.UpdateJobStateRequest;
  return proto.pps.UpdateJobStateRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.UpdateJobStateRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.UpdateJobStateRequest}
 */
proto.pps.UpdateJobStateRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Job;
      reader.readMessage(value,proto.pps.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 2:
      var value = /** @type {!proto.pps.JobState} */ (reader.readEnum());
      msg.setState(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setReason(value);
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
      msg.setDataFailed(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataRecovered(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDataTotal(value);
      break;
    case 10:
      var value = new proto.pps.ProcessStats;
      reader.readMessage(value,proto.pps.ProcessStats.deserializeBinaryFromReader);
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
proto.pps.UpdateJobStateRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.UpdateJobStateRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.UpdateJobStateRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.UpdateJobStateRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Job.serializeBinaryToWriter
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
  f = message.getDataFailed();
  if (f !== 0) {
    writer.writeInt64(
      7,
      f
    );
  }
  f = message.getDataRecovered();
  if (f !== 0) {
    writer.writeInt64(
      8,
      f
    );
  }
  f = message.getDataTotal();
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
      proto.pps.ProcessStats.serializeBinaryToWriter
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps.Job}
 */
proto.pps.UpdateJobStateRequest.prototype.getJob = function() {
  return /** @type{?proto.pps.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps.Job, 1));
};


/**
 * @param {?proto.pps.Job|undefined} value
 * @return {!proto.pps.UpdateJobStateRequest} returns this
*/
proto.pps.UpdateJobStateRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.UpdateJobStateRequest} returns this
 */
proto.pps.UpdateJobStateRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.UpdateJobStateRequest.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional JobState state = 2;
 * @return {!proto.pps.JobState}
 */
proto.pps.UpdateJobStateRequest.prototype.getState = function() {
  return /** @type {!proto.pps.JobState} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {!proto.pps.JobState} value
 * @return {!proto.pps.UpdateJobStateRequest} returns this
 */
proto.pps.UpdateJobStateRequest.prototype.setState = function(value) {
  return jspb.Message.setProto3EnumField(this, 2, value);
};


/**
 * optional string reason = 3;
 * @return {string}
 */
proto.pps.UpdateJobStateRequest.prototype.getReason = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.UpdateJobStateRequest} returns this
 */
proto.pps.UpdateJobStateRequest.prototype.setReason = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional uint64 restart = 4;
 * @return {number}
 */
proto.pps.UpdateJobStateRequest.prototype.getRestart = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.UpdateJobStateRequest} returns this
 */
proto.pps.UpdateJobStateRequest.prototype.setRestart = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional int64 data_processed = 5;
 * @return {number}
 */
proto.pps.UpdateJobStateRequest.prototype.getDataProcessed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.UpdateJobStateRequest} returns this
 */
proto.pps.UpdateJobStateRequest.prototype.setDataProcessed = function(value) {
  return jspb.Message.setProto3IntField(this, 5, value);
};


/**
 * optional int64 data_skipped = 6;
 * @return {number}
 */
proto.pps.UpdateJobStateRequest.prototype.getDataSkipped = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.UpdateJobStateRequest} returns this
 */
proto.pps.UpdateJobStateRequest.prototype.setDataSkipped = function(value) {
  return jspb.Message.setProto3IntField(this, 6, value);
};


/**
 * optional int64 data_failed = 7;
 * @return {number}
 */
proto.pps.UpdateJobStateRequest.prototype.getDataFailed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.UpdateJobStateRequest} returns this
 */
proto.pps.UpdateJobStateRequest.prototype.setDataFailed = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
};


/**
 * optional int64 data_recovered = 8;
 * @return {number}
 */
proto.pps.UpdateJobStateRequest.prototype.getDataRecovered = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.UpdateJobStateRequest} returns this
 */
proto.pps.UpdateJobStateRequest.prototype.setDataRecovered = function(value) {
  return jspb.Message.setProto3IntField(this, 8, value);
};


/**
 * optional int64 data_total = 9;
 * @return {number}
 */
proto.pps.UpdateJobStateRequest.prototype.getDataTotal = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.UpdateJobStateRequest} returns this
 */
proto.pps.UpdateJobStateRequest.prototype.setDataTotal = function(value) {
  return jspb.Message.setProto3IntField(this, 9, value);
};


/**
 * optional ProcessStats stats = 10;
 * @return {?proto.pps.ProcessStats}
 */
proto.pps.UpdateJobStateRequest.prototype.getStats = function() {
  return /** @type{?proto.pps.ProcessStats} */ (
    jspb.Message.getWrapperField(this, proto.pps.ProcessStats, 10));
};


/**
 * @param {?proto.pps.ProcessStats|undefined} value
 * @return {!proto.pps.UpdateJobStateRequest} returns this
*/
proto.pps.UpdateJobStateRequest.prototype.setStats = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.UpdateJobStateRequest} returns this
 */
proto.pps.UpdateJobStateRequest.prototype.clearStats = function() {
  return this.setStats(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.UpdateJobStateRequest.prototype.hasStats = function() {
  return jspb.Message.getField(this, 10) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.GetLogsRequest.repeatedFields_ = [3];



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
proto.pps.GetLogsRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.GetLogsRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.GetLogsRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.GetLogsRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f),
    job: (f = msg.getJob()) && proto.pps.Job.toObject(includeInstance, f),
    dataFiltersList: (f = jspb.Message.getRepeatedField(msg, 3)) == null ? undefined : f,
    datum: (f = msg.getDatum()) && proto.pps.Datum.toObject(includeInstance, f),
    master: jspb.Message.getBooleanFieldWithDefault(msg, 5, false),
    follow: jspb.Message.getBooleanFieldWithDefault(msg, 7, false),
    tail: jspb.Message.getFieldWithDefault(msg, 8, 0),
    useLokiBackend: jspb.Message.getBooleanFieldWithDefault(msg, 9, false),
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
 * @return {!proto.pps.GetLogsRequest}
 */
proto.pps.GetLogsRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.GetLogsRequest;
  return proto.pps.GetLogsRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.GetLogsRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.GetLogsRequest}
 */
proto.pps.GetLogsRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 2:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 1:
      var value = new proto.pps.Job;
      reader.readMessage(value,proto.pps.Job.deserializeBinaryFromReader);
      msg.setJob(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.addDataFilters(value);
      break;
    case 6:
      var value = new proto.pps.Datum;
      reader.readMessage(value,proto.pps.Datum.deserializeBinaryFromReader);
      msg.setDatum(value);
      break;
    case 5:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setMaster(value);
      break;
    case 7:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setFollow(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setTail(value);
      break;
    case 9:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setUseLokiBackend(value);
      break;
    case 10:
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
proto.pps.GetLogsRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.GetLogsRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.GetLogsRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.GetLogsRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Job.serializeBinaryToWriter
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
      6,
      f,
      proto.pps.Datum.serializeBinaryToWriter
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
      7,
      f
    );
  }
  f = message.getTail();
  if (f !== 0) {
    writer.writeInt64(
      8,
      f
    );
  }
  f = message.getUseLokiBackend();
  if (f) {
    writer.writeBool(
      9,
      f
    );
  }
  f = message.getSince();
  if (f != null) {
    writer.writeMessage(
      10,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
};


/**
 * optional Pipeline pipeline = 2;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.GetLogsRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 2));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.GetLogsRequest} returns this
*/
proto.pps.GetLogsRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.GetLogsRequest} returns this
 */
proto.pps.GetLogsRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.GetLogsRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional Job job = 1;
 * @return {?proto.pps.Job}
 */
proto.pps.GetLogsRequest.prototype.getJob = function() {
  return /** @type{?proto.pps.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps.Job, 1));
};


/**
 * @param {?proto.pps.Job|undefined} value
 * @return {!proto.pps.GetLogsRequest} returns this
*/
proto.pps.GetLogsRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.GetLogsRequest} returns this
 */
proto.pps.GetLogsRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.GetLogsRequest.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * repeated string data_filters = 3;
 * @return {!Array<string>}
 */
proto.pps.GetLogsRequest.prototype.getDataFiltersList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 3));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps.GetLogsRequest} returns this
 */
proto.pps.GetLogsRequest.prototype.setDataFiltersList = function(value) {
  return jspb.Message.setField(this, 3, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps.GetLogsRequest} returns this
 */
proto.pps.GetLogsRequest.prototype.addDataFilters = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 3, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.GetLogsRequest} returns this
 */
proto.pps.GetLogsRequest.prototype.clearDataFiltersList = function() {
  return this.setDataFiltersList([]);
};


/**
 * optional Datum datum = 6;
 * @return {?proto.pps.Datum}
 */
proto.pps.GetLogsRequest.prototype.getDatum = function() {
  return /** @type{?proto.pps.Datum} */ (
    jspb.Message.getWrapperField(this, proto.pps.Datum, 6));
};


/**
 * @param {?proto.pps.Datum|undefined} value
 * @return {!proto.pps.GetLogsRequest} returns this
*/
proto.pps.GetLogsRequest.prototype.setDatum = function(value) {
  return jspb.Message.setWrapperField(this, 6, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.GetLogsRequest} returns this
 */
proto.pps.GetLogsRequest.prototype.clearDatum = function() {
  return this.setDatum(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.GetLogsRequest.prototype.hasDatum = function() {
  return jspb.Message.getField(this, 6) != null;
};


/**
 * optional bool master = 5;
 * @return {boolean}
 */
proto.pps.GetLogsRequest.prototype.getMaster = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 5, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.GetLogsRequest} returns this
 */
proto.pps.GetLogsRequest.prototype.setMaster = function(value) {
  return jspb.Message.setProto3BooleanField(this, 5, value);
};


/**
 * optional bool follow = 7;
 * @return {boolean}
 */
proto.pps.GetLogsRequest.prototype.getFollow = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 7, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.GetLogsRequest} returns this
 */
proto.pps.GetLogsRequest.prototype.setFollow = function(value) {
  return jspb.Message.setProto3BooleanField(this, 7, value);
};


/**
 * optional int64 tail = 8;
 * @return {number}
 */
proto.pps.GetLogsRequest.prototype.getTail = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.GetLogsRequest} returns this
 */
proto.pps.GetLogsRequest.prototype.setTail = function(value) {
  return jspb.Message.setProto3IntField(this, 8, value);
};


/**
 * optional bool use_loki_backend = 9;
 * @return {boolean}
 */
proto.pps.GetLogsRequest.prototype.getUseLokiBackend = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 9, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.GetLogsRequest} returns this
 */
proto.pps.GetLogsRequest.prototype.setUseLokiBackend = function(value) {
  return jspb.Message.setProto3BooleanField(this, 9, value);
};


/**
 * optional google.protobuf.Duration since = 10;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps.GetLogsRequest.prototype.getSince = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 10));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps.GetLogsRequest} returns this
*/
proto.pps.GetLogsRequest.prototype.setSince = function(value) {
  return jspb.Message.setWrapperField(this, 10, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.GetLogsRequest} returns this
 */
proto.pps.GetLogsRequest.prototype.clearSince = function() {
  return this.setSince(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.GetLogsRequest.prototype.hasSince = function() {
  return jspb.Message.getField(this, 10) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.LogMessage.repeatedFields_ = [4];



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
proto.pps.LogMessage.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.LogMessage.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.LogMessage} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.LogMessage.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipelineName: jspb.Message.getFieldWithDefault(msg, 1, ""),
    jobId: jspb.Message.getFieldWithDefault(msg, 3, ""),
    workerId: jspb.Message.getFieldWithDefault(msg, 7, ""),
    datumId: jspb.Message.getFieldWithDefault(msg, 9, ""),
    master: jspb.Message.getBooleanFieldWithDefault(msg, 10, false),
    dataList: jspb.Message.toObjectList(msg.getDataList(),
    proto.pps.InputFile.toObject, includeInstance),
    user: jspb.Message.getBooleanFieldWithDefault(msg, 8, false),
    ts: (f = msg.getTs()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    message: jspb.Message.getFieldWithDefault(msg, 6, "")
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
 * @return {!proto.pps.LogMessage}
 */
proto.pps.LogMessage.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.LogMessage;
  return proto.pps.LogMessage.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.LogMessage} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.LogMessage}
 */
proto.pps.LogMessage.deserializeBinaryFromReader = function(msg, reader) {
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
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setJobId(value);
      break;
    case 7:
      var value = /** @type {string} */ (reader.readString());
      msg.setWorkerId(value);
      break;
    case 9:
      var value = /** @type {string} */ (reader.readString());
      msg.setDatumId(value);
      break;
    case 10:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setMaster(value);
      break;
    case 4:
      var value = new proto.pps.InputFile;
      reader.readMessage(value,proto.pps.InputFile.deserializeBinaryFromReader);
      msg.addData(value);
      break;
    case 8:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setUser(value);
      break;
    case 5:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setTs(value);
      break;
    case 6:
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
proto.pps.LogMessage.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.LogMessage.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.LogMessage} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.LogMessage.serializeBinaryToWriter = function(message, writer) {
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
      3,
      f
    );
  }
  f = message.getWorkerId();
  if (f.length > 0) {
    writer.writeString(
      7,
      f
    );
  }
  f = message.getDatumId();
  if (f.length > 0) {
    writer.writeString(
      9,
      f
    );
  }
  f = message.getMaster();
  if (f) {
    writer.writeBool(
      10,
      f
    );
  }
  f = message.getDataList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      4,
      f,
      proto.pps.InputFile.serializeBinaryToWriter
    );
  }
  f = message.getUser();
  if (f) {
    writer.writeBool(
      8,
      f
    );
  }
  f = message.getTs();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getMessage();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
};


/**
 * optional string pipeline_name = 1;
 * @return {string}
 */
proto.pps.LogMessage.prototype.getPipelineName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.LogMessage} returns this
 */
proto.pps.LogMessage.prototype.setPipelineName = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string job_id = 3;
 * @return {string}
 */
proto.pps.LogMessage.prototype.getJobId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.LogMessage} returns this
 */
proto.pps.LogMessage.prototype.setJobId = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional string worker_id = 7;
 * @return {string}
 */
proto.pps.LogMessage.prototype.getWorkerId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 7, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.LogMessage} returns this
 */
proto.pps.LogMessage.prototype.setWorkerId = function(value) {
  return jspb.Message.setProto3StringField(this, 7, value);
};


/**
 * optional string datum_id = 9;
 * @return {string}
 */
proto.pps.LogMessage.prototype.getDatumId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 9, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.LogMessage} returns this
 */
proto.pps.LogMessage.prototype.setDatumId = function(value) {
  return jspb.Message.setProto3StringField(this, 9, value);
};


/**
 * optional bool master = 10;
 * @return {boolean}
 */
proto.pps.LogMessage.prototype.getMaster = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 10, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.LogMessage} returns this
 */
proto.pps.LogMessage.prototype.setMaster = function(value) {
  return jspb.Message.setProto3BooleanField(this, 10, value);
};


/**
 * repeated InputFile data = 4;
 * @return {!Array<!proto.pps.InputFile>}
 */
proto.pps.LogMessage.prototype.getDataList = function() {
  return /** @type{!Array<!proto.pps.InputFile>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps.InputFile, 4));
};


/**
 * @param {!Array<!proto.pps.InputFile>} value
 * @return {!proto.pps.LogMessage} returns this
*/
proto.pps.LogMessage.prototype.setDataList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 4, value);
};


/**
 * @param {!proto.pps.InputFile=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps.InputFile}
 */
proto.pps.LogMessage.prototype.addData = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 4, opt_value, proto.pps.InputFile, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.LogMessage} returns this
 */
proto.pps.LogMessage.prototype.clearDataList = function() {
  return this.setDataList([]);
};


/**
 * optional bool user = 8;
 * @return {boolean}
 */
proto.pps.LogMessage.prototype.getUser = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 8, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.LogMessage} returns this
 */
proto.pps.LogMessage.prototype.setUser = function(value) {
  return jspb.Message.setProto3BooleanField(this, 8, value);
};


/**
 * optional google.protobuf.Timestamp ts = 5;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps.LogMessage.prototype.getTs = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 5));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps.LogMessage} returns this
*/
proto.pps.LogMessage.prototype.setTs = function(value) {
  return jspb.Message.setWrapperField(this, 5, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.LogMessage} returns this
 */
proto.pps.LogMessage.prototype.clearTs = function() {
  return this.setTs(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.LogMessage.prototype.hasTs = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional string message = 6;
 * @return {string}
 */
proto.pps.LogMessage.prototype.getMessage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.LogMessage} returns this
 */
proto.pps.LogMessage.prototype.setMessage = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.RestartDatumRequest.repeatedFields_ = [2];



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
proto.pps.RestartDatumRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.RestartDatumRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.RestartDatumRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.RestartDatumRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps.Job.toObject(includeInstance, f),
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
 * @return {!proto.pps.RestartDatumRequest}
 */
proto.pps.RestartDatumRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.RestartDatumRequest;
  return proto.pps.RestartDatumRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.RestartDatumRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.RestartDatumRequest}
 */
proto.pps.RestartDatumRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Job;
      reader.readMessage(value,proto.pps.Job.deserializeBinaryFromReader);
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
proto.pps.RestartDatumRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.RestartDatumRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.RestartDatumRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.RestartDatumRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Job.serializeBinaryToWriter
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
 * @return {?proto.pps.Job}
 */
proto.pps.RestartDatumRequest.prototype.getJob = function() {
  return /** @type{?proto.pps.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps.Job, 1));
};


/**
 * @param {?proto.pps.Job|undefined} value
 * @return {!proto.pps.RestartDatumRequest} returns this
*/
proto.pps.RestartDatumRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.RestartDatumRequest} returns this
 */
proto.pps.RestartDatumRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.RestartDatumRequest.prototype.hasJob = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * repeated string data_filters = 2;
 * @return {!Array<string>}
 */
proto.pps.RestartDatumRequest.prototype.getDataFiltersList = function() {
  return /** @type {!Array<string>} */ (jspb.Message.getRepeatedField(this, 2));
};


/**
 * @param {!Array<string>} value
 * @return {!proto.pps.RestartDatumRequest} returns this
 */
proto.pps.RestartDatumRequest.prototype.setDataFiltersList = function(value) {
  return jspb.Message.setField(this, 2, value || []);
};


/**
 * @param {string} value
 * @param {number=} opt_index
 * @return {!proto.pps.RestartDatumRequest} returns this
 */
proto.pps.RestartDatumRequest.prototype.addDataFilters = function(value, opt_index) {
  return jspb.Message.addToRepeatedField(this, 2, value, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.RestartDatumRequest} returns this
 */
proto.pps.RestartDatumRequest.prototype.clearDataFiltersList = function() {
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
proto.pps.InspectDatumRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.InspectDatumRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.InspectDatumRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.InspectDatumRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    datum: (f = msg.getDatum()) && proto.pps.Datum.toObject(includeInstance, f)
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
 * @return {!proto.pps.InspectDatumRequest}
 */
proto.pps.InspectDatumRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.InspectDatumRequest;
  return proto.pps.InspectDatumRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.InspectDatumRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.InspectDatumRequest}
 */
proto.pps.InspectDatumRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Datum;
      reader.readMessage(value,proto.pps.Datum.deserializeBinaryFromReader);
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
proto.pps.InspectDatumRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.InspectDatumRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.InspectDatumRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.InspectDatumRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getDatum();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Datum.serializeBinaryToWriter
    );
  }
};


/**
 * optional Datum datum = 1;
 * @return {?proto.pps.Datum}
 */
proto.pps.InspectDatumRequest.prototype.getDatum = function() {
  return /** @type{?proto.pps.Datum} */ (
    jspb.Message.getWrapperField(this, proto.pps.Datum, 1));
};


/**
 * @param {?proto.pps.Datum|undefined} value
 * @return {!proto.pps.InspectDatumRequest} returns this
*/
proto.pps.InspectDatumRequest.prototype.setDatum = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.InspectDatumRequest} returns this
 */
proto.pps.InspectDatumRequest.prototype.clearDatum = function() {
  return this.setDatum(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.InspectDatumRequest.prototype.hasDatum = function() {
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
proto.pps.ListDatumRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.ListDatumRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.ListDatumRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ListDatumRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    job: (f = msg.getJob()) && proto.pps.Job.toObject(includeInstance, f)
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
 * @return {!proto.pps.ListDatumRequest}
 */
proto.pps.ListDatumRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.ListDatumRequest;
  return proto.pps.ListDatumRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.ListDatumRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.ListDatumRequest}
 */
proto.pps.ListDatumRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Job;
      reader.readMessage(value,proto.pps.Job.deserializeBinaryFromReader);
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
proto.pps.ListDatumRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.ListDatumRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.ListDatumRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ListDatumRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getJob();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Job.serializeBinaryToWriter
    );
  }
};


/**
 * optional Job job = 1;
 * @return {?proto.pps.Job}
 */
proto.pps.ListDatumRequest.prototype.getJob = function() {
  return /** @type{?proto.pps.Job} */ (
    jspb.Message.getWrapperField(this, proto.pps.Job, 1));
};


/**
 * @param {?proto.pps.Job|undefined} value
 * @return {!proto.pps.ListDatumRequest} returns this
*/
proto.pps.ListDatumRequest.prototype.setJob = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.ListDatumRequest} returns this
 */
proto.pps.ListDatumRequest.prototype.clearJob = function() {
  return this.setJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.ListDatumRequest.prototype.hasJob = function() {
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
proto.pps.ChunkSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.ChunkSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.ChunkSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ChunkSpec.toObject = function(includeInstance, msg) {
  var f, obj = {
    number: jspb.Message.getFieldWithDefault(msg, 1, 0),
    sizeBytes: jspb.Message.getFieldWithDefault(msg, 2, 0)
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
 * @return {!proto.pps.ChunkSpec}
 */
proto.pps.ChunkSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.ChunkSpec;
  return proto.pps.ChunkSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.ChunkSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.ChunkSpec}
 */
proto.pps.ChunkSpec.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.ChunkSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.ChunkSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.ChunkSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ChunkSpec.serializeBinaryToWriter = function(message, writer) {
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
};


/**
 * optional int64 number = 1;
 * @return {number}
 */
proto.pps.ChunkSpec.prototype.getNumber = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.ChunkSpec} returns this
 */
proto.pps.ChunkSpec.prototype.setNumber = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional int64 size_bytes = 2;
 * @return {number}
 */
proto.pps.ChunkSpec.prototype.getSizeBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.ChunkSpec} returns this
 */
proto.pps.ChunkSpec.prototype.setSizeBytes = function(value) {
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
proto.pps.SchedulingSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.SchedulingSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.SchedulingSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.SchedulingSpec.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.SchedulingSpec}
 */
proto.pps.SchedulingSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.SchedulingSpec;
  return proto.pps.SchedulingSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.SchedulingSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.SchedulingSpec}
 */
proto.pps.SchedulingSpec.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.SchedulingSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.SchedulingSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.SchedulingSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.SchedulingSpec.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.SchedulingSpec.prototype.getNodeSelectorMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      null));
};


/**
 * Clears values from the map. The map will be non-null.
 * @return {!proto.pps.SchedulingSpec} returns this
 */
proto.pps.SchedulingSpec.prototype.clearNodeSelectorMap = function() {
  this.getNodeSelectorMap().clear();
  return this;};


/**
 * optional string priority_class_name = 2;
 * @return {string}
 */
proto.pps.SchedulingSpec.prototype.getPriorityClassName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.SchedulingSpec} returns this
 */
proto.pps.SchedulingSpec.prototype.setPriorityClassName = function(value) {
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
proto.pps.CreatePipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.CreatePipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.CreatePipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.CreatePipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f),
    tfJob: (f = msg.getTfJob()) && proto.pps.TFJob.toObject(includeInstance, f),
    transform: (f = msg.getTransform()) && proto.pps.Transform.toObject(includeInstance, f),
    parallelismSpec: (f = msg.getParallelismSpec()) && proto.pps.ParallelismSpec.toObject(includeInstance, f),
    egress: (f = msg.getEgress()) && proto.pps.Egress.toObject(includeInstance, f),
    update: jspb.Message.getBooleanFieldWithDefault(msg, 5, false),
    outputBranch: jspb.Message.getFieldWithDefault(msg, 10, ""),
    s3Out: jspb.Message.getBooleanFieldWithDefault(msg, 36, false),
    resourceRequests: (f = msg.getResourceRequests()) && proto.pps.ResourceSpec.toObject(includeInstance, f),
    resourceLimits: (f = msg.getResourceLimits()) && proto.pps.ResourceSpec.toObject(includeInstance, f),
    sidecarResourceLimits: (f = msg.getSidecarResourceLimits()) && proto.pps.ResourceSpec.toObject(includeInstance, f),
    input: (f = msg.getInput()) && proto.pps.Input.toObject(includeInstance, f),
    description: jspb.Message.getFieldWithDefault(msg, 14, ""),
    cacheSize: jspb.Message.getFieldWithDefault(msg, 16, ""),
    enableStats: jspb.Message.getBooleanFieldWithDefault(msg, 17, false),
    reprocess: jspb.Message.getBooleanFieldWithDefault(msg, 18, false),
    maxQueueSize: jspb.Message.getFieldWithDefault(msg, 20, 0),
    service: (f = msg.getService()) && proto.pps.Service.toObject(includeInstance, f),
    spout: (f = msg.getSpout()) && proto.pps.Spout.toObject(includeInstance, f),
    chunkSpec: (f = msg.getChunkSpec()) && proto.pps.ChunkSpec.toObject(includeInstance, f),
    datumTimeout: (f = msg.getDatumTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    jobTimeout: (f = msg.getJobTimeout()) && google_protobuf_duration_pb.Duration.toObject(includeInstance, f),
    salt: jspb.Message.getFieldWithDefault(msg, 26, ""),
    standby: jspb.Message.getBooleanFieldWithDefault(msg, 27, false),
    datumTries: jspb.Message.getFieldWithDefault(msg, 28, 0),
    schedulingSpec: (f = msg.getSchedulingSpec()) && proto.pps.SchedulingSpec.toObject(includeInstance, f),
    podSpec: jspb.Message.getFieldWithDefault(msg, 30, ""),
    podPatch: jspb.Message.getFieldWithDefault(msg, 32, ""),
    specCommit: (f = msg.getSpecCommit()) && pfs_pfs_pb.Commit.toObject(includeInstance, f),
    metadata: (f = msg.getMetadata()) && proto.pps.Metadata.toObject(includeInstance, f),
    noSkip: jspb.Message.getBooleanFieldWithDefault(msg, 48, false)
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
 * @return {!proto.pps.CreatePipelineRequest}
 */
proto.pps.CreatePipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.CreatePipelineRequest;
  return proto.pps.CreatePipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.CreatePipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.CreatePipelineRequest}
 */
proto.pps.CreatePipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 35:
      var value = new proto.pps.TFJob;
      reader.readMessage(value,proto.pps.TFJob.deserializeBinaryFromReader);
      msg.setTfJob(value);
      break;
    case 2:
      var value = new proto.pps.Transform;
      reader.readMessage(value,proto.pps.Transform.deserializeBinaryFromReader);
      msg.setTransform(value);
      break;
    case 7:
      var value = new proto.pps.ParallelismSpec;
      reader.readMessage(value,proto.pps.ParallelismSpec.deserializeBinaryFromReader);
      msg.setParallelismSpec(value);
      break;
    case 9:
      var value = new proto.pps.Egress;
      reader.readMessage(value,proto.pps.Egress.deserializeBinaryFromReader);
      msg.setEgress(value);
      break;
    case 5:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setUpdate(value);
      break;
    case 10:
      var value = /** @type {string} */ (reader.readString());
      msg.setOutputBranch(value);
      break;
    case 36:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setS3Out(value);
      break;
    case 12:
      var value = new proto.pps.ResourceSpec;
      reader.readMessage(value,proto.pps.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceRequests(value);
      break;
    case 22:
      var value = new proto.pps.ResourceSpec;
      reader.readMessage(value,proto.pps.ResourceSpec.deserializeBinaryFromReader);
      msg.setResourceLimits(value);
      break;
    case 47:
      var value = new proto.pps.ResourceSpec;
      reader.readMessage(value,proto.pps.ResourceSpec.deserializeBinaryFromReader);
      msg.setSidecarResourceLimits(value);
      break;
    case 13:
      var value = new proto.pps.Input;
      reader.readMessage(value,proto.pps.Input.deserializeBinaryFromReader);
      msg.setInput(value);
      break;
    case 14:
      var value = /** @type {string} */ (reader.readString());
      msg.setDescription(value);
      break;
    case 16:
      var value = /** @type {string} */ (reader.readString());
      msg.setCacheSize(value);
      break;
    case 17:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEnableStats(value);
      break;
    case 18:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setReprocess(value);
      break;
    case 20:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setMaxQueueSize(value);
      break;
    case 21:
      var value = new proto.pps.Service;
      reader.readMessage(value,proto.pps.Service.deserializeBinaryFromReader);
      msg.setService(value);
      break;
    case 33:
      var value = new proto.pps.Spout;
      reader.readMessage(value,proto.pps.Spout.deserializeBinaryFromReader);
      msg.setSpout(value);
      break;
    case 23:
      var value = new proto.pps.ChunkSpec;
      reader.readMessage(value,proto.pps.ChunkSpec.deserializeBinaryFromReader);
      msg.setChunkSpec(value);
      break;
    case 24:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setDatumTimeout(value);
      break;
    case 25:
      var value = new google_protobuf_duration_pb.Duration;
      reader.readMessage(value,google_protobuf_duration_pb.Duration.deserializeBinaryFromReader);
      msg.setJobTimeout(value);
      break;
    case 26:
      var value = /** @type {string} */ (reader.readString());
      msg.setSalt(value);
      break;
    case 27:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setStandby(value);
      break;
    case 28:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setDatumTries(value);
      break;
    case 29:
      var value = new proto.pps.SchedulingSpec;
      reader.readMessage(value,proto.pps.SchedulingSpec.deserializeBinaryFromReader);
      msg.setSchedulingSpec(value);
      break;
    case 30:
      var value = /** @type {string} */ (reader.readString());
      msg.setPodSpec(value);
      break;
    case 32:
      var value = /** @type {string} */ (reader.readString());
      msg.setPodPatch(value);
      break;
    case 34:
      var value = new pfs_pfs_pb.Commit;
      reader.readMessage(value,pfs_pfs_pb.Commit.deserializeBinaryFromReader);
      msg.setSpecCommit(value);
      break;
    case 46:
      var value = new proto.pps.Metadata;
      reader.readMessage(value,proto.pps.Metadata.deserializeBinaryFromReader);
      msg.setMetadata(value);
      break;
    case 48:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setNoSkip(value);
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
proto.pps.CreatePipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.CreatePipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.CreatePipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.CreatePipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getTfJob();
  if (f != null) {
    writer.writeMessage(
      35,
      f,
      proto.pps.TFJob.serializeBinaryToWriter
    );
  }
  f = message.getTransform();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.pps.Transform.serializeBinaryToWriter
    );
  }
  f = message.getParallelismSpec();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      proto.pps.ParallelismSpec.serializeBinaryToWriter
    );
  }
  f = message.getEgress();
  if (f != null) {
    writer.writeMessage(
      9,
      f,
      proto.pps.Egress.serializeBinaryToWriter
    );
  }
  f = message.getUpdate();
  if (f) {
    writer.writeBool(
      5,
      f
    );
  }
  f = message.getOutputBranch();
  if (f.length > 0) {
    writer.writeString(
      10,
      f
    );
  }
  f = message.getS3Out();
  if (f) {
    writer.writeBool(
      36,
      f
    );
  }
  f = message.getResourceRequests();
  if (f != null) {
    writer.writeMessage(
      12,
      f,
      proto.pps.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getResourceLimits();
  if (f != null) {
    writer.writeMessage(
      22,
      f,
      proto.pps.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getSidecarResourceLimits();
  if (f != null) {
    writer.writeMessage(
      47,
      f,
      proto.pps.ResourceSpec.serializeBinaryToWriter
    );
  }
  f = message.getInput();
  if (f != null) {
    writer.writeMessage(
      13,
      f,
      proto.pps.Input.serializeBinaryToWriter
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
      16,
      f
    );
  }
  f = message.getEnableStats();
  if (f) {
    writer.writeBool(
      17,
      f
    );
  }
  f = message.getReprocess();
  if (f) {
    writer.writeBool(
      18,
      f
    );
  }
  f = message.getMaxQueueSize();
  if (f !== 0) {
    writer.writeInt64(
      20,
      f
    );
  }
  f = message.getService();
  if (f != null) {
    writer.writeMessage(
      21,
      f,
      proto.pps.Service.serializeBinaryToWriter
    );
  }
  f = message.getSpout();
  if (f != null) {
    writer.writeMessage(
      33,
      f,
      proto.pps.Spout.serializeBinaryToWriter
    );
  }
  f = message.getChunkSpec();
  if (f != null) {
    writer.writeMessage(
      23,
      f,
      proto.pps.ChunkSpec.serializeBinaryToWriter
    );
  }
  f = message.getDatumTimeout();
  if (f != null) {
    writer.writeMessage(
      24,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getJobTimeout();
  if (f != null) {
    writer.writeMessage(
      25,
      f,
      google_protobuf_duration_pb.Duration.serializeBinaryToWriter
    );
  }
  f = message.getSalt();
  if (f.length > 0) {
    writer.writeString(
      26,
      f
    );
  }
  f = message.getStandby();
  if (f) {
    writer.writeBool(
      27,
      f
    );
  }
  f = message.getDatumTries();
  if (f !== 0) {
    writer.writeInt64(
      28,
      f
    );
  }
  f = message.getSchedulingSpec();
  if (f != null) {
    writer.writeMessage(
      29,
      f,
      proto.pps.SchedulingSpec.serializeBinaryToWriter
    );
  }
  f = message.getPodSpec();
  if (f.length > 0) {
    writer.writeString(
      30,
      f
    );
  }
  f = message.getPodPatch();
  if (f.length > 0) {
    writer.writeString(
      32,
      f
    );
  }
  f = message.getSpecCommit();
  if (f != null) {
    writer.writeMessage(
      34,
      f,
      pfs_pfs_pb.Commit.serializeBinaryToWriter
    );
  }
  f = message.getMetadata();
  if (f != null) {
    writer.writeMessage(
      46,
      f,
      proto.pps.Metadata.serializeBinaryToWriter
    );
  }
  f = message.getNoSkip();
  if (f) {
    writer.writeBool(
      48,
      f
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.CreatePipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 1));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional TFJob tf_job = 35;
 * @return {?proto.pps.TFJob}
 */
proto.pps.CreatePipelineRequest.prototype.getTfJob = function() {
  return /** @type{?proto.pps.TFJob} */ (
    jspb.Message.getWrapperField(this, proto.pps.TFJob, 35));
};


/**
 * @param {?proto.pps.TFJob|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setTfJob = function(value) {
  return jspb.Message.setWrapperField(this, 35, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearTfJob = function() {
  return this.setTfJob(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasTfJob = function() {
  return jspb.Message.getField(this, 35) != null;
};


/**
 * optional Transform transform = 2;
 * @return {?proto.pps.Transform}
 */
proto.pps.CreatePipelineRequest.prototype.getTransform = function() {
  return /** @type{?proto.pps.Transform} */ (
    jspb.Message.getWrapperField(this, proto.pps.Transform, 2));
};


/**
 * @param {?proto.pps.Transform|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setTransform = function(value) {
  return jspb.Message.setWrapperField(this, 2, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearTransform = function() {
  return this.setTransform(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasTransform = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional ParallelismSpec parallelism_spec = 7;
 * @return {?proto.pps.ParallelismSpec}
 */
proto.pps.CreatePipelineRequest.prototype.getParallelismSpec = function() {
  return /** @type{?proto.pps.ParallelismSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ParallelismSpec, 7));
};


/**
 * @param {?proto.pps.ParallelismSpec|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setParallelismSpec = function(value) {
  return jspb.Message.setWrapperField(this, 7, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearParallelismSpec = function() {
  return this.setParallelismSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasParallelismSpec = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * optional Egress egress = 9;
 * @return {?proto.pps.Egress}
 */
proto.pps.CreatePipelineRequest.prototype.getEgress = function() {
  return /** @type{?proto.pps.Egress} */ (
    jspb.Message.getWrapperField(this, proto.pps.Egress, 9));
};


/**
 * @param {?proto.pps.Egress|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setEgress = function(value) {
  return jspb.Message.setWrapperField(this, 9, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearEgress = function() {
  return this.setEgress(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasEgress = function() {
  return jspb.Message.getField(this, 9) != null;
};


/**
 * optional bool update = 5;
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.getUpdate = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 5, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setUpdate = function(value) {
  return jspb.Message.setProto3BooleanField(this, 5, value);
};


/**
 * optional string output_branch = 10;
 * @return {string}
 */
proto.pps.CreatePipelineRequest.prototype.getOutputBranch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 10, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setOutputBranch = function(value) {
  return jspb.Message.setProto3StringField(this, 10, value);
};


/**
 * optional bool s3_out = 36;
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.getS3Out = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 36, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setS3Out = function(value) {
  return jspb.Message.setProto3BooleanField(this, 36, value);
};


/**
 * optional ResourceSpec resource_requests = 12;
 * @return {?proto.pps.ResourceSpec}
 */
proto.pps.CreatePipelineRequest.prototype.getResourceRequests = function() {
  return /** @type{?proto.pps.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ResourceSpec, 12));
};


/**
 * @param {?proto.pps.ResourceSpec|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setResourceRequests = function(value) {
  return jspb.Message.setWrapperField(this, 12, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearResourceRequests = function() {
  return this.setResourceRequests(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasResourceRequests = function() {
  return jspb.Message.getField(this, 12) != null;
};


/**
 * optional ResourceSpec resource_limits = 22;
 * @return {?proto.pps.ResourceSpec}
 */
proto.pps.CreatePipelineRequest.prototype.getResourceLimits = function() {
  return /** @type{?proto.pps.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ResourceSpec, 22));
};


/**
 * @param {?proto.pps.ResourceSpec|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 22, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearResourceLimits = function() {
  return this.setResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasResourceLimits = function() {
  return jspb.Message.getField(this, 22) != null;
};


/**
 * optional ResourceSpec sidecar_resource_limits = 47;
 * @return {?proto.pps.ResourceSpec}
 */
proto.pps.CreatePipelineRequest.prototype.getSidecarResourceLimits = function() {
  return /** @type{?proto.pps.ResourceSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ResourceSpec, 47));
};


/**
 * @param {?proto.pps.ResourceSpec|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setSidecarResourceLimits = function(value) {
  return jspb.Message.setWrapperField(this, 47, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearSidecarResourceLimits = function() {
  return this.setSidecarResourceLimits(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasSidecarResourceLimits = function() {
  return jspb.Message.getField(this, 47) != null;
};


/**
 * optional Input input = 13;
 * @return {?proto.pps.Input}
 */
proto.pps.CreatePipelineRequest.prototype.getInput = function() {
  return /** @type{?proto.pps.Input} */ (
    jspb.Message.getWrapperField(this, proto.pps.Input, 13));
};


/**
 * @param {?proto.pps.Input|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setInput = function(value) {
  return jspb.Message.setWrapperField(this, 13, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearInput = function() {
  return this.setInput(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasInput = function() {
  return jspb.Message.getField(this, 13) != null;
};


/**
 * optional string description = 14;
 * @return {string}
 */
proto.pps.CreatePipelineRequest.prototype.getDescription = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 14, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setDescription = function(value) {
  return jspb.Message.setProto3StringField(this, 14, value);
};


/**
 * optional string cache_size = 16;
 * @return {string}
 */
proto.pps.CreatePipelineRequest.prototype.getCacheSize = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 16, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setCacheSize = function(value) {
  return jspb.Message.setProto3StringField(this, 16, value);
};


/**
 * optional bool enable_stats = 17;
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.getEnableStats = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 17, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setEnableStats = function(value) {
  return jspb.Message.setProto3BooleanField(this, 17, value);
};


/**
 * optional bool reprocess = 18;
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.getReprocess = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 18, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setReprocess = function(value) {
  return jspb.Message.setProto3BooleanField(this, 18, value);
};


/**
 * optional int64 max_queue_size = 20;
 * @return {number}
 */
proto.pps.CreatePipelineRequest.prototype.getMaxQueueSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 20, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setMaxQueueSize = function(value) {
  return jspb.Message.setProto3IntField(this, 20, value);
};


/**
 * optional Service service = 21;
 * @return {?proto.pps.Service}
 */
proto.pps.CreatePipelineRequest.prototype.getService = function() {
  return /** @type{?proto.pps.Service} */ (
    jspb.Message.getWrapperField(this, proto.pps.Service, 21));
};


/**
 * @param {?proto.pps.Service|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setService = function(value) {
  return jspb.Message.setWrapperField(this, 21, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearService = function() {
  return this.setService(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasService = function() {
  return jspb.Message.getField(this, 21) != null;
};


/**
 * optional Spout spout = 33;
 * @return {?proto.pps.Spout}
 */
proto.pps.CreatePipelineRequest.prototype.getSpout = function() {
  return /** @type{?proto.pps.Spout} */ (
    jspb.Message.getWrapperField(this, proto.pps.Spout, 33));
};


/**
 * @param {?proto.pps.Spout|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setSpout = function(value) {
  return jspb.Message.setWrapperField(this, 33, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearSpout = function() {
  return this.setSpout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasSpout = function() {
  return jspb.Message.getField(this, 33) != null;
};


/**
 * optional ChunkSpec chunk_spec = 23;
 * @return {?proto.pps.ChunkSpec}
 */
proto.pps.CreatePipelineRequest.prototype.getChunkSpec = function() {
  return /** @type{?proto.pps.ChunkSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.ChunkSpec, 23));
};


/**
 * @param {?proto.pps.ChunkSpec|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setChunkSpec = function(value) {
  return jspb.Message.setWrapperField(this, 23, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearChunkSpec = function() {
  return this.setChunkSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasChunkSpec = function() {
  return jspb.Message.getField(this, 23) != null;
};


/**
 * optional google.protobuf.Duration datum_timeout = 24;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps.CreatePipelineRequest.prototype.getDatumTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 24));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setDatumTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 24, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearDatumTimeout = function() {
  return this.setDatumTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasDatumTimeout = function() {
  return jspb.Message.getField(this, 24) != null;
};


/**
 * optional google.protobuf.Duration job_timeout = 25;
 * @return {?proto.google.protobuf.Duration}
 */
proto.pps.CreatePipelineRequest.prototype.getJobTimeout = function() {
  return /** @type{?proto.google.protobuf.Duration} */ (
    jspb.Message.getWrapperField(this, google_protobuf_duration_pb.Duration, 25));
};


/**
 * @param {?proto.google.protobuf.Duration|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setJobTimeout = function(value) {
  return jspb.Message.setWrapperField(this, 25, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearJobTimeout = function() {
  return this.setJobTimeout(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasJobTimeout = function() {
  return jspb.Message.getField(this, 25) != null;
};


/**
 * optional string salt = 26;
 * @return {string}
 */
proto.pps.CreatePipelineRequest.prototype.getSalt = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 26, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setSalt = function(value) {
  return jspb.Message.setProto3StringField(this, 26, value);
};


/**
 * optional bool standby = 27;
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.getStandby = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 27, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setStandby = function(value) {
  return jspb.Message.setProto3BooleanField(this, 27, value);
};


/**
 * optional int64 datum_tries = 28;
 * @return {number}
 */
proto.pps.CreatePipelineRequest.prototype.getDatumTries = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 28, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setDatumTries = function(value) {
  return jspb.Message.setProto3IntField(this, 28, value);
};


/**
 * optional SchedulingSpec scheduling_spec = 29;
 * @return {?proto.pps.SchedulingSpec}
 */
proto.pps.CreatePipelineRequest.prototype.getSchedulingSpec = function() {
  return /** @type{?proto.pps.SchedulingSpec} */ (
    jspb.Message.getWrapperField(this, proto.pps.SchedulingSpec, 29));
};


/**
 * @param {?proto.pps.SchedulingSpec|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setSchedulingSpec = function(value) {
  return jspb.Message.setWrapperField(this, 29, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearSchedulingSpec = function() {
  return this.setSchedulingSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasSchedulingSpec = function() {
  return jspb.Message.getField(this, 29) != null;
};


/**
 * optional string pod_spec = 30;
 * @return {string}
 */
proto.pps.CreatePipelineRequest.prototype.getPodSpec = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 30, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setPodSpec = function(value) {
  return jspb.Message.setProto3StringField(this, 30, value);
};


/**
 * optional string pod_patch = 32;
 * @return {string}
 */
proto.pps.CreatePipelineRequest.prototype.getPodPatch = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 32, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setPodPatch = function(value) {
  return jspb.Message.setProto3StringField(this, 32, value);
};


/**
 * optional pfs.Commit spec_commit = 34;
 * @return {?proto.pfs.Commit}
 */
proto.pps.CreatePipelineRequest.prototype.getSpecCommit = function() {
  return /** @type{?proto.pfs.Commit} */ (
    jspb.Message.getWrapperField(this, pfs_pfs_pb.Commit, 34));
};


/**
 * @param {?proto.pfs.Commit|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setSpecCommit = function(value) {
  return jspb.Message.setWrapperField(this, 34, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearSpecCommit = function() {
  return this.setSpecCommit(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasSpecCommit = function() {
  return jspb.Message.getField(this, 34) != null;
};


/**
 * optional Metadata metadata = 46;
 * @return {?proto.pps.Metadata}
 */
proto.pps.CreatePipelineRequest.prototype.getMetadata = function() {
  return /** @type{?proto.pps.Metadata} */ (
    jspb.Message.getWrapperField(this, proto.pps.Metadata, 46));
};


/**
 * @param {?proto.pps.Metadata|undefined} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
*/
proto.pps.CreatePipelineRequest.prototype.setMetadata = function(value) {
  return jspb.Message.setWrapperField(this, 46, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.clearMetadata = function() {
  return this.setMetadata(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.hasMetadata = function() {
  return jspb.Message.getField(this, 46) != null;
};


/**
 * optional bool no_skip = 48;
 * @return {boolean}
 */
proto.pps.CreatePipelineRequest.prototype.getNoSkip = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 48, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.CreatePipelineRequest} returns this
 */
proto.pps.CreatePipelineRequest.prototype.setNoSkip = function(value) {
  return jspb.Message.setProto3BooleanField(this, 48, value);
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
proto.pps.InspectPipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.InspectPipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.InspectPipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.InspectPipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f)
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
 * @return {!proto.pps.InspectPipelineRequest}
 */
proto.pps.InspectPipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.InspectPipelineRequest;
  return proto.pps.InspectPipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.InspectPipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.InspectPipelineRequest}
 */
proto.pps.InspectPipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
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
proto.pps.InspectPipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.InspectPipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.InspectPipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.InspectPipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.InspectPipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 1));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.InspectPipelineRequest} returns this
*/
proto.pps.InspectPipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.InspectPipelineRequest} returns this
 */
proto.pps.InspectPipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.InspectPipelineRequest.prototype.hasPipeline = function() {
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
proto.pps.ListPipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.ListPipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.ListPipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ListPipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f),
    history: jspb.Message.getFieldWithDefault(msg, 2, 0),
    allowIncomplete: jspb.Message.getBooleanFieldWithDefault(msg, 3, false),
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
 * @return {!proto.pps.ListPipelineRequest}
 */
proto.pps.ListPipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.ListPipelineRequest;
  return proto.pps.ListPipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.ListPipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.ListPipelineRequest}
 */
proto.pps.ListPipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setHistory(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAllowIncomplete(value);
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
proto.pps.ListPipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.ListPipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.ListPipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ListPipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getHistory();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
  f = message.getAllowIncomplete();
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
 * @return {?proto.pps.Pipeline}
 */
proto.pps.ListPipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 1));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.ListPipelineRequest} returns this
*/
proto.pps.ListPipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.ListPipelineRequest} returns this
 */
proto.pps.ListPipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.ListPipelineRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional int64 history = 2;
 * @return {number}
 */
proto.pps.ListPipelineRequest.prototype.getHistory = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.ListPipelineRequest} returns this
 */
proto.pps.ListPipelineRequest.prototype.setHistory = function(value) {
  return jspb.Message.setProto3IntField(this, 2, value);
};


/**
 * optional bool allow_incomplete = 3;
 * @return {boolean}
 */
proto.pps.ListPipelineRequest.prototype.getAllowIncomplete = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 3, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.ListPipelineRequest} returns this
 */
proto.pps.ListPipelineRequest.prototype.setAllowIncomplete = function(value) {
  return jspb.Message.setProto3BooleanField(this, 3, value);
};


/**
 * optional string jqFilter = 4;
 * @return {string}
 */
proto.pps.ListPipelineRequest.prototype.getJqfilter = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.ListPipelineRequest} returns this
 */
proto.pps.ListPipelineRequest.prototype.setJqfilter = function(value) {
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
proto.pps.DeletePipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.DeletePipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.DeletePipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.DeletePipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f),
    all: jspb.Message.getBooleanFieldWithDefault(msg, 4, false),
    force: jspb.Message.getBooleanFieldWithDefault(msg, 5, false),
    keepRepo: jspb.Message.getBooleanFieldWithDefault(msg, 6, false)
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
 * @return {!proto.pps.DeletePipelineRequest}
 */
proto.pps.DeletePipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.DeletePipelineRequest;
  return proto.pps.DeletePipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.DeletePipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.DeletePipelineRequest}
 */
proto.pps.DeletePipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setAll(value);
      break;
    case 5:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setForce(value);
      break;
    case 6:
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
proto.pps.DeletePipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.DeletePipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.DeletePipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.DeletePipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getAll();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
  f = message.getForce();
  if (f) {
    writer.writeBool(
      5,
      f
    );
  }
  f = message.getKeepRepo();
  if (f) {
    writer.writeBool(
      6,
      f
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.DeletePipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 1));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.DeletePipelineRequest} returns this
*/
proto.pps.DeletePipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.DeletePipelineRequest} returns this
 */
proto.pps.DeletePipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.DeletePipelineRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional bool all = 4;
 * @return {boolean}
 */
proto.pps.DeletePipelineRequest.prototype.getAll = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 4, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.DeletePipelineRequest} returns this
 */
proto.pps.DeletePipelineRequest.prototype.setAll = function(value) {
  return jspb.Message.setProto3BooleanField(this, 4, value);
};


/**
 * optional bool force = 5;
 * @return {boolean}
 */
proto.pps.DeletePipelineRequest.prototype.getForce = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 5, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.DeletePipelineRequest} returns this
 */
proto.pps.DeletePipelineRequest.prototype.setForce = function(value) {
  return jspb.Message.setProto3BooleanField(this, 5, value);
};


/**
 * optional bool keep_repo = 6;
 * @return {boolean}
 */
proto.pps.DeletePipelineRequest.prototype.getKeepRepo = function() {
  return /** @type {boolean} */ (jspb.Message.getBooleanFieldWithDefault(this, 6, false));
};


/**
 * @param {boolean} value
 * @return {!proto.pps.DeletePipelineRequest} returns this
 */
proto.pps.DeletePipelineRequest.prototype.setKeepRepo = function(value) {
  return jspb.Message.setProto3BooleanField(this, 6, value);
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
proto.pps.StartPipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.StartPipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.StartPipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.StartPipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f)
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
 * @return {!proto.pps.StartPipelineRequest}
 */
proto.pps.StartPipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.StartPipelineRequest;
  return proto.pps.StartPipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.StartPipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.StartPipelineRequest}
 */
proto.pps.StartPipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
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
proto.pps.StartPipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.StartPipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.StartPipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.StartPipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.StartPipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 1));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.StartPipelineRequest} returns this
*/
proto.pps.StartPipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.StartPipelineRequest} returns this
 */
proto.pps.StartPipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.StartPipelineRequest.prototype.hasPipeline = function() {
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
proto.pps.StopPipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.StopPipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.StopPipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.StopPipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f)
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
 * @return {!proto.pps.StopPipelineRequest}
 */
proto.pps.StopPipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.StopPipelineRequest;
  return proto.pps.StopPipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.StopPipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.StopPipelineRequest}
 */
proto.pps.StopPipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
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
proto.pps.StopPipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.StopPipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.StopPipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.StopPipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.StopPipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 1));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.StopPipelineRequest} returns this
*/
proto.pps.StopPipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.StopPipelineRequest} returns this
 */
proto.pps.StopPipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.StopPipelineRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.RunPipelineRequest.repeatedFields_ = [2];



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
proto.pps.RunPipelineRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.RunPipelineRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.RunPipelineRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.RunPipelineRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f),
    provenanceList: jspb.Message.toObjectList(msg.getProvenanceList(),
    pfs_pfs_pb.CommitProvenance.toObject, includeInstance),
    jobId: jspb.Message.getFieldWithDefault(msg, 4, "")
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
 * @return {!proto.pps.RunPipelineRequest}
 */
proto.pps.RunPipelineRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.RunPipelineRequest;
  return proto.pps.RunPipelineRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.RunPipelineRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.RunPipelineRequest}
 */
proto.pps.RunPipelineRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
      msg.setPipeline(value);
      break;
    case 2:
      var value = new pfs_pfs_pb.CommitProvenance;
      reader.readMessage(value,pfs_pfs_pb.CommitProvenance.deserializeBinaryFromReader);
      msg.addProvenance(value);
      break;
    case 4:
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
proto.pps.RunPipelineRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.RunPipelineRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.RunPipelineRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.RunPipelineRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
  f = message.getProvenanceList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      pfs_pfs_pb.CommitProvenance.serializeBinaryToWriter
    );
  }
  f = message.getJobId();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.RunPipelineRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 1));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.RunPipelineRequest} returns this
*/
proto.pps.RunPipelineRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.RunPipelineRequest} returns this
 */
proto.pps.RunPipelineRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.RunPipelineRequest.prototype.hasPipeline = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * repeated pfs.CommitProvenance provenance = 2;
 * @return {!Array<!proto.pfs.CommitProvenance>}
 */
proto.pps.RunPipelineRequest.prototype.getProvenanceList = function() {
  return /** @type{!Array<!proto.pfs.CommitProvenance>} */ (
    jspb.Message.getRepeatedWrapperField(this, pfs_pfs_pb.CommitProvenance, 2));
};


/**
 * @param {!Array<!proto.pfs.CommitProvenance>} value
 * @return {!proto.pps.RunPipelineRequest} returns this
*/
proto.pps.RunPipelineRequest.prototype.setProvenanceList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.pfs.CommitProvenance=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pfs.CommitProvenance}
 */
proto.pps.RunPipelineRequest.prototype.addProvenance = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.pfs.CommitProvenance, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.RunPipelineRequest} returns this
 */
proto.pps.RunPipelineRequest.prototype.clearProvenanceList = function() {
  return this.setProvenanceList([]);
};


/**
 * optional string job_id = 4;
 * @return {string}
 */
proto.pps.RunPipelineRequest.prototype.getJobId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.RunPipelineRequest} returns this
 */
proto.pps.RunPipelineRequest.prototype.setJobId = function(value) {
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
proto.pps.RunCronRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.RunCronRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.RunCronRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.RunCronRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    pipeline: (f = msg.getPipeline()) && proto.pps.Pipeline.toObject(includeInstance, f)
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
 * @return {!proto.pps.RunCronRequest}
 */
proto.pps.RunCronRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.RunCronRequest;
  return proto.pps.RunCronRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.RunCronRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.RunCronRequest}
 */
proto.pps.RunCronRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Pipeline;
      reader.readMessage(value,proto.pps.Pipeline.deserializeBinaryFromReader);
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
proto.pps.RunCronRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.RunCronRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.RunCronRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.RunCronRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPipeline();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Pipeline.serializeBinaryToWriter
    );
  }
};


/**
 * optional Pipeline pipeline = 1;
 * @return {?proto.pps.Pipeline}
 */
proto.pps.RunCronRequest.prototype.getPipeline = function() {
  return /** @type{?proto.pps.Pipeline} */ (
    jspb.Message.getWrapperField(this, proto.pps.Pipeline, 1));
};


/**
 * @param {?proto.pps.Pipeline|undefined} value
 * @return {!proto.pps.RunCronRequest} returns this
*/
proto.pps.RunCronRequest.prototype.setPipeline = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.RunCronRequest} returns this
 */
proto.pps.RunCronRequest.prototype.clearPipeline = function() {
  return this.setPipeline(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.RunCronRequest.prototype.hasPipeline = function() {
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
proto.pps.CreateSecretRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.CreateSecretRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.CreateSecretRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.CreateSecretRequest.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.CreateSecretRequest}
 */
proto.pps.CreateSecretRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.CreateSecretRequest;
  return proto.pps.CreateSecretRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.CreateSecretRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.CreateSecretRequest}
 */
proto.pps.CreateSecretRequest.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.CreateSecretRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.CreateSecretRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.CreateSecretRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.CreateSecretRequest.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.CreateSecretRequest.prototype.getFile = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * optional bytes file = 1;
 * This is a type-conversion wrapper around `getFile()`
 * @return {string}
 */
proto.pps.CreateSecretRequest.prototype.getFile_asB64 = function() {
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
proto.pps.CreateSecretRequest.prototype.getFile_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getFile()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pps.CreateSecretRequest} returns this
 */
proto.pps.CreateSecretRequest.prototype.setFile = function(value) {
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
proto.pps.DeleteSecretRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.DeleteSecretRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.DeleteSecretRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.DeleteSecretRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    secret: (f = msg.getSecret()) && proto.pps.Secret.toObject(includeInstance, f)
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
 * @return {!proto.pps.DeleteSecretRequest}
 */
proto.pps.DeleteSecretRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.DeleteSecretRequest;
  return proto.pps.DeleteSecretRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.DeleteSecretRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.DeleteSecretRequest}
 */
proto.pps.DeleteSecretRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Secret;
      reader.readMessage(value,proto.pps.Secret.deserializeBinaryFromReader);
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
proto.pps.DeleteSecretRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.DeleteSecretRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.DeleteSecretRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.DeleteSecretRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSecret();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Secret.serializeBinaryToWriter
    );
  }
};


/**
 * optional Secret secret = 1;
 * @return {?proto.pps.Secret}
 */
proto.pps.DeleteSecretRequest.prototype.getSecret = function() {
  return /** @type{?proto.pps.Secret} */ (
    jspb.Message.getWrapperField(this, proto.pps.Secret, 1));
};


/**
 * @param {?proto.pps.Secret|undefined} value
 * @return {!proto.pps.DeleteSecretRequest} returns this
*/
proto.pps.DeleteSecretRequest.prototype.setSecret = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.DeleteSecretRequest} returns this
 */
proto.pps.DeleteSecretRequest.prototype.clearSecret = function() {
  return this.setSecret(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.DeleteSecretRequest.prototype.hasSecret = function() {
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
proto.pps.InspectSecretRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.InspectSecretRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.InspectSecretRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.InspectSecretRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    secret: (f = msg.getSecret()) && proto.pps.Secret.toObject(includeInstance, f)
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
 * @return {!proto.pps.InspectSecretRequest}
 */
proto.pps.InspectSecretRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.InspectSecretRequest;
  return proto.pps.InspectSecretRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.InspectSecretRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.InspectSecretRequest}
 */
proto.pps.InspectSecretRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Secret;
      reader.readMessage(value,proto.pps.Secret.deserializeBinaryFromReader);
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
proto.pps.InspectSecretRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.InspectSecretRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.InspectSecretRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.InspectSecretRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSecret();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Secret.serializeBinaryToWriter
    );
  }
};


/**
 * optional Secret secret = 1;
 * @return {?proto.pps.Secret}
 */
proto.pps.InspectSecretRequest.prototype.getSecret = function() {
  return /** @type{?proto.pps.Secret} */ (
    jspb.Message.getWrapperField(this, proto.pps.Secret, 1));
};


/**
 * @param {?proto.pps.Secret|undefined} value
 * @return {!proto.pps.InspectSecretRequest} returns this
*/
proto.pps.InspectSecretRequest.prototype.setSecret = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.InspectSecretRequest} returns this
 */
proto.pps.InspectSecretRequest.prototype.clearSecret = function() {
  return this.setSecret(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.InspectSecretRequest.prototype.hasSecret = function() {
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
proto.pps.Secret.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.Secret.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.Secret} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Secret.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.Secret}
 */
proto.pps.Secret.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.Secret;
  return proto.pps.Secret.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.Secret} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.Secret}
 */
proto.pps.Secret.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.Secret.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.Secret.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.Secret} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.Secret.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.Secret.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.Secret} returns this
 */
proto.pps.Secret.prototype.setName = function(value) {
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
proto.pps.SecretInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.SecretInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.SecretInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.SecretInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    secret: (f = msg.getSecret()) && proto.pps.Secret.toObject(includeInstance, f),
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
 * @return {!proto.pps.SecretInfo}
 */
proto.pps.SecretInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.SecretInfo;
  return proto.pps.SecretInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.SecretInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.SecretInfo}
 */
proto.pps.SecretInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.Secret;
      reader.readMessage(value,proto.pps.Secret.deserializeBinaryFromReader);
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
proto.pps.SecretInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.SecretInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.SecretInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.SecretInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSecret();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.pps.Secret.serializeBinaryToWriter
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
 * @return {?proto.pps.Secret}
 */
proto.pps.SecretInfo.prototype.getSecret = function() {
  return /** @type{?proto.pps.Secret} */ (
    jspb.Message.getWrapperField(this, proto.pps.Secret, 1));
};


/**
 * @param {?proto.pps.Secret|undefined} value
 * @return {!proto.pps.SecretInfo} returns this
*/
proto.pps.SecretInfo.prototype.setSecret = function(value) {
  return jspb.Message.setWrapperField(this, 1, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.SecretInfo} returns this
 */
proto.pps.SecretInfo.prototype.clearSecret = function() {
  return this.setSecret(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.SecretInfo.prototype.hasSecret = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional string type = 2;
 * @return {string}
 */
proto.pps.SecretInfo.prototype.getType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pps.SecretInfo} returns this
 */
proto.pps.SecretInfo.prototype.setType = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional google.protobuf.Timestamp creation_timestamp = 3;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.pps.SecretInfo.prototype.getCreationTimestamp = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 3));
};


/**
 * @param {?proto.google.protobuf.Timestamp|undefined} value
 * @return {!proto.pps.SecretInfo} returns this
*/
proto.pps.SecretInfo.prototype.setCreationTimestamp = function(value) {
  return jspb.Message.setWrapperField(this, 3, value);
};


/**
 * Clears the message field making it undefined.
 * @return {!proto.pps.SecretInfo} returns this
 */
proto.pps.SecretInfo.prototype.clearCreationTimestamp = function() {
  return this.setCreationTimestamp(undefined);
};


/**
 * Returns whether this field is set.
 * @return {boolean}
 */
proto.pps.SecretInfo.prototype.hasCreationTimestamp = function() {
  return jspb.Message.getField(this, 3) != null;
};



/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.pps.SecretInfos.repeatedFields_ = [1];



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
proto.pps.SecretInfos.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.SecretInfos.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.SecretInfos} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.SecretInfos.toObject = function(includeInstance, msg) {
  var f, obj = {
    secretInfoList: jspb.Message.toObjectList(msg.getSecretInfoList(),
    proto.pps.SecretInfo.toObject, includeInstance)
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
 * @return {!proto.pps.SecretInfos}
 */
proto.pps.SecretInfos.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.SecretInfos;
  return proto.pps.SecretInfos.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.SecretInfos} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.SecretInfos}
 */
proto.pps.SecretInfos.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.pps.SecretInfo;
      reader.readMessage(value,proto.pps.SecretInfo.deserializeBinaryFromReader);
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
proto.pps.SecretInfos.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.SecretInfos.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.SecretInfos} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.SecretInfos.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSecretInfoList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.pps.SecretInfo.serializeBinaryToWriter
    );
  }
};


/**
 * repeated SecretInfo secret_info = 1;
 * @return {!Array<!proto.pps.SecretInfo>}
 */
proto.pps.SecretInfos.prototype.getSecretInfoList = function() {
  return /** @type{!Array<!proto.pps.SecretInfo>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.pps.SecretInfo, 1));
};


/**
 * @param {!Array<!proto.pps.SecretInfo>} value
 * @return {!proto.pps.SecretInfos} returns this
*/
proto.pps.SecretInfos.prototype.setSecretInfoList = function(value) {
  return jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.pps.SecretInfo=} opt_value
 * @param {number=} opt_index
 * @return {!proto.pps.SecretInfo}
 */
proto.pps.SecretInfos.prototype.addSecretInfo = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.pps.SecretInfo, opt_index);
};


/**
 * Clears the list making it empty but non-null.
 * @return {!proto.pps.SecretInfos} returns this
 */
proto.pps.SecretInfos.prototype.clearSecretInfoList = function() {
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
proto.pps.GarbageCollectRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.GarbageCollectRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.GarbageCollectRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.GarbageCollectRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    memoryBytes: jspb.Message.getFieldWithDefault(msg, 1, 0)
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
 * @return {!proto.pps.GarbageCollectRequest}
 */
proto.pps.GarbageCollectRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.GarbageCollectRequest;
  return proto.pps.GarbageCollectRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.GarbageCollectRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.GarbageCollectRequest}
 */
proto.pps.GarbageCollectRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setMemoryBytes(value);
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
proto.pps.GarbageCollectRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.GarbageCollectRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.GarbageCollectRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.GarbageCollectRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getMemoryBytes();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
};


/**
 * optional int64 memory_bytes = 1;
 * @return {number}
 */
proto.pps.GarbageCollectRequest.prototype.getMemoryBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pps.GarbageCollectRequest} returns this
 */
proto.pps.GarbageCollectRequest.prototype.setMemoryBytes = function(value) {
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
proto.pps.GarbageCollectResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.GarbageCollectResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.GarbageCollectResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.GarbageCollectResponse.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.GarbageCollectResponse}
 */
proto.pps.GarbageCollectResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.GarbageCollectResponse;
  return proto.pps.GarbageCollectResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.GarbageCollectResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.GarbageCollectResponse}
 */
proto.pps.GarbageCollectResponse.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.GarbageCollectResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.GarbageCollectResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.GarbageCollectResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.GarbageCollectResponse.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.ActivateAuthRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.ActivateAuthRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.ActivateAuthRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ActivateAuthRequest.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.ActivateAuthRequest}
 */
proto.pps.ActivateAuthRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.ActivateAuthRequest;
  return proto.pps.ActivateAuthRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.ActivateAuthRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.ActivateAuthRequest}
 */
proto.pps.ActivateAuthRequest.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.ActivateAuthRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.ActivateAuthRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.ActivateAuthRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ActivateAuthRequest.serializeBinaryToWriter = function(message, writer) {
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
proto.pps.ActivateAuthResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.pps.ActivateAuthResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pps.ActivateAuthResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ActivateAuthResponse.toObject = function(includeInstance, msg) {
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
 * @return {!proto.pps.ActivateAuthResponse}
 */
proto.pps.ActivateAuthResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pps.ActivateAuthResponse;
  return proto.pps.ActivateAuthResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pps.ActivateAuthResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pps.ActivateAuthResponse}
 */
proto.pps.ActivateAuthResponse.deserializeBinaryFromReader = function(msg, reader) {
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
proto.pps.ActivateAuthResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pps.ActivateAuthResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pps.ActivateAuthResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pps.ActivateAuthResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};


/**
 * @enum {number}
 */
proto.pps.JobState = {
  JOB_STARTING: 0,
  JOB_RUNNING: 1,
  JOB_FAILURE: 2,
  JOB_SUCCESS: 3,
  JOB_KILLED: 4,
  JOB_EGRESSING: 6
};

/**
 * @enum {number}
 */
proto.pps.DatumState = {
  FAILED: 0,
  SUCCESS: 1,
  SKIPPED: 2,
  STARTING: 3,
  RECOVERED: 4
};

/**
 * @enum {number}
 */
proto.pps.WorkerState = {
  POD_RUNNING: 0,
  POD_SUCCESS: 1,
  POD_FAILED: 2
};

/**
 * @enum {number}
 */
proto.pps.PipelineState = {
  PIPELINE_STARTING: 0,
  PIPELINE_RUNNING: 1,
  PIPELINE_RESTARTING: 2,
  PIPELINE_FAILURE: 3,
  PIPELINE_PAUSED: 4,
  PIPELINE_STANDBY: 5,
  PIPELINE_CRASHING: 6
};

goog.object.extend(exports, proto.pps);
