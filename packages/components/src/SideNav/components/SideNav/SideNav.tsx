import classnames from 'classnames';
import React, {useState, useMemo, useCallback, useEffect} from 'react';

import useBreakpoint from '../../../hooks/useBreakpoint';
import {ChevronDoubleRightSVG} from '../../../Svg';
import SideNavContext from '../../SideNavContext';

import styles from './SideNav.module.css';

type Props = {
  breakpoint: number;
  styleMode?: 'light' | 'dark';
};

const SideNav: React.FC<Props> = ({
  children,
  breakpoint,
  styleMode = 'dark',
}) => {
  const [minimized, setMinimized] = useState(false);
  const isMobile = useBreakpoint(breakpoint);

  const sideNavContext = useMemo(
    () => ({
      minimized,
      setMinimized,
    }),
    [minimized, setMinimized],
  );

  const toggleMinimized = useCallback(
    (e: React.SyntheticEvent) => {
      e.preventDefault();
      if (!isMobile) setMinimized(!minimized);
    },
    [setMinimized, minimized, isMobile],
  );

  useEffect(() => {
    if (isMobile) {
      setMinimized(true);
    }
  }, [setMinimized, isMobile]);

  const className = classnames(styles.base, {
    [styles.collapsed]: minimized,
    [styles[styleMode]]: true,
  });

  return (
    <SideNavContext.Provider value={sideNavContext}>
      <nav data-testid="SideNav__nav" className={className}>
        {children}

        {!isMobile && (
          <button
            className={styles.collapse}
            onClick={toggleMinimized}
            data-testid="SideNav__toggle"
            aria-label={`${minimized ? 'Open' : 'Close'} navigation`}
          >
            <div aria-hidden={true} className={styles.collapseIcon}>
              <ChevronDoubleRightSVG
                className={!minimized ? styles.flipSVG : ''}
              />
            </div>
            {!minimized && 'Collapse'}
          </button>
        )}
      </nav>
    </SideNavContext.Provider>
  );
};

export default SideNav;
