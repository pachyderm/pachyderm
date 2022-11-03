import React, {AnchorHTMLAttributes, useRef} from 'react';

import useTab from 'Tabs/hooks/useTab';

import styles from './Tab.module.css';

export interface TabProps extends AnchorHTMLAttributes<HTMLAnchorElement> {
  id: string;
}

const Tab: React.FC<TabProps> = ({children, id, ...rest}) => {
  const ref = useRef<HTMLAnchorElement>(null);
  const {isShown, selectTab, handleKeyDown} = useTab(id, ref);

  return (
    <li role="presentation" className={styles.base}>
      <a
        {...rest}
        ref={ref}
        role="tab"
        onClick={selectTab}
        id={id}
        href={`#${id}`}
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
