import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import ProjectSidebar from '../ProjectSidebar';

describe('ProjectSidebar', () => {
  const Project = withContextProviders(ProjectSidebar);

  beforeEach(() => {
    window.history.replaceState('', '', '/');
  });

  it('should display job list', async () => {
    window.history.replaceState('', '', '/project/1/jobs');

    const {queryByTestId, findByTestId} = render(<Project />);

    expect(queryByTestId('JobListSkeleton__list')).toBeInTheDocument();
    expect(await findByTestId('JobList__project1')).toBeInTheDocument();
  });

  it('should display pipeline details', async () => {
    // TODO: update this with _actual_ pipeline details
    window.history.replaceState('', '', '/project/1/repo/1');

    const {findByText} = render(<Project />);

    expect(await findByText('TODO: Repo'));
  });

  it('should display repo details', async () => {
    // TODO: update this with _actual_ repo details
    window.history.replaceState('', '', '/project/1/pipeline/1');

    const {findByText} = render(<Project />);

    expect(await findByText('TODO: Pipeline'));
  });

  it('should not display the job list if not on jobs route', async () => {
    window.history.replaceState('', '', '/project/1');

    const {queryByTestId} = render(<Project />);

    expect(queryByTestId('JobListSkeleton__list')).not.toBeInTheDocument();
  });
});
