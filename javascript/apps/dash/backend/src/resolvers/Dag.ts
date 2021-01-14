import {Input} from '@pachyderm/proto/pb/client/pps/pps_pb';
import flatten from 'lodash/flatten';
import flattenDeep from 'lodash/flattenDeep';
import keyBy from 'lodash/keyBy';

import {NodeType, QueryResolvers} from 'generated/types';
import pfs from 'grpc/pfs';
import pps from 'grpc/pps';
import {LinkInputData} from 'lib/types';

interface DagResolver {
  Query: {
    dag: QueryResolvers['dag'];
  };
}

const flattenPipelineInput = (input: Input.AsObject): string[] => {
  const result = [];

  // TODO: Deal with git and cron input types
  if (input.pfs) {
    result.push(`${input.pfs.repo}_repo`);
  }

  // TODO: Update to indicate which elemets are crossed, unioned, joined, grouped, with.
  return flattenDeep([
    result,
    input.crossList.map((i) => flattenPipelineInput(i)),
    input.unionList.map((i) => flattenPipelineInput(i)),
    input.joinList.map((i) => flattenPipelineInput(i)),
    input.groupList.map((i) => flattenPipelineInput(i)),
  ]);
};

const dagResolver: DagResolver = {
  Query: {
    dag: async (_field, _args, {pachdAddress = '', authToken = ''}) => {
      // TODO: Error handling
      const [repos, pipelines] = await Promise.all([
        pfs(pachdAddress, authToken).listRepo(),
        pps(pachdAddress, authToken).listPipeline(),
      ]);

      const pipelineMap = keyBy(pipelines, (p) => p.pipeline?.name);

      const repoNodes = repos.map((r) => ({
        name: `${r.repo?.name}_repo`,
        type: NodeType.Repo,
        // detect out output repos as those name matching a pipeline
        parents: r.repo && pipelineMap[r.repo.name] ? [r.repo.name] : [],
      }));

      const pipelineNodes = pipelines.map((p) => ({
        name: p.pipeline?.name || '',
        type: NodeType.Pipeline,
        parents: p.input ? flattenPipelineInput(p.input) : [],
      }));

      const allNodes = [...repoNodes, ...pipelineNodes];

      // Calculate the indicies of the nodes in the DAG, as
      // the "link" object requires the source & target attributes
      // to reference the index from the "nodes" list. This prevents
      // us from needing to run "indexOf" on each individual node
      // when building the list of links for the client.
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

      return {
        nodes: dataNodes,
        links: flatten(allLinks),
      };
    },
  },
};

export default dagResolver;
