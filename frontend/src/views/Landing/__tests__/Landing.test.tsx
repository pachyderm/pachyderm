import {render, waitFor, within, screen} from '@testing-library/react';
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
    render(<Landing />);

    expect(
      await screen.findByRole('heading', {
        name: 'Data-Cleaning-Process',
        level: 5,
      }),
    ).toBeInTheDocument();

    expect(await screen.findAllByRole('row', {})).toHaveLength(9);

    expect(
      await screen.findByRole('tab', {
        name: /projects/i,
      }),
    ).toBeInTheDocument();

    expect(screen.getByRole('tab')).toBeInTheDocument();
  });

  it('should allow users to non-case sensitive search for projects by name', async () => {
    render(<Landing />);

    expect(
      await screen.findByRole('heading', {
        name: 'Data-Cleaning-Process',
        level: 5,
      }),
    ).toBeInTheDocument();
    expect(
      await screen.findByRole('heading', {
        name: 'Solar-Power-Data-Logger-Team-Collab',
        level: 5,
      }),
    ).toBeInTheDocument();

    const searchBox = await screen.findByRole('searchbox');

    await type(searchBox, 'Data-Cleaning-Process');

    await waitFor(() =>
      expect(
        screen.queryByRole('heading', {
          name: 'Solar-Power-Data-Logger-Team-Collab',
          level: 5,
        }),
      ).not.toBeInTheDocument(),
    );
    await screen.findByRole('heading', {
      name: 'Data-Cleaning-Process',
      level: 5,
    });
  });

  it('should display the project names', async () => {
    render(<Landing />);
    expect(
      await screen.findByRole('heading', {
        name: 'Solar-Power-Data-Logger-Team-Collab',
        level: 5,
      }),
    ).toBeInTheDocument();
  });

  it('should display the project status', async () => {
    render(<Landing />);

    expect(await screen.findAllByTestId('ProjectStatus__HEALTHY')).toHaveLength(
      5,
    );
    expect(
      await screen.findAllByTestId('ProjectStatus__UNHEALTHY'),
    ).toHaveLength(4);
  });

  it('should display project descriptions', async () => {
    render(<Landing />);
    expect(
      await screen.findAllByText(
        'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
      ),
    ).toHaveLength(6);
  });

  it('should allow a user to view a project based in default lineage view', async () => {
    render(<Landing />);

    expect(window.location.pathname).not.toBe('/lineage/Data-Cleaning-Process');

    const cell = await screen.findByRole('cell', {
      name: /Data-Cleaning-Process view project project status description/i,
    });

    const viewProjectButton = within(cell).getByRole('button', {
      name: /view project/i,
    });
    await click(viewProjectButton);
    expect(window.location.pathname).toBe('/lineage/Data-Cleaning-Process');
  });

  it('should allow the user to sort by name', async () => {
    render(<Landing />);

    expect(
      await screen.findByRole('heading', {
        name: 'Data-Cleaning-Process',
        level: 5,
      }),
    ).toBeInTheDocument();

    const projectsPanel = screen.getByRole('tabpanel', {
      name: /projects 9/i,
    });
    const projectNamesAZ = within(projectsPanel).getAllByRole('heading', {
      level: 5,
    });

    // Starts in A-Z order
    expect(
      projectNamesAZ.map((projectName) => projectName.textContent),
    ).toEqual([
      'Data-Cleaning-Process',
      'Egress-Examples',
      'Empty-Project',
      'Multi-Project-Pipeline-A',
      'Multi-Project-Pipeline-B',
      'Solar-Panel-Data-Sorting',
      'Solar-Power-Data-Logger-Team-Collab',
      'Solar-Price-Prediction-Modal',
      'Trait-Discovery',
    ]);

    const sortDropdown = await screen.findByRole('button', {
      name: 'Sort by: Name A-Z',
    });
    await click(sortDropdown);
    const nameSort = await screen.findByRole('menuitem', {name: 'Name Z-A'});
    await click(nameSort);

    const projectNamesZA = within(projectsPanel).getAllByRole('heading', {
      level: 5,
    });
    expect(
      projectNamesZA.map((projectName) => projectName.textContent),
    ).toEqual([
      'Trait-Discovery',
      'Solar-Price-Prediction-Modal',
      'Solar-Power-Data-Logger-Team-Collab',
      'Solar-Panel-Data-Sorting',
      'Multi-Project-Pipeline-B',
      'Multi-Project-Pipeline-A',
      'Empty-Project',
      'Egress-Examples',
      'Data-Cleaning-Process',
    ]);
  });

  it('should allow the user to filter projects by status', async () => {
    render(<Landing />);
    const projects = await screen.findByTestId('Landing__view');
    expect(
      await within(projects).findAllByTestId('ProjectStatus__HEALTHY'),
    ).toHaveLength(5);
    expect(
      await within(projects).findAllByTestId('ProjectStatus__UNHEALTHY'),
    ).toHaveLength(4);

    const filterDropdown = await screen.findByRole('button', {
      name: 'Show: All',
    });
    await click(filterDropdown);
    const healthyButton = await screen.findByLabelText('Healthy');
    await click(healthyButton);

    expect(
      await screen.findByRole('button', {name: 'Show: Unhealthy'}),
    ).toBeInTheDocument();

    expect(
      await within(projects).findAllByTestId('ProjectStatus__UNHEALTHY'),
    ).toHaveLength(4);
    expect(
      within(projects).queryByTestId('ProjectStatus__HEALTHY'),
    ).not.toBeInTheDocument();

    const unHealthyButton = await screen.findByLabelText('Unhealthy');
    await click(unHealthyButton);

    expect(
      await screen.findByRole('button', {name: 'Show: None'}),
    ).toBeInTheDocument();

    expect(
      within(projects).queryByTestId('ProjectStatus__HEALTHY'),
    ).not.toBeInTheDocument();
    expect(
      within(projects).queryByTestId('ProjectStatus__UNHEALTHY'),
    ).not.toBeInTheDocument();

    await click(healthyButton);

    expect(
      await screen.findByRole('button', {name: 'Show: Healthy'}),
    ).toBeInTheDocument();

    expect(
      await within(projects).findAllByTestId('ProjectStatus__HEALTHY'),
    ).toHaveLength(5);
    expect(
      within(projects).queryByTestId('ProjectStatus__UNHEALTHY'),
    ).not.toBeInTheDocument();
  });

  it('should display the project details', async () => {
    render(<Landing />);
    click(await screen.findByText('Solar-Panel-Data-Sorting'));

    expect(await screen.findByText('3/2')).toBeInTheDocument();
    expect(await screen.findByText('3 kB')).toBeInTheDocument();
  });

  it('should not display the project details when the project is empty', async () => {
    render(<Landing />);
    await click(await screen.findByText('Empty-Project'));

    expect(
      await screen.findByText('Create your first repo/pipeline!'),
    ).toBeInTheDocument();
  });
});
