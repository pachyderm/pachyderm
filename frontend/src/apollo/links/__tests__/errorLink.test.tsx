import {createServiceError} from '@dash-backend/testHelpers';
import {render, waitFor} from '@testing-library/react';
import Cookies from 'js-cookie';
import React from 'react';

import {useDAGData} from '@dash-frontend/hooks/useDAGData';
import useProject from '@dash-frontend/hooks/useProject';
import {DagDirection} from '@dash-frontend/lib/types';
import {mockServer, withContextProviders} from '@dash-frontend/testHelpers';

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
      const {findByText} = render(<TestBed />);

      expect(
        await findByText(
          `Unable to locate this resource, are you sure it exists?`,
        ),
      ).toBeInTheDocument();
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

    it('should show an error page if the project query is unsuccessful', async () => {
      mockServer.setError(createServiceError({code: 13}));
      const {findByText} = render(<TestBed />);

      expect(
        await findByText(`Looks like this API call can't be completed.`),
      ).toBeInTheDocument();
    });
  });
});
