import {PipelineInfoPipelineType} from '@dash-frontend/api/pps';

import readablePipelineType from '../readablePipelineType';

describe('readablePipelineState', () => {
  const cases: [PipelineInfoPipelineType | undefined | string, string][] = [
    [PipelineInfoPipelineType.PIPELINE_TYPE_SERVICE, 'Service'],
    [PipelineInfoPipelineType.PIPELINE_TYPE_SPOUT, 'Spout'],
    [PipelineInfoPipelineType.PIPELINE_TYPE_TRANSFORM, 'Standard'],
    [PipelineInfoPipelineType.PIPELINT_TYPE_UNKNOWN, 'Unknown'],
    [undefined, ''],
    ['', ''],
  ];
  test.each(cases)('%p returns %p', (input, result) => {
    expect(
      readablePipelineType(input as PipelineInfoPipelineType | undefined),
    ).toBe(result);
  });
});
