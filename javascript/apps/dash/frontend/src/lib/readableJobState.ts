import capitalize from 'lodash/capitalize';

import {PipelineJobState} from '@graphqlTypes';

const readableJobState = (pipelineJobState: PipelineJobState | string) => {
  const state = pipelineJobState.toString().replace('JOB_', '');
  return capitalize(state);
};

export default readableJobState;
