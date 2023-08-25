import classnames from 'classnames';
import React, {useState, useMemo, useCallback, useEffect} from 'react';

import useBreakpoint from '../../../hooks/useBreakpoint';
import {Icon} from '../../../Icon';
import {PanelLeftSVG, PanelRightSVG} from '../../../Svg';
import SideNavContext from '../../SideNavContext';

import styles from './SideNav.module.css';

type Props = {
  children?: React.ReactNode;
  breakpoint: number;
};

const SideNav: React.FC<Props> = ({children, breakpoint}) => {
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
            <Icon aria-hidden={true} className={styles.collapseIcon} small>
              {minimized ? <PanelRightSVG /> : <PanelLeftSVG />}
            </Icon>
            {!minimized && 'Collapse'}
          </button>
        )}
      </nav>
    </SideNavContext.Provider>
  );
};

export default SideNav;
