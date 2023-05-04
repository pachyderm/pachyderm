import {render, screen} from '@testing-library/react';
import {GraphQLError} from 'graphql';
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

  it('should handle a GraphQLError stack trace', () => {
    const error = new GraphQLError('NotFoundError: projects x not found', {
      path: ['project'],
      extensions: {
        code: 'NOT_FOUND',
      },
    });
    render(<ErrorView stackTrace={error} />);

    expect(screen.getByText('NOT_FOUND', {exact: false})).toBeInTheDocument();
    expect(
      screen.getByText('NotFoundError: projects x not found', {exact: false}),
    ).toBeInTheDocument();
    expect(screen.getByText('extensions', {exact: false})).toBeInTheDocument();
    expect(screen.getByText('path', {exact: false})).toBeInTheDocument();
  });
});
