import {Node, PipelineState} from '@graphqlTypes';

const convertNodeStateToDagState = (state: Node['state']) => {
  if (!state) return '';

  switch (state) {
    case PipelineState.PIPELINE_RUNNING:
      return 'idle';
    case PipelineState.PIPELINE_PAUSED:
      return 'paused';
    case PipelineState.PIPELINE_STANDBY:
    case PipelineState.PIPELINE_STARTING:
    case PipelineState.PIPELINE_RESTARTING:
      return 'busy';
    case PipelineState.PIPELINE_FAILURE:
    case PipelineState.PIPELINE_CRASHING:
      return 'error';
    default:
      return 'idle';
  }
};

export default convertNodeStateToDagState;
