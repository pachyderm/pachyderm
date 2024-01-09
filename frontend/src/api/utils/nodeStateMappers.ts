import {PipelineState, JobState} from '@dash-frontend/api/pps';
import {NodeState} from '@dash-frontend/lib/types';

export const restPipelineStateToNodeState = (pipelineState?: PipelineState) => {
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

export const restJobStateToNodeState = (jobState?: JobState) => {
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

export const nodeStateToJobStateEnum = (nodeState: NodeState) => {
  switch (nodeState) {
    case NodeState.SUCCESS:
      return [JobState.JOB_SUCCESS];
    case NodeState.RUNNING:
      return [
        JobState.JOB_CREATED,
        JobState.JOB_RUNNING,
        JobState.JOB_EGRESSING,
        JobState.JOB_STARTING,
        JobState.JOB_FINISHING,
      ];
    case NodeState.ERROR:
      return [
        JobState.JOB_FAILURE,
        JobState.JOB_KILLED,
        JobState.JOB_UNRUNNABLE,
      ];
    default:
      return [JobState.JOB_STATE_UNKNOWN];
  }
};
