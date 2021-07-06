import {ExtractRouteParams, generatePath as rrGeneratePath} from 'react-router';

const generatePathWithSearch = <S extends string>(
  pathTemplate: string,
  params?: ExtractRouteParams<S>,
) => {
  const path = encodeURI(rrGeneratePath(pathTemplate, {...params}));
  return `${path}${window.location.search}`;
};

export default generatePathWithSearch;
