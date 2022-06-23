import {JobState} from '@graphqlTypes';
import {render} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import {
  click,
  getUrlState,
  withContextProviders,
} from '@dash-frontend/testHelpers';

import Search from '../Search';

describe('Search', () => {
  const renderTestBed = () => {
    const SearchComponent = withContextProviders(() => {
      return <Search />;
    });

    const testUtils = render(<SearchComponent />);

    const assertDropdown = () => {
      const dropdown = testUtils.getByTestId('Search__dropdown');

      return {
        toBeShown: () => expect(dropdown).toHaveClass('open'),
        toBeHidden: () => expect(dropdown).not.toHaveClass('open'),
      };
    };
    return {
      assertDropdown,
      ...testUtils,
    };
  };

  beforeEach(() => {
    window.history.replaceState({}, '', '/project/1');
  });
  afterEach(() => {
    window.localStorage.removeItem('pachyderm-console-1');
  });

  it('should display empty state messages', async () => {
    window.history.replaceState({}, '', '/project/6');
    const {findByRole, getByText, findByText, assertDropdown} = renderTestBed();

    const searchBar = await findByRole('searchbox');
    assertDropdown().toBeHidden();

    searchBar.focus();
    assertDropdown().toBeShown();
    expect(getByText('There are no recent searches.')).toBeInTheDocument();
    expect(getByText('There are no jobs on this project.')).toBeInTheDocument();
    userEvent.type(searchBar, 'anything');
    expect(
      await findByText('No matching pipelines, jobs or global ID found.'),
    ).toBeInTheDocument();
    assertDropdown().toBeShown();
  });

  it('should underline matched search term', async () => {
    const {findByRole, findByText, assertDropdown} = renderTestBed();

    const searchBar = await findByRole('searchbox');
    assertDropdown().toBeHidden();

    userEvent.type(searchBar, 'im');
    assertDropdown().toBeShown();
    expect(await findByText('im')).toHaveClass('underline');
    expect(await findByText('ages')).not.toHaveClass('underline');
  });

  it('should save recent searches and populate search bar', async () => {
    const {findByRole, assertDropdown} = renderTestBed();

    const searchBar = await findByRole('searchbox');
    assertDropdown().toBeHidden();

    userEvent.type(searchBar, 'images');
    assertDropdown().toBeShown();
    const searchResult = await findByRole('button', {name: 'See Commits'});
    click(searchResult);
    assertDropdown().toBeHidden();

    searchBar.focus();
    assertDropdown().toBeShown();
    const savedSearch = await findByRole('button', {name: 'images'});
    click(savedSearch);
    expect(searchBar).toHaveValue('images');
    assertDropdown().toBeShown();
  });

  it('should clear saved search history', async () => {
    const {findByRole, getByText, queryByText, assertDropdown} =
      renderTestBed();
    expect(getByText('There are no recent searches.')).toBeInTheDocument();

    const searchBar = await findByRole('searchbox');
    assertDropdown().toBeHidden();

    userEvent.type(searchBar, 'edges');
    assertDropdown().toBeShown();
    click(await findByRole('button', {name: 'See Jobs'}));
    assertDropdown().toBeHidden();

    searchBar.focus();
    assertDropdown().toBeShown();
    expect(
      queryByText('There are no recent searches.'),
    ).not.toBeInTheDocument();
    click(await findByRole('button', {name: 'Clear'}));
    expect(getByText('There are no recent searches.')).toBeInTheDocument();
    assertDropdown().toBeShown();
  });

  it('should route to selected repo', async () => {
    const {findByRole, assertDropdown} = renderTestBed();

    const searchBar = await findByRole('searchbox');
    assertDropdown().toBeHidden();

    userEvent.type(searchBar, 'images');
    assertDropdown().toBeShown();
    const searchResult = await findByRole('button', {name: 'See Commits'});
    click(searchResult);

    assertDropdown().toBeHidden();
    expect(window.location.pathname).toBe(
      '/project/1/repos/images/branch/default',
    );
  });

  it('should route to selected pipeline and pipeline jobs', async () => {
    const {findByRole, queryAllByText, assertDropdown} = renderTestBed();
    const searchBar = await findByRole('searchbox');
    assertDropdown().toBeHidden();

    userEvent.type(searchBar, 'edges');
    assertDropdown().toBeShown();
    click(await findByRole('button', {name: 'See Jobs'}));
    expect(window.location.pathname).toBe('/project/1/pipelines/edges/jobs');
    assertDropdown().toBeHidden();

    searchBar.focus();
    assertDropdown().toBeShown();
    click(await findByRole('button', {name: 'edges'}));
    click(await queryAllByText('edges')[1]);

    assertDropdown().toBeHidden();
    expect(window.location.pathname).toBe('/project/1/pipelines/edges');
  });

  it('should route to jobs for selected id', async () => {
    const {findByRole, findByText, assertDropdown} = renderTestBed();

    const searchBar = await findByRole('searchbox');
    assertDropdown().toBeHidden();

    userEvent.type(searchBar, '23b9af7d5d4343219bc8e02ff44cd55a');
    assertDropdown().toBeShown();
    click(await findByText('23b9af7d5d4343219bc8e02ff44cd55a'));

    assertDropdown().toBeHidden();
    expect(window.location.pathname).toBe(
      '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a',
    );
  });

  it('should route to jobs view when chips are clicked', async () => {
    const {findByRole, assertDropdown} = renderTestBed();

    const searchBar = await findByRole('searchbox');
    assertDropdown().toBeHidden();

    searchBar.focus();
    assertDropdown().toBeShown();
    click(await findByRole('button', {name: 'All (2)'}));
    expect(window.location.pathname).toBe('/project/1/jobs');
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
    ]);
    assertDropdown().toBeHidden();
  });
});
