import flattenDeep from 'lodash/flattenDeep';
import objectHash from 'object-hash';

import {Input} from '@dash-backend/proto';
import {VertexIdentifier} from '@graphqlTypes';

const flattenPipelineInput = (input: Input.AsObject): VertexIdentifier[] => {
  const result = [];

  if (input.pfs) {
    const {repo: name, project} = input.pfs;

    result.push({
      name,
      project,
      id: objectHash({project, name}),
    });
  }

  if (input.cron) {
    const {repo: name, project} = input.cron;

    result.push({
      name,
      project,
      id: objectHash({project, name}),
    });
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
