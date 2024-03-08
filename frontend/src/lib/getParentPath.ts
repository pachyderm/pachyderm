const getParentPath = (path?: string) => {
  if (!path) return '/';

  return path
    .replace(/\/$/, '') // Remove trailing slash
    .split('/')
    .slice(0, -1)
    .join('/')
    .concat('/');
};

export default getParentPath;
