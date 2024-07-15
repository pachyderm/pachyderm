import * as sentryReact from '@sentry/react';
import {render, waitFor, screen} from '@testing-library/react';
import Cookies from 'js-cookie';
import React from 'react';

import {
  captureTrackingCookies,
  getTrackingCookies,
  fireClick,
  fireIdentify,
  firePageView,
  fireUTM,
  initClickTracker,
  initPageTracker,
  CLICK_TIMEOUT,
} from '../analytics';

jest.spyOn(sentryReact, 'captureException').mockImplementation(jest.fn());

describe('lib/analytics', () => {
  let identify = jest.fn(),
    page = jest.fn(),
    track = jest.fn();

  beforeEach(() => {
    window.history.pushState({}, '', '');
    Cookies.remove('latest_utm_source');
    Cookies.remove('latest_utm_content');
    Cookies.remove('source_utm_source');
    Cookies.remove('source_utm_content');
    identify = jest.fn();
    page = jest.fn();
    track = jest.fn();
  });

  afterEach(() => {
    jest.useRealTimers();
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
    fireClick('mock-click-id', track);

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
      identify,
      track,
    );

    expect(identify).toHaveBeenCalledWith('7', {
      email: 'cloud@avalanche.org',
      created_at: '2020-11-18T17:03:58.392Z',
      latest_utm_source: 'shinra',
      latest_utm_content: 'mako',
      source_utm_source: 'shinra',
      source_utm_content: 'mako',
    });
    expect(track).toHaveBeenCalledWith('authenticated', {
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
    firePageView(page);

    expect(page).toHaveBeenCalledTimes(1);
  });

  it('should fire a page event when the document title changes', async () => {
    document.title = 'Hello World 1';

    const observer = initPageTracker(page);

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

  it('should fire a UTM event', () => {
    window.history.pushState({}, '', '?utm_source=google&utm_content=hello');
    captureTrackingCookies();

    fireUTM(track);

    expect(track).toHaveBeenCalledWith('UTM', {
      context: {
        traits: {
          latest_utm_source: 'google',
          latest_utm_content: 'hello',
        },
      },
    });
  });

  it('should track specified clicks', async () => {
    jest.useFakeTimers();

    render(
      <>
        {/* eslint-disable-next-line testing-library/consistent-data-testid */}
        <button data-testid="Custom__contactUs">Contact Us</button>
        <button>Random</button>
      </>,
    );
    const contactButton = screen.getByTestId('Custom__contactUs');
    const randomButton = screen.getByText('Random');

    initClickTracker(track);

    randomButton.click();
    expect(track).toHaveBeenCalledTimes(0);

    jest.advanceTimersByTime(CLICK_TIMEOUT);
    contactButton.click();
    expect(track).toHaveBeenCalledTimes(1);

    jest.advanceTimersByTime(CLICK_TIMEOUT);
    contactButton.click();
    expect(track).toHaveBeenCalledTimes(2);
    jest.useRealTimers();
  });

  it('should send an event to sentry when a click event fails', () => {
    track.mockImplementationOnce(() => {
      throw new Error('Analytics exploded!');
    });

    fireClick('Button__click', track);
    expect(sentryReact.captureException).toHaveBeenCalledWith(
      '[Analytics Error]: Operation: track, Event: click, ID: Button__click, Error: Analytics exploded!',
    );
  });

  it('should send an event to sentry when an identify event fails', () => {
    track.mockImplementationOnce(() => {
      throw new Error('Analytics exploded!');
    });

    fireIdentify(
      '7',
      'cloud@avalanche.org',
      new Date(1605719038392).getTime() / 1000,
      identify,
      track,
    );
    expect(sentryReact.captureException).toHaveBeenCalledWith(
      '[Analytics Error]: Operation: track, Event: identify, ID: 7, Error: Analytics exploded!',
    );
  });

  it('should send an event to sentry when a UTM event fails', () => {
    track.mockImplementationOnce(() => {
      throw new Error('Analytics exploded!');
    });

    fireUTM(track);
    expect(sentryReact.captureException).toHaveBeenCalledWith(
      '[Analytics Error]: Operation: track, Event: UTM, Error: Analytics exploded!',
    );
  });
});
