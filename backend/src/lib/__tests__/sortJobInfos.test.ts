import sortJobInfos from '@dash-backend/lib/sortJobInfos';
import {JobState} from '@dash-backend/proto';
import {jobInfoFromObject} from '@dash-backend/proto/builders/pps';

describe('sortJobInfos', () => {
  it('should sortJobInfos by createdAt and pipeline name', () => {
    const jobInfos = [
      jobInfoFromObject({
        job: {id: '1', pipeline: {name: 'A'}},
        state: JobState.JOB_SUCCESS,
        createdAt: {seconds: 1631196482, nanos: 0},
      }).toObject(),
      jobInfoFromObject({
        job: {id: '3', pipeline: {name: 'D'}},
        state: JobState.JOB_SUCCESS,
        createdAt: {seconds: 1631196522, nanos: 400},
      }).toObject(),
      jobInfoFromObject({
        job: {id: '4', pipeline: {name: 'C'}},
        state: JobState.JOB_SUCCESS,
        createdAt: {seconds: 1631196522, nanos: 400},
      }).toObject(),
      jobInfoFromObject({
        job: {id: '2', pipeline: {name: 'B'}},
        state: JobState.JOB_SUCCESS,
        createdAt: {seconds: 1631196522, nanos: 0},
      }).toObject(),
    ];

    const sortedJobInfos = sortJobInfos(jobInfos);

    expect(sortedJobInfos).toEqual([
      jobInfoFromObject({
        job: {id: '1', pipeline: {name: 'A'}},
        state: JobState.JOB_SUCCESS,
        createdAt: {seconds: 1631196482, nanos: 0},
      }).toObject(),
      jobInfoFromObject({
        job: {id: '2', pipeline: {name: 'B'}},
        state: JobState.JOB_SUCCESS,
        createdAt: {seconds: 1631196522, nanos: 0},
      }).toObject(),
      jobInfoFromObject({
        job: {id: '4', pipeline: {name: 'C'}},
        state: JobState.JOB_SUCCESS,
        createdAt: {seconds: 1631196522, nanos: 400},
      }).toObject(),
      jobInfoFromObject({
        job: {id: '3', pipeline: {name: 'D'}},
        state: JobState.JOB_SUCCESS,
        createdAt: {seconds: 1631196522, nanos: 400},
      }).toObject(),
    ]);
  });
});
