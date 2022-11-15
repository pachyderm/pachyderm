import noop from 'lodash/noop';
import React from 'react';

import NotificationBanner from './NotificationBanner';

export default {title: 'Notification Banner'};

export const Success = () => {
  return (
    <NotificationBanner duration={1e9} type="success" remove={noop}>
      This is a notification!
    </NotificationBanner>
  );
};
export const Error = () => {
  return (
    <NotificationBanner duration={1e9} type="error" remove={noop}>
      This is a notification!
    </NotificationBanner>
  );
};
