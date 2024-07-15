import {rest} from 'msw';

import {
  GetStateRequest,
  GetStateResponse,
  State,
} from '@dash-frontend/api/enterprise';
import {Empty} from '@dash-frontend/api/googleTypes';

export const getFutureTimeStamp = (days = 30) => {
  const currentDate = new Date();

  currentDate.setDate(currentDate.getDate() + days);

  return Math.floor(currentDate.getTime() / 1000);
};

export const getFutureISO = (days = 30) => {
  const currentDate = new Date();

  currentDate.setDate(currentDate.getDate() + days);

  return currentDate.toISOString();
};

export const mockGetEnterpriseInfo = () =>
  rest.post<GetStateRequest, Empty, GetStateResponse>(
    '/api/enterprise_v2.API/GetState',
    (_req, res, ctx) => {
      return res(
        ctx.json({
          activationCode: '',
          state: State.ACTIVE,
          info: {
            expires: getFutureISO(),
          },
        }),
      );
    },
  );

export const mockGetEnterpriseInfoInactive = () =>
  rest.post<GetStateRequest, Empty, GetStateResponse>(
    '/api/enterprise_v2.API/GetState',
    (_req, res, ctx) => {
      return res(
        ctx.json({
          activationCode: '',
          state: State.NONE,
        }),
      );
    },
  );

export const mockGetEnterpriseInfoExpiring = () =>
  rest.post<GetStateRequest, Empty, GetStateResponse>(
    '/api/enterprise_v2.API/GetState',
    (_req, res, ctx) => {
      return res(
        ctx.json({
          activationCode: '',
          state: State.ACTIVE,
          info: {
            expires: getFutureISO(1),
          },
        }),
      );
    },
  );
