import {render, screen} from '@testing-library/react';
import React from 'react';

import {click} from '@dash-frontend/testHelpers';

import ExpandableText from '../ExpandableText';

describe('ExpandableText', () => {
  it('should expand content on button click', async () => {
    Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
      configurable: true,
      value: 50, // larger than two lines
    });
    render(<ExpandableText text="description" />);

    expect(screen.getByText('description')).toHaveClass('minimized');

    await click(screen.getByRole('button', {name: /read more/i}));

    expect(screen.getByText('description')).not.toHaveClass('minimized');
  });

  it('should not show button for less than 2 lines', async () => {
    Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
      configurable: true,
      value: 40, // smaller than two lines
    });
    render(<ExpandableText text="description" />);

    expect(screen.getByText('description')).not.toHaveClass('minimized');
  });
});
