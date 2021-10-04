import {useMemo} from 'react';

import {DagDirection} from '@dash-frontend/lib/types';
interface LocalProjectPreferencesProps {
  projectId: string;
}

const useLocalProjectPreferences = ({
  projectId,
}: LocalProjectPreferencesProps) => {
  const dagDirection = useMemo(() => {
    const localDirection = localStorage.getItem(`dag-direction-${projectId}`);
    return localDirection === DagDirection.DOWN
      ? DagDirection.DOWN
      : DagDirection.RIGHT;
  }, [projectId]);

  const sidebarWidth = useMemo(() => {
    const localWidth = localStorage.getItem(`dag-sidebar-width-${projectId}`);
    return localWidth ? +localWidth : null;
  }, [projectId]);

  const getSkipCenterOnSelect = () => {
    const localSkipCenterOnSelect = localStorage.getItem(
      `dag-skip-centering-${projectId}`,
    );
    return localSkipCenterOnSelect === 'true';
  };

  const handleUpdateSidebarWidth = (width: number) => {
    localStorage.setItem(`dag-sidebar-width-${projectId}`, `${width}`);
  };

  const handleUpdateDagDirection = (direction: DagDirection) => {
    localStorage.setItem(`dag-direction-${projectId}`, direction);
  };

  const handleUpdateCenterOnSelect = (skipCentering: boolean) => {
    localStorage.setItem(`dag-skip-centering-${projectId}`, `${skipCentering}`);
  };

  return {
    dagDirection,
    sidebarWidth,
    getSkipCenterOnSelect,
    handleUpdateDagDirection,
    handleUpdateSidebarWidth,
    handleUpdateCenterOnSelect,
  };
};

export default useLocalProjectPreferences;
