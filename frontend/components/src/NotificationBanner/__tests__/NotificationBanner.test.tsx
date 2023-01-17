/* eslint-disable @typescript-eslint/no-explicit-any */
import {render, waitFor, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React, {ReactElement} from 'react';

import {click} from '@dash-frontend/testHelpers';

import {useNotificationBanner, NotificationBannerProvider} from '../';

const TestComponent: React.FC<{
  duration?: number;
  type?: 'success' | 'error';
}> = ({duration, type}) => {
  const {add} = useNotificationBanner();

  return (
    <div id="main" onClick={() => add('Test Banner', type, duration)}>
      Create Banner
    </div>
  );
};

const withContextProviders = (
  Component: React.ElementType,
): ((props: any) => ReactElement) => {
  // eslint-disable-next-line react/display-name
  return (props: any): any => {
    return (
      <NotificationBannerProvider>
        <Component {...props} />
      </NotificationBannerProvider>
    );
  };
};

const WrappedTestComponent = withContextProviders(TestComponent);

describe('NotificationBanner', () => {
  it.skip('should show a notification banner and default to removing it after 3 seconds', async () => {
    jest.useFakeTimers('legacy');
    const user = userEvent.setup({
      advanceTimers: jest.advanceTimersByTime,
    });

    render(<WrappedTestComponent />);

    const bannerButton = screen.getByText('Create Banner');

    await user.click(bannerButton);

    const banner = await screen.findByText('Test Banner');
    expect(banner).toBeInTheDocument();

    await waitFor(() =>
      expect(screen.queryByText('Test Banner')).not.toBeInTheDocument(),
    );
    jest.useRealTimers();
  });

  it.skip('should show the notification banner for the specified duration', async () => {
    jest.useFakeTimers('legacy');
    const user = userEvent.setup({
      advanceTimers: jest.advanceTimersByTime,
    });
    render(<WrappedTestComponent type="success" duration={2000} />);

    const bannerButton = screen.getByText('Create Banner');

    await user.click(bannerButton);

    const banner = await screen.findByText('Test Banner');
    expect(banner).toBeInTheDocument();

    await waitFor(() =>
      expect(screen.queryByText('Test Banner')).not.toBeInTheDocument(),
    );
    jest.useRealTimers();
  });

  it('should be able to show a success banner', async () => {
    render(<WrappedTestComponent type="success" />);

    const bannerButton = screen.getByText('Create Banner');

    await click(bannerButton);
    expect(
      await screen.findByTestId('NotificationBanner__checkmark'),
    ).toBeInTheDocument();
  });

  it('should be able to show an error banner', async () => {
    render(<WrappedTestComponent type="error" />);

    const bannerButton = screen.getByText('Create Banner');

    await click(bannerButton);
    expect(
      await screen.findByTestId('NotificationBanner__error'),
    ).toBeInTheDocument();
  });
});
