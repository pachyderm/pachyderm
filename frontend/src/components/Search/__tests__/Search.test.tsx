import {JobState} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import React from 'react';

import {
  click,
  getUrlState,
  withContextProviders,
  type,
} from '@dash-frontend/testHelpers';

import Search from '../Search';

describe('Search', () => {
  const renderTestbed = () => {
    const SearchComponent = withContextProviders(() => {
      return <Search />;
    });

    render(<SearchComponent />);
  };

  const assertDropdown = () => {
    const dropdown = screen.getByTestId('Search__dropdown');

    return {
      toBeShown: () => expect(dropdown).toHaveClass('open'),
      toBeHidden: () => expect(dropdown).not.toHaveClass('open'),
    };
  };

  beforeEach(() => {
    window.history.replaceState({}, '', '/project/Solar-Panel-Data-Sorting');
  });
  afterEach(() => {
    window.localStorage.removeItem(
      'pachyderm-console-Solar-Panel-Data-Sorting',
    );
  });

  it('should display empty state messages', async () => {
    window.history.replaceState({}, '', '/project/Empty-Project');
    renderTestbed();

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    searchBar.focus();
    assertDropdown().toBeShown();
    expect(
      screen.getByText('There are no recent searches.'),
    ).toBeInTheDocument();
    expect(
      screen.getByText('There are no jobs on this project.'),
    ).toBeInTheDocument();
    await type(searchBar, 'anything');
    expect(
      await screen.findByText(
        'No matching pipelines, jobs or global ID found.',
      ),
    ).toBeInTheDocument();
    assertDropdown().toBeShown();
  });

  it('should underline matched search term', async () => {
    renderTestbed();

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, 'im');
    assertDropdown().toBeShown();
    expect(await screen.findByText('im')).toHaveClass('underline');
    expect(await screen.findByText('ages')).not.toHaveClass('underline');
  });

  it('should save recent searches and populate search bar', async () => {
    renderTestbed();

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, 'images');
    assertDropdown().toBeShown();
    const searchResult = await screen.findByRole('button', {
      name: 'See Commits',
    });
    await click(searchResult);
    assertDropdown().toBeHidden();

    searchBar.focus();
    assertDropdown().toBeShown();
    const savedSearch = await screen.findByRole('button', {name: 'images'});
    await click(savedSearch);
    expect(searchBar).toHaveValue('images');
    assertDropdown().toBeShown();
  });

  it('should clear saved search history', async () => {
    renderTestbed();
    expect(
      screen.getByText('There are no recent searches.'),
    ).toBeInTheDocument();

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, 'images');
    assertDropdown().toBeShown();
    await click(await screen.findByRole('button', {name: 'See Commits'}));
    assertDropdown().toBeHidden();

    searchBar.focus();
    assertDropdown().toBeShown();
    expect(
      screen.queryByText('There are no recent searches.'),
    ).not.toBeInTheDocument();
    await click(await screen.findByRole('button', {name: 'Clear'}));
    expect(
      screen.getByText('There are no recent searches.'),
    ).toBeInTheDocument();
    assertDropdown().toBeShown();
  });

  it('should route to selected repo', async () => {
    renderTestbed();

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, 'images');
    assertDropdown().toBeShown();
    const searchResult = await screen.findByRole('button', {
      name: 'See Commits',
    });
    await click(searchResult);

    assertDropdown().toBeHidden();
    expect(window.location.pathname).toBe(
      '/project/Solar-Panel-Data-Sorting/repos/images/branch/default',
    );
  });

  it('should route to selected pipeline', async () => {
    renderTestbed();
    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, 'edges');
    assertDropdown().toBeShown();
    await click((await screen.findAllByText('edges'))[1]);

    assertDropdown().toBeHidden();
    expect(window.location.pathname).toBe(
      '/project/Solar-Panel-Data-Sorting/pipelines/edges',
    );
    assertDropdown().toBeHidden();
  });

  it('should route to jobs for selected id', async () => {
    renderTestbed();

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, '23b9af7d5d4343219bc8e02ff44cd55a');
    assertDropdown().toBeShown();
    await click(await screen.findByText('23b9af7d5d4343219bc8e02ff44cd55a'));

    assertDropdown().toBeHidden();
    expect(window.location.pathname).toBe(
      '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a',
    );
  });

  it('should route to jobs view when chips are clicked', async () => {
    renderTestbed();

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    searchBar.focus();
    assertDropdown().toBeShown();
    await click(await screen.findByRole('button', {name: 'All (4)'}));
    expect(window.location.pathname).toBe(
      '/project/Solar-Panel-Data-Sorting/jobs',
    );
    expect(getUrlState().jobFilters).toEqual([
      JobState.JOB_CREATED,
      JobState.JOB_EGRESSING,
      JobState.JOB_FAILURE,
      JobState.JOB_FINISHING,
      JobState.JOB_KILLED,
      JobState.JOB_RUNNING,
      JobState.JOB_STARTING,
      JobState.JOB_STATE_UNKNOWN,
      JobState.JOB_SUCCESS,
      JobState.JOB_UNRUNNABLE,
    ]);
    assertDropdown().toBeHidden();
  });
});
