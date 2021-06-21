import {render} from '@testing-library/react';
import React from 'react';

import {Switch} from '../';
import {click} from '../../testHelpers';

describe('Switch', () => {
  it('should toggle the input value when clicked', () => {
    const {getByTestId} = render(<Switch name="switchTest" />);
    const button = getByTestId('Switch__button');
    const input = getByTestId('Switch__buttonInput');

    expect(input).not.toBeChecked();
    click(button);
    expect(input).toBeChecked();
  });
});
