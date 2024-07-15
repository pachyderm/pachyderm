import {useContext} from 'react';

import notificationBannerContext from '../NotificationBannerContext';

const useNotificationBanner = () => {
  const {add, remove} = useContext(notificationBannerContext);

  return {add, remove};
};

export default useNotificationBanner;
