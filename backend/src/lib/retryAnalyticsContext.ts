import {pachydermClient} from '@pachyderm/node-pachyderm';
import {operation} from 'retry';

import {enterpriseInfoToGQLInfo} from '@dash-backend/resolvers/builders/enterprise';
import {EnterpriseState} from '@graphqlTypes';

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
        const pachClient = pachydermClient({
          pachdAddress: process.env.PACHD_ADDRESS,
          ssl: false,
        });

        const [clusterInfo, enterpriseInfo] = await Promise.all([
          pachClient.admin().inspectCluster(),
          pachClient.enterprise().getState(),
        ]);

        const {expiration, state} = enterpriseInfoToGQLInfo(enterpriseInfo);
        return resolve({
          anonymousId: clusterInfo.id,
          expiration: expiration,
          enterpriseState: state,
        });
      } catch (err) {
        if (!retryObj.retry(new Error())) {
          return reject('Maximum retries reached');
        }
      }
    });
  });
}

export default retryAnalyticsContext;
