import {JobState} from '@dash-frontend/api/pps';
import {readableJobState, getJobRuntime} from '@dash-frontend/lib/jobs';

describe('readableJobState', () => {
  const cases: [JobState | string, string][] = [
    [JobState.JOB_CREATED, 'Created'],
    [JobState.JOB_EGRESSING, 'Egressing'],
    [JobState.JOB_FAILURE, 'Failure'],
    [JobState.JOB_FINISHING, 'Finishing'],
    [JobState.JOB_KILLED, 'Killed'],
    [JobState.JOB_RUNNING, 'Running'],
    [JobState.JOB_STARTING, 'Starting'],
    [JobState.JOB_STATE_UNKNOWN, 'Unknown'],
    [JobState.JOB_SUCCESS, 'Success'],
    [JobState.JOB_UNRUNNABLE, 'Unrunnable'],
    ['', ''],
  ];
  test.each(cases)('%p returns %p', (input, result) => {
    expect(readableJobState(input)).toBe(result);
  });
});

describe('getJobRuntime', () => {
  const cases: [number | null, number | null, string][] = [
    [null, null, 'In Progress'],
    [1689125549, null, ' - In Progress'],
    [null, 1689126549, 'In Progress'],
    [1689125549, 1689126549, '16 mins 40 s'],
  ];
  test.each(cases)('%p returns %p', (input1, input2, result) => {
    expect(getJobRuntime(input1, input2)).toContain(result);
  });
});
