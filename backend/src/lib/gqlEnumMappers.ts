import {ApolloError} from 'apollo-server-errors';

import {
  FileType,
  JobState,
  PipelineState,
  OriginKind,
  DatumState,
  State,
  ResourceType,
  Permission,
} from '@dash-backend/proto';
import {
  FileType as GQLFileType,
  PipelineState as GQLPipelineState,
  JobState as GQLJobState,
  OriginKind as GQLOriginKind,
  DatumState as GQLDatumState,
  DatumFilter as GQLDatumFilter,
  EnterpriseState as GQLEnterpriseState,
  ResourceType as GQLResourceType,
  Permission as GQLPermission,
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
    case JobState.JOB_UNRUNNABLE:
      return GQLJobState.JOB_UNRUNNABLE;
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

export const toGQLCommitOrigin = (originKind?: OriginKind) => {
  switch (originKind) {
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
      return GQLDatumState.UNKNOWN;
    default:
      throw new ApolloError(`Uknown datum state ${state}`);
  }
};

export const toProtoDatumState = (state: GQLDatumFilter) => {
  switch (state) {
    case GQLDatumFilter.FAILED:
      return DatumState.FAILED;
    case GQLDatumFilter.RECOVERED:
      return DatumState.RECOVERED;
    case GQLDatumFilter.SKIPPED:
      return DatumState.SKIPPED;
    case GQLDatumFilter.SUCCESS:
      return DatumState.SUCCESS;
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

export const toProtoResourceType = (resourceType: GQLResourceType) => {
  switch (resourceType) {
    case GQLResourceType.RESOURCE_TYPE_UNKNOWN:
      return ResourceType.RESOURCE_TYPE_UNKNOWN;
    case GQLResourceType.CLUSTER:
      return ResourceType.CLUSTER;
    case GQLResourceType.REPO:
      return ResourceType.REPO;
    case GQLResourceType.SPEC_REPO:
      return ResourceType.SPEC_REPO;
    case GQLResourceType.PROJECT:
      return ResourceType.PROJECT;
    default:
      throw new ApolloError(`Unknown resource type ${resourceType}`);
  }
};

export const toProtoPermissionType = (permission: GQLPermission) => {
  switch (permission) {
    case GQLPermission.PERMISSION_UNKNOWN:
      return Permission.PERMISSION_UNKNOWN;
    case GQLPermission.CLUSTER_MODIFY_BINDINGS:
      return Permission.CLUSTER_MODIFY_BINDINGS;
    case GQLPermission.CLUSTER_GET_BINDINGS:
      return Permission.CLUSTER_GET_BINDINGS;
    case GQLPermission.CLUSTER_GET_PACHD_LOGS:
      return Permission.CLUSTER_GET_PACHD_LOGS;
    case GQLPermission.CLUSTER_AUTH_ACTIVATE:
      return Permission.CLUSTER_AUTH_ACTIVATE;
    case GQLPermission.CLUSTER_AUTH_DEACTIVATE:
      return Permission.CLUSTER_AUTH_DEACTIVATE;
    case GQLPermission.CLUSTER_AUTH_GET_CONFIG:
      return Permission.CLUSTER_AUTH_GET_CONFIG;
    case GQLPermission.CLUSTER_AUTH_SET_CONFIG:
      return Permission.CLUSTER_AUTH_SET_CONFIG;
    case GQLPermission.CLUSTER_AUTH_GET_ROBOT_TOKEN:
      return Permission.CLUSTER_AUTH_GET_ROBOT_TOKEN;
    case GQLPermission.CLUSTER_AUTH_MODIFY_GROUP_MEMBERS:
      return Permission.CLUSTER_AUTH_MODIFY_GROUP_MEMBERS;
    case GQLPermission.CLUSTER_AUTH_GET_GROUPS:
      return Permission.CLUSTER_AUTH_GET_GROUPS;
    case GQLPermission.CLUSTER_AUTH_GET_GROUP_USERS:
      return Permission.CLUSTER_AUTH_GET_GROUP_USERS;
    case GQLPermission.CLUSTER_AUTH_EXTRACT_TOKENS:
      return Permission.CLUSTER_AUTH_EXTRACT_TOKENS;
    case GQLPermission.CLUSTER_AUTH_RESTORE_TOKEN:
      return Permission.CLUSTER_AUTH_RESTORE_TOKEN;
    case GQLPermission.CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL:
      return Permission.CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL;
    case GQLPermission.CLUSTER_AUTH_DELETE_EXPIRED_TOKENS:
      return Permission.CLUSTER_AUTH_DELETE_EXPIRED_TOKENS;
    case GQLPermission.CLUSTER_AUTH_REVOKE_USER_TOKENS:
      return Permission.CLUSTER_AUTH_REVOKE_USER_TOKENS;
    case GQLPermission.CLUSTER_AUTH_ROTATE_ROOT_TOKEN:
      return Permission.CLUSTER_AUTH_ROTATE_ROOT_TOKEN;
    case GQLPermission.CLUSTER_ENTERPRISE_ACTIVATE:
      return Permission.CLUSTER_ENTERPRISE_ACTIVATE;
    case GQLPermission.CLUSTER_ENTERPRISE_HEARTBEAT:
      return Permission.CLUSTER_ENTERPRISE_HEARTBEAT;
    case GQLPermission.CLUSTER_ENTERPRISE_GET_CODE:
      return Permission.CLUSTER_ENTERPRISE_GET_CODE;
    case GQLPermission.CLUSTER_ENTERPRISE_DEACTIVATE:
      return Permission.CLUSTER_ENTERPRISE_DEACTIVATE;
    case GQLPermission.CLUSTER_ENTERPRISE_PAUSE:
      return Permission.CLUSTER_ENTERPRISE_PAUSE;
    case GQLPermission.CLUSTER_IDENTITY_SET_CONFIG:
      return Permission.CLUSTER_IDENTITY_SET_CONFIG;
    case GQLPermission.CLUSTER_IDENTITY_GET_CONFIG:
      return Permission.CLUSTER_IDENTITY_GET_CONFIG;
    case GQLPermission.CLUSTER_IDENTITY_CREATE_IDP:
      return Permission.CLUSTER_IDENTITY_CREATE_IDP;
    case GQLPermission.CLUSTER_IDENTITY_UPDATE_IDP:
      return Permission.CLUSTER_IDENTITY_UPDATE_IDP;
    case GQLPermission.CLUSTER_IDENTITY_LIST_IDPS:
      return Permission.CLUSTER_IDENTITY_LIST_IDPS;
    case GQLPermission.CLUSTER_IDENTITY_GET_IDP:
      return Permission.CLUSTER_IDENTITY_GET_IDP;
    case GQLPermission.CLUSTER_IDENTITY_DELETE_IDP:
      return Permission.CLUSTER_IDENTITY_DELETE_IDP;
    case GQLPermission.CLUSTER_IDENTITY_CREATE_OIDC_CLIENT:
      return Permission.CLUSTER_IDENTITY_CREATE_OIDC_CLIENT;
    case GQLPermission.CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT:
      return Permission.CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT;
    case GQLPermission.CLUSTER_IDENTITY_LIST_OIDC_CLIENTS:
      return Permission.CLUSTER_IDENTITY_LIST_OIDC_CLIENTS;
    case GQLPermission.CLUSTER_IDENTITY_GET_OIDC_CLIENT:
      return Permission.CLUSTER_IDENTITY_GET_OIDC_CLIENT;
    case GQLPermission.CLUSTER_IDENTITY_DELETE_OIDC_CLIENT:
      return Permission.CLUSTER_IDENTITY_DELETE_OIDC_CLIENT;
    case GQLPermission.CLUSTER_DEBUG_DUMP:
      return Permission.CLUSTER_DEBUG_DUMP;
    case GQLPermission.CLUSTER_LICENSE_ACTIVATE:
      return Permission.CLUSTER_LICENSE_ACTIVATE;
    case GQLPermission.CLUSTER_LICENSE_GET_CODE:
      return Permission.CLUSTER_LICENSE_GET_CODE;
    case GQLPermission.CLUSTER_LICENSE_ADD_CLUSTER:
      return Permission.CLUSTER_LICENSE_ADD_CLUSTER;
    case GQLPermission.CLUSTER_LICENSE_UPDATE_CLUSTER:
      return Permission.CLUSTER_LICENSE_UPDATE_CLUSTER;
    case GQLPermission.CLUSTER_LICENSE_DELETE_CLUSTER:
      return Permission.CLUSTER_LICENSE_DELETE_CLUSTER;
    case GQLPermission.CLUSTER_LICENSE_LIST_CLUSTERS:
      return Permission.CLUSTER_LICENSE_LIST_CLUSTERS;
    case GQLPermission.CLUSTER_CREATE_SECRET:
      return Permission.CLUSTER_CREATE_SECRET;
    case GQLPermission.CLUSTER_LIST_SECRETS:
      return Permission.CLUSTER_LIST_SECRETS;
    case GQLPermission.SECRET_DELETE:
      return Permission.SECRET_DELETE;
    case GQLPermission.SECRET_INSPECT:
      return Permission.SECRET_INSPECT;
    case GQLPermission.CLUSTER_DELETE_ALL:
      return Permission.CLUSTER_DELETE_ALL;
    case GQLPermission.REPO_READ:
      return Permission.REPO_READ;
    case GQLPermission.REPO_WRITE:
      return Permission.REPO_WRITE;
    case GQLPermission.REPO_MODIFY_BINDINGS:
      return Permission.REPO_MODIFY_BINDINGS;
    case GQLPermission.REPO_DELETE:
      return Permission.REPO_DELETE;
    case GQLPermission.REPO_INSPECT_COMMIT:
      return Permission.REPO_INSPECT_COMMIT;
    case GQLPermission.REPO_LIST_COMMIT:
      return Permission.REPO_LIST_COMMIT;
    case GQLPermission.REPO_DELETE_COMMIT:
      return Permission.REPO_DELETE_COMMIT;
    case GQLPermission.REPO_CREATE_BRANCH:
      return Permission.REPO_CREATE_BRANCH;
    case GQLPermission.REPO_LIST_BRANCH:
      return Permission.REPO_LIST_BRANCH;
    case GQLPermission.REPO_DELETE_BRANCH:
      return Permission.REPO_DELETE_BRANCH;
    case GQLPermission.REPO_INSPECT_FILE:
      return Permission.REPO_INSPECT_FILE;
    case GQLPermission.REPO_LIST_FILE:
      return Permission.REPO_LIST_FILE;
    case GQLPermission.REPO_ADD_PIPELINE_READER:
      return Permission.REPO_ADD_PIPELINE_READER;
    case GQLPermission.REPO_REMOVE_PIPELINE_READER:
      return Permission.REPO_REMOVE_PIPELINE_READER;
    case GQLPermission.REPO_ADD_PIPELINE_WRITER:
      return Permission.REPO_ADD_PIPELINE_WRITER;
    case GQLPermission.PIPELINE_LIST_JOB:
      return Permission.PIPELINE_LIST_JOB;
    case GQLPermission.PROJECT_CREATE:
      return Permission.PROJECT_CREATE;
    case GQLPermission.PROJECT_DELETE:
      return Permission.PROJECT_DELETE;
    case GQLPermission.PROJECT_LIST_REPO:
      return Permission.PROJECT_LIST_REPO;
    case GQLPermission.PROJECT_CREATE_REPO:
      return Permission.PROJECT_CREATE_REPO;
    case GQLPermission.PROJECT_MODIFY_BINDINGS:
      return Permission.PROJECT_MODIFY_BINDINGS;
    case GQLPermission.CLUSTER_GET_LOKI_LOGS:
      return Permission.CLUSTER_GET_LOKI_LOGS;
    case GQLPermission.CLUSTER_SET_DEFAULTS:
      return Permission.CLUSTER_SET_DEFAULTS;
    default:
      throw new ApolloError(`Unknown GQL Permission ${permission}`);
  }
};

export const toGQLPermissionType = (permission: Permission) => {
  switch (permission) {
    case Permission.PERMISSION_UNKNOWN:
      return GQLPermission.PERMISSION_UNKNOWN;
    case Permission.CLUSTER_MODIFY_BINDINGS:
      return GQLPermission.CLUSTER_MODIFY_BINDINGS;
    case Permission.CLUSTER_GET_BINDINGS:
      return GQLPermission.CLUSTER_GET_BINDINGS;
    case Permission.CLUSTER_GET_PACHD_LOGS:
      return GQLPermission.CLUSTER_GET_PACHD_LOGS;
    case Permission.CLUSTER_AUTH_ACTIVATE:
      return GQLPermission.CLUSTER_AUTH_ACTIVATE;
    case Permission.CLUSTER_AUTH_DEACTIVATE:
      return GQLPermission.CLUSTER_AUTH_DEACTIVATE;
    case Permission.CLUSTER_AUTH_GET_CONFIG:
      return GQLPermission.CLUSTER_AUTH_GET_CONFIG;
    case Permission.CLUSTER_AUTH_SET_CONFIG:
      return GQLPermission.CLUSTER_AUTH_SET_CONFIG;
    case Permission.CLUSTER_AUTH_GET_ROBOT_TOKEN:
      return GQLPermission.CLUSTER_AUTH_GET_ROBOT_TOKEN;
    case Permission.CLUSTER_AUTH_MODIFY_GROUP_MEMBERS:
      return GQLPermission.CLUSTER_AUTH_MODIFY_GROUP_MEMBERS;
    case Permission.CLUSTER_AUTH_GET_GROUPS:
      return GQLPermission.CLUSTER_AUTH_GET_GROUPS;
    case Permission.CLUSTER_AUTH_GET_GROUP_USERS:
      return GQLPermission.CLUSTER_AUTH_GET_GROUP_USERS;
    case Permission.CLUSTER_AUTH_EXTRACT_TOKENS:
      return GQLPermission.CLUSTER_AUTH_EXTRACT_TOKENS;
    case Permission.CLUSTER_AUTH_RESTORE_TOKEN:
      return GQLPermission.CLUSTER_AUTH_RESTORE_TOKEN;
    case Permission.CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL:
      return GQLPermission.CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL;
    case Permission.CLUSTER_AUTH_DELETE_EXPIRED_TOKENS:
      return GQLPermission.CLUSTER_AUTH_DELETE_EXPIRED_TOKENS;
    case Permission.CLUSTER_AUTH_REVOKE_USER_TOKENS:
      return GQLPermission.CLUSTER_AUTH_REVOKE_USER_TOKENS;
    case Permission.CLUSTER_AUTH_ROTATE_ROOT_TOKEN:
      return GQLPermission.CLUSTER_AUTH_ROTATE_ROOT_TOKEN;
    case Permission.CLUSTER_ENTERPRISE_ACTIVATE:
      return GQLPermission.CLUSTER_ENTERPRISE_ACTIVATE;
    case Permission.CLUSTER_ENTERPRISE_HEARTBEAT:
      return GQLPermission.CLUSTER_ENTERPRISE_HEARTBEAT;
    case Permission.CLUSTER_ENTERPRISE_GET_CODE:
      return GQLPermission.CLUSTER_ENTERPRISE_GET_CODE;
    case Permission.CLUSTER_ENTERPRISE_DEACTIVATE:
      return GQLPermission.CLUSTER_ENTERPRISE_DEACTIVATE;
    case Permission.CLUSTER_ENTERPRISE_PAUSE:
      return GQLPermission.CLUSTER_ENTERPRISE_PAUSE;
    case Permission.CLUSTER_IDENTITY_SET_CONFIG:
      return GQLPermission.CLUSTER_IDENTITY_SET_CONFIG;
    case Permission.CLUSTER_IDENTITY_GET_CONFIG:
      return GQLPermission.CLUSTER_IDENTITY_GET_CONFIG;
    case Permission.CLUSTER_IDENTITY_CREATE_IDP:
      return GQLPermission.CLUSTER_IDENTITY_CREATE_IDP;
    case Permission.CLUSTER_IDENTITY_UPDATE_IDP:
      return GQLPermission.CLUSTER_IDENTITY_UPDATE_IDP;
    case Permission.CLUSTER_IDENTITY_LIST_IDPS:
      return GQLPermission.CLUSTER_IDENTITY_LIST_IDPS;
    case Permission.CLUSTER_IDENTITY_GET_IDP:
      return GQLPermission.CLUSTER_IDENTITY_GET_IDP;
    case Permission.CLUSTER_IDENTITY_DELETE_IDP:
      return GQLPermission.CLUSTER_IDENTITY_DELETE_IDP;
    case Permission.CLUSTER_IDENTITY_CREATE_OIDC_CLIENT:
      return GQLPermission.CLUSTER_IDENTITY_CREATE_OIDC_CLIENT;
    case Permission.CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT:
      return GQLPermission.CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT;
    case Permission.CLUSTER_IDENTITY_LIST_OIDC_CLIENTS:
      return GQLPermission.CLUSTER_IDENTITY_LIST_OIDC_CLIENTS;
    case Permission.CLUSTER_IDENTITY_GET_OIDC_CLIENT:
      return GQLPermission.CLUSTER_IDENTITY_GET_OIDC_CLIENT;
    case Permission.CLUSTER_IDENTITY_DELETE_OIDC_CLIENT:
      return GQLPermission.CLUSTER_IDENTITY_DELETE_OIDC_CLIENT;
    case Permission.CLUSTER_DEBUG_DUMP:
      return GQLPermission.CLUSTER_DEBUG_DUMP;
    case Permission.CLUSTER_LICENSE_ACTIVATE:
      return GQLPermission.CLUSTER_LICENSE_ACTIVATE;
    case Permission.CLUSTER_LICENSE_GET_CODE:
      return GQLPermission.CLUSTER_LICENSE_GET_CODE;
    case Permission.CLUSTER_LICENSE_ADD_CLUSTER:
      return GQLPermission.CLUSTER_LICENSE_ADD_CLUSTER;
    case Permission.CLUSTER_LICENSE_UPDATE_CLUSTER:
      return GQLPermission.CLUSTER_LICENSE_UPDATE_CLUSTER;
    case Permission.CLUSTER_LICENSE_DELETE_CLUSTER:
      return GQLPermission.CLUSTER_LICENSE_DELETE_CLUSTER;
    case Permission.CLUSTER_LICENSE_LIST_CLUSTERS:
      return GQLPermission.CLUSTER_LICENSE_LIST_CLUSTERS;
    case Permission.CLUSTER_CREATE_SECRET:
      return GQLPermission.CLUSTER_CREATE_SECRET;
    case Permission.CLUSTER_LIST_SECRETS:
      return GQLPermission.CLUSTER_LIST_SECRETS;
    case Permission.SECRET_DELETE:
      return GQLPermission.SECRET_DELETE;
    case Permission.SECRET_INSPECT:
      return GQLPermission.SECRET_INSPECT;
    case Permission.CLUSTER_DELETE_ALL:
      return GQLPermission.CLUSTER_DELETE_ALL;
    case Permission.REPO_READ:
      return GQLPermission.REPO_READ;
    case Permission.REPO_WRITE:
      return GQLPermission.REPO_WRITE;
    case Permission.REPO_MODIFY_BINDINGS:
      return GQLPermission.REPO_MODIFY_BINDINGS;
    case Permission.REPO_DELETE:
      return GQLPermission.REPO_DELETE;
    case Permission.REPO_INSPECT_COMMIT:
      return GQLPermission.REPO_INSPECT_COMMIT;
    case Permission.REPO_LIST_COMMIT:
      return GQLPermission.REPO_LIST_COMMIT;
    case Permission.REPO_DELETE_COMMIT:
      return GQLPermission.REPO_DELETE_COMMIT;
    case Permission.REPO_CREATE_BRANCH:
      return GQLPermission.REPO_CREATE_BRANCH;
    case Permission.REPO_LIST_BRANCH:
      return GQLPermission.REPO_LIST_BRANCH;
    case Permission.REPO_DELETE_BRANCH:
      return GQLPermission.REPO_DELETE_BRANCH;
    case Permission.REPO_INSPECT_FILE:
      return GQLPermission.REPO_INSPECT_FILE;
    case Permission.REPO_LIST_FILE:
      return GQLPermission.REPO_LIST_FILE;
    case Permission.REPO_ADD_PIPELINE_READER:
      return GQLPermission.REPO_ADD_PIPELINE_READER;
    case Permission.REPO_REMOVE_PIPELINE_READER:
      return GQLPermission.REPO_REMOVE_PIPELINE_READER;
    case Permission.REPO_ADD_PIPELINE_WRITER:
      return GQLPermission.REPO_ADD_PIPELINE_WRITER;
    case Permission.PIPELINE_LIST_JOB:
      return GQLPermission.PIPELINE_LIST_JOB;
    case Permission.PROJECT_CREATE:
      return GQLPermission.PROJECT_CREATE;
    case Permission.PROJECT_DELETE:
      return GQLPermission.PROJECT_DELETE;
    case Permission.PROJECT_LIST_REPO:
      return GQLPermission.PROJECT_LIST_REPO;
    case Permission.PROJECT_CREATE_REPO:
      return GQLPermission.PROJECT_CREATE_REPO;
    case Permission.PROJECT_MODIFY_BINDINGS:
      return GQLPermission.PROJECT_MODIFY_BINDINGS;
    case Permission.CLUSTER_GET_LOKI_LOGS:
      return GQLPermission.CLUSTER_GET_LOKI_LOGS;
    case Permission.CLUSTER_SET_DEFAULTS:
      return GQLPermission.CLUSTER_SET_DEFAULTS;
    default:
      throw new ApolloError(`Unknown Proto Permission ${permission}`);
  }
};
