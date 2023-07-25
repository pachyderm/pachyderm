import {NodeType} from '@graphqlTypes';

import {deriveNameFromNodeNameAndType} from '@dash-frontend/lib/deriveRepoNameFromNode';

describe('deriveRepoNameFromNode', () => {
  const cases: [string, NodeType, string][] = [
    ['images_repo', NodeType.INPUT_REPO, 'images'],
    ['images_repo', NodeType.OUTPUT_REPO, 'images'],
    ['images_pipeline_repo', NodeType.PIPELINE, 'images_pipeline_repo'],
  ];
  test.each(cases)('%p returns %p', (input1, input2, result) => {
    expect(deriveNameFromNodeNameAndType(input1, input2)).toBe(result);
  });
});
