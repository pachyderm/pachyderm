import React from 'react';
import {render, fireEvent} from '@testing-library/react';

import SortableList from '../SortableList';
import {requestAPI} from '../../../../handler';
jest.mock('../../../../handler');

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
    const repos = [
      {
        repo: 'images',
        branches: [
          {
            branch: 'master',
            mount: {
              name: null,
              state: 'unmounted',
              status: null,
              mode: null,
              mountpoint: null,
            },
          },
        ],
      },
    ];

    const {getByTestId, getByText} = render(
      <SortableList open={open} repos={repos} />,
    );
    const listItem = getByTestId('ListItem__branches');
    expect(listItem).toHaveTextContent('images');

    getByText('Mount').click();
    expect(requestAPI).toHaveBeenCalledWith(
      'repos/images/master/_mount?name=images&mode=ro',
      'PUT',
    );
  });

  it('should allow user to unmount mounted repo', async () => {
    const open = jest.fn();

    const repos = [
      {
        repo: 'images',
        branches: [
          {
            branch: 'master',
            mount: {
              name: null,
              state: 'mounted',
              status: null,
              mode: null,
              mountpoint: null,
            },
          },
        ],
      },
    ];

    const {getByTestId, getByText} = render(
      <SortableList open={open} repos={repos} />,
    );
    const listItem = getByTestId('ListItem__branches');
    expect(listItem).toHaveTextContent('images');
    getByText('Unmount').click();
    expect(requestAPI).toHaveBeenCalledWith(
      'repos/images/master/_unmount?name=images',
      'PUT',
    );
  });

  it('should open mounted repo on click', async () => {
    const open = jest.fn();

    const repos = [
      {
        repo: 'images',
        branches: [
          {
            branch: 'master',
            mount: {
              name: null,
              state: 'mounted',
              status: null,
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

    const repos = [
      {
        repo: 'images',
        branches: [
          {
            branch: 'master',
            mount: {
              name: null,
              state: 'unmounted',
              status: null,
              mode: null,
              mountpoint: null,
            },
          },
          {
            branch: 'develop',
            mount: {
              name: null,
              state: 'unmounted',
              status: null,
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
});
