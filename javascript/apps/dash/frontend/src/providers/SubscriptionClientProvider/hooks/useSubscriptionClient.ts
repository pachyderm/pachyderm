import {useContext} from 'react';

import SubscriptionClientContext from '../contexts/SubscriptionClientContext';

const useSubscriptionClient = () => {
  const {client} = useContext(SubscriptionClientContext);

  return {
    client,
  };
};

export default useSubscriptionClient;
