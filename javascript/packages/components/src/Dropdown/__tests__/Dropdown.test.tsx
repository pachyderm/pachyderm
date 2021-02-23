import {render, waitFor, act, fireEvent} from '@testing-library/react';
import React from 'react';

import {click, type} from 'testHelpers';

import Dropdown from '../';

describe('Dropdown', () => {
  const renderTestbed = ({
    closeOnClick = false,
  }: {closeOnClick?: boolean} = {}) => {
    const onSelect = jest.fn();

    const renderResults = render(
      <>
        <div>Outside</div>

        <Dropdown onSelect={onSelect}>
          <Dropdown.Button data-testid={'Dropdown__button'}>
            Button
          </Dropdown.Button>

          <Dropdown.Menu>
            <Dropdown.MenuItem id="disabled1" disabled>
              Disabled 1
            </Dropdown.MenuItem>
            <Dropdown.MenuItem id="link1" closeOnClick={closeOnClick}>
              Link 1
            </Dropdown.MenuItem>
            <Dropdown.MenuItem id="link2" closeOnClick={closeOnClick}>
              Link 2
            </Dropdown.MenuItem>
          </Dropdown.Menu>
        </Dropdown>
      </>,
    );

    return {
      ...renderResults,
      onSelect,
    };
  };

  it('should toggle dropdown when DropdownButton is clicked', () => {
    const {queryByRole, getByText} = renderTestbed();

    const button = getByText('Button');

    expect(queryByRole('menu')).toBeNull();
    click(button);
    expect(queryByRole('menu')).not.toBeNull();
    click(button);
    expect(queryByRole('menu')).toBeNull();
  });

  it('should close the dropdown if a user clicks away from the menu', async () => {
    const {queryByRole, getByText} = renderTestbed();

    const button = getByText('Button');
    const outside = getByText('Outside');

    click(button);
    await waitFor(() => expect(queryByRole('menu')).not.toBeNull());

    act(() => {
      click(outside);
    });

    await waitFor(() => expect(queryByRole('menu')).toBeNull());
  });

  it('should invoke onSelect callback with the selected element', () => {
    const {onSelect, getByText} = renderTestbed();

    const button = getByText('Button');
    const link1 = getByText('Link 1');
    const link2 = getByText('Link 2');

    click(button);
    click(link1);
    expect(onSelect).toHaveBeenCalledWith('link1');
    click(link2);
    expect(onSelect).toHaveBeenCalledWith('link2');
  });

  it('should optionally close the dropdown menu on click of an item', async () => {
    const {getByText, queryByRole, getByTestId} = renderTestbed({
      closeOnClick: true,
    });

    const button = getByText('Button');
    const link1 = getByText('Link 1');

    click(button);
    click(link1);

    await waitFor(() => expect(queryByRole('menu')).toBeNull());
    expect(getByTestId('Dropdown__button')).toHaveFocus();
  });

  describe('keyboard interactions', () => {
    it('should allow users to open the dropdown with the down key', () => {
      const {queryByRole, getByText} = renderTestbed();

      const button = getByText('Button');

      fireEvent.keyDown(button, {key: 'ArrowDown'});
      expect(queryByRole('menu')).not.toBeNull();
    });

    it('should allow users to close the dropdown with the up key', () => {
      const {queryByRole, getByText} = renderTestbed();

      const button = getByText('Button');

      click(button);

      fireEvent.keyDown(button, {key: 'ArrowUp'});
      expect(queryByRole('menu')).toBeNull();
    });

    it('should allow users to close the dropdown with the escape key', () => {
      const {queryByRole, getByText} = renderTestbed();

      const button = getByText('Button');

      click(button);

      fireEvent.keyDown(button, {key: 'Escape'});
      expect(queryByRole('menu')).toBeNull();
    });

    it('should allow users to navigate to the first active item with the down key', async () => {
      const {onSelect, getByText} = renderTestbed();

      const button = getByText('Button');

      fireEvent.keyDown(button, {key: 'ArrowDown'});
      fireEvent.keyDown(button, {key: 'ArrowDown'});

      await type(document.activeElement as HTMLElement, '{enter}');
      expect(onSelect).toHaveBeenCalledWith('link1');
    });

    it('should allow users to descend through active menu items with the down key', () => {
      const {onSelect, getByText} = renderTestbed();

      const button = getByText('Button');

      fireEvent.keyDown(button, {key: 'ArrowDown'});
      fireEvent.keyDown(button, {key: 'ArrowDown'});

      click(document.activeElement as HTMLElement);
      expect(onSelect).toHaveBeenLastCalledWith('link1');

      fireEvent.keyDown(document.activeElement as HTMLElement, {
        key: 'ArrowDown',
      });

      click(document.activeElement as HTMLElement);
      expect(onSelect).toHaveBeenLastCalledWith('link2');

      fireEvent.keyDown(document.activeElement as HTMLElement, {
        key: 'ArrowDown',
      });

      click(document.activeElement as HTMLElement);
      expect(onSelect).toHaveBeenLastCalledWith('link1');
    });

    it('should allow users to ascend through active menu items with the up key', () => {
      const {onSelect, getByText} = renderTestbed();

      const button = getByText('Button');

      fireEvent.keyDown(button, {key: 'ArrowDown'});
      fireEvent.keyDown(button, {key: 'ArrowDown'});

      fireEvent.keyDown(document.activeElement as HTMLElement, {
        key: 'ArrowUp',
      });

      click(document.activeElement as HTMLElement);
      expect(onSelect).toHaveBeenLastCalledWith('link2');

      fireEvent.keyDown(document.activeElement as HTMLElement, {
        key: 'ArrowUp',
      });

      click(document.activeElement as HTMLElement);
      expect(onSelect).toHaveBeenLastCalledWith('link1');
    });

    it('should allow users to close the menu from an item by pressing escape', async () => {
      const {getByText, queryByRole, getByTestId} = renderTestbed();

      const button = getByText('Button');

      fireEvent.keyDown(button, {key: 'ArrowDown'});
      fireEvent.keyDown(button, {key: 'ArrowDown'});

      await type(document.activeElement as HTMLElement, '{esc}');

      expect(queryByRole('menu')).toBeNull();
      expect(getByTestId('Dropdown__button')).toHaveFocus();
    });
  });
});
