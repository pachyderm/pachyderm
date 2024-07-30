import Cookies from 'js-cookie';

const logout = () => {
  window.localStorage.removeItem('auth-token');
  window.localStorage.removeItem('id-token');
  Cookies.remove('dashAuthToken');
  Cookies.remove('dashAddress');
  window.location.href = '/';
};

export default logout;
