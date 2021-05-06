import {useParams, useHistory} from 'react-router';

import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {DagDirection} from '@graphqlTypes';

export const useProjectView = (nodeWidth: number, nodeHeight: number) => {
  const {projectId} = useParams<{projectId: string}>();
  const {viewState, setUrlFromViewState} = useUrlQueryState();
  const browserHistory = useHistory();

  const dagDirection = viewState.dagDirection || DagDirection.RIGHT;

  const rotateDag = () => {
    switch (dagDirection) {
      case DagDirection.DOWN:
        browserHistory.push(
          setUrlFromViewState({dagDirection: DagDirection.LEFT}),
        );
        break;
      case DagDirection.LEFT:
        browserHistory.push(
          setUrlFromViewState({dagDirection: DagDirection.UP}),
        );
        break;
      case DagDirection.UP:
        browserHistory.push(
          setUrlFromViewState({dagDirection: DagDirection.RIGHT}),
        );
        break;
      case DagDirection.RIGHT:
        browserHistory.push(
          setUrlFromViewState({dagDirection: DagDirection.DOWN}),
        );
        break;
    }
  };

  const {dags, loading, error} = useProjectDagsData({
    projectId,
    nodeHeight,
    nodeWidth,
    direction: dagDirection,
  });

  return {
    dags,
    error,
    loading,
    rotateDag,
    dagDirection,
  };
};
