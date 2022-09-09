import {render, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {act} from 'react-dom/test-utils';

import {withContextProviders} from 'testHelpers';

import {useNotificationBanner} from '../';

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

const WrappedTestComponent = withContextProviders(TestComponent);

describe('NotificationBanner', () => {
  it('should show a notification banner and default to removing it after 3 seconds', () => {
    const {getByText, queryByText} = render(<WrappedTestComponent />);

    const bannerButton = getByText('Create Banner');

    userEvent.click(bannerButton);
    let banner = queryByText('Test Banner');

    expect(banner).not.toBeNull();

    act(() => {
      jest.advanceTimersByTime(3000);
    });

    banner = queryByText('Test Banner');

    expect(banner).toBeNull();
  });

  it('should show the notification banner for the specified duration', () => {
    const {getByText, queryByText} = render(
      <WrappedTestComponent type="success" duration={2000} />,
    );

    const bannerButton = getByText('Create Banner');

    userEvent.click(bannerButton);
    let banner = queryByText('Test Banner');

    expect(banner).not.toBeNull();

    act(() => {
      jest.advanceTimersByTime(2001);
    });

    banner = queryByText('Test Banner');

    expect(banner).toBeNull();
  });

  it('should be able to show a success banner', async () => {
    const {getByText, queryByTestId} = render(
      <WrappedTestComponent type="success" />,
    );

    const bannerButton = getByText('Create Banner');

    userEvent.click(bannerButton);
    await waitFor(() =>
      expect(queryByTestId('NotificationBanner__checkmark')).not.toBeNull(),
    );
  });

  it('should be able to show an error banner', async () => {
    const {getByText, queryByTestId} = render(
      <WrappedTestComponent type="error" />,
    );

    const bannerButton = getByText('Create Banner');

    userEvent.click(bannerButton);
    await waitFor(() =>
      expect(queryByTestId('NotificationBanner__error')).not.toBeNull(),
    );
  });
});
