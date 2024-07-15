import React, {useEffect, useState, useCallback} from 'react';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useSidebarInfo from '@dash-frontend/hooks/useSidebarInfo';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const SIDEBAR_MIN_WIDTH = 304;
const SIDEBAR_MAX_WIDTH = 700;

const useSidebar = () => {
  const [isOpen, setIsOpen] = useState(false);
  const {projectId, pipelineId, repoId} = useUrlState();
  const [sidebarWidthSetting, handleUpdateSidebarWidth] =
    useLocalProjectSettings({projectId, key: 'sidebar_width'});
  const {sidebarSize} = useSidebarInfo();
  const [sidebarWidth, setSidebarWidth] = useState(sidebarSize);
  useEffect(() => {
    if (!sidebarWidthSetting) {
      setSidebarWidth(sidebarSize);
    }
  }, [sidebarSize, sidebarWidthSetting]);

  const [dragging, setDragging] = useState(false);

  const throttleMouseEvent = useCallback(
    (
      callback: (e: React.MouseEvent<HTMLDivElement, MouseEvent>) => void,
      interval: number,
    ) => {
      let enableCall = true;

      return (e: React.MouseEvent<HTMLDivElement, MouseEvent>) => {
        if (!enableCall || !dragging) return;

        enableCall = false;
        callback(e);
        setTimeout(() => (enableCall = true), interval);
      };
    },
    [dragging],
  );

  const applyMousePosition = useCallback(
    (e: React.MouseEvent<HTMLDivElement, MouseEvent>) => {
      const newWidth = window.innerWidth - e.clientX;
      if (newWidth >= SIDEBAR_MIN_WIDTH && newWidth <= SIDEBAR_MAX_WIDTH) {
        setSidebarWidth(newWidth);
      }
    },
    [],
  );

  const onDragEnd = useCallback(() => {
    if (dragging) {
      handleUpdateSidebarWidth(sidebarWidth);
    }
    setDragging(false);
  }, [dragging, handleUpdateSidebarWidth, sidebarWidth]);

  useEffect(() => {
    setIsOpen(true);
  }, []);

  return {
    isOpen,
    sidebarWidth,
    dragging,
    setDragging,
    throttleMouseEvent,
    applyMousePosition,
    onDragEnd,
    pipelineId,
    repoId,
  };
};

export default useSidebar;
