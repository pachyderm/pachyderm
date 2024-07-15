import React from 'react';

import ErrorView from './ErrorView';
import useErrorView, {ErrorViewType} from './hooks/useErrorView';

type ErrorViewProps = {
  error?: unknown;
};

const AuthenticatedRouteErrorView: React.FC<ErrorViewProps> = ({error}) => {
  const {errorType, errorMessage, errorDetails} = useErrorView(error);

  return (
    <ErrorView
      errorMessage={errorMessage}
      errorDetails={errorDetails}
      showBackHomeButton={errorType !== ErrorViewType.UNAUTHENTICATED}
    />
  );
};

export default AuthenticatedRouteErrorView;
