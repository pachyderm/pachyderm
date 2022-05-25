import {
  FileType,
  JobState,
  PipelineState,
  ProjectStatus,
  OriginKind,
  DatumState,
  State,
} from '@pachyderm/node-pachyderm';
import {ApolloError} from 'apollo-server-errors';

import {
  ProjectStatus as GQLProjectStatus,
  FileType as GQLFileType,
  PipelineState as GQLPipelineState,
  JobState as GQLJobState,
  OriginKind as GQLOriginKind,
  DatumState as GQLDatumState,
  EnterpriseState as GQLEnterpriseState,
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
    case JobState.JOB_FINISHING:
      return GQLJobState.JOB_FINISHING;
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

export const toGQLCommitOrigin = (originKind?: OriginKind) => {
  switch (originKind) {
    case OriginKind.ALIAS:
      return GQLOriginKind.ALIAS;
    case OriginKind.AUTO:
      return GQLOriginKind.AUTO;
    case OriginKind.FSCK:
      return GQLOriginKind.FSCK;
    case OriginKind.USER:
      return GQLOriginKind.USER;
    case OriginKind.ORIGIN_KIND_UNKNOWN:
      return GQLOriginKind.ORIGIN_KIND_UNKNOWN;
    default:
      throw new ApolloError(`Unknown origin kind ${originKind}`);
  }
};

export const toProtoCommitOrigin = (originKind?: GQLOriginKind): OriginKind => {
  switch (originKind) {
    case GQLOriginKind.ALIAS:
      return OriginKind.ALIAS;
    case GQLOriginKind.AUTO:
      return OriginKind.AUTO;
    case GQLOriginKind.FSCK:
      return OriginKind.FSCK;
    case GQLOriginKind.USER:
      return OriginKind.USER;
    case GQLOriginKind.ORIGIN_KIND_UNKNOWN:
      return OriginKind.ORIGIN_KIND_UNKNOWN;
    default:
      throw new ApolloError(`Unknown origin kind ${originKind}`);
  }
};

export const toGQLDatumState = (state: DatumState) => {
  switch (state) {
    case DatumState.FAILED:
      return GQLDatumState.FAILED;
    case DatumState.RECOVERED:
      return GQLDatumState.RECOVERED;
    case DatumState.SKIPPED:
      return GQLDatumState.SKIPPED;
    case DatumState.STARTING:
      return GQLDatumState.STARTING;
    case DatumState.SUCCESS:
      return GQLDatumState.SUCCESS;
    case DatumState.UNKNOWN:
      return GQLDatumState.UNKOWN;
    default:
      throw new ApolloError(`Uknown datum state ${state}`);
  }
};

export const toGQLEnterpriseState = (state: State) => {
  switch (state) {
    case State.ACTIVE:
      return GQLEnterpriseState.ACTIVE;
    case State.EXPIRED:
      return GQLEnterpriseState.EXPIRED;
    case State.HEARTBEAT_FAILED:
      return GQLEnterpriseState.HEARTBEAT_FAILED;
    case State.NONE:
      return GQLEnterpriseState.NONE;
    default:
      throw new ApolloError(`Unknown enterprise state ${state}`);
  }
};
