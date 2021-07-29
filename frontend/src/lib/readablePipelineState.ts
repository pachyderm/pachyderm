import {PipelineState} from '@graphqlTypes';
import capitalize from 'lodash/capitalize';

const readablePipelineState = (PipelineState: PipelineState | string) => {
  const state = PipelineState.toString().replace('PIPELINE_', '');
  return capitalize(state);
};

export default readablePipelineState;
