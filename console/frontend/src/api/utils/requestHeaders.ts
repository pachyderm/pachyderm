export const getHeaders = () => {
  const authToken = window.localStorage.getItem('auth-token') || '';

  return {
    'Content-Type': 'application/json',
    'authn-token': authToken,
  };
};

export const getRequestOptions = () => {
  return {
    pathPrefix: '/api',
    headers: getHeaders(),
  };
};
