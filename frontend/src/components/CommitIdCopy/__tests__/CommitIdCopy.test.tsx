import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import {click} from '@dash-frontend/testHelpers';

import CommitIdCopy from '../CommitIdCopy';

describe('CommitIdCopy', () => {
  it('should copy path on action click', async () => {
    render(
      <CommitIdCopy
        repo="test"
        branch="master"
        commit="0918ac9d5daa76b86e3bb5e88e4c43a4"
        clickable
      />,
    );

    const copyAction = await screen.findByTestId('CommitIdCopy_copy');
    await click(copyAction);

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      'test@master=0918ac9d',
    );

    expect(
      await screen.findByLabelText('You have successfully copied the id'),
    ).toBeInTheDocument();

    // Moving the mouse away hides the icon
    await userEvent.unhover(copyAction);
    expect(
      screen.queryByLabelText('You have successfully copied the id'),
    ).not.toBeInTheDocument();
  });

  it('should handle no repo or branch with a long commit', async () => {
    render(
      <CommitIdCopy longId small commit="0918ac9d5daa76b86e3bb5e88e4c43a4" />,
    );

    const copyAction = await screen.findByTestId('CommitIdCopy_copy');
    await click(copyAction);

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      '0918ac9d5daa76b86e3bb5e88e4c43a4',
    );
  });
});
