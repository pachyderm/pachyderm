import {Button, ButtonGroup, CloseSVG} from '@pachyderm/components';
import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';
import {Route} from 'react-router';

import {SidebarSize} from '@dash-frontend/lib/types';
import DeletePipelineButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/DeletePipelineButton';
import DeleteRepoButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/DeleteRepoButton';
import ReadLogsButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/ReadLogsButton';
import UploadFilesButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/UploadFilesButton';
import {
  PROJECT_PATH,
  LINEAGE_PATH,
  PROJECT_REPO_PATH,
  PROJECT_PIPELINE_PATH,
  LINEAGE_REPO_PATH,
  LINEAGE_PIPELINE_PATH,
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
      className={classnames(styles.container, {
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
        <Route path={[PROJECT_PATH, LINEAGE_PATH]}>
          <div className={styles.sideBarToolbar}>
            <ButtonGroup>
              <Route path={[PROJECT_REPO_PATH, LINEAGE_REPO_PATH]}>
                <UploadFilesButton />
                <DeleteRepoButton />
              </Route>
              <Route path={[PROJECT_PIPELINE_PATH, LINEAGE_PIPELINE_PATH]}>
                <ReadLogsButton />
              </Route>
              <Route path={[PROJECT_PIPELINE_PATH, LINEAGE_PIPELINE_PATH]}>
                <DeletePipelineButton />
              </Route>
              {onClose && (
                <Button
                  IconSVG={CloseSVG}
                  onClick={onClose}
                  aria-label="Close"
                  color="black"
                  buttonType="ghost"
                />
              )}
            </ButtonGroup>
          </div>
        </Route>

        {children}
      </div>
    </div>
  );
};

export default Sidebar;
