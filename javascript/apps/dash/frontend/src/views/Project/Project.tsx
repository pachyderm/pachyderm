import React, {useState} from 'react';
import {Redirect} from 'react-router';

import View from '@dash-frontend/components/View';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import DAG from './components/DAG';
import ProjectHeader from './components/ProjectHeader';
import ProjectSidebar from './components/ProjectSidebar';
import {NODE_HEIGHT, NODE_WIDTH} from './constants/nodeSizes';
import {useProjectView} from './hooks/useProjectView';
import {dagRoute} from './utils/routes';

const Project: React.FC = () => {
  const {dags, error, loading} = useProjectView(NODE_WIDTH, NODE_HEIGHT);
  const {dagId, projectId} = useUrlState();
  const [largestDagScale, setLargestDagScale] = useState<number | null>(null);

  if (error)
    return (
      <View>
        <h1>{JSON.stringify(error)}</h1>
      </View>
    );
  if (loading || !dags)
    return (
      <View>
        <h1>Loading...</h1>
      </View>
    );

  const dagFromRoute = dags.find((dag) => dag.id === dagId);
  const dagsToShow = dagFromRoute ? [dagFromRoute] : dags;

  if (!dagFromRoute && dagsToShow.length === 1) {
    return <Redirect to={dagRoute({projectId, dagId: dagsToShow[0].id})} />;
  }

  return (
    <>
      <ProjectHeader totalDags={dags.length} />
      <View>
        {dagsToShow.map((dag) => {
          return (
            <DAG
              data={dag}
              key={dag.id}
              id={dag.id}
              nodeWidth={NODE_WIDTH}
              nodeHeight={NODE_HEIGHT}
              count={dagsToShow.length}
              isInteractive={dagsToShow.length === 1}
              largestDagScale={largestDagScale}
              setLargestDagScale={setLargestDagScale}
            />
          );
        })}

        <ProjectSidebar />
      </View>
    </>
  );
};

export default Project;
