import React, {useState} from 'react';

import LoggedInContext from './contexts/LoggedInContext';

const LoggedInProvider: React.FC = ({children}) => {
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
