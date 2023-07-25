import {mockGetEnterpriseInfoQuery, EnterpriseState} from '@graphqlTypes';

// 30 days into the future
const currentDate = new Date();
currentDate.setDate(currentDate.getDate() + 30);
const timestamp = Math.floor(currentDate.getTime() / 1000);

export const mockGetEnterpriseInfo = () =>
  mockGetEnterpriseInfoQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        enterpriseInfo: {
          state: EnterpriseState.ACTIVE,
          expiration: timestamp,
        },
      }),
    );
  });

export const mockGetEnterpriseInfoExpired = () =>
  mockGetEnterpriseInfoQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        enterpriseInfo: {
          state: EnterpriseState.EXPIRED,
          expiration: -1,
        },
      }),
    );
  });

export const mockGetEnterpriseInfoHeartbeatFailed = () =>
  mockGetEnterpriseInfoQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        enterpriseInfo: {
          state: EnterpriseState.HEARTBEAT_FAILED,
          expiration: -1,
        },
      }),
    );
  });
