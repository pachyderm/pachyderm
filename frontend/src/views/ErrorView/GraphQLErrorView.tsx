import {GraphQLError} from 'graphql';
import React from 'react';

import ErrorView from './ErrorView';
import useErrorView, {ErrorViewType} from './hooks/useErrorView';

type ErrorViewProps = {
  graphQLError?: GraphQLError;
};

const GraphQLErrorView: React.FC<ErrorViewProps> = ({graphQLError}) => {
  const {errorType, errorMessage, errorDetails} = useErrorView(graphQLError);

  return (
    <ErrorView
      errorMessage={errorMessage}
      errorDetails={errorDetails}
      source={graphQLError?.source?.name}
      stackTrace={graphQLError}
      showBackHomeButton={errorType !== ErrorViewType.UNAUTHENTICATED}
    />
  );
};

export default GraphQLErrorView;
