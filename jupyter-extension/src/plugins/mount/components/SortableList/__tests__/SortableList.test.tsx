import React from 'react';
import {render, fireEvent, waitFor} from '@testing-library/react';

import SortableList from '../SortableList';
import {Repo, Mount} from 'plugins/mount/types';
import * as requestAPI from '../../../../../handler';
import {mockedRequestAPI} from 'utils/testUtils';
jest.mock('../../../../../handler');

describe('sortable list components', () => {
  let open = jest.fn();
  let updateData = jest.fn();
  const mockRequestAPI = requestAPI as jest.Mocked<typeof requestAPI>;

  beforeEach(() => {
    open = jest.fn();
    updateData = jest.fn();
    mockRequestAPI.requestAPI.mockImplementation(mockedRequestAPI([]));
  });

  it('should disable mount for repo with no unmounted branches', async () => {
    const items: Repo[] = [
      {
        repo: 'images',
        authorization: 'off',
        branches: [],
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );
    const listItem = getByTestId('ListItem__noBranches');
    expect(listItem).toHaveTextContent('images');
    expect(listItem).toHaveTextContent('No branches');
    expect(listItem).toHaveAttribute(
      'title',
      "Either all branches are mounted or the repo doesn't have a branch",
    );
  });

  it('should disable mount for repo with no access', async () => {
    const items: Repo[] = [
      {
        repo: 'images',
        authorization: 'none',
        branches: [],
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );
    const listItem = getByTestId('ListItem__unauthorized');
    expect(listItem).toHaveTextContent('images');
    expect(listItem).toHaveTextContent('No read access');
    expect(listItem).toHaveAttribute(
      'title',
      "You don't have the correct permissions to access this repository",
    );
  });

  it('should sort repos by name', async () => {
    const items: Repo[] = [
      {
        repo: 'images',
        authorization: 'off',
        branches: [],
      },
      {
        repo: 'data',
        authorization: 'off',
        branches: [],
      },
    ];

    const {getByText, getAllByTestId} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );
    let listItems = getAllByTestId('ListItem__noBranches');
    expect(listItems.length).toEqual(2);
    expect(listItems[0]).toHaveTextContent('data');
    expect(listItems[1]).toHaveTextContent('images');

    getByText('Name').click();

    listItems = getAllByTestId('ListItem__noBranches');
    expect(listItems.length).toEqual(2);
    expect(listItems[0]).toHaveTextContent('images');
    expect(listItems[1]).toHaveTextContent('data');
  });

  it('should allow user to mount unmounted repo', async () => {
    const items: Repo[] = [
      {
        repo: 'images',
        authorization: 'off',
        branches: ['master'],
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );
    const listItem = getByTestId('ListItem__branches');
    const mountButton = getByTestId('ListItem__mount');

    expect(listItem).toHaveTextContent('images');
    expect(mountButton).not.toBeDisabled();
    mountButton.click();

    await waitFor(() => {
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith('_mount', 'PUT', {
        mounts: [
          {
            name: 'images',
            repo: 'images',
            branch: 'master',
            mode: 'ro',
          },
        ],
      });
    });
    expect(mountButton).toBeDisabled();
    expect(updateData).toBeCalledWith([]);
  });

  it('should allow user to unmount mounted repo', async () => {
    const items: Mount[] = [
      {
        name: 'images',
        repo: 'images',
        branch: 'master',
        commit: null,
        glob: null,
        mode: null,
        state: 'mounted',
        status: '',
        mountpoint: null,
        how_many_commits_behind: 0,
        actual_mounted_commit: 'a1b2c3',
        latest_commit: 'a1b2c3',
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );
    const listItem = getByTestId('ListItem__branches');
    const unmountButton = getByTestId('ListItem__unmount');

    expect(listItem).toHaveTextContent('images');
    expect(unmountButton).not.toBeDisabled();

    unmountButton.click();

    await waitFor(() => {
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith(
        '_unmount',
        'PUT',
        {
          mounts: ['images'],
        },
      );
    });
    expect(unmountButton).toBeDisabled();
    expect(updateData).toBeCalledWith([]);
  });

  it('should open mounted repo on click', async () => {
    const items: Mount[] = [
      {
        name: 'images',
        repo: 'images',
        branch: 'master',
        commit: null,
        glob: null,
        mode: null,
        state: 'mounted',
        status: '',
        mountpoint: null,
        how_many_commits_behind: 0,
        actual_mounted_commit: 'a1b2c3',
        latest_commit: 'a1b2c3',
      },
    ];

    const {getByText} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );
    getByText('images').click();
    expect(open).toHaveBeenCalledWith('images');
  });

  it('should allow user to select a branch to mount', async () => {
    const items: Repo[] = [
      {
        repo: 'images',
        authorization: 'off',
        branches: ['master', 'develop'],
      },
    ];

    const {getByText, getByTestId} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );
    const select = getByTestId('ListItem__select') as HTMLSelectElement;

    expect(select.value).toEqual('master');
    fireEvent.change(select, {target: {value: 'develop'}});

    getByText('Mount').click();
    expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith('_mount', 'PUT', {
      mounts: [
        {
          name: 'images_develop',
          repo: 'images',
          branch: 'develop',
          mode: 'ro',
        },
      ],
    });
  });

  it('should display state and status of mounted brach', async () => {
    const items: Mount[] = [
      {
        name: 'edges',
        repo: 'edges',
        branch: 'master',
        commit: null,
        glob: null,
        mode: null,
        state: 'error',
        status: 'error mounting branch',
        mountpoint: null,
        how_many_commits_behind: 0,
        actual_mounted_commit: 'a1b2c3',
        latest_commit: 'a1b2c3',
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );

    const statusIcon = getByTestId('ListItem__statusIcon');
    expect(statusIcon.title).toEqual('Error: error mounting branch');
    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('up to date');
  });

  it('should disable item when it is in a loading state', async () => {
    const items: Mount[] = [
      {
        name: '1',
        repo: '1',
        branch: 'master',
        commit: null,
        glob: null,
        mode: null,
        state: 'error',
        status: 'error mounting branch',
        mountpoint: null,
        how_many_commits_behind: 0,
        actual_mounted_commit: 'a1b2c3',
        latest_commit: 'a1b2c3',
      },
      {
        name: '2',
        repo: '2',
        branch: 'master',
        commit: null,
        glob: null,
        mode: null,
        state: 'unmounting',
        status: '',
        mountpoint: null,
        how_many_commits_behind: 0,
        actual_mounted_commit: 'a1b2c3',
        latest_commit: 'a1b2c3',
      },
      {
        name: '3',
        repo: '3',
        branch: 'master',
        commit: null,
        glob: null,
        mode: null,
        state: 'mounting',
        status: '',
        mountpoint: null,
        how_many_commits_behind: 0,
        actual_mounted_commit: 'a1b2c3',
        latest_commit: 'a1b2c3',
      },
    ];
    const {getAllByTestId} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );
    const unmountButtons = getAllByTestId('ListItem__unmount');
    expect(unmountButtons[0]).toBeDisabled();
    expect(unmountButtons[1]).toBeDisabled();
    expect(unmountButtons[2]).toBeDisabled();
  });

  it('should show up to date when behindness is zero', async () => {
    const items: Mount[] = [
      {
        name: 'edges',
        repo: 'edges',
        branch: 'master',
        commit: null,
        glob: null,
        mode: null,
        state: 'mounted',
        status: 'all is well, la la la',
        mountpoint: null,
        how_many_commits_behind: 0,
        actual_mounted_commit: 'a1b2c3',
        latest_commit: 'a1b2c3',
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );

    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('up to date');
  });

  it('should show behind when one commit behind', async () => {
    const items: Mount[] = [
      {
        name: 'edges',
        repo: 'edges',
        branch: 'master',
        commit: null,
        glob: null,
        mode: null,
        state: 'mounted',
        status: 'all is well, la la la',
        mountpoint: null,
        how_many_commits_behind: 1,
        actual_mounted_commit: 'a1b2c3',
        latest_commit: 'a1b2c3',
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );

    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('1 commit behind');
  });

  it('should show behind when two commits behind', async () => {
    const items: Mount[] = [
      {
        name: 'edges',
        repo: 'edges',
        branch: 'master',
        commit: null,
        glob: null,
        mode: null,
        state: 'mounted',
        status: 'all is well, la la la',
        mountpoint: null,
        how_many_commits_behind: 2,
        actual_mounted_commit: 'a1b2c3',
        latest_commit: 'a1b2c3',
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} items={items} updateData={updateData} />,
    );

    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('2 commits behind');
  });
});
