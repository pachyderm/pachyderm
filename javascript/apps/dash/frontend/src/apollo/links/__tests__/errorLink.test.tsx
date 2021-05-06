import {render, waitFor} from '@testing-library/react';
import Cookies from 'js-cookie';
import React from 'react';

import {createServiceError} from '@dash-backend/testHelpers';
import {useDAGData} from '@dash-frontend/hooks/useDAGData';
import useProject from '@dash-frontend/hooks/useProject';
import {mockServer, withContextProviders} from '@dash-frontend/testHelpers';
import {DagDirection} from '@graphqlTypes';

const windowLocation = window.location;

describe('errorLink', () => {
  describe('unauthenticated', () => {
    const TestBed = withContextProviders(() => {
      useDAGData({
        projectId: '1',
        nodeHeight: 60,
        nodeWidth: 120,
        direction: DagDirection.RIGHT,
      });

      return <>test</>;
    });

    afterEach(() => {
      window.location = windowLocation;
      Cookies.remove('dashAuthToken');
      Cookies.remove('dashAddress');
      window.localStorage.removeItem('auth-token');
    });

    it('should log the user out if they receive an unauthenticated error', async () => {
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
      expect(hrefSpy).toBeCalledWith('/');
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
      render(<TestBed />);

      await waitFor(() => expect(window.location.pathname).toBe('/not-found'));
    });
  });

  describe('generic error', () => {
    const TestBed = withContextProviders(() => {
      useProject({id: 'bogus'});
      return <>test</>;
    });

    afterEach(() => {
      window.history.replaceState('', '', '/');
    });

    it('should redirect the user to /error if there is a service error', async () => {
      mockServer.setProjectsError(createServiceError({code: 13}));
      render(<TestBed />);

      await waitFor(() => expect(window.location.pathname).toBe('/error'));
    });
  });
});
