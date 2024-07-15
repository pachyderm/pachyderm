import {render, screen} from '@testing-library/react';
import React from 'react';

import {click, unhover} from '@dash-frontend/testHelpers';

import GlobalIdCopy from '../GlobalIdCopy';

describe('GlobalIdCopy', () => {
  it('should copy id on click', async () => {
    render(<GlobalIdCopy id="0918ac9d5daa76b86e3bb5e88e4c43a4" />);

    const copyButton = screen.getByRole('button', {name: 'Copy ID'});
    await click(copyButton);

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      '0918ac9d5daa76b86e3bb5e88e4c43a4',
    );

    expect(
      await screen.findByLabelText('ID copied successfully'),
    ).toBeInTheDocument();

    // Moving the mouse away hides the icon
    await unhover(screen.getByTestId('GlobalIdCopy__id'));
    expect(
      screen.queryByLabelText('ID copied successfully'),
    ).not.toBeInTheDocument();
  });

  it('should show short id', async () => {
    render(<GlobalIdCopy id="0918ac9d5daa76b86e3bb5e88e4c43a4" shortenId />);

    expect(await screen.findByText('0918ac...')).toBeInTheDocument();
  });
});
