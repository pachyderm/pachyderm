import {setContext} from '@apollo/client/link/context';

export const contextLink = () =>
  setContext((_, {headers}) => {
    return {
      headers: {
        ...headers,
        // The below headers should eventually be pulled from localStorage
        'auth-token': process.env.REACT_APP_PACHD_AUTH,
        'pachd-address': process.env.REACT_APP_PACHD_ADDRESS,
      },
    };
  });
