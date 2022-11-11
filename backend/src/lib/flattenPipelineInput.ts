import flattenDeep from 'lodash/flattenDeep';

import {Input} from '@dash-backend/proto';

const flattenPipelineInput = (input: Input.AsObject): string[] => {
  const result = [];

  if (input.pfs) {
    const {repo} = input.pfs;
    result.push(repo);
  }

  if (input.cron) {
    const {repo} = input.cron;

    result.push(repo);
  }

  // TODO: Update to indicate which elements are crossed, unioned, joined, grouped, with.
  return flattenDeep([
    result,
    input.crossList.map((i) => flattenPipelineInput(i)),
    input.unionList.map((i) => flattenPipelineInput(i)),
    input.joinList.map((i) => flattenPipelineInput(i)),
    input.groupList.map((i) => flattenPipelineInput(i)),
  ]);
};

export default flattenPipelineInput;
