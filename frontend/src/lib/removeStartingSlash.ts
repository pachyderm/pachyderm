const removeStartingSlash = (path = '') => {
  if (!path || path === '/') return '/';

  return path.replace(/^\//, '');
};

export default removeStartingSlash;
