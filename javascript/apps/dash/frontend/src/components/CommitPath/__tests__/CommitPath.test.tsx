import {render} from '@testing-library/react';
import React from 'react';

import {click} from '@dash-frontend/testHelpers';

import CommitPath from '../CommitPath';

describe('CommitPath', () => {
  const Commit = (
    <CommitPath
      repo="test"
      branch="master"
      commit="0918ac9d5daa76b86e3bb5e88e4c43a4"
    />
  );

  afterEach(() => {
    window.history.pushState({}, document.title, '/');
  });

  it('should copy path on action click', async () => {
    const {findByTestId} = render(Commit);

    const copyAction = await findByTestId('CommitPath_copy');
    click(copyAction);

    expect(window.document.execCommand).toHaveBeenCalledWith('copy');
  });
});
