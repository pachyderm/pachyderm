import {render, waitFor} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import ProjectSideNav from '../../ProjectSideNav';

describe('project sidenav', () => {
  const Header = withContextProviders(ProjectSideNav);

  it('should display notification badge if the project has unhealthy jobs', async () => {
    window.history.replaceState('', '', '/project/2');

    const {findByLabelText} = render(<Header />);

    expect(await findByLabelText('Number of failed jobs')).toHaveTextContent(
      '1',
    );
  });

  it('should not display notification badge for projects with no unhealthy jobs', async () => {
    window.history.replaceState('', '', '/project/3');

    const {queryByLabelText, queryByTestId} = render(<Header />);

    await waitFor(() =>
      expect(queryByTestId('ProjectHeader__projectNameLoader')).toBeNull(),
    );
    expect(queryByLabelText('Number of failed')).not.toBeInTheDocument();
  });
});
