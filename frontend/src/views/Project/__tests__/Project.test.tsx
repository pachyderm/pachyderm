import {render, waitFor} from '@testing-library/react';
import React from 'react';

import {LETS_START_TITLE} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import {withContextProviders} from '@dash-frontend/testHelpers';

import ProjectComponent from '../Project';

describe('Project', () => {
  const Project = withContextProviders(() => {
    return <ProjectComponent />;
  });

  it('should display an empty state prompt when no repos are found', async () => {
    window.history.replaceState({}, '', '/project/6/repos');

    const {findByText, queryByTestId} = render(<Project />);

    expect(await findByText(LETS_START_TITLE)).toBeInTheDocument();
    expect(queryByTestId('ListItem__row')).not.toBeInTheDocument();
  });

  it('should display repos', async () => {
    window.history.replaceState({}, '', '/project/2/repos');

    const {queryByTestId, findAllByTestId} = render(<Project />);

    await waitFor(() =>
      expect(queryByTestId('Title__name')).toHaveTextContent('samples'),
    );

    expect(await findAllByTestId('ListItem__row')).toHaveLength(14);
  });

  it('should display pipelines', async () => {
    window.history.replaceState({}, '', '/project/2/pipelines');

    const {queryByTestId, findAllByTestId} = render(<Project />);

    await waitFor(() =>
      expect(queryByTestId('Title__name')).toHaveTextContent('likelihoods'),
    );

    expect(await findAllByTestId('ListItem__row')).toHaveLength(8);
  });
});
