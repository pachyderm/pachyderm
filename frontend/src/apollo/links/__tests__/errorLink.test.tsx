import {mockProjectDetailsQuery, mockProjectQuery} from '@graphqlTypes';
import {render, waitFor, screen} from '@testing-library/react';
import Cookies from 'js-cookie';
import {setupServer} from 'msw/node';
import React from 'react';

import useProject from '@dash-frontend/hooks/useProject';
import {useProjectDetails} from '@dash-frontend/hooks/useProjectDetails';
import {withContextProviders} from '@dash-frontend/testHelpers';

const windowLocation = window.location;

describe('errorLink', () => {
  const server = setupServer();

  beforeAll(() => {
    server.listen();
  });

  afterAll(() => server.close());

  describe('unauthenticated', () => {
    const TestBed = withContextProviders(() => {
      useProjectDetails('bogus');

      return <>test</>;
    });

    afterEach(() => {
      window.location = windowLocation;
      Cookies.remove('dashAuthToken');
      Cookies.remove('dashAddress');
      window.localStorage.removeItem('auth-token');
    });

    it('should log the user out if they receive an unauthenticated error', async () => {
      server.use(
        mockProjectDetailsQuery((_req, res, ctx) => {
          return res(
            ctx.errors([
              {
                message: 'User is not authenticated',
                locations: [
                  {
                    line: 2,
                    column: 3,
                  },
                ],
                path: ['projectDetails'],
                extensions: {
                  code: 'UNAUTHENTICATED',
                  exception: {
                    stacktrace: [
                      'AuthenticationError: User is not authenticated',
                    ],
                  },
                },
              },
            ]),
          );
        }),
      );

      window.localStorage.setItem('auth-token', 'expired');
      Cookies.set('dashAuthToken', 'dashAuthToken');
      Cookies.set('dashAddress', 'dashAddress');

      Object.defineProperty(window, 'location', {
        configurable: true,
        writable: true,
        value: {
          ...windowLocation,
        },
      });

      const hrefSpy = jest.fn();
      Object.defineProperty(window.location, 'href', {
        set: hrefSpy,
        get: () => '/',
      });

      render(<TestBed />);

      await waitFor(() =>
        expect(window.localStorage.getItem('auth-token')).toBeFalsy(),
      );
      expect(hrefSpy).toHaveBeenCalledWith('/');
      expect(Cookies.get('dashAuthToken')).toBeUndefined();
      expect(Cookies.get('dashAddress')).toBeUndefined();
    });
  });

  describe('not found', () => {
    const TestBed = withContextProviders(() => {
      useProject({id: 'bogus'});
      return <>test</>;
    });

    afterEach(() => {
      window.history.replaceState('', '', '/');
    });

    it('should redirect the user to /not-found if a resource does not exist', async () => {
      server.use(
        mockProjectQuery((_req, res, ctx) => {
          return res(
            ctx.errors([
              {
                message: 'projects bogus not found',
                locations: [
                  {
                    line: 2,
                    column: 3,
                  },
                ],
                path: ['project'],
                extensions: {
                  code: 'NOT_FOUND',
                  exception: {
                    stacktrace: ['NotFoundError: projects bogus not found'],
                  },
                },
              },
            ]),
          );
        }),
      );

      render(<TestBed />);

      expect(
        await screen.findByText(
          `Unable to locate this resource, are you sure it exists?`,
        ),
      ).toBeInTheDocument();
    });
  });
});
