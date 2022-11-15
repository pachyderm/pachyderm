import noop from 'lodash/noop';
import React, {createContext} from 'react';

export interface NotificationBannerContextInterface {
  add: (
    content: React.ReactNode,
    type?: 'success' | 'error',
    duration?: number,
  ) => void;
  remove: (id: string) => void;
}

const NotificationBannerContext =
  createContext<NotificationBannerContextInterface>({
    add: noop,
    remove: noop,
  });

export default NotificationBannerContext;
