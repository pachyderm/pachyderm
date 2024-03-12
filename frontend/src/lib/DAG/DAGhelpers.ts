import flattenDeep from 'lodash/flattenDeep';
import sortBy from 'lodash/sortBy';
import objectHash from 'object-hash';

import {
  JobInfo,
  Egress,
  Input,
  PipelineInfoDetails,
} from '@dash-frontend/api/pps';
import {getUnixSecondsFromISOString} from '@dash-frontend/lib/dateTime';

import {EgressType} from '../types';

type EgressInput = Egress & {URL?: string};

export const egressNodeName = (egress?: EgressInput) => {
  if (egress) {
    return (
      egress.URL ||
      egress.uRL ||
      egress.sqlDatabase?.url ||
      egress.objectStorage?.url
    );
  }
};

export const egressType = (details?: PipelineInfoDetails) => {
  if (details) {
    if (details.s3Out) {
      return EgressType.S3;
    } else if (details.egress) {
      if (details.egress.objectStorage) {
        return EgressType.STORAGE;
      } else if (details.egress.sqlDatabase) {
        return EgressType.DB;
      }
    }
  }
};

export const generateVertexId = (project: string, name: string) => {
  return objectHash({
    project,
    name,
  });
};

export const sortJobInfos = (jobInfos: JobInfo[]) => {
  const filteredJobInfos = jobInfos.filter((j) => !!j);
  return sortBy(filteredJobInfos, [
    (jobInfo) =>
      jobInfo.created ? getUnixSecondsFromISOString(jobInfo.created) : 0,
    (jobInfo) => jobInfo.job?.pipeline?.name || '',
  ]);
};

export enum VertexIdentifierType {
  DEFAULT = 'DEFAULT',
  CRON = 'CRON',
}

export type VertexIdentifier = {
  id: string;
  name: string;
  project: string;
  type: VertexIdentifierType;
};

export const flattenPipelineInput = (input: Input): VertexIdentifier[] => {
  const result: VertexIdentifier[] = [];

  if (input.pfs) {
    const {repo: name, project} = input.pfs;

    result.push({
      name: name || '',
      project: project || '',
      id: objectHash({project, name}),
      type: VertexIdentifierType.DEFAULT,
    });
  }

  if (input.cron) {
    const {repo: name, project} = input.cron;

    result.push({
      name: name || '',
      project: project || '',
      id: objectHash({project, name}),
      type: VertexIdentifierType.CRON,
    });
  }

  // TODO: Update to indicate which elements are crossed, unioned, joined, grouped, with.
  return flattenDeep([
    result,
    input.cross?.map((i) => flattenPipelineInput(i)) || [],
    input.union?.map((i) => flattenPipelineInput(i)) || [],
    input.join?.map((i) => flattenPipelineInput(i)) || [],
    input.group?.map((i) => flattenPipelineInput(i)) || [],
  ]);
};
