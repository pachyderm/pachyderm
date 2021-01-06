import {useQuery} from '@apollo/client';
import flatten from 'lodash/flatten';
import {useMemo, useCallback} from 'react';

import {LinkInputData} from 'lib/DAGTypes';
import {Query} from 'lib/graphqlTypes';
import {GET_PACH_QUERY} from 'queries/GetPachQuery';

export const useDAGData = () => {
  const {data: pach, error, loading} = useQuery(GET_PACH_QUERY);

  // Data manipulation to convert API schema to a {node[], link[]} format for d3-force
  const formatDAGData = useCallback((pach: Query) => {
    const nodeRepos = pach.repos.map((r) => ({
      name: `${r.name}_repo`,
      type: 'repo',
      // detect out output repos as those name matching a pipeline
      parents: pach.pipelines.some((pl) => pl.name === r.name) ? [r.name] : [],
    }));

    const nodePipelines = pach.pipelines.map((r) => ({
      name: r.id,
      type: 'pipeline',
      parents: r.inputs.map((i) => `${i.id}_repo`),
    }));

    const allNodes = [...nodeRepos, ...nodePipelines];

    const correspondingIndex: {[key: string]: number} = allNodes.reduce(
      (acc, val, index) => ({
        ...acc,
        [val.name]: index,
      }),
      {},
    );

    const allLinks = allNodes.map((node) =>
      node.parents.reduce((acc, parentName) => {
        const source = correspondingIndex[parentName || ''];
        if (Number.isInteger(source)) {
          return [
            ...acc,
            {
              source: source,
              target: correspondingIndex[node.name],
              error: false,
              active: false,
            },
          ];
        }
        return acc;
      }, [] as LinkInputData[]),
    );

    const dataNodes = allNodes.map((node) => ({
      name: node.name,
      type: node.type,
      error: false,
      access: true,
    }));

    const data = {
      nodes: dataNodes,
      links: flatten(allLinks),
    };

    return data;
  }, []);

  const formattedDAGData = useMemo(() => {
    if (!error && !loading) {
      return formatDAGData(pach);
    }
    return {
      nodes: [],
      links: [],
    };
  }, [pach, loading, error, formatDAGData]);

  return {
    error,
    loading,
    pach,
    formattedDAGData,
  };
};
