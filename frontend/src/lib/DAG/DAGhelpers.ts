import flattenDeep from 'lodash/flattenDeep';
import sortBy from 'lodash/sortBy';
import objectHash from 'object-hash';

import {JobInfo, Egress, Input} from '@dash-frontend/api/pps';
import {getUnixSecondsFromISOString} from '@dash-frontend/lib/dateTime';

type EgressInput = Egress & {URL?: string};

export const egressNodeName = (egress: EgressInput) => {
  return (
    egress.URL ||
    egress.uRL ||
    egress.sqlDatabase?.url ||
    egress.objectStorage?.url
  );
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
export type VertexIdentifier = {
  id: string;
  name: string;
  project: string;
};

export const flattenPipelineInput = (input: Input): VertexIdentifier[] => {
  const result: VertexIdentifier[] = [];

  if (input.pfs) {
    const {repo: name, project} = input.pfs;

    result.push({
      name: name || '',
      project: project || '',
      id: objectHash({project, name}),
    });
  }

  if (input.cron) {
    const {repo: name, project} = input.cron;

    result.push({
      name: name || '',
      project: project || '',
      id: objectHash({project, name}),
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
