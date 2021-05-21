import {FileType} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {PipelineJobState, PipelineState} from '@pachyderm/proto/pb/pps/pps_pb';
import {ProjectStatus} from '@pachyderm/proto/pb/projects/projects_pb';
import {ApolloError} from 'apollo-server-errors';

import {
  ProjectStatus as GQLProjectStatus,
  FileType as GQLFileType,
  PipelineState as GQLPipelineState,
  PipelineJobState as GQLPipelineJobState,
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

export const toGQLPipelineJobState = (jobState: PipelineJobState) => {
  switch (jobState) {
    case PipelineJobState.JOB_EGRESSING:
      return GQLPipelineJobState.JOB_EGRESSING;
    case PipelineJobState.JOB_FAILURE:
      return GQLPipelineJobState.JOB_FAILURE;
    case PipelineJobState.JOB_KILLED:
      return GQLPipelineJobState.JOB_KILLED;
    case PipelineJobState.JOB_RUNNING:
      return GQLPipelineJobState.JOB_RUNNING;
    case PipelineJobState.JOB_STARTING:
      return GQLPipelineJobState.JOB_STARTING;
    case PipelineJobState.JOB_SUCCESS:
      return GQLPipelineJobState.JOB_SUCCESS;
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
