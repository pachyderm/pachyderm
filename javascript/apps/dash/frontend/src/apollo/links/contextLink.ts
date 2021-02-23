import {setContext} from '@apollo/client/link/context';

export const contextLink = () =>
  setContext((_, {headers}) => {
    const authToken = window.localStorage.getItem('auth-token') || '';

    return {
      headers: {
        ...headers,
        'auth-token': authToken,
        'pachd-address': process.env.REACT_APP_PACHD_ADDRESS,
      },
    };
  });
