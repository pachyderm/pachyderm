import capitalize from 'lodash/capitalize';

import {PipelineState} from '@dash-frontend/api/pps';

const readablePipelineState = (PipelineState?: PipelineState | string) => {
  if (!PipelineState) return '';
  const state = PipelineState.toString().replace(/PIPELINE_(STATE_)?/, '');
  return capitalize(state);
};

export default readablePipelineState;
