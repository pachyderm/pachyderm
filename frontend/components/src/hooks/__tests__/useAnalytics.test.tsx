import {render} from '@testing-library/react';
import React, {useEffect} from 'react';

import * as analytics from '../../lib/analytics';
import useAnalytics, {UseAnalyticsProps} from '../useAnalytics';

jest.spyOn(analytics, 'captureTrackingCookies').mockImplementation(jest.fn());
jest.spyOn(analytics, 'fireUTM').mockImplementation(jest.fn());
jest.spyOn(analytics, 'initClickTracker').mockImplementation(jest.fn());
jest.spyOn(analytics, 'initPageTracker').mockImplementation(jest.fn());
jest.spyOn(analytics, 'fireIdentify').mockImplementation(jest.fn());
jest.spyOn(analytics, 'fireClusterInfo').mockImplementation(jest.fn());

type AnalyticsComponentProps = Omit<UseAnalyticsProps, 'provider'>;

const AnalyticsComponent: React.FC<AnalyticsComponentProps> = ({
  createdAt,
  email,
  id,
  clusterId,
}) => {
  const provider = {
    getAnonymousId: jest.fn(),
    identify: jest.fn(),
    page: jest.fn(),
    track: jest.fn(),
  };
  const {init} = useAnalytics({createdAt, email, id, clusterId, provider});

  useEffect(() => {
    init();
  }, [init]);

  return null;
};

describe('hooks/useAnalytics', () => {
  beforeEach(() => {
    window.analyticsInitialized = false;
    window.analyticsIdentified = false;
    window.clusterIdentified = false;
  });

  it('should initialize only once', async () => {
    render(
      <AnalyticsComponent
        createdAt={1629823988932}
        email="cloud@avalanche.org"
        id="1"
        clusterId="12e345"
      />,
    );

    expect(analytics.captureTrackingCookies).toHaveBeenCalledTimes(1);
    expect(analytics.initPageTracker).toHaveBeenCalledTimes(1);
    expect(analytics.initClickTracker).toHaveBeenCalledTimes(1);
    expect(analytics.fireUTM).toHaveBeenCalledTimes(1);
    expect(analytics.fireIdentify).toHaveBeenCalledTimes(1);
    expect(analytics.fireClusterInfo).toHaveBeenCalledTimes(1);

    render(
      <AnalyticsComponent
        createdAt={1629823988932}
        email="barret@avalanche.org"
        id="2"
        clusterId="12345"
      />,
    );

    expect(analytics.captureTrackingCookies).toHaveBeenCalledTimes(1);
    expect(analytics.initPageTracker).toHaveBeenCalledTimes(1);
    expect(analytics.initClickTracker).toHaveBeenCalledTimes(1);
    expect(analytics.fireUTM).toHaveBeenCalledTimes(1);
    expect(analytics.fireIdentify).toHaveBeenCalledTimes(1);
    expect(analytics.fireClusterInfo).toHaveBeenCalledTimes(1);
  });

  it('should not fire identify when not given account info', () => {
    render(<AnalyticsComponent />);

    expect(analytics.captureTrackingCookies).toHaveBeenCalledTimes(1);
    expect(analytics.initPageTracker).toHaveBeenCalledTimes(1);
    expect(analytics.initClickTracker).toHaveBeenCalledTimes(1);
    expect(analytics.fireUTM).toHaveBeenCalledTimes(1);
    expect(analytics.fireIdentify).toHaveBeenCalledTimes(0);

    render(
      <AnalyticsComponent
        createdAt={1629823988932}
        email="cloud@avalanche.org"
        id="1"
      />,
    );

    expect(analytics.captureTrackingCookies).toHaveBeenCalledTimes(1);
    expect(analytics.initPageTracker).toHaveBeenCalledTimes(1);
    expect(analytics.initClickTracker).toHaveBeenCalledTimes(1);
    expect(analytics.fireUTM).toHaveBeenCalledTimes(1);
    expect(analytics.fireIdentify).toHaveBeenCalledTimes(1);
  });

  it('should not fire cluster info when not given clusterId', () => {
    render(<AnalyticsComponent />);

    expect(analytics.captureTrackingCookies).toHaveBeenCalledTimes(1);
    expect(analytics.initPageTracker).toHaveBeenCalledTimes(1);
    expect(analytics.initClickTracker).toHaveBeenCalledTimes(1);
    expect(analytics.fireUTM).toHaveBeenCalledTimes(1);
    expect(analytics.fireClusterInfo).toHaveBeenCalledTimes(0);

    render(<AnalyticsComponent clusterId="12323" />);

    expect(analytics.captureTrackingCookies).toHaveBeenCalledTimes(1);
    expect(analytics.initPageTracker).toHaveBeenCalledTimes(1);
    expect(analytics.initClickTracker).toHaveBeenCalledTimes(1);
    expect(analytics.fireUTM).toHaveBeenCalledTimes(1);
    expect(analytics.fireClusterInfo).toHaveBeenCalledTimes(1);
  });
});
