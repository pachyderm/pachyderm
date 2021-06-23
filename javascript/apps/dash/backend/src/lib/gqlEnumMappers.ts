import {FileType} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {JobState, PipelineState} from '@pachyderm/proto/pb/pps/pps_pb';
import {ProjectStatus} from '@pachyderm/proto/pb/projects/projects_pb';
import {ApolloError} from 'apollo-server-errors';

import {
  ProjectStatus as GQLProjectStatus,
  FileType as GQLFileType,
  PipelineState as GQLPipelineState,
  JobState as GQLJobState,
} from '@graphqlTypes';

/*
 * NOTE: TS String enums do not support reverse mapping,
 * otherwise, we could attempt to look up the enum value
 * based on the rpc key.
 */

export const toGQLPipelineState = (
  pipelineState: PipelineState,
): GQLPipelineState => {
  switch (pipelineState) {
    case PipelineState.PIPELINE_STATE_UNKNOWN:
      return GQLPipelineState.PIPELINE_STATE_UNKNOWN;
    case PipelineState.PIPELINE_CRASHING:
      return GQLPipelineState.PIPELINE_CRASHING;
    case PipelineState.PIPELINE_FAILURE:
      return GQLPipelineState.PIPELINE_FAILURE;
    case PipelineState.PIPELINE_PAUSED:
      return GQLPipelineState.PIPELINE_PAUSED;
    case PipelineState.PIPELINE_RESTARTING:
      return GQLPipelineState.PIPELINE_RESTARTING;
    case PipelineState.PIPELINE_RUNNING:
      return GQLPipelineState.PIPELINE_RUNNING;
    case PipelineState.PIPELINE_STANDBY:
      return GQLPipelineState.PIPELINE_STANDBY;
    case PipelineState.PIPELINE_STARTING:
      return GQLPipelineState.PIPELINE_STARTING;
    default:
      throw new ApolloError(`Unknown pipeline state ${pipelineState}`);
  }
};

export const toGQLJobState = (jobState: JobState) => {
  switch (jobState) {
    case JobState.JOB_STATE_UNKNOWN:
      return GQLJobState.JOB_STATE_UNKNOWN;
    case JobState.JOB_CREATED:
      return GQLJobState.JOB_CREATED;
    case JobState.JOB_EGRESSING:
      return GQLJobState.JOB_EGRESSING;
    case JobState.JOB_FAILURE:
      return GQLJobState.JOB_FAILURE;
    case JobState.JOB_KILLED:
      return GQLJobState.JOB_KILLED;
    case JobState.JOB_RUNNING:
      return GQLJobState.JOB_RUNNING;
    case JobState.JOB_STARTING:
      return GQLJobState.JOB_STARTING;
    case JobState.JOB_SUCCESS:
      return GQLJobState.JOB_SUCCESS;
    default:
      throw new ApolloError(`Unknown job state ${jobState}`);
  }
};

export const toGQLFileType = (fileType: FileType) => {
  switch (fileType) {
    case FileType.DIR:
      return GQLFileType.DIR;
    case FileType.FILE:
      return GQLFileType.FILE;
    case FileType.RESERVED:
      return GQLFileType.RESERVED;
    default:
      throw new ApolloError(`Unknown file type ${fileType}`);
  }
};

export const toGQLProjectStatus = (projectStatus: ProjectStatus) => {
  switch (projectStatus) {
    case ProjectStatus.HEALTHY:
      return GQLProjectStatus.HEALTHY;
    case ProjectStatus.UNHEALTHY:
      return GQLProjectStatus.UNHEALTHY;
    default:
      throw new ApolloError(`Unknown project status ${projectStatus}`);
  }
};
