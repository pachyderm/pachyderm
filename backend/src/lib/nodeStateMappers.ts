import {JobState as ProtoJobState} from '@dash-backend/proto';
import {NodeState, PipelineState, InputMaybe, JobState} from '@graphqlTypes';

export const gqlPipelineStateToNodeState = (pipelineState: PipelineState) => {
  switch (pipelineState) {
    case PipelineState.PIPELINE_RUNNING:
    case PipelineState.PIPELINE_STANDBY:
      return NodeState.IDLE;
    case PipelineState.PIPELINE_PAUSED:
      return NodeState.PAUSED;
    case PipelineState.PIPELINE_STARTING:
    case PipelineState.PIPELINE_RESTARTING:
      return NodeState.BUSY;
    case PipelineState.PIPELINE_FAILURE:
    case PipelineState.PIPELINE_CRASHING:
      return NodeState.ERROR;
    default:
      return NodeState.IDLE;
  }
};

export const gqlJobStateToNodeState = (jobState: JobState) => {
  switch (jobState) {
    case JobState.JOB_SUCCESS:
      return NodeState.SUCCESS;
    case JobState.JOB_CREATED:
    case JobState.JOB_RUNNING:
    case JobState.JOB_EGRESSING:
    case JobState.JOB_STARTING:
    case JobState.JOB_FINISHING:
      return NodeState.RUNNING;
    case JobState.JOB_FAILURE:
    case JobState.JOB_KILLED:
    case JobState.JOB_UNRUNNABLE:
      return NodeState.ERROR;
    default:
      return NodeState.IDLE;
  }
};

export const nodeStateToJobStateEnum = (nodeState: InputMaybe<NodeState>) => {
  switch (nodeState) {
    case NodeState.SUCCESS:
      return [ProtoJobState.JOB_SUCCESS];
    case NodeState.RUNNING:
      return [
        ProtoJobState.JOB_CREATED,
        ProtoJobState.JOB_RUNNING,
        ProtoJobState.JOB_EGRESSING,
        ProtoJobState.JOB_STARTING,
        ProtoJobState.JOB_FINISHING,
      ];
    case NodeState.ERROR:
      return [
        ProtoJobState.JOB_FAILURE,
        ProtoJobState.JOB_KILLED,
        ProtoJobState.JOB_UNRUNNABLE,
      ];
    default:
      return [ProtoJobState.JOB_STATE_UNKNOWN];
  }
};
