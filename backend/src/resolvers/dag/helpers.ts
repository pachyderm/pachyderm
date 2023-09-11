import objectHash from 'object-hash';

import {Egress} from '@dash-backend/proto';

export const egressNodeName = (egress: Egress.AsObject) =>
  egress.url || egress.sqlDatabase?.url || egress.objectStorage?.url;

export const generateVertexId = (project: string, name: string) =>
  objectHash({
    project,
    name,
  });
