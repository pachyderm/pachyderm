/* eslint-disable @typescript-eslint/no-explicit-any */
import {render, waitFor} from '@testing-library/react';
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
  it('should show a notification banner and default to removing it after 3 seconds', async () => {
    const {getByText, queryByText} = render(<WrappedTestComponent />);

    const bannerButton = getByText('Create Banner');

    await click(bannerButton);
    const banner = queryByText('Test Banner');

    expect(banner).not.toBeNull();

    waitFor(() => expect(queryByText('Test Banner')).toBeNull());
  });

  it('should show the notification banner for the specified duration', async () => {
    const {getByText, queryByText} = render(
      <WrappedTestComponent type="success" duration={2000} />,
    );

    const bannerButton = getByText('Create Banner');

    await click(bannerButton);
    const banner = queryByText('Test Banner');

    expect(banner).not.toBeNull();
    waitFor(() => expect(queryByText('Test Banner')).toBeNull());
  });

  it('should be able to show a success banner', async () => {
    const {getByText, queryByTestId} = render(
      <WrappedTestComponent type="success" />,
    );

    const bannerButton = getByText('Create Banner');

    await click(bannerButton);
    await waitFor(() =>
      expect(queryByTestId('NotificationBanner__checkmark')).not.toBeNull(),
    );
  });

  it('should be able to show an error banner', async () => {
    const {getByText, queryByTestId} = render(
      <WrappedTestComponent type="error" />,
    );

    const bannerButton = getByText('Create Banner');

    await click(bannerButton);
    await waitFor(() =>
      expect(queryByTestId('NotificationBanner__error')).not.toBeNull(),
    );
  });
});
