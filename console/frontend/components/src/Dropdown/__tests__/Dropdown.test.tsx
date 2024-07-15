import {fireEvent, render, waitFor, screen} from '@testing-library/react';
import React, {useState} from 'react';

import {click, type} from '@dash-frontend/testHelpers';

import {DropdownItem, DropdownProps, SearchableDropdown} from '../';

describe('Dropdown', () => {
  const renderTestbed = ({
    closeOnClick = false,
    disabled = false,
  }: {closeOnClick?: boolean; disabled?: boolean} = {}) => {
    const onSelect = jest.fn();

    const items: DropdownItem[] = [
      {id: 'disabled1', content: 'Disabled 1', disabled: true},
      {id: 'link1', content: 'Link 1', closeOnClick},
      {id: 'link2', content: 'Link 2', closeOnClick},
    ];

    render(
      <>
        <div>Outside</div>

        <SearchableDropdown
          onSelect={onSelect}
          items={items}
          emptyResultsContent={'No results found.'}
          buttonOpts={{disabled}}
        >
          Button
        </SearchableDropdown>
      </>,
    );

    return {
      onSelect,
    };
  };

  it('should toggle dropdown when DropdownButton is clicked', async () => {
    renderTestbed();

    const button = screen.getByText('Button');

    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    await click(button);
    expect(screen.getByRole('menu')).toBeInTheDocument();
    await click(button);
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });

  it('should close the dropdown if a user clicks away from the menu', async () => {
    renderTestbed();

    const button = screen.getByText('Button');
    const outside = screen.getByText('Outside');

    await click(button);
    await screen.findByRole('menu');

    await click(outside);

    await waitFor(() =>
      expect(screen.queryByRole('menu')).not.toBeInTheDocument(),
    );
  });

  it('should invoke onSelect callback with the selected element', async () => {
    const {onSelect} = renderTestbed();

    const button = screen.getByText('Button');
    const link1 = screen.getByText('Link 1');
    const link2 = screen.getByText('Link 2');

    await click(button);
    await click(link1);
    expect(onSelect).toHaveBeenCalledWith('link1');
    await click(link2);
    expect(onSelect).toHaveBeenCalledWith('link2');
  });

  it('should optionally close the dropdown menu on click of an item', async () => {
    renderTestbed({
      closeOnClick: true,
    });

    const button = screen.getByText('Button');
    const link1 = screen.getByText('Link 1');

    await click(button);
    await click(link1);

    await waitFor(() =>
      expect(screen.queryByRole('menu')).not.toBeInTheDocument(),
    );
    expect(screen.getByTestId('DropdownButton__button')).toHaveFocus();
  });

  it('should disable the dropdown when buttonOpts disabled is true', async () => {
    renderTestbed({
      disabled: true,
    });

    const button = screen.getByText('Button');
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    await click(button);
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });

  describe('keyboard interactions', () => {
    it('should allow users to open the dropdown with the down key', () => {
      renderTestbed();

      const button = screen.getByText('Button');

      fireEvent.keyDown(button, {key: 'ArrowDown'});
      expect(screen.getByRole('menu')).toBeInTheDocument();
    });

    it('should allow users to close the dropdown with the up key', async () => {
      renderTestbed();

      const button = screen.getByText('Button');

      await click(button);

      fireEvent.keyDown(button, {key: 'ArrowUp'});
      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('should allow users to close the dropdown with the escape key', async () => {
      renderTestbed();

      const button = screen.getByText('Button');

      await click(button);

      fireEvent.keyDown(button, {key: 'Escape'});
      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('should allow users to navigate to the search bar with the down key', async () => {
      renderTestbed();

      const button = screen.getByText('Button');

      fireEvent.keyDown(button, {key: 'ArrowDown'});
      fireEvent.keyDown(button, {key: 'ArrowDown'});

      expect(document.activeElement?.nodeName).toBe('INPUT');
    });

    it('should allow users to navigate to the first active item with the down key', async () => {
      const {onSelect} = renderTestbed();

      const button = screen.getByText('Button');

      fireEvent.keyDown(button, {key: 'ArrowDown'});
      fireEvent.keyDown(button, {key: 'ArrowDown'});

      const searchBar = screen.getByLabelText('Search');
      fireEvent.keyDown(searchBar, {key: 'ArrowDown'});

      await type(document.activeElement as HTMLElement, '{enter}');
      expect(onSelect).toHaveBeenCalledWith('link1');
    });

    it('should allow users to descend through active menu items with the down key', async () => {
      const {onSelect} = renderTestbed();

      const button = screen.getByText('Button');

      fireEvent.keyDown(button, {key: 'ArrowDown'});
      fireEvent.keyDown(button, {key: 'ArrowDown'});

      const searchBar = screen.getByLabelText('Search');
      fireEvent.keyDown(searchBar, {key: 'ArrowDown'});

      await click(document.activeElement as HTMLElement);
      expect(onSelect).toHaveBeenLastCalledWith('link1');

      fireEvent.keyDown(document.activeElement as HTMLElement, {
        key: 'ArrowDown',
      });

      await click(document.activeElement as HTMLElement);
      expect(onSelect).toHaveBeenLastCalledWith('link2');

      fireEvent.keyDown(document.activeElement as HTMLElement, {
        key: 'ArrowDown',
      });

      expect(document.activeElement?.nodeName).toBe('INPUT');
    });

    it('should allow users to ascend through active menu items with the up key', async () => {
      const {onSelect} = renderTestbed();

      const button = screen.getByText('Button');

      fireEvent.keyDown(button, {key: 'ArrowDown'});
      fireEvent.keyDown(button, {key: 'ArrowDown'});

      fireEvent.keyDown(document.activeElement as HTMLElement, {
        key: 'ArrowUp',
      });

      await click(document.activeElement as HTMLElement);
      expect(onSelect).toHaveBeenLastCalledWith('link2');

      fireEvent.keyDown(document.activeElement as HTMLElement, {
        key: 'ArrowUp',
      });

      await click(document.activeElement as HTMLElement);
      expect(onSelect).toHaveBeenLastCalledWith('link1');
    });

    it('should allow users to close the menu from an item by pressing escape', async () => {
      renderTestbed();

      const button = screen.getByText('Button');

      fireEvent.keyDown(button, {key: 'ArrowDown'});
      fireEvent.keyDown(button, {key: 'ArrowDown'});

      fireEvent.keyDown(document.activeElement as HTMLElement, {
        key: 'Escape',
      });

      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      expect(screen.getByTestId('DropdownButton__button')).toHaveFocus();
    });
  });

  describe('search', () => {
    const TestBed = (props: DropdownProps) => {
      const [branch, setBranch] = useState('master');

      const items: DropdownItem[] = [
        {id: '0', value: 'master', content: 'master'},
        {id: '1', value: 'staging', content: 'staging'},
        {id: '2', value: 'development', content: 'development'},
      ];

      return (
        <SearchableDropdown
          selectedId={branch}
          onSelect={setBranch}
          items={items}
          emptyResultsContent={'No results found.'}
          {...props}
        >
          Select branch
        </SearchableDropdown>
      );
    };

    it('should filter results when search bar is present', async () => {
      render(<TestBed />);

      await click(screen.getByText('Select branch'));

      const searchBar = screen.getByLabelText('Search');

      await type(searchBar, 'mas');

      expect(screen.getByText('master')).toBeInTheDocument();
      expect(screen.queryByText('staging')).not.toBeInTheDocument();
      expect(screen.queryByText('development')).not.toBeInTheDocument();
    });

    it('should allow users to pass a custom filter', async () => {
      render(<TestBed filter={({id}, value) => id === value} />);

      await click(screen.getByText('Select branch'));

      const searchBar = screen.getByLabelText('Search');

      await type(searchBar, '2');

      expect(screen.getByText('development')).toBeInTheDocument();
      expect(screen.queryByText('master')).not.toBeInTheDocument();
      expect(screen.queryByText('staging')).not.toBeInTheDocument();
    });

    it('should allow user to clear search input with clear button', async () => {
      render(<TestBed />);

      await click(screen.getByText('Select branch'));

      const searchBar = screen.getByLabelText('Search');
      const clearButton = screen.getByLabelText('Clear');

      await type(searchBar, 'stuff');
      expect(searchBar).toHaveValue('stuff');

      await click(clearButton);

      expect(searchBar).toHaveValue('');
    });

    it('should display empty result component when all results are filtered', async () => {
      render(<TestBed />);

      await click(screen.getByText('Select branch'));

      const searchBar = screen.getByLabelText('Search');

      await type(searchBar, 'what');

      expect(screen.getByText('No results found.')).toBeInTheDocument();
    });
  });
});
