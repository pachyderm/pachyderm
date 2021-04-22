import {createContext} from 'react';
import {SubscriptionClient} from 'subscriptions-transport-ws/dist/client';

export default createContext<{client: SubscriptionClient | null}>({
  client: null,
});
