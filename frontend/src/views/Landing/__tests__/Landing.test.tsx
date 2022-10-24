import {
  createServiceError,
  mockServer,
  status,
} from '@dash-backend/testHelpers';
import {render, waitFor, within} from '@testing-library/react';
import React from 'react';

import {withContextProviders, click, type} from '@dash-frontend/testHelpers';

import LandingComponent from '../Landing';

describe('Landing', () => {
  const Landing = withContextProviders(() => {
    return <LandingComponent />;
  });

  beforeEach(() => {
    window.localStorage.setItem(
      'pachyderm-console-account-data',
      '{"tutorial_introduction_seen": true}',
    );
  });

  afterEach(() => {
    window.localStorage.removeItem('pachyderm-console-account-data');
  });

  it('should display all projects with only a single tab', async () => {
    const {findAllByRole, findByRole, getAllByRole} = render(<Landing />);

    expect(
      await findByRole('heading', {name: 'Data Cleaning Process', level: 5}),
    ).toBeInTheDocument();

    expect(await findAllByRole('row', {})).toHaveLength(7);

    expect(
      await findByRole('tab', {
        name: /projects/i,
      }),
    ).toBeInTheDocument();

    expect(getAllByRole('tab')).toHaveLength(1);
  });

  it('should allow users to non-case sensitive search for projects by name', async () => {
    const {queryByRole, findByRole} = render(<Landing />);

    expect(
      await findByRole('heading', {name: 'Data Cleaning Process', level: 5}),
    ).toBeInTheDocument();
    expect(
      await findByRole('heading', {
        name: 'Solar Power Data Logger Team Collab',
        level: 5,
      }),
    ).toBeInTheDocument();

    const searchBox = await findByRole('searchbox');

    await type(searchBox, 'data CLeaning process');

    await waitFor(() =>
      expect(
        queryByRole('heading', {
          name: 'Solar Power Data Logger Team Collab',
          level: 5,
        }),
      ).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        queryByRole('heading', {name: 'Data Cleaning Process', level: 5}),
      ).toBeInTheDocument(),
    );
  });

  it('should display the project names', async () => {
    const {findByRole} = render(<Landing />);
    expect(
      await findByRole('heading', {
        name: 'Solar Power Data Logger Team Collab',
        level: 5,
      }),
    ).toBeInTheDocument();
  });

  it('should display the project status', async () => {
    const {findAllByTestId} = render(<Landing />);

    expect(await findAllByTestId('ProjectStatus__HEALTHY')).toHaveLength(5);
    expect(await findAllByTestId('ProjectStatus__UNHEALTHY')).toHaveLength(2);
  });

  it('should display project creation date in MM/DD/YYY format', async () => {
    const {findByText} = render(<Landing />);
    expect(await findByText('02/28/2021')).toBeInTheDocument();
  });

  it('should display project descriptions', async () => {
    const {findAllByText} = render(<Landing />);
    expect(
      await findAllByText(
        'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
      ),
    ).toHaveLength(6);
  });

  it('should allow a user to view a project based in default lineage view', async () => {
    const {findAllByRole} = render(<Landing />);

    expect(window.location.pathname).not.toEqual('/lineage/2');
    const viewProjectButtons = await findAllByRole('button', {
      name: 'View Project',
    });
    await click(viewProjectButtons[0]);
    expect(window.location.pathname).toEqual('/lineage/2');
  });

  it('should allow a user to view a project based on the preferred view', async () => {
    window.localStorage.setItem(
      'pachyderm-console-2',
      '{"list_view_default": true}',
    );

    const {findAllByRole} = render(<Landing />);

    expect(window.location.pathname).not.toEqual('/project/2/repos');
    const viewProjectButtons = await findAllByRole('button', {
      name: 'View Project',
    });
    await click(viewProjectButtons[0]);
    expect(window.location.pathname).toEqual('/project/2/repos');

    localStorage.removeItem('pachyderm-console-2');
  });

  it('should initially sort the projects by creation', async () => {
    const {findAllByTestId} = render(<Landing />);

    const projectCreations = await findAllByTestId('ProjectRow__created');

    expect(projectCreations[0].textContent).toEqual('02/28/2021');
    expect(projectCreations[6].textContent).toEqual('02/22/2021');
  });

  it('should allow the user to sort by name', async () => {
    const {findAllByRole, findByRole} = render(<Landing />);

    expect(
      await findByRole('heading', {name: 'Data Cleaning Process', level: 5}),
    ).toBeInTheDocument();
    const projectNames = await findAllByRole('heading', {
      level: 5,
    });

    expect(projectNames[1].textContent).toEqual('Data Cleaning Process');
    expect(projectNames[2].textContent).toEqual(
      'Solar Power Data Logger Team Collab',
    );
    expect(projectNames[3].textContent).toEqual('Solar Price Prediction Modal');
    expect(projectNames[4].textContent).toEqual('Egress Examples');
    expect(projectNames[5].textContent).toEqual('Empty Project');
    expect(projectNames[6].textContent).toEqual('Trait Discovery');

    expect(projectNames[7].textContent).toEqual('Solar Panel Data Sorting');

    const sortDropdown = await findByRole('button', {
      name: 'Sort by: Newest',
    });
    await click(sortDropdown);
    const nameSort = await findByRole('menuitem', {name: 'Name A-Z'});
    await click(nameSort);

    const nameSortedProjectNames = await findAllByRole('heading', {level: 5});

    expect(nameSortedProjectNames[1].textContent).toEqual(
      'Data Cleaning Process',
    );
    expect(nameSortedProjectNames[2].textContent).toEqual('Egress Examples');
    expect(nameSortedProjectNames[3].textContent).toEqual('Empty Project');
    expect(nameSortedProjectNames[4].textContent).toEqual(
      'Solar Panel Data Sorting',
    );
    expect(nameSortedProjectNames[5].textContent).toEqual(
      'Solar Power Data Logger Team Collab',
    );
    expect(nameSortedProjectNames[6].textContent).toEqual(
      'Solar Price Prediction Modal',
    );
    expect(nameSortedProjectNames[7].textContent).toEqual('Trait Discovery');
  });

  it('should allow the user to filter projects by status', async () => {
    const {findByRole, findByTestId, findByLabelText} = render(<Landing />);
    const projects = await findByTestId('Landing__view');
    expect(
      await within(projects).findAllByTestId('ProjectStatus__HEALTHY'),
    ).toHaveLength(5);
    expect(
      await within(projects).findAllByTestId('ProjectStatus__UNHEALTHY'),
    ).toHaveLength(2);

    const filterDropdown = await findByRole('button', {name: 'Show: All'});
    await click(filterDropdown);
    const healthyButton = await findByLabelText('Healthy');
    await click(healthyButton);

    expect(
      await findByRole('button', {name: 'Show: Unhealthy'}),
    ).toBeInTheDocument();

    expect(
      await within(projects).findAllByTestId('ProjectStatus__UNHEALTHY'),
    ).toHaveLength(2);
    expect(
      within(projects).queryByTestId('ProjectStatus__HEALTHY'),
    ).not.toBeInTheDocument();

    const unHealthyButton = await findByLabelText('Unhealthy');
    await click(unHealthyButton);

    expect(
      await findByRole('button', {name: 'Show: None'}),
    ).toBeInTheDocument();

    expect(
      within(projects).queryByTestId('ProjectStatus__HEALTHY'),
    ).not.toBeInTheDocument();
    expect(
      within(projects).queryByTestId('ProjectStatus__UNHEALTHY'),
    ).not.toBeInTheDocument();

    await click(healthyButton);

    expect(
      await findByRole('button', {name: 'Show: Healthy'}),
    ).toBeInTheDocument();

    expect(
      await within(projects).findAllByTestId('ProjectStatus__HEALTHY'),
    ).toHaveLength(5);
    expect(
      within(projects).queryByTestId('ProjectStatus__UNHEALTHY'),
    ).not.toBeInTheDocument();
  });

  it('should display an all tab when viewing just the default project', async () => {
    const error = createServiceError({code: status.UNIMPLEMENTED});
    mockServer.setError(error);

    const {findByRole} = render(<Landing />);

    expect(
      await findByRole('tab', {
        name: /projects/i,
      }),
    ).toBeInTheDocument();
  });

  it('should display the project details', async () => {
    const error = createServiceError({code: status.UNIMPLEMENTED});
    mockServer.setError(error);

    const {findByText} = render(<Landing />);

    expect(await findByText('17/10')).toBeInTheDocument();
    expect(await findByText('3 kB')).toBeInTheDocument();
  });

  it('should not display the project details when the project is empty', async () => {
    const {findByText} = render(<Landing />);
    await click(await findByText('Empty Project'));

    expect(
      await findByText('Create your first repo/pipeline!'),
    ).toBeInTheDocument();
  });
});
