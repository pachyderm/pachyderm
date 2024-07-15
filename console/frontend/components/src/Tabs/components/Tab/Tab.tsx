import classnames from 'classnames';
import React, {AnchorHTMLAttributes, useRef} from 'react';

import useTab from '@pachyderm/components/Tabs/hooks/useTab';

import styles from './Tab.module.css';

export interface TabProps extends AnchorHTMLAttributes<HTMLAnchorElement> {
  id: string;
  disabled?: boolean;
}

const Tab: React.FC<TabProps> = ({children, id, disabled, ...rest}) => {
  const ref = useRef<HTMLAnchorElement>(null);
  const {isShown, selectTab, handleKeyDown} = useTab(id, ref);

  return (
    <li
      role="presentation"
      className={classnames(styles.base, {[styles.disabled]: disabled})}
    >
      <a
        {...rest}
        ref={ref}
        role="tab"
        onClick={!disabled ? selectTab : undefined}
        id={id}
        href={!disabled ? `#${id}` : undefined}
        aria-selected={isShown}
        tabIndex={!isShown ? -1 : undefined}
        data-testid={`Tab__${id}`}
        onKeyDown={handleKeyDown}
      >
        {children}
      </a>
    </li>
  );
};

export default Tab;
