import React from 'react';
import {SubscriptionClient} from 'subscriptions-transport-ws/dist/client';

import SubscriptionClientContext from '@dash-frontend/providers/SubscriptionClientProvider/contexts/SubscriptionClientContext';

type SubscriptionClientProviderProps = {
  client: SubscriptionClient;
};

const SubscriptionClientProvider: React.FC<SubscriptionClientProviderProps> = ({
  children,
  client,
}) => {
  return (
    <SubscriptionClientContext.Provider value={{client}}>
      {children}
    </SubscriptionClientContext.Provider>
  );
};

export default SubscriptionClientProvider;
