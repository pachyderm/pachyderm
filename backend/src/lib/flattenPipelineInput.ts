import flattenDeep from 'lodash/flattenDeep';

import {Input} from '@dash-backend/proto';

export const flattenPipelineInputObj = (
  input: Input.AsObject,
): {project: string; repo: string}[] => {
  const result = [];

  if (input.pfs) {
    const {repo, project} = input.pfs;

    result.push({repo, project});
  }

  if (input.cron) {
    const {repo, project} = input.cron;

    result.push({repo, project});
  }

  // TODO: Update to indicate which elements are crossed, unioned, joined, grouped, with.
  return flattenDeep([
    result,
    input.crossList.map((i) => flattenPipelineInputObj(i)),
    input.unionList.map((i) => flattenPipelineInputObj(i)),
    input.joinList.map((i) => flattenPipelineInputObj(i)),
    input.groupList.map((i) => flattenPipelineInputObj(i)),
  ]);
};

const flattenPipelineInput = (input: Input.AsObject): string[] => {
  const res = flattenPipelineInputObj(input);
  return res.map(({project, repo}) => `${project}_${repo}`);
};

export default flattenPipelineInput;
