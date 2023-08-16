import {mockGetEnterpriseInfoQuery, EnterpriseState} from '@graphqlTypes';

export const getFutureTimeStamp = (days = 30) => {
  const currentDate = new Date();

  currentDate.setDate(currentDate.getDate() + days);

  return Math.floor(currentDate.getTime() / 1000);
};

export const mockGetEnterpriseInfo = () =>
  mockGetEnterpriseInfoQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        enterpriseInfo: {
          state: EnterpriseState.ACTIVE,
          expiration: getFutureTimeStamp(),
        },
      }),
    );
  });

export const mockGetEnterpriseInfoInactive = () =>
  mockGetEnterpriseInfoQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        enterpriseInfo: {
          state: EnterpriseState.NONE,
          expiration: -1,
        },
      }),
    );
  });

export const mockGetEnterpriseInfoExpiring = () =>
  mockGetEnterpriseInfoQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        enterpriseInfo: {
          state: EnterpriseState.ACTIVE,
          expiration: getFutureTimeStamp(1),
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
