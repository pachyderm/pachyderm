import {render} from '@testing-library/react';
import React, {useEffect} from 'react';

import {
  captureTrackingCookies,
  fireIdentify,
  initPageTracker,
  initClickTracker,
  fireUTM,
} from '../../lib/analytics';
import useAnalytics, {UseAnalyticsProps} from '../useAnalytics';

jest.mock('../../lib/analytics', () => ({
  captureTrackingCookies: jest.fn(),
  fireUTM: jest.fn(),
  initClickTracker: jest.fn(),
  initPageTracker: jest.fn(),
  fireIdentify: jest.fn(),
}));

type AnalyticsComponentProps = Omit<UseAnalyticsProps, 'provider'>;

const AnalyticsComponent: React.FC<AnalyticsComponentProps> = ({
  createdAt,
  email,
  id,
}) => {
  const provider = {
    getAnonymousId: jest.fn(),
    identify: jest.fn(),
    page: jest.fn(),
    track: jest.fn(),
  };
  const {init} = useAnalytics({createdAt, email, id, provider});

  useEffect(() => {
    init();
  }, [init]);

  return null;
};

describe('hooks/useAnalytics', () => {
  beforeEach(() => {
    window.analyticsInitialized = false;
    window.analyticsIdentified = false;
  });

  it('should initialize only once', () => {
    render(
      <AnalyticsComponent
        createdAt={1629823988932}
        email="cloud@avalanche.org"
        id="1"
      />,
    );

    expect(captureTrackingCookies).toHaveBeenCalledTimes(1);
    expect(initPageTracker).toHaveBeenCalledTimes(1);
    expect(initClickTracker).toHaveBeenCalledTimes(1);
    expect(fireUTM).toHaveBeenCalledTimes(1);
    expect(fireIdentify).toHaveBeenCalledTimes(1);

    render(
      <AnalyticsComponent
        createdAt={1629823988932}
        email="barret@avalanche.org"
        id="2"
      />,
    );

    expect(captureTrackingCookies).toHaveBeenCalledTimes(1);
    expect(initPageTracker).toHaveBeenCalledTimes(1);
    expect(initClickTracker).toHaveBeenCalledTimes(1);
    expect(fireUTM).toHaveBeenCalledTimes(1);
    expect(fireIdentify).toHaveBeenCalledTimes(1);
  });

  it('should not fire identify when not given account info', () => {
    render(<AnalyticsComponent />);

    expect(captureTrackingCookies).toHaveBeenCalledTimes(1);
    expect(initPageTracker).toHaveBeenCalledTimes(1);
    expect(initClickTracker).toHaveBeenCalledTimes(1);
    expect(fireUTM).toHaveBeenCalledTimes(1);
    expect(fireIdentify).toHaveBeenCalledTimes(0);

    render(
      <AnalyticsComponent
        createdAt={1629823988932}
        email="cloud@avalanche.org"
        id="1"
      />,
    );

    expect(captureTrackingCookies).toHaveBeenCalledTimes(1);
    expect(initPageTracker).toHaveBeenCalledTimes(1);
    expect(initClickTracker).toHaveBeenCalledTimes(1);
    expect(fireUTM).toHaveBeenCalledTimes(1);
    expect(fireIdentify).toHaveBeenCalledTimes(1);
  });
});
