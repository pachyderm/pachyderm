import {rest} from 'msw';
import {setupServer} from 'msw/node';

import {
  account,
  authenticate,
  config,
  Account,
  AuthConfig,
  AuthenticateRequest,
  AuthenticateResponse,
  Exchange,
  WhoAmIRequest,
} from '@dash-frontend/api/auth';
import {Empty, CODES} from '@dash-frontend/api/googleTypes';
import {RequestError} from '@dash-frontend/api/utils/error';
import {
  mockAuthConfigError,
  mockAuthConfigNotConfigured,
  mockExchangeCode,
  mockWhoAmINotActivated,
} from '@dash-frontend/mocks';

describe('api/rest', () => {
  const server = setupServer();

  beforeAll(() => server.listen());

  beforeEach(() => {
    localStorage.removeItem('id-token');
    localStorage.removeItem('auth-token');
    server.resetHandlers();
  });

  afterAll(() => server.close());

  describe('config', () => {
    it('should return an auth config', async () => {
      server.use(
        rest.get<Empty, Empty, AuthConfig>('/auth/config', (_req, res, ctx) => {
          return res(
            ctx.json({
              authEndpoint: '/dex',
              clientId: 'client',
              pachdClientId: 'pachd',
            }),
          );
        }),
      );

      expect(await config()).toEqual({
        authEndpoint: '/dex',
        clientId: 'client',
        pachdClientId: 'pachd',
      });
    });

    it('should return an empty config when auth is disabled', async () => {
      server.use(mockAuthConfigNotConfigured());
      server.use(mockWhoAmINotActivated());
      expect(await config()).toEqual({
        authEndpoint: '',
        clientId: '',
        pachdClientId: '',
      });
    });

    it('should throw other errors when auth is enabled', async () => {
      server.use(mockAuthConfigError());
      server.use(
        rest.post<WhoAmIRequest, Empty, RequestError>(
          '/api/auth_v2.API/WhoAmI',
          (_req, res, ctx) => {
            return res(
              ctx.status(500),
              ctx.json({
                code: 1,
              }),
            );
          },
        ),
      );

      await expect(() => config()).rejects.toThrow(
        'Authentication Error: Issuer is misconfigured.',
      );
    });
  });

  describe('account', () => {
    it('should return an account', async () => {
      localStorage.setItem('id-token', '789');
      server.use(
        rest.post<Empty, Empty, Account>('/auth/account', (_req, res, ctx) => {
          return res(
            ctx.json({
              id: '123',
              email: 'test@test.com',
              name: 'test',
            }),
          );
        }),
      );

      expect(await account()).toEqual({
        id: '123',
        email: 'test@test.com',
        name: 'test',
      });
    });

    it('should return an empty account when auth is disabled', async () => {
      server.use(mockWhoAmINotActivated());

      expect(await account()).toEqual({
        id: 'unauthenticated',
        email: '',
        name: 'User',
      });
    });

    it('throw other errors when auth is enabled', async () => {
      server.use(
        rest.post<WhoAmIRequest, Empty, RequestError>(
          '/api/auth_v2.API/WhoAmI',
          (_req, res, ctx) => {
            return res(
              ctx.status(401),
              ctx.json({
                message: 'some error',
              }),
            );
          },
        ),
      );

      await expect(() => account()).rejects.toThrow(
        'Authentication Error: Could not retrieve an account.',
      );
    });
  });

  describe('authenticate', () => {
    it('should return auth tokens', async () => {
      server.use(
        rest.post<Empty, Empty, Exchange>(
          '/auth/exchange',
          (_req, res, ctx) => {
            return res(
              ctx.json({
                idToken: '1234',
              }),
            );
          },
        ),
      );
      server.use(
        rest.post<AuthenticateRequest, Empty, AuthenticateResponse>(
          '/api/auth_v2.API/Authenticate',
          (_req, res, ctx) => {
            return res(ctx.json({pachToken: '0987'}));
          },
        ),
      );

      expect(await authenticate('321')).toEqual({
        idToken: '1234',
        pachToken: '0987',
      });
    });

    it('should throw an error when it did not receive an id token', async () => {
      server.use(
        rest.post<Empty, Empty, RequestError>(
          '/auth/exchange',
          (_req, res, ctx) => {
            return res(
              ctx.json({
                message: 'Authenticate Error',
                details: ['Could not exchange codes.'],
              }),
            );
          },
        ),
      );

      await expect(() => authenticate('321')).rejects.toThrow(
        'Authenticate Error: Could not exchange codes.',
      );
    });

    it('should throw an error when it did not receive a pach token', async () => {
      server.use(
        rest.post<Empty, Empty, Exchange>(
          '/auth/exchange',
          (_req, res, ctx) => {
            return res(
              ctx.json({
                idToken: '1234',
              }),
            );
          },
        ),
      );
      server.use(
        rest.post<AuthenticateRequest, Empty, AuthenticateResponse>(
          '/api/auth_v2.API/Authenticate',
          (_req, res, ctx) => {
            return res(ctx.json({pachToken: ''}));
          },
        ),
      );

      await expect(() => authenticate('321')).rejects.toThrow(
        'Authenticate Error: Did not receive a pachToken.',
      );
    });

    it('should return empty tokens when auth is disabled', async () => {
      server.use(mockExchangeCode());
      server.use(
        rest.post<AuthenticateRequest, Empty, RequestError>(
          '/api/auth_v2.API/Authenticate',
          (_req, res, ctx) => {
            return res(
              ctx.status(501),
              ctx.json({
                code: CODES.Unimplemented,
              }),
            );
          },
        ),
      );

      expect(await authenticate('321')).toEqual({
        idToken: '',
        pachToken: '',
      });
    });
  });
});
