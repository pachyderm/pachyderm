import {useContext} from 'react';

import {LoggedInContext} from '@dash-frontend/providers/LoggedInProvider';

const useLoggedIn = () => {
  return useContext(LoggedInContext);
};

export default useLoggedIn;
