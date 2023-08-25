import React, {useState} from 'react';

import LoggedInContext from './contexts/LoggedInContext';

const LoggedInProvider = ({children}: {children?: React.ReactNode}) => {
  const [loggedIn, setLoggedIn] = useState(
    Boolean(
      window.localStorage.getItem('auth-token') !== null &&
        window.localStorage.getItem('id-token') !== null,
    ),
  );

  const ctx = {
    loggedIn,
    setLoggedIn,
  };

  return (
    <LoggedInContext.Provider value={ctx}>{children}</LoggedInContext.Provider>
  );
};

export default LoggedInProvider;
