import {render, screen} from '@testing-library/react';
import React from 'react';

import {click} from '@dash-frontend/testHelpers';

import {Switch} from '../';

describe('Switch', () => {
  it('should toggle the input value when clicked', async () => {
    render(<Switch name="switchTest" />);
    const button = screen.getByTestId('Switch__button');
    const input = screen.getByTestId('Switch__buttonInput');

    expect(input).not.toBeChecked();
    await click(button);
    expect(input).toBeChecked();
  });
});
