import {rest} from 'msw';

import {
  AuthorizeRequest,
  AuthorizeResponse,
  GetRoleBindingRequest,
  GetRoleBindingResponse,
  AuthenticateRequest,
  AuthenticateResponse,
  Permission,
  WhoAmIRequest,
} from '@dash-frontend/api/auth';
import {CODES, Empty} from '@dash-frontend/api/googleTypes';
import {RequestError} from '@dash-frontend/api/utils/error';

export const mockEmptyGetAuthorize = () =>
  rest.post<AuthorizeRequest, Empty, AuthorizeResponse>(
    '/api/auth_v2.API/Authorize',
    (_req, res, ctx) => {
      return res(
        ctx.json({
          authorized: undefined,
          satisfied: [],
          missing: [],
          principal: '',
        }),
      );
    },
  );

export const mockFalseGetAuthorize = () =>
  rest.post<AuthorizeRequest, Empty, AuthorizeResponse>(
    '/api/auth_v2.API/Authorize',
    (_req, res, ctx) => {
      return res(
        ctx.json({
          authorized: false,
          satisfied: [],
          missing: [],
          principal: '',
        }),
      );
    },
  );

export const mockTrueGetAuthorize = (permissions: Permission[] = []) =>
  rest.post<AuthorizeRequest, Empty, AuthorizeResponse>(
    '/api/auth_v2.API/Authorize',
    (_req, res, ctx) => {
      return res(
        ctx.json({
          authorized: true,
          satisfied: permissions,
          missing: [],
          principal: '',
        }),
      );
    },
  );

export const mockGetAccountAuth = () =>
  rest.post('/auth/account', (_req, res, ctx) => {
    return res(
      ctx.json({
        email: 'email@user.com',
        id: 'TestUsername',
        name: 'User Test',
      }),
    );
  });

export const mockEmptyGetRoles = () =>
  rest.post<GetRoleBindingRequest, Empty, GetRoleBindingResponse>(
    '/api/auth_v2.API/GetRoleBinding',
    (_req, res, ctx) => {
      return res(
        ctx.json({
          binding: {
            entries: {},
          },
        }),
      );
    },
  );

export const mockAuthConfig = () =>
  rest.get('/auth/config', (_req, res, ctx) => {
    return res(
      ctx.json({
        authEndpoint: '/dex/auth',
        clientId: 'console-test',
        pachdClientId: 'pachd',
      }),
    );
  });

export const mockExchangeCode = () =>
  rest.post('/auth/exchange', (_req, res, ctx) => {
    return res(
      ctx.json({
        idToken: '123',
      }),
    );
  });

export const mockAuthenticate = () =>
  rest.post<AuthenticateRequest, Empty, AuthenticateResponse>(
    '/api/auth_v2.API/Authenticate',
    (_req, res, ctx) => {
      return res(
        ctx.json({
          pachToken: 'abc',
        }),
      );
    },
  );

export const mockWhoAmINotActivated = () =>
  rest.post<WhoAmIRequest, Empty, RequestError>(
    '/api/auth_v2.API/WhoAmI',
    (_req, res, ctx) => {
      return res(
        ctx.status(501),
        ctx.json({
          code: CODES.Unimplemented,
          message: 'the auth service is not activated',
          details: [],
        }),
      );
    },
  );

export const mockAuthConfigError = () =>
  rest.get<never, never, RequestError>('/auth/config', (_req, res, ctx) => {
    return res(
      ctx.status(401),
      ctx.json({
        message: 'Authentication Error',
        details: ['Issuer is misconfigured.'],
      }),
    );
  });

export const mockAuthConfigNotConfigured = () =>
  rest.get<never, never, RequestError>('/auth/config', (_req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        message: 'Authentication Error',
        details: [
          'Unable to connect to authorization issuer.',
          'OPError: expected 200 OK, got: 503 Service Unavailable',
        ],
      }),
    );
  });

export const mockAuthExchanceError = () =>
  rest.post<never, never, RequestError>('/auth/exchange', (_req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        message: 'Authentication Error',
        details: [
          'OPError: invalid_grant (Invalid or expired code parameter.)',
        ],
      }),
    );
  });
