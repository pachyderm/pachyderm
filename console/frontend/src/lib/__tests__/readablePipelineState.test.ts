import {PipelineState} from '@dash-frontend/api/pps';
import readablePipelineState from '@dash-frontend/lib/readablePipelineState';

describe('readablePipelineState', () => {
  const cases: [PipelineState | string, string][] = [
    [PipelineState.PIPELINE_CRASHING, 'Crashing'],
    [PipelineState.PIPELINE_FAILURE, 'Failure'],
    [PipelineState.PIPELINE_PAUSED, 'Paused'],
    [PipelineState.PIPELINE_RESTARTING, 'Restarting'],
    [PipelineState.PIPELINE_RUNNING, 'Running'],
    [PipelineState.PIPELINE_STANDBY, 'Standby'],
    [PipelineState.PIPELINE_STARTING, 'Starting'],
    [PipelineState.PIPELINE_STATE_UNKNOWN, 'Unknown'],
    ['', ''],
  ];
  test.each(cases)('%p returns %p', (input, result) => {
    expect(readablePipelineState(input)).toBe(result);
  });
});
