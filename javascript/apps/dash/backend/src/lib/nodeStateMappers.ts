import {NodeState, PipelineState, JobState} from '@graphqlTypes';

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
    case JobState.JOB_RUNNING:
    case JobState.JOB_EGRESSING:
    case JobState.JOB_STARTING:
      return NodeState.BUSY;
    case JobState.JOB_FAILURE:
    case JobState.JOB_KILLED:
      return NodeState.ERROR;
    default:
      return NodeState.IDLE;
  }
};
