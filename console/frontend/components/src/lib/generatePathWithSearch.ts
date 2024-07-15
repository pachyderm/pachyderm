import {generatePath as rrGeneratePath} from 'react-router-dom';

const generatePathWithSearch = (
  pathTemplate: string,
  params?: {
    [x: string]: string | number | boolean | undefined;
  },
) => {
  const path = rrGeneratePath(pathTemplate, {...params});
  return `${path}${window.location.search}`;
};

export default generatePathWithSearch;
