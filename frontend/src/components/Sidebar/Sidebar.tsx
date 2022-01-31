import {ButtonLink, CloseSVG} from '@pachyderm/components';
import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';
import {Route} from 'react-router';

import {SidebarSize} from '@dash-frontend/lib/types';
import DeleteRepoButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/DeleteRepoButton';
import {
  PROJECT_REPO_PATH,
  LINEAGE_REPO_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';

import useSidebar from './hooks/useSidebar';
import styles from './Sidebar.module.css';
interface SidebarProps extends HTMLAttributes<HTMLDivElement> {
  overlay?: boolean;
  fixed?: boolean;
  onClose?: () => void;
  defaultSize?: SidebarSize;
  resizable?: boolean;
}

const Sidebar: React.FC<SidebarProps> = ({
  children,
  overlay = false,
  onClose,
  className,
  fixed = false,
  defaultSize = 'sm',
  resizable = false,
  ...rest
}) => {
  const {
    isOpen,
    sidebarWidth,
    dragging,
    setDragging,
    throttleMouseEvent,
    applyMousePosition,
    onDragEnd,
  } = useSidebar({defaultSize});

  return (
    <div
      onMouseMove={throttleMouseEvent(applyMousePosition, 50)}
      onMouseUp={onDragEnd}
      onMouseLeave={onDragEnd}
      className={classnames({
        [styles.resizable]: resizable,
        [styles.active]: dragging,
      })}
    >
      {overlay && (
        <div
          className={classnames(styles.overlay, {
            [styles.open]: isOpen,
          })}
        />
      )}

      <div
        className={classnames(
          styles.base,
          {
            [styles.fixed]: fixed,
            [styles.open]: isOpen,
            [styles[defaultSize]]: true,
          },
          className,
        )}
        style={{width: resizable ? sidebarWidth : '100%'}}
        {...rest}
      >
        {resizable && (
          <div
            className={classnames(styles.dragBar, {[styles.active]: dragging})}
            onMouseDown={() => setDragging(true)}
          />
        )}
        {onClose && (
          <div className={styles.sideBarToolbar}>
            <Route path={[PROJECT_REPO_PATH, LINEAGE_REPO_PATH]}>
              <DeleteRepoButton />
            </Route>
            <ButtonLink className={styles.closeButton} onClick={onClose}>
              <CloseSVG aria-label="Close" className={styles.closeSvg} />
            </ButtonLink>
          </div>
        )}

        {children}
      </div>
    </div>
  );
};

export default Sidebar;
