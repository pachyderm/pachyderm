const CHARSET =
  'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_';

const getRandomString = (stringLength: number) => {
  const data = new Uint8Array(stringLength);

  window.crypto.getRandomValues(data);

  return Array.from(data, (x) => CHARSET[x % CHARSET.length]).join('');
};

export default getRandomString;
