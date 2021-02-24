import classNames from 'classnames';
import noop from 'lodash/noop';
import React, {useState, useMemo, useCallback, useRef} from 'react';

import useOutsideClick from 'hooks/useOutsideClick';

import {DropdownButton} from './components/DropdownButton';
import {DropdownMenu} from './components/DropdownMenu';
import {DropdownMenuItem} from './components/DropdownMenuItem';
import DropdownContext from './contexts/DropdownContext';
import styles from './Dropdown.module.css';

export interface DropdownProps {
  className?: string;
  onSelect?: (id: string) => void;
  initialSelectId?: string;
  storeSelected?: boolean;
}

export const Dropdown: React.FC<DropdownProps> = ({
  children,
  className,
  onSelect = noop,
  initialSelectId = '',
  storeSelected = false,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [selectedId, selectId] = useState(initialSelectId);

  const setSelectedId = useCallback(
    (id: string) => {
      if (storeSelected) selectId(id);
    },
    [storeSelected, selectId],
  );

  const ctxValue = useMemo(
    () => ({
      isOpen,
      setIsOpen,
      selectedId,
      setSelectedId,
      onSelect,
    }),
    [isOpen, setIsOpen, selectedId, setSelectedId, onSelect],
  );

  const handleOutsideClick = useCallback(() => {
    if (isOpen) {
      setIsOpen(false);
    }
  }, [isOpen, setIsOpen]);
  useOutsideClick(containerRef, handleOutsideClick);

  return (
    <DropdownContext.Provider value={ctxValue}>
      <div className={classNames(styles.base, className)} ref={containerRef}>
        {children}
      </div>
    </DropdownContext.Provider>
  );
};

export default Object.assign(Dropdown, {
  Button: DropdownButton,
  Menu: DropdownMenu,
  MenuItem: DropdownMenuItem,
});
