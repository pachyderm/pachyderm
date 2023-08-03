import {operation} from 'retry';

import {enterpriseInfoToGQLInfo} from '@dash-backend/resolvers/builders/enterprise';
import {EnterpriseState} from '@graphqlTypes';

import {generateConsoleTraceUuid} from './generateTrace';
import getPachClientAndAttachHeaders from './getPachClient';
interface retryAnalyticsContextResponse {
  anonymousId?: string;
  expiration?: number;
  enterpriseState?: EnterpriseState;
}

const MAX_RETRY = 10;

async function retryAnalyticsContext() {
  const retryObj = operation({
    retries: MAX_RETRY,
  });
  return new Promise<retryAnalyticsContextResponse>((resolve, reject) => {
    retryObj.attempt(async () => {
      try {
        const pachClient = getPachClientAndAttachHeaders({
          requestId: generateConsoleTraceUuid(),
        });

        const [clusterInfo, enterpriseInfo] = await Promise.all([
          pachClient.admin.inspectCluster(),
          pachClient.enterprise.getState(),
        ]);

        const {expiration, state} = enterpriseInfoToGQLInfo(enterpriseInfo);
        return resolve({
          anonymousId: clusterInfo.id,
          expiration: expiration,
          enterpriseState: state,
        });
      } catch (err) {
        if (!retryObj.retry(new Error())) {
          return reject();
        }
      }
    });
  });
}

export default retryAnalyticsContext;
