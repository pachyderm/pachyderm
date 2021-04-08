import {render, waitFor} from '@testing-library/react';
import Cookies from 'js-cookie';
import React from 'react';

import {useDAGData} from '@dash-frontend/hooks/useDAGData';
import {withContextProviders} from '@dash-frontend/testHelpers';

const windowLocation = window.location;

describe('errorLink', () => {
  const TestBed = withContextProviders(() => {
    useDAGData({projectId: '1', nodeHeight: 60, nodeWidth: 120});

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
