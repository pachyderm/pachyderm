import React from 'react';
import {useCombobox} from 'downshift';

export type DropdownComboboxProps = {
  initialSelectedItem?: string | null;
  items: string[];
  onSelectedItemChange: (selectedItem: string | null) => void;
  placeholder?: string;
  testIdPrefix?: string;
};

export const DropdownCombobox: React.FC<DropdownComboboxProps> = ({
  initialSelectedItem,
  items,
  onSelectedItemChange,
  placeholder,
  testIdPrefix,
}) => {
  const [inputItems, setInputItems] = React.useState(items);

  testIdPrefix = testIdPrefix || '';
  const {
    isOpen,
    getMenuProps,
    getInputProps,
    highlightedIndex,
    getItemProps,
    selectedItem,
    selectItem,
  } = useCombobox({
    items: inputItems,
    initialIsOpen: initialSelectedItem ? false : true,
    initialSelectedItem: initialSelectedItem,
    onInputValueChange: ({inputValue}) => {
      setInputItems(
        inputValue
          ? items.filter((item) =>
              item.toLowerCase().startsWith(inputValue.toLowerCase()),
            )
          : items,
      );
    },
    onIsOpenChange: ({isOpen, selectedItem}) => {
      // Clear selectedItem on opening if there is a selected item
      if (isOpen && selectedItem) {
        selectItem(null);
      }
    },
    onSelectedItemChange: ({selectedItem}) => {
      onSelectedItemChange(selectedItem || null);
    },
  });

  return (
    <div className="pachyderm-DropdownCombobox">
      <div className="bp3-input-group jp-InputGroup">
        <input
          {...getInputProps()}
          className="bp3-input"
          placeholder={placeholder}
          data-testid={`${testIdPrefix}DropdownCombobox-input`}
        />
      </div>
      <ul
        {...getMenuProps()}
        data-testid={`${testIdPrefix}DropdownCombobox-ul`}
      >
        {(isOpen || !selectedItem) &&
          inputItems.map((item, index) => (
            <li
              style={{
                backgroundColor: highlightedIndex === index ? '#bde4ff' : '',
              }}
              key={`${item}${index}`}
              {...getItemProps({
                item,
                index,
              })}
              data-testid={`${testIdPrefix}DropdownCombobox-li-${item}`}
            >
              {item}
            </li>
          ))}
      </ul>
    </div>
  );
};

export default DropdownCombobox;
