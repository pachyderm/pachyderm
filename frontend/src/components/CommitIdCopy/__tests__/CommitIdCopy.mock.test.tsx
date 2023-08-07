import {render, screen} from '@testing-library/react';
import React from 'react';

import {click} from '@dash-frontend/testHelpers';

import CommitIdCopy from '../CommitIdCopy';

describe('CommitIdCopy', () => {
  const Commit = (
    <CommitIdCopy
      repo="test"
      branch="master"
      commit="0918ac9d5daa76b86e3bb5e88e4c43a4"
    />
  );

  afterEach(() => {
    window.history.pushState({}, document.title, '/');
  });

  it('should copy path on action click', async () => {
    render(Commit);

    const copyAction = await screen.findByTestId('CommitIdCopy_copy');
    await click(copyAction);

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      'test@master=0918ac9d',
    );
  });
});
