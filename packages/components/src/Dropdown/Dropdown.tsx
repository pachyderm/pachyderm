import classNames from 'classnames';
import noop from 'lodash/noop';
import React, {useState, useMemo, useCallback, useRef} from 'react';
import {FormProvider, useForm, UseFormReturn} from 'react-hook-form';

import useOutsideClick from 'hooks/useOutsideClick';

import {DropdownButton, DropdownButtonProps} from './components/DropdownButton';
import {DropdownMenu, DropdownMenuProps} from './components/DropdownMenu';
import DropdownMenuEmptyResults from './components/DropdownMenuEmptyResults';
import {DropdownMenuEmptyResultsProps} from './components/DropdownMenuEmptyResults/DropdownMenuEmptyResults';
import {
  DropdownMenuItem,
  DropdownMenuItemProps,
} from './components/DropdownMenuItem';
import DropdownMenuSearchBar from './components/DropdownMenuSearchBar';
import {DropdownMenuSearchBarProps} from './components/DropdownMenuSearchBar/DropdownMenuSearchBar';
import DropdownContext from './contexts/DropdownContext';
import FilteredResultsContext from './contexts/FilteredResultsContext';
import SelectedIdContext from './contexts/SelectedIdContext';
import styles from './Dropdown.module.css';

export interface ItemObject {
  id: string;
  value?: string;
}
export interface DropdownProps {
  className?: string;
  onSelect?: (id: string) => void;
  initialSelectId?: string;
  storeSelected?: boolean;
  filter?: (items: ItemObject, searchValue: string) => boolean;
  selectedId?: string;
  formCtx?: UseFormReturn;
}

const defaultFilter = (item: ItemObject, searchValue: string) => {
  return (
    !searchValue ||
    (item.value || item.id).startsWith(searchValue.toLowerCase())
  );
};

export const Dropdown: React.FC<DropdownProps> = ({
  children,
  className,
  onSelect = noop,
  initialSelectId = '',
  storeSelected = false,
  filter = defaultFilter,
  formCtx,
  selectedId: managedSelectedId,
}) => {
  const defaultFormContext = useForm();
  const containerRef = useRef<HTMLDivElement>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [internalSelectedId, selectId] = useState(initialSelectId);
  const [filteredResults, setFilteredResults] = useState<ItemObject[]>([]);

  const selectedId = managedSelectedId || internalSelectedId;

  const setSelectedId = useCallback(
    (id: string) => {
      if (storeSelected) selectId(id);

      onSelect(id);
    },
    [storeSelected, selectId, onSelect],
  );

  const dropdownContext = useMemo(
    () => ({
      isOpen,
      setIsOpen,
      filter,
      setFilteredResults,
    }),
    [isOpen, setIsOpen, filter, setFilteredResults],
  );

  const selectedIdContext = useMemo(
    () => ({
      selectedId,
      setSelectedId,
    }),
    [selectedId, setSelectedId],
  );

  const filteredResultsContext = useMemo(
    () => ({
      filteredResults,
    }),
    [filteredResults],
  );

  const handleOutsideClick = useCallback(() => {
    if (isOpen) {
      setIsOpen(false);
    }
  }, [isOpen, setIsOpen]);
  useOutsideClick(containerRef, handleOutsideClick);

  return (
    <DropdownContext.Provider value={dropdownContext}>
      <SelectedIdContext.Provider value={selectedIdContext}>
        <FilteredResultsContext.Provider value={filteredResultsContext}>
          <FormProvider {...defaultFormContext} {...formCtx}>
            <div
              className={classNames(styles.base, className)}
              ref={containerRef}
            >
              {children}
            </div>
          </FormProvider>
        </FilteredResultsContext.Provider>
      </SelectedIdContext.Provider>
    </DropdownContext.Provider>
  );
};

export type DropdownItem = ItemObject &
  Omit<DropdownMenuItemProps, 'children'> & {
    content: Required<DropdownMenuItemProps['children']>;
  };
interface DefaultDropdownProps extends DropdownProps {
  items: DropdownItem[];
  buttonOpts?: DropdownButtonProps;
  menuOpts?: DropdownMenuProps;
}

export const DefaultDropdown: React.FC<DefaultDropdownProps> = ({
  items,
  children,
  buttonOpts,
  menuOpts,
  ...rest
}) => {
  return (
    <Dropdown {...rest}>
      <DropdownButton {...buttonOpts}>{children}</DropdownButton>

      <DropdownMenu {...menuOpts}>
        {items.map(({id, content, ...props}) => (
          <DropdownMenuItem key={id} id={id} {...props}>
            {content}
          </DropdownMenuItem>
        ))}
      </DropdownMenu>
    </Dropdown>
  );
};

interface SearchableDropdownProps extends DefaultDropdownProps {
  searchOpts?: DropdownMenuSearchBarProps;
  emptyResultsContent: Required<DropdownMenuEmptyResultsProps['children']>;
}

export const SearchableDropdown: React.FC<SearchableDropdownProps> = ({
  items,
  children,
  buttonOpts,
  menuOpts,
  searchOpts,
  emptyResultsContent,
  ...rest
}) => {
  return (
    <Dropdown {...rest}>
      <DropdownButton {...buttonOpts}>{children}</DropdownButton>

      <DropdownMenu {...menuOpts}>
        <DropdownMenuSearchBar {...searchOpts} />

        {items.map(({id, content, ...props}) => (
          <DropdownMenuItem key={id} id={id} {...props}>
            {content}
          </DropdownMenuItem>
        ))}

        <DropdownMenuEmptyResults>
          {emptyResultsContent}
        </DropdownMenuEmptyResults>
      </DropdownMenu>
    </Dropdown>
  );
};

export default Object.assign(Dropdown, {
  Button: DropdownButton,
  Menu: DropdownMenu,
  MenuItem: DropdownMenuItem,
  Search: DropdownMenuSearchBar,
  EmptyResults: DropdownMenuEmptyResults,
});
