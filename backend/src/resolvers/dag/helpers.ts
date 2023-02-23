import {Egress} from '@dash-backend/proto';

export const postfixNameWithRepo = (name?: string) => `${name}_repo`;
export const egressNodeName = (egress: Egress.AsObject) =>
  egress.url || egress.sqlDatabase?.url || egress.objectStorage?.url;
