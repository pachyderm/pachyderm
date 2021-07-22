import {setContext} from '@apollo/client/link/context';

export const contextLink = () =>
  setContext((_, {headers}) => {
    const authToken = window.localStorage.getItem('auth-token') || '';
    const idToken = window.localStorage.getItem('id-token') || '';

    return {
      headers: {
        ...headers,
        'auth-token': authToken,
        'id-token': idToken,
      },
    };
  });
