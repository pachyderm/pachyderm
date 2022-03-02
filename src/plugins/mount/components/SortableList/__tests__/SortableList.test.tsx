import React from 'react';
import {act, render, screen, fireEvent} from '@testing-library/react';

import SortableList from '../SortableList';
import {requestAPI} from '../../../../../handler';
import userEvent from '@testing-library/user-event';
import {Repo} from 'plugins/mount/types';
jest.mock('../../../../../handler');

describe('mount components', () => {
  it('should disable mount for repo with no branches', async () => {
    const open = jest.fn();
    const repos = [
      {
        repo: 'images',
        branches: [],
      },
    ];

    const {getByTestId} = render(<SortableList open={open} repos={repos} />);
    const listItem = getByTestId('ListItem__noBranches');
    expect(listItem).toHaveTextContent('images');
    expect(listItem).toHaveTextContent('No Branches');
  });

  it('should sort repos by name', async () => {
    const open = jest.fn();

    const repos = [
      {
        repo: 'images',
        branches: [],
      },
      {
        repo: 'data',
        branches: [],
      },
    ];

    const {getByText, getAllByTestId} = render(
      <SortableList open={open} repos={repos} />,
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
    const open = jest.fn();
    const repos: Repo[] = [
      {
        repo: 'images',
        branches: [
          {
            branch: 'master',
            mount: {
              name: '',
              state: 'unmounted',
              status: '',
              mode: null,
              mountpoint: null,
            },
          },
        ],
      },
    ];

    const {getByTestId} = render(<SortableList open={open} repos={repos} />);
    const listItem = getByTestId('ListItem__branches');
    const mountButton = getByTestId('ListItem__mount');

    expect(listItem).toHaveTextContent('images');
    expect(mountButton).not.toBeDisabled();
    mountButton.click();

    expect(mountButton).toBeDisabled();
    expect(requestAPI).toHaveBeenCalledWith(
      'repos/images/master/_mount?name=images&mode=ro',
      'PUT',
    );
  });

  it('should allow user to unmount mounted repo', async () => {
    const open = jest.fn();

    const repos: Repo[] = [
      {
        repo: 'images',
        branches: [
          {
            branch: 'master',
            mount: {
              name: '',
              state: 'mounted',
              status: '',
              mode: null,
              mountpoint: null,
            },
          },
        ],
      },
    ];

    const {getByTestId} = render(<SortableList open={open} repos={repos} />);
    const listItem = getByTestId('ListItem__branches');
    const unmountButton = getByTestId('ListItem__unmount');

    expect(listItem).toHaveTextContent('images');
    expect(unmountButton).not.toBeDisabled();

    unmountButton.click();

    expect(unmountButton).toBeDisabled();
    expect(requestAPI).toHaveBeenCalledWith(
      'repos/images/master/_unmount?name=images',
      'PUT',
    );
  });

  it('should open mounted repo on click', async () => {
    const open = jest.fn();

    const repos: Repo[] = [
      {
        repo: 'images',
        branches: [
          {
            branch: 'master',
            mount: {
              name: '',
              state: 'mounted',
              status: '',
              mode: null,
              mountpoint: null,
            },
          },
        ],
      },
    ];

    const {getByText} = render(<SortableList open={open} repos={repos} />);
    getByText('images').click();
    expect(open).toHaveBeenCalledWith('images');
  });

  it('should allow user to select a branch to mount', async () => {
    const open = jest.fn();

    const repos: Repo[] = [
      {
        repo: 'images',
        branches: [
          {
            branch: 'master',
            mount: {
              name: '',
              state: 'unmounted',
              status: '',
              mode: null,
              mountpoint: null,
            },
          },
          {
            branch: 'develop',
            mount: {
              name: '',
              state: 'unmounted',
              status: '',
              mode: null,
              mountpoint: null,
            },
          },
        ],
      },
    ];

    const {getByText, getByTestId} = render(
      <SortableList open={open} repos={repos} />,
    );
    const select = getByTestId('ListItem__select') as HTMLSelectElement;

    expect(select.value).toEqual('master');
    fireEvent.change(select, {target: {value: 'develop'}});

    getByText('Mount').click();
    expect(requestAPI).toHaveBeenCalledWith(
      'repos/images/develop/_mount?name=images&mode=ro',
      'PUT',
    );
  });

  it('should display state and status of mounted brach', async () => {
    const open = jest.fn();

    const repos: Repo[] = [
      {
        repo: 'edges',
        branches: [
          {
            branch: 'master',
            mount: {
              name: '',
              state: 'error',
              status: 'error mounting branch',
              mode: null,
              mountpoint: null,
            },
          },
        ],
      },
    ];
    act(() => {
      render(<SortableList open={open} repos={repos} />);
    });
    const statusIcon = screen.getByTestId('ListItem__statusIcon');

    userEvent.hover(statusIcon);
    expect(screen.queryByTestId('tooltip-branch-status')).toHaveTextContent(
      'Error: error mounting branch',
    );
  });

  it('should disable item when it is in a loading state', async () => {
    const repos: Repo[] = [
      {
        repo: '1',
        branches: [
          {
            branch: 'master',
            mount: {
              name: '',
              state: 'error',
              status: 'error mounting branch',
              mode: null,
              mountpoint: null,
            },
          },
        ],
      },
      {
        repo: '2',
        branches: [
          {
            branch: 'master',
            mount: {
              name: '',
              state: 'unmounting',
              status: '',
              mode: null,
              mountpoint: null,
            },
          },
        ],
      },
      {
        repo: '3',
        branches: [
          {
            branch: 'master',
            mount: {
              name: '',
              state: 'mounting',
              status: '',
              mode: null,
              mountpoint: null,
            },
          },
        ],
      },
    ];
    const {getAllByTestId} = render(<SortableList open={open} repos={repos} />);
    const unmountButtons = getAllByTestId('ListItem__unmount');
    expect(unmountButtons[0]).toBeDisabled();
    expect(unmountButtons[1]).toBeDisabled();
    expect(unmountButtons[2]).toBeDisabled();
  });
});
