/* eslint-disable @typescript-eslint/no-explicit-any */
import {act} from '@testing-library/react';
import userEvent, {TargetElement, ITypeOpts} from '@testing-library/user-event';
import React, {ReactElement} from 'react';

import {NotificationBannerProvider} from './NotificationBanner';

export const click: typeof userEvent.click = (...args) => {
  act(() => userEvent.click(...args));
};

export const paste: typeof userEvent.paste = (...args) => {
  act(() => userEvent.paste(...args));
};

export const type = async (
  element: TargetElement,
  text: string,
  userOpts?: ITypeOpts,
) => {
  await act(async () => {
    await userEvent.type(element, text, userOpts);
  });
};

export const withContextProviders = (
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
