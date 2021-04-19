import capitalize from 'lodash/capitalize';

import {PipelineState} from '@graphqlTypes';

const readablePipelineState = (PipelineState: PipelineState | string) => {
  const state = PipelineState.toString().replace('PIPELINE_', '');
  return capitalize(state);
};

export default readablePipelineState;
