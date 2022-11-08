import {render} from '@testing-library/react';
import React from 'react';

import {click} from '@dash-frontend/testHelpers';

import {Switch} from '../';

describe('Switch', () => {
  it('should toggle the input value when clicked', async () => {
    const {getByTestId} = render(<Switch name="switchTest" />);
    const button = getByTestId('Switch__button');
    const input = getByTestId('Switch__buttonInput');

    expect(input).not.toBeChecked();
    await click(button);
    expect(input).toBeChecked();
  });
});
