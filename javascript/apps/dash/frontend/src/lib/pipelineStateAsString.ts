import {PipelineState} from '@graphqlTypes';

const pipelineStateAsString = (state: PipelineState): string => {
  return state.toString().replace('PIPELINE_', '');
};

export default pipelineStateAsString;
