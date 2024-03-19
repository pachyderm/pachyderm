import React from 'react';
import {render, fireEvent} from '@testing-library/react';

import DropdownCombobox from '../DropdownCombobox';

describe('DropdownCombobox', () => {
  it('should render placeholder', async () => {
    const {getByTestId} = render(
      <DropdownCombobox
        items={[]}
        placeholder="placeholder"
        onSelectedItemChange={() => {
          // Do nothing
        }}
      />,
    );

    const input = getByTestId('DropdownCombobox-input');
    expect(input).toHaveAttribute('placeholder', 'placeholder');
  });

  it('should select initial item', async () => {
    const {getByTestId} = render(
      <DropdownCombobox
        items={['item1', 'item2', 'item3']}
        initialSelectedItem="item1"
        onSelectedItemChange={() => {
          // Do nothing
        }}
      />,
    );

    const input = getByTestId('DropdownCombobox-input');
    expect(input).toHaveValue('item1');
    const ul = getByTestId('DropdownCombobox-ul');
    expect(ul.children).toHaveLength(0);
  });

  it('should filter items based on input', async () => {
    const {getByTestId} = render(
      <DropdownCombobox
        items={['foo', 'bar']}
        onSelectedItemChange={() => {
          // Do nothing
        }}
      />,
    );

    let ul = getByTestId('DropdownCombobox-ul');
    expect(ul.children).toHaveLength(2);
    const liFoo = getByTestId('DropdownCombobox-li-foo');
    expect(liFoo).toHaveTextContent('foo');
    let liBar = getByTestId('DropdownCombobox-li-bar');
    expect(liBar).toHaveTextContent('bar');

    const input = getByTestId('DropdownCombobox-input');
    fireEvent.change(input, {target: {value: 'ba'}});

    ul = getByTestId('DropdownCombobox-ul');
    expect(ul.children).toHaveLength(1);
    liBar = getByTestId('DropdownCombobox-li-bar');
    expect(liBar).toHaveTextContent('bar');
  });

  it('should select new item', async () => {
    let selectedItem: string | null = null;
    const {getByTestId} = render(
      <DropdownCombobox
        items={['foo', 'bar']}
        onSelectedItemChange={(newSelectedItem) => {
          selectedItem = newSelectedItem;
        }}
      />,
    );

    const input = getByTestId('DropdownCombobox-input');
    fireEvent.change(input, {target: {value: 'ba'}});
    fireEvent.click(getByTestId('DropdownCombobox-li-bar'));
    expect(selectedItem).toBe('bar');
  });

  it('should be open by default if no initial selected item', async () => {
    const {getByTestId} = render(
      <DropdownCombobox
        items={['foo', 'bar']}
        onSelectedItemChange={(newSelectedItem) => {
          // Do nothing
        }}
      />,
    );

    const ul = getByTestId('DropdownCombobox-ul');
    expect(ul.children).toHaveLength(2);
  });

  it('should clear selected item if opened with a selected item', async () => {
    let selectedItem: string | null = null;
    const {getByTestId} = render(
      <DropdownCombobox
        items={['foo', 'bar']}
        onSelectedItemChange={(newSelectedItem) => {
          selectedItem = newSelectedItem;
        }}
      />,
    );

    const input = getByTestId('DropdownCombobox-input');
    fireEvent.change(input, {target: {value: 'ba'}});
    fireEvent.click(getByTestId('DropdownCombobox-li-bar'));
    expect(selectedItem).toBe('bar');

    fireEvent.click(getByTestId('DropdownCombobox-input'));
    expect(selectedItem).toBeNull();
  });

  it('should highlight items on mouse over', async () => {
    const {getByTestId} = render(
      <DropdownCombobox
        items={['foo', 'bar']}
        onSelectedItemChange={(newSelectedItem) => {
          // Do nothing
        }}
      />,
    );

    const liBar = getByTestId('DropdownCombobox-li-bar');
    fireEvent.mouseOver(liBar);
    expect(liBar).toHaveStyle('background-color: bde4ff');
  });

  it('should set selected item to selectItem parameter in onSelectedItemChange', async () => {
    let selectedItem: string | null = null;
    const {getByTestId} = render(
      <DropdownCombobox
        items={['foo', 'bar']}
        onSelectedItemChange={(newSelectedItem, selectItem) => {
          selectedItem = newSelectedItem;
          if (newSelectedItem) {
            selectItem(null);
          }
        }}
      />,
    );

    const input = getByTestId('DropdownCombobox-input');
    fireEvent.change(input, {target: {value: 'ba'}});
    fireEvent.click(getByTestId('DropdownCombobox-li-bar'));
    expect(selectedItem).toBeNull();
  });
});
