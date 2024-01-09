import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';
import {Route} from 'react-router';

import PipelineActionsMenu from '@dash-frontend/components/PipelineActionsMenu';
import RepoActionsMenu from '@dash-frontend/components/RepoActionsMenu';
import {DEFAULT_SIDEBAR_SIZE} from '@dash-frontend/hooks/useSidebarInfo';
import InspectCommitsButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/InspectCommitsButton';
import ReadLogsButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/ReadLogsButton';
import {
  PROJECT_PATH,
  LINEAGE_PATH,
  LINEAGE_REPO_PATH,
  LINEAGE_PIPELINE_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {Button, ButtonGroup, CloseSVG, Group} from '@pachyderm/components';

import useSidebar from './hooks/useSidebar';
import styles from './Sidebar.module.css';
interface SidebarProps extends HTMLAttributes<HTMLDivElement> {
  overlay?: boolean;
  fixed?: boolean;
  onClose?: () => void;
  resizable?: boolean;
}

const Sidebar: React.FC<SidebarProps> = ({
  children,
  overlay = false,
  onClose,
  className,
  fixed = false,
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
    pipelineId,
    repoId,
  } = useSidebar();

  return (
    <section
      aria-label="project-sidebar"
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
          },
          className,
        )}
        style={{width: resizable ? sidebarWidth : DEFAULT_SIDEBAR_SIZE}}
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
              <Route path={LINEAGE_PIPELINE_PATH}>
                <Group spacing={8}>
                  <ReadLogsButton />
                  <PipelineActionsMenu pipelineId={pipelineId} />
                </Group>
              </Route>
              <Route path={LINEAGE_REPO_PATH}>
                <Group spacing={8}>
                  <InspectCommitsButton />
                  <RepoActionsMenu repoId={repoId} />
                </Group>
              </Route>
            </ButtonGroup>
            <ButtonGroup>
              <Route path={[LINEAGE_PIPELINE_PATH, LINEAGE_REPO_PATH]}>
                {onClose && (
                  <Button
                    IconSVG={CloseSVG}
                    onClick={onClose}
                    aria-label="Close sidebar"
                    color="black"
                    buttonType="ghost"
                  />
                )}
              </Route>
            </ButtonGroup>
          </div>
        </Route>

        {children}
      </div>
    </section>
  );
};

export default Sidebar;
