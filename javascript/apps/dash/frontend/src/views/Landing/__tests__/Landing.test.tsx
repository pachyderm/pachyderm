import {render, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import {
  createServiceError,
  mockServer,
  status,
} from '@dash-backend/testHelpers';
import {withContextProviders} from '@dash-frontend/testHelpers';

import LandingComponent from '../Landing';

describe('Landing', () => {
  const Landing = withContextProviders(() => {
    return <LandingComponent />;
  });

  it('should display projects', async () => {
    const {findAllByRole} = render(<Landing />);

    expect(await findAllByRole('row')).toHaveLength(5);
  });

  it('should allow users to search for projects by name', async () => {
    const {queryByRole, findByRole} = render(<Landing />);

    expect(
      await findByRole('heading', {name: 'Data Cleaning Process', level: 3}),
    ).toBeInTheDocument();
    expect(
      await findByRole('heading', {
        name: 'Solar Power Data Logger Team Collab',
        level: 3,
      }),
    ).toBeInTheDocument();

    const searchBox = await findByRole('searchbox');

    await userEvent.type(searchBox, 'Data Cleaning Process');

    await waitFor(() =>
      expect(
        queryByRole('heading', {
          name: 'Solar Power Data Logger Team Collab',
          level: 3,
        }),
      ).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        queryByRole('heading', {name: 'Data Cleaning Process', level: 3}),
      ).toBeInTheDocument(),
    );
  });

  it('should display the project names', async () => {
    const {findByRole} = render(<Landing />);
    expect(
      await findByRole('heading', {
        name: 'Solar Power Data Logger Team Collab',
        level: 3,
      }),
    ).toBeInTheDocument();
  });

  it('should display the project status', async () => {
    const {findAllByText} = render(<Landing />);

    // one extra due to filter dropdown
    expect(await findAllByText('Healthy')).toHaveLength(4);
    expect(await findAllByText('Unhealthy')).toHaveLength(3);
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
    ).toHaveLength(5);
  });

  it('should allow a user to view a project', async () => {
    const {findAllByRole} = render(<Landing />);

    expect(window.location.pathname).not.toEqual('/project/2');
    const viewProjectButtons = await findAllByRole('button', {
      name: 'View Project',
    });
    userEvent.click(viewProjectButtons[0]);
    expect(window.location.pathname).toEqual('/project/2');
  });

  it('should initially sort the projects by creation', async () => {
    const {findAllByTestId} = render(<Landing />);

    const projectCreations = await findAllByTestId('ProjectRow__created');

    expect(projectCreations[0].textContent).toEqual('02/28/2021');
    expect(projectCreations[4].textContent).toEqual('02/22/2021');
  });

  it('should allow the user to sort by name', async () => {
    const {findAllByRole, findByRole} = render(<Landing />);

    const projectNames = await findAllByRole('heading', {level: 3});

    expect(projectNames[0].textContent).toEqual('Data Cleaning Process');
    expect(projectNames[1].textContent).toEqual(
      'Solar Power Data Logger Team Collab',
    );
    expect(projectNames[2].textContent).toEqual('Solar Price Prediction Modal');
    expect(projectNames[3].textContent).toEqual('Solar Industry Analysis 2020');
    expect(projectNames[4].textContent).toEqual('Solar Panel Data Sorting');

    const sortDropdown = await findByRole('button', {
      name: 'Sort by: Created On',
    });
    userEvent.click(sortDropdown);
    const nameSort = await findByRole('menuitem', {name: 'Name A-Z'});
    userEvent.click(nameSort);

    const nameSortedProjectNames = await findAllByRole('heading', {level: 3});

    expect(nameSortedProjectNames[0].textContent).toEqual(
      'Data Cleaning Process',
    );
    expect(nameSortedProjectNames[1].textContent).toEqual(
      'Solar Industry Analysis 2020',
    );
    expect(nameSortedProjectNames[2].textContent).toEqual(
      'Solar Panel Data Sorting',
    );
    expect(nameSortedProjectNames[3].textContent).toEqual(
      'Solar Power Data Logger Team Collab',
    );
    expect(nameSortedProjectNames[4].textContent).toEqual(
      'Solar Price Prediction Modal',
    );
  });

  it('should allow the user to filter projects by status', async () => {
    const {findAllByText, findByRole, findByLabelText} = render(<Landing />);

    expect(await findAllByText('Healthy')).toHaveLength(4);
    expect(await findAllByText('Unhealthy')).toHaveLength(3);

    const filterDropdown = await findByRole('button', {name: 'Show: All'});
    userEvent.click(filterDropdown);
    const healthyButton = await findByLabelText('Healthy');
    userEvent.click(healthyButton);

    expect(
      await findByRole('button', {name: 'Show: Unhealthy'}),
    ).toBeInTheDocument();

    expect(await findAllByText('Healthy')).toHaveLength(1);
    expect(await findAllByText('Unhealthy')).toHaveLength(3);

    const unHealthyButton = await findByLabelText('Unhealthy');

    userEvent.click(unHealthyButton);

    expect(
      await findByRole('button', {name: 'Show: None'}),
    ).toBeInTheDocument();

    expect(await findAllByText('Healthy')).toHaveLength(1);
    expect(await findAllByText('Unhealthy')).toHaveLength(1);

    userEvent.click(healthyButton);

    expect(
      await findByRole('button', {name: 'Show: Healthy'}),
    ).toBeInTheDocument();

    expect(await findAllByText('Healthy')).toHaveLength(4);
    expect(await findAllByText('Unhealthy')).toHaveLength(1);
  });

  it('should display an all tab when viewing just the default project', async () => {
    const error = createServiceError({code: status.UNIMPLEMENTED});
    mockServer.setProjectsError(error);

    const {findByText} = render(<Landing />);

    expect(await findByText('All')).toBeInTheDocument();
  });

  it('should display the project details', async () => {
    const error = createServiceError({code: status.UNIMPLEMENTED});
    mockServer.setProjectsError(error);

    const {findByLabelText} = render(<Landing />);

    const repoPipelineCount = await findByLabelText(
      'Total No. of Repos/Pipelines',
    );
    const dataSize = await findByLabelText('Total Data Size');
    const pipelineStatus = await findByLabelText('Pipeline Status');

    expect(repoPipelineCount.textContent).toEqual('17/17');
    expect(dataSize.textContent).toEqual('2.93 KB');
    expect(pipelineStatus.textContent).toEqual('UnhealthyInspect');
  });
});
