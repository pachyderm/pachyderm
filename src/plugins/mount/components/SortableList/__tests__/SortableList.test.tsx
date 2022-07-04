import React from 'react';
import {render, fireEvent, waitFor} from '@testing-library/react';

import SortableList from '../SortableList';
import {Repo} from 'plugins/mount/types';
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

  it('should disable mount for repo with no branches', async () => {
    const repos: Repo[] = [
      {
        repo: 'images',
        authorization: 'off',
        branches: [],
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} repos={repos} updateData={updateData} />,
    );
    const listItem = getByTestId('ListItem__noBranches');
    expect(listItem).toHaveTextContent('images');
    expect(listItem).toHaveTextContent('No Branches');
    expect(listItem).toHaveAttribute(
      'title',
      'A repository must have a branch in order to mount it',
    );
  });

  it('should disable mount for repo with no access', async () => {
    const repos: Repo[] = [
      {
        repo: 'images',
        authorization: 'none',
        branches: [],
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} repos={repos} updateData={updateData} />,
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
    const repos: Repo[] = [
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
      <SortableList open={open} repos={repos} updateData={updateData} />,
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
    const repos: Repo[] = [
      {
        repo: 'images',
        authorization: 'off',
        branches: [
          {
            branch: 'master',
            mount: [
              {
                name: '',
                state: 'unmounted',
                status: '',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 0,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
        ],
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} repos={repos} updateData={updateData} />,
    );
    const listItem = getByTestId('ListItem__branches');
    const mountButton = getByTestId('ListItem__mount');

    expect(listItem).toHaveTextContent('images');
    expect(mountButton).not.toBeDisabled();
    mountButton.click();

    await waitFor(() => {
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith(
        'repos/images/master/_mount?name=images&mode=ro',
        'PUT',
      );
    });
    expect(mountButton).toBeDisabled();
    expect(updateData).toBeCalledWith([]);
  });

  it('should allow user to unmount mounted repo', async () => {
    const repos: Repo[] = [
      {
        repo: 'images',
        authorization: 'off',
        branches: [
          {
            branch: 'master',
            mount: [
              {
                name: '',
                state: 'mounted',
                status: '',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 0,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
        ],
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} repos={repos} updateData={updateData} />,
    );
    const listItem = getByTestId('ListItem__branches');
    const unmountButton = getByTestId('ListItem__unmount');

    expect(listItem).toHaveTextContent('images');
    expect(unmountButton).not.toBeDisabled();

    unmountButton.click();

    await waitFor(() => {
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith(
        'repos/images/master/_unmount?name=images',
        'PUT',
      );
    });
    expect(unmountButton).toBeDisabled();
    expect(updateData).toBeCalledWith([]);
  });

  it('should open mounted repo on click', async () => {
    const repos: Repo[] = [
      {
        repo: 'images',
        authorization: 'off',
        branches: [
          {
            branch: 'master',
            mount: [
              {
                name: '',
                state: 'mounted',
                status: '',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 0,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
        ],
      },
    ];

    const {getByText} = render(
      <SortableList open={open} repos={repos} updateData={updateData} />,
    );
    getByText('images').click();
    expect(open).toHaveBeenCalledWith('images');
  });

  it('should allow user to select a branch to mount', async () => {
    const repos: Repo[] = [
      {
        repo: 'images',
        authorization: 'off',
        branches: [
          {
            branch: 'master',
            mount: [
              {
                name: '',
                state: 'unmounted',
                status: '',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 0,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
          {
            branch: 'develop',
            mount: [
              {
                name: '',
                state: 'unmounted',
                status: '',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 0,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
        ],
      },
    ];

    const {getByText, getByTestId} = render(
      <SortableList open={open} repos={repos} updateData={updateData} />,
    );
    const select = getByTestId('ListItem__select') as HTMLSelectElement;

    expect(select.value).toEqual('master');
    fireEvent.change(select, {target: {value: 'develop'}});

    getByText('Mount').click();
    expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith(
      'repos/images/develop/_mount?name=images&mode=ro',
      'PUT',
    );
  });

  it('should display state and status of mounted brach', async () => {
    const repos: Repo[] = [
      {
        repo: 'edges',
        authorization: 'off',
        branches: [
          {
            branch: 'master',
            mount: [
              {
                name: '',
                state: 'error',
                status: 'error mounting branch',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 0,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
        ],
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} repos={repos} updateData={updateData} />,
    );

    const statusIcon = getByTestId('ListItem__statusIcon');
    expect(statusIcon.title).toEqual('Error: error mounting branch');
    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('up to date');
  });

  it('should disable item when it is in a loading state', async () => {
    const repos: Repo[] = [
      {
        repo: '1',
        authorization: 'off',
        branches: [
          {
            branch: 'master',
            mount: [
              {
                name: '',
                state: 'error',
                status: 'error mounting branch',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 0,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
        ],
      },
      {
        repo: '2',
        authorization: 'off',
        branches: [
          {
            branch: 'master',
            mount: [
              {
                name: '',
                state: 'unmounting',
                status: '',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 0,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
        ],
      },
      {
        repo: '3',
        authorization: 'off',
        branches: [
          {
            branch: 'master',
            mount: [
              {
                name: '',
                state: 'mounting',
                status: '',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 0,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
        ],
      },
    ];
    const {getAllByTestId} = render(
      <SortableList open={open} repos={repos} updateData={updateData} />,
    );
    const unmountButtons = getAllByTestId('ListItem__unmount');
    expect(unmountButtons[0]).toBeDisabled();
    expect(unmountButtons[1]).toBeDisabled();
    expect(unmountButtons[2]).toBeDisabled();
  });
  it('should show up to date when behindness is zero', async () => {
    const repos: Repo[] = [
      {
        repo: 'edges',
        authorization: 'off',
        branches: [
          {
            branch: 'master',
            mount: [
              {
                name: '',
                state: 'mounted',
                status: 'all is well, la la la',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 0,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
        ],
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} repos={repos} updateData={updateData} />,
    );

    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('up to date');
  });
  it('should show behind when one commit behind', async () => {
    const repos: Repo[] = [
      {
        repo: 'edges',
        authorization: 'off',
        branches: [
          {
            branch: 'master',
            mount: [
              {
                name: '',
                state: 'mounted',
                status: 'all is well, la la la',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 1,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
        ],
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} repos={repos} updateData={updateData} />,
    );

    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('1 commit behind');
  });
  it('should show behind when two commits behind', async () => {
    const repos: Repo[] = [
      {
        repo: 'edges',
        authorization: 'off',
        branches: [
          {
            branch: 'master',
            mount: [
              {
                name: '',
                state: 'mounted',
                status: 'all is well, la la la',
                mode: null,
                mountpoint: null,
                mount_key: null,
                how_many_commits_behind: 2,
                actual_mounted_commit: 'a1b2c3',
                latest_commit: 'a1b2c3',
              },
            ],
          },
        ],
      },
    ];

    const {getByTestId} = render(
      <SortableList open={open} repos={repos} updateData={updateData} />,
    );

    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('2 commits behind');
  });
});
