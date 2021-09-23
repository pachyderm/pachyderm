import {useMemo} from 'react';

import {DagDirection} from '@dash-frontend/lib/types';

export interface LocalPreferences {
  dagDirection: DagDirection;
  handleUpdateDagDirection: (direction: DagDirection) => void;
}

interface LocalProjectPreferencesProps {
  projectId: string;
}

const useLocalProjectPreferences = ({
  projectId,
}: LocalProjectPreferencesProps): LocalPreferences => {
  const dagDirection = useMemo(() => {
    const localDirection = localStorage.getItem(`dag-direction-${projectId}`);
    return localDirection === DagDirection.DOWN
      ? DagDirection.DOWN
      : DagDirection.RIGHT;
  }, [projectId]);

  const handleUpdateDagDirection = (direction: DagDirection) => {
    localStorage.setItem(`dag-direction-${projectId}`, direction);
  };

  return {
    dagDirection,
    handleUpdateDagDirection,
  };
};

export default useLocalProjectPreferences;
