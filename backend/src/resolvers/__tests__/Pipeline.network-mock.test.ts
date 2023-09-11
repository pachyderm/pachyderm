import {CREATE_PIPELINE_V2_MUTATION} from '@dash-frontend/mutations/CreatePipelineV2';
import {SET_CLUSTER_DEFAULTS} from '@dash-frontend/mutations/SetClusterDefaults';
import {GET_CLUSTER_DEFAULTS} from '@dash-frontend/queries/GetClusterDefaults';

import {grpcServer} from '@dash-backend/mock/networkMock';
import {PpsAPIService, PpsIAPIServer, Project} from '@dash-backend/proto';
import {
  CreatePipelineV2Response,
  GetClusterDefaultsResponse,
  Pipeline,
  SetClusterDefaultsResponse,
} from '@dash-backend/proto/proto/pps/pps_pb';
import {executeMutation, executeQuery} from '@dash-backend/testHelpers';
import {
  CreatePipelineV2Mutation,
  GetClusterDefaultsResp,
  SetClusterDefaultsResp,
} from '@graphqlTypes';

describe('createPipelineV2', () => {
  it('returns the effective spec', async () => {
    const service: Pick<PpsIAPIServer, 'createPipelineV2'> = {
      createPipelineV2: (_, callback) => {
        const resp =
          new CreatePipelineV2Response().setEffectiveCreatePipelineRequestJson(
            '{"pipeline":{"project":{"name":"default"}, "name":"asdf"}, "transform":{}}',
          );
        callback(null, resp);
      },
    };
    grpcServer.addService(PpsAPIService, service);

    const {data, errors} = await executeMutation<CreatePipelineV2Mutation>(
      CREATE_PIPELINE_V2_MUTATION,
      {
        args: {}, // Inputs are not tested TODO
      },
    );

    expect(errors).toBeUndefined();
    expect(data).toEqual(
      expect.objectContaining({
        createPipelineV2: {
          __typename: 'CreatePipelineV2Response',
          effectiveCreatePipelineRequestJson:
            '{"pipeline":{"project":{"name":"default"}, "name":"asdf"}, "transform":{}}',
        },
      }),
    );
  });
});

describe('getClusterDefaults', () => {
  it('returns cluster defaults', async () => {
    const service: Pick<PpsIAPIServer, 'getClusterDefaults'> = {
      getClusterDefaults: (_, callback) => {
        const resp = new GetClusterDefaultsResponse().setClusterDefaultsJson(
          '"{clusterDefaultsJson": "{"createPipelineRequest":{"resourceRequests":{"cpu":1, "memory":"256Mi", "disk":"1Gi"}, "sidecarResourceRequests":{"cpu":1, "memory":"256Mi", "disk":"1Gi"}}}}"',
        );
        callback(null, resp);
      },
    };
    grpcServer.addService(PpsAPIService, service);

    const {data, errors} = await executeQuery<GetClusterDefaultsResp>(
      GET_CLUSTER_DEFAULTS,
      {},
    );

    expect(errors).toBeUndefined();
    expect(data).toEqual(
      expect.objectContaining({
        getClusterDefaults: {
          __typename: 'GetClusterDefaultsResp',
          clusterDefaultsJson:
            '"{clusterDefaultsJson": "{"createPipelineRequest":{"resourceRequests":{"cpu":1, "memory":"256Mi", "disk":"1Gi"}, "sidecarResourceRequests":{"cpu":1, "memory":"256Mi", "disk":"1Gi"}}}}"',
        },
      }),
    );
  });
});

describe('setClusterDefaults', () => {
  it('returns affected pipelines', async () => {
    const service: Pick<PpsIAPIServer, 'setClusterDefaults'> = {
      setClusterDefaults: (_, callback) => {
        const resp = new SetClusterDefaultsResponse().setAffectedPipelinesList([
          new Pipeline()
            .setName('foo')
            .setProject(new Project().setName('bar')),
        ]);
        callback(null, resp);
      },
    };
    grpcServer.addService(PpsAPIService, service);

    const {data, errors} = await executeMutation<SetClusterDefaultsResp>(
      SET_CLUSTER_DEFAULTS,
      {args: {}},
    );

    expect(errors).toBeUndefined();
    expect(data).toEqual(
      expect.objectContaining({
        setClusterDefaults: {
          __typename: 'SetClusterDefaultsResp',
          affectedPipelinesList: [
            {
              __typename: 'PipelineObject',
              name: 'foo',
              project: {
                __typename: 'ProjectObject',
                name: 'bar',
              },
            },
          ],
        },
      }),
    );
  });
});
