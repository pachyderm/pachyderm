import {useLocation} from 'react-router';

export enum ErrorViewType {
  NOT_FOUND = 'NOT_FOUND',
  UNAUTHENTICATED = 'UNAUTHENTICATED',
  GENERIC = 'GENERIC',
}

const pathToErrorViewType: {[key: string]: ErrorViewType} = {
  '/not-found': ErrorViewType.NOT_FOUND,
  '/unauthenticated': ErrorViewType.UNAUTHENTICATED,
};

const useErrorView = () => {
  const {pathname} = useLocation();

  const errorType = pathToErrorViewType[pathname] || ErrorViewType.GENERIC;

  return {errorType};
};

export default useErrorView;
