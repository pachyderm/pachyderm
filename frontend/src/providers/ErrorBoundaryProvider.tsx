import React, {Component, ReactNode} from 'react';

import ErrorView from '@dash-frontend/views/ErrorView';

interface Props {
  children?: ReactNode;
}

interface State {
  error?: Error;
}

class ErrorBoundaryProvider extends Component<Props, State> {
  state: State = {
    error: undefined,
  };

  static getDerivedStateFromError(error: Error): State {
    return {error};
  }

  render() {
    const {error} = this.state;

    if (error) {
      return (
        <ErrorView
          errorMessage={`Uncaught error: ${error.message}`}
          source={error.name}
          stackTrace={error.stack}
        />
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundaryProvider;
