import {render, screen} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import ErrorViewComponent from '../ErrorView';

describe('ErrorView', () => {
  const ErrorView = withContextProviders(ErrorViewComponent);

  it('should handle a string stack trace', () => {
    const error = `
      ReferenceError: x is not defined
        at Landing (http://localhost:4001/src/views/Landing/Landing.tsx?t=1682536309934:57:3)
    `;

    render(<ErrorView stackTrace={error} />);

    expect(
      screen.getByText('ReferenceError: x is not defined'),
    ).toBeInTheDocument();
    expect(screen.getByText('at Landing', {exact: false})).toBeInTheDocument();
  });

  it('should show error message, details, source and stack trace', () => {
    render(
      <ErrorView
        errorMessage={'errorMessage'}
        errorDetails={'errorDetails'}
        source={'source'}
        stackTrace={'stackTrace'}
      />,
    );

    expect(screen.getByText('errorMessage')).toBeInTheDocument();
    expect(screen.getByText('errorDetails')).toBeInTheDocument();
    expect(screen.getByText('Source: source')).toBeInTheDocument();
    expect(screen.getByText('stackTrace')).toBeInTheDocument();
  });

  it('should handle object stack trace', () => {
    render(<ErrorView stackTrace={{foo: 'bar'}} />);

    expect(screen.getByText('"foo"')).toBeInTheDocument();
    expect(screen.getByText(':')).toBeInTheDocument();
    expect(screen.getByText('"bar"')).toBeInTheDocument();
  });

  it('should handle array stack trace', () => {
    render(<ErrorView stackTrace={['foo', 'bar']} />);

    expect(screen.getByText('"foo"')).toBeInTheDocument();
    expect(screen.getByText('"bar"')).toBeInTheDocument();
  });

  it('should show back button', () => {
    render(<ErrorView showBackHomeButton={true} />);

    expect(
      screen.getByRole('button', {name: 'Go Back Home'}),
    ).toBeInTheDocument();
  });

  it('should not show back button', () => {
    render(
      <ErrorView errorMessage={'errorMessage'} showBackHomeButton={false} />,
    );

    expect(screen.getByText('errorMessage')).toBeInTheDocument();
    expect(
      screen.queryByRole('button', {name: 'Go Back Home'}),
    ).not.toBeInTheDocument();
  });
});
