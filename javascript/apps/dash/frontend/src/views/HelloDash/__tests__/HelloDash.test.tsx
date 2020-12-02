import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from 'testHelpers';

import HelloDashComponent from '../';

const HelloDash = withContextProviders(HelloDashComponent);

describe('views/HelloDash', () => {
  it('should show HelloDash', async () => {
    const {findByText} = render(<HelloDash />);
    const message = await findByText('Hello Dash!');

    expect(message).toBeInTheDocument();
  });
});
