import React from 'react';
import {render, fireEvent, waitFor} from '@testing-library/react';

import SortableList from '../SortableList';
import {Repo, Mount, ProjectInfo} from 'plugins/mount/types';
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
        project: 'default',
        authorization: 'off',
        branches: [],
      },
    ];

    const {getByTestId} = render(
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'unmounted'}
        projects={[]}
      />,
    );
    const listItem = getByTestId('ListItem__noBranches');
    expect(listItem).toHaveTextContent('images');
    expect(listItem).toHaveTextContent('No branches');
    expect(listItem).toHaveAttribute('title', "Repo doesn't have a branch");
  });

  it('should disable mount for repo with no access', async () => {
    const items: Repo[] = [
      {
        repo: 'images',
        project: 'default',
        authorization: 'none',
        branches: [],
      },
    ];

    const {getByTestId} = render(
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'unmounted'}
        projects={[]}
      />,
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
        project: 'default',
        authorization: 'off',
        branches: [],
      },
      {
        repo: 'data',
        project: 'default',
        authorization: 'off',
        branches: [],
      },
    ];

    const {getByText, getAllByTestId} = render(
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'unmounted'}
        projects={[]}
      />,
    );
    let listItems = getAllByTestId('ListItem__noBranches');
    expect(listItems).toHaveLength(2);
    expect(listItems[0]).toHaveTextContent('data');
    expect(listItems[1]).toHaveTextContent('images');

    getByText('Name').click();

    listItems = getAllByTestId('ListItem__noBranches');
    expect(listItems).toHaveLength(2);
    expect(listItems[0]).toHaveTextContent('images');
    expect(listItems[1]).toHaveTextContent('data');
  });

  it('should allow user to mount unmounted repo', async () => {
    const items: Repo[] = [
      {
        repo: 'images',
        project: 'default',
        authorization: 'off',
        branches: ['master'],
      },
    ];

    const {getByTestId} = render(
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'unmounted'}
        projects={[]}
      />,
    );
    const listItem = getByTestId('ListItem__repo');
    const mountButton = getByTestId('ListItem__mount');

    expect(listItem).toHaveTextContent('images');
    expect(mountButton).not.toBeDisabled();
    mountButton.click();

    await waitFor(() => {
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith('_mount', 'PUT', {
        mounts: [
          {
            name: 'default_images',
            project: 'default',
            repo: 'images',
            branch: 'master',
            mode: 'ro',
          },
        ],
      });
    });
    expect(mountButton).toBeDisabled();
    expect(updateData).toHaveBeenCalledWith([]);
  });

  it('should allow user to unmount mounted repo', async () => {
    const items: Mount[] = [
      {
        name: 'images',
        repo: 'images',
        project: 'default',
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
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'mounted'}
        projects={[]}
      />,
    );
    const listItem = getByTestId('ListItem__repo');
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
    expect(updateData).toHaveBeenCalledWith([]);
  });

  it('should open mounted repo on click', async () => {
    const items: Mount[] = [
      {
        name: 'images',
        repo: 'images',
        project: 'default',
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
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'mounted'}
        projects={[]}
      />,
    );
    getByText('images').click();
    expect(open).toHaveBeenCalledWith('images');
  });

  it('should allow user to select a branch to mount', async () => {
    const items: Repo[] = [
      {
        repo: 'images',
        project: 'default',
        authorization: 'off',
        branches: ['master', 'develop'],
      },
    ];

    const {getByText, getByTestId} = render(
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'unmounted'}
        projects={[]}
      />,
    );
    const select = getByTestId('ListItem__select') as HTMLSelectElement;

    expect(select.value).toBe('master');
    fireEvent.change(select, {target: {value: 'develop'}});

    getByText('Mount').click();
    expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith('_mount', 'PUT', {
      mounts: [
        {
          name: 'default_images_develop',
          repo: 'images',
          project: 'default',
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
        project: 'default',
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
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'mounted'}
        projects={[]}
      />,
    );

    const statusIcon = getByTestId('ListItem__statusIcon');
    expect(statusIcon.title).toBe('Error: error mounting branch');
    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('up to date');
  });

  it('should disable item when it is in a loading state', async () => {
    const items: Mount[] = [
      {
        name: '1',
        repo: '1',
        project: 'default',
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
        project: 'default',
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
        project: 'default',
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
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'mounted'}
        projects={[]}
      />,
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
        project: 'default',
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
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'mounted'}
        projects={[]}
      />,
    );

    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('up to date');
  });

  it('should show behind when one commit behind', async () => {
    const items: Mount[] = [
      {
        name: 'edges',
        repo: 'edges',
        project: 'default',
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
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'mounted'}
        projects={[]}
      />,
    );

    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('1 commit behind');
  });

  it('should show behind when two commits behind', async () => {
    const items: Mount[] = [
      {
        name: 'edges',
        repo: 'edges',
        project: 'default',
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
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'mounted'}
        projects={[]}
      />,
    );

    const commitBehindnessText = getByTestId('ListItem__commitBehindness');
    expect(commitBehindnessText.textContent).toContain('2 commits behind');
  });

  it('should grey out mount button for selected branches in dropdown that are already mounted', async () => {
    const mountedItems: Mount[] = [
      {
        name: 'edges',
        repo: 'edges',
        project: 'default',
        branch: 'mounted_branch',
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

    const items: Repo[] = [
      {
        repo: 'edges',
        project: 'default',
        authorization: 'off',
        branches: ['dev', 'master', 'mounted_branch'],
      },
    ];

    const {getByTestId} = render(
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={mountedItems}
        type={'unmounted'}
        projects={[]}
      />,
    );

    fireEvent.change(getByTestId('ListItem__select'), {
      target: {value: 'mounted_branch'},
    });
    expect(getByTestId('ListItem__mount')).toBeDisabled();

    fireEvent.change(getByTestId('ListItem__select'), {target: {value: 'dev'}});
    expect(getByTestId('ListItem__mount')).not.toBeDisabled();
  });

  it('project filtering', async () => {
    const items: Repo[] = [
      {
        repo: 'images',
        project: 'p1',
        authorization: 'off',
        branches: ['master'],
      },
      {
        repo: 'edges',
        project: 'p1',
        authorization: 'off',
        branches: ['master'],
      },
      {
        repo: 'edges',
        project: 'default',
        authorization: 'off',
        branches: ['master'],
      },
    ];

    const projects: ProjectInfo[] = [
      {
        project: {
          name: 'p1',
        },
        auth: {
          permissions: [0, 1, 2],
          roles: ['foo', 'bar'],
        },
      },
      {
        project: {
          name: 'default',
        },
        auth: {
          permissions: [3, 4, 5],
          roles: ['foo', 'bar', 'baz'],
        },
      },
    ];

    const {getAllByTestId, getByTestId} = render(
      <SortableList
        open={open}
        items={items}
        updateData={updateData}
        mountedItems={[]}
        type={'unmounted'}
        projects={projects}
      />,
    );

    let listItems = getAllByTestId('ListItem__repo');
    expect(listItems).toHaveLength(3);

    fireEvent.change(getByTestId('ProjectList__select'), {
      target: {value: 'p1'},
    });
    listItems = getAllByTestId('ListItem__repo');
    expect(listItems).toHaveLength(2);

    fireEvent.change(getByTestId('ProjectList__select'), {
      target: {value: 'default'},
    });
    listItems = getAllByTestId('ListItem__repo');
    expect(listItems).toHaveLength(1);

    fireEvent.change(getByTestId('ProjectList__select'), {
      target: {value: 'All projects'},
    });
    listItems = getAllByTestId('ListItem__repo');
    expect(listItems).toHaveLength(3);
  });
});
