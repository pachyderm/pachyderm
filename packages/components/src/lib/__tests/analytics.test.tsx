/* eslint-disable @typescript-eslint/naming-convention */
/* eslint-disable testing-library/consistent-data-testid */
import {captureException} from '@sentry/react';
import {render, waitFor} from '@testing-library/react';
import Cookies from 'js-cookie';
import React from 'react';
import {identify, page, track} from 'rudder-sdk-js';
import {mocked} from 'ts-jest/utils';

import {click} from 'testHelpers';

import {
  captureTrackingCookies,
  CLICK_TIMEOUT,
  getTrackingCookies,
  fireClick,
  fireIdentify,
  firePageView,
  firePromoApplied,
  fireUTM,
  initClickTracker,
  initPageTracker,
} from '../analytics';

jest.mock('rudder-sdk-js', () => ({
  getAnonymousId: jest.fn(() => 'mock-anonymous-id'),
  identify: jest.fn(),
  page: jest.fn(),
  track: jest.fn(),
}));

jest.mock('@sentry/react', () => ({
  captureException: jest.fn(),
}));

describe('lib/analytics', () => {
  beforeEach(() => {
    window.history.pushState({}, '', '');
    Cookies.remove('latest_utm_source');
    Cookies.remove('latest_utm_content');
    Cookies.remove('source_utm_source');
    Cookies.remove('source_utm_content');
  });

  it('should capture and get tracking cookies', () => {
    window.history.pushState({}, '', '?utm_source=google&utm_content=hello');
    captureTrackingCookies();

    expect(getTrackingCookies()).toEqual({
      latest_utm_source: 'google',
      latest_utm_content: 'hello',
      source_utm_source: 'google',
      source_utm_content: 'hello',
    });
  });

  it('should fire a click event', () => {
    fireClick('mock-click-id');

    expect(track).toHaveBeenCalledWith('click', {
      clickId: 'mock-click-id',
    });
  });

  it('should fire an identify event', () => {
    window.history.pushState({}, '', '?utm_source=shinra&utm_content=mako');
    captureTrackingCookies();

    fireIdentify(
      '7',
      'cloud@avalanche.org',
      new Date(1605719038392).getTime() / 1000,
    );

    expect(identify).toHaveBeenCalledWith('7', {
      anonymous_id: 'mock-anonymous-id',
      email: 'cloud@avalanche.org',
      hub_created_at: '2020-11-18T17:03:58.392Z',
      hub_user_id: '7',
      latest_utm_source: 'shinra',
      latest_utm_content: 'mako',
      source_utm_source: 'shinra',
      source_utm_content: 'mako',
    });
    expect(track).toHaveBeenCalledWith('authenticated', {
      anonymousId: 'mock-anonymous-id',
      authCreatedAt: '2020-11-18T17:03:58.392Z',
      authEmail: 'cloud@avalanche.org',
      authId: '7',
      latest_utm_source: 'shinra',
      latest_utm_content: 'mako',
      source_utm_source: 'shinra',
      source_utm_content: 'mako',
    });
  });

  it('should fire a page event', () => {
    firePageView();

    expect(page).toHaveBeenCalledTimes(1);
  });

  it('should fire a page event when the document title changes', async () => {
    document.title = 'Hello World 1';

    const observer = initPageTracker();

    expect(page).toHaveBeenCalledTimes(0);

    document.title = 'Hello World 2';
    await waitFor(() => {
      expect(page).toHaveBeenCalledTimes(1);
    });

    document.title = 'Hello World 3';
    await waitFor(() => {
      expect(page).toHaveBeenCalledTimes(2);
    });

    observer.disconnect();
  });

  it('should fire a promo applied event', () => {
    firePromoApplied('mock-promo', '1');

    expect(identify).toHaveBeenCalledWith('1', {
      hub_promo_code: 'mock-promo',
    });
    expect(track).toHaveBeenCalledWith('Promo', {
      context: {
        traits: {
          hub_promo_code: 'mock-promo',
        },
      },
    });
  });

  it('should fire a UTM event', () => {
    window.history.pushState({}, '', '?utm_source=google&utm_content=hello');
    captureTrackingCookies();

    fireUTM();

    expect(track).toHaveBeenCalledWith('UTM', {
      context: {
        traits: {
          latest_utm_source: 'google',
          latest_utm_content: 'hello',
        },
      },
    });
  });

  it('should track specified clicks', () => {
    const {getByTestId, getByText} = render(
      <>
        <button data-testid="Custom__contactUs">Contact Us</button>
        <button>Random</button>
      </>,
    );
    const contactButton = getByTestId('Custom__contactUs');
    const randomButton = getByText('Random');

    initClickTracker();

    click(randomButton);
    jest.advanceTimersByTime(CLICK_TIMEOUT);
    expect(track).toHaveBeenCalledTimes(0);

    click(contactButton);
    jest.advanceTimersByTime(CLICK_TIMEOUT);
    expect(track).toHaveBeenCalledTimes(1);

    click(contactButton);
    jest.advanceTimersByTime(CLICK_TIMEOUT);
    expect(track).toHaveBeenCalledTimes(2);
  });

  it('should send an event to sentry when a click event fails', () => {
    mocked(track).mockImplementationOnce(() => {
      throw new Error('Rudderstack exploded!');
    });

    fireClick('Button__click');
    expect(captureException).toHaveBeenCalledWith(
      '[Rudderstack Error]: Operation: track, Event: click, ID: Button__click, Error: Rudderstack exploded!',
    );
  });

  it('should send an event to sentry when an identify event fails', () => {
    mocked(track).mockImplementationOnce(() => {
      throw new Error('Rudderstack exploded!');
    });

    fireIdentify(
      '7',
      'cloud@avalanche.org',
      new Date(1605719038392).getTime() / 1000,
    );
    expect(captureException).toHaveBeenCalledWith(
      '[Rudderstack Error]: Operation: track, Event: identify, ID: 7, Error: Rudderstack exploded!',
    );
  });

  it('should send an event to sentry when a promo event fails', () => {
    mocked(track).mockImplementationOnce(() => {
      throw new Error('Rudderstack exploded!');
    });

    firePromoApplied('mock-promo', '1');
    expect(captureException).toHaveBeenCalledWith(
      '[Rudderstack Error]: Operation: track, Event: promo, ID: 1, Promo: mock-promo, Error: Rudderstack exploded!',
    );
  });

  it('should send an event to sentry when a UTM event fails', () => {
    mocked(track).mockImplementationOnce(() => {
      throw new Error('Rudderstack exploded!');
    });

    fireUTM();
    expect(captureException).toHaveBeenCalledWith(
      '[Rudderstack Error]: Operation: track, Event: UTM, Error: Rudderstack exploded!',
    );
  });
});
