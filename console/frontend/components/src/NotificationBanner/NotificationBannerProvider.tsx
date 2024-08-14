import uniqueId from 'lodash/uniqueId';
import React, {useState, useMemo, useCallback} from 'react';
import {createPortal} from 'react-dom';

import NotificationBanner from './NotificationBanner';
import NotificationBannerContext from './NotificationBannerContext';

type Notification = {
  id: string;
  content: React.ReactNode;
  duration: number;
  type: 'success' | 'error';
};

const NotificationBannerProvider = ({
  children,
}: {
  children?: React.ReactNode;
}) => {
  const [notifications, setNotifications] = useState<Notification[]>([]);

  const add = useCallback(
    (
      content: React.ReactNode,
      type: Notification['type'] = 'success',
      duration = 6000,
    ) => {
      const id = uniqueId();
      setNotifications([...notifications, {id, content, duration, type}]);
    },
    [notifications],
  );

  const remove = useCallback(
    (id: string) => {
      setNotifications(notifications.filter((t) => t.id !== id));
    },
    [notifications],
  );

  const providerValue = useMemo(
    () => ({
      add,
      remove,
    }),
    [add, remove],
  );

  const renderNotificationBanners = useMemo(() => {
    const parentNode = document.getElementById('main');
    if (parentNode) {
      return createPortal(
        notifications.map((n) => {
          return (
            <NotificationBanner
              key={n.id}
              remove={() => remove(n.id)}
              duration={n.duration}
              type={n.type}
            >
              {n.content}
            </NotificationBanner>
          );
        }),
        parentNode,
      );
    }
    return null;
  }, [notifications, remove]);

  return (
    <NotificationBannerContext.Provider value={providerValue}>
      {renderNotificationBanners}
      {children}
    </NotificationBannerContext.Provider>
  );
};

export default NotificationBannerProvider;
