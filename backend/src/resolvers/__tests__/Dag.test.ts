import {ObservableSubscription} from '@apollo/client';
import {GET_DAG_QUERY} from '@dash-frontend/queries/GetDagQuery';
import {GET_DAGS_QUERY} from '@dash-frontend/queries/GetDagsQuery';

import {
  createSubscriptionClients,
  executeQuery,
  mockServer,
} from '@dash-backend/testHelpers';
import {Vertex} from '@graphqlTypes';

describe('Dag resolver', () => {
  let subscription: ObservableSubscription | null = null;

  afterAll(() => {
    if (subscription) {
      subscription.unsubscribe();
      subscription = null;
    }
  });
  it('should resolve dag data', async () => {
    const {data} = await executeQuery<{dag: Vertex[]}>(GET_DAG_QUERY, {
      args: {
        projectId: 'Solar-Panel-Data-Sorting',
      },
    });

    const vertices = data?.dag;
    expect(vertices).toHaveLength(4);
    expect(vertices?.[0]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'acb880d6f43bb784017bed136a1418072252cc9a',
        project: 'Solar-Panel-Data-Sorting',
        name: 'images',
        state: null,
        access: true,
        parents: [],
        type: 'REPO',
        jobState: null,
        createdAt: 1614116189,
      }),
    );
    expect(vertices?.[1]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'fe5ee514dd88c55a06181b12fdaf7b29e40982e9',
        project: 'Solar-Panel-Data-Sorting',
        name: 'montage',
        state: 'PIPELINE_FAILURE',
        nodeState: 'ERROR',
        access: true,
        parents: [
          {
            id: '48d741f82638e73365b8232d3f3de144fab28eab',
            project: 'Solar-Panel-Data-Sorting',
            name: 'edges',
            __typename: 'VertexIdentifier',
          },
          {
            id: 'acb880d6f43bb784017bed136a1418072252cc9a',
            project: 'Solar-Panel-Data-Sorting',
            name: 'images',
            __typename: 'VertexIdentifier',
          },
        ],
        type: 'PIPELINE',
        jobNodeState: 'RUNNING',
        jobState: 'JOB_CREATED',
        createdAt: null,
      }),
    );
    expect(vertices?.[2]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: '6399ad582692c0ed48e6aee1f139a36db0da0fd0',
        project: 'Solar-Panel-Data-Sorting',
        name: 'https://egress.com',
        state: null,
        access: true,
        parents: [
          {
            id: 'fe5ee514dd88c55a06181b12fdaf7b29e40982e9',
            project: 'Solar-Panel-Data-Sorting',
            name: 'montage',
            __typename: 'VertexIdentifier',
          },
        ],
        type: 'EGRESS',
        jobState: null,
        createdAt: null,
      }),
    );
    expect(vertices?.[3]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: '48d741f82638e73365b8232d3f3de144fab28eab',
        project: 'Solar-Panel-Data-Sorting',
        name: 'edges',
        state: 'PIPELINE_RUNNING',
        nodeState: 'IDLE',
        access: true,
        parents: [
          {
            __typename: 'VertexIdentifier',
            id: 'acb880d6f43bb784017bed136a1418072252cc9a',
            name: 'images',
            project: 'Solar-Panel-Data-Sorting',
          },
        ],
        type: 'PIPELINE',
        jobState: 'JOB_CREATED',
        jobNodeState: 'RUNNING',
        createdAt: null,
      }),
    );
  });

  it('should correctly return access data to a given node', async () => {
    mockServer.setAccount('2');

    const {data} = await executeQuery<{dag: Vertex[]}>(GET_DAG_QUERY, {
      args: {
        projectId: 'Solar-Panel-Data-Sorting',
      },
    });

    const imagesRepo = data?.dag.find(
      (vertex) => vertex.id === 'acb880d6f43bb784017bed136a1418072252cc9a',
    );

    const montagePipeline = data?.dag.find(
      (vertex) => vertex.id === 'fe5ee514dd88c55a06181b12fdaf7b29e40982e9',
    );

    expect(imagesRepo?.access).toBe(true);
    expect(montagePipeline?.access).toBe(false);
  });

  it('should correctly filter sub-dag for jobsets', async () => {
    const {observable} = createSubscriptionClients<{
      data: {dags: Vertex[]};
    }>(GET_DAGS_QUERY, {
      args: {
        projectId: 'Solar-Panel-Data-Sorting',
        jobSetId: '33b9af7d5d4343219bc8e02ff44cd55a',
      },
    });

    await new Promise<void>((resolve) => {
      subscription = observable.subscribe({
        next: (data: {data: {dags: Vertex[]}}) => {
          const vertices = data.data?.dags;

          expect(vertices).toHaveLength(3);
          expect(vertices?.[0]).toEqual(
            expect.objectContaining({
              id: '48d741f82638e73365b8232d3f3de144fab28eab',
              project: 'Solar-Panel-Data-Sorting',
              name: 'edges',
              state: null,
              nodeState: null,
              access: true,
              parents: [],
              type: 'REPO',
              jobState: null,
              jobNodeState: null,
              createdAt: null,
            }),
          );
          expect(vertices?.[1]).toEqual(
            expect.objectContaining({
              id: 'acb880d6f43bb784017bed136a1418072252cc9a',
              project: 'Solar-Panel-Data-Sorting',
              name: 'images',
              state: null,
              nodeState: null,
              access: true,
              parents: [],
              type: 'REPO',
              jobState: null,
              jobNodeState: null,
              createdAt: null,
            }),
          );
          expect(vertices?.[2]).toEqual(
            expect.objectContaining({
              id: 'fe5ee514dd88c55a06181b12fdaf7b29e40982e9',
              project: 'Solar-Panel-Data-Sorting',
              name: 'montage',
              state: null,
              nodeState: null,
              access: true,
              parents: [
                {
                  id: '48d741f82638e73365b8232d3f3de144fab28eab',
                  project: 'Solar-Panel-Data-Sorting',
                  name: 'edges',
                },
                {
                  id: 'acb880d6f43bb784017bed136a1418072252cc9a',
                  project: 'Solar-Panel-Data-Sorting',
                  name: 'images',
                },
              ],
              type: 'PIPELINE',
              jobState: 'JOB_FAILURE',
              jobNodeState: 'ERROR',
              createdAt: null,
            }),
          );

          resolve();
        },
      });
    });
    subscription?.unsubscribe();
  });

  it('should add cross project nodes', async () => {
    const {data} = await executeQuery<{dag: Vertex[]}>(GET_DAG_QUERY, {
      args: {
        projectId: 'Multi-Project-Pipeline-A',
      },
    });

    const vertices = data?.dag;
    expect(vertices).toHaveLength(3);
    expect(vertices?.[0]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: '72c853c677398684a3cbda4fa7cb8f39e31c2a5b',
        project: 'Multi-Project-Pipeline-A',
        name: 'Node_1',
        state: null,
        nodeState: null,
        access: true,
        parents: [],
        type: 'REPO',
        jobState: null,
        jobNodeState: null,
        createdAt: 1614126189,
      }),
    );
    expect(vertices?.[1]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'b82485a18fac187149a107b34be15c14da3b39a4',
        project: 'Multi-Project-Pipeline-A',
        name: 'Node_2',
        state: 'PIPELINE_STANDBY',
        nodeState: 'IDLE',
        access: true,
        parents: [
          {
            __typename: 'VertexIdentifier',
            id: 'b198c3ed00295bad947fc8e68d68bf285c8fbf8a',
            name: 'Node_1',
            project: 'Multi-Project-Pipeline-B',
          },
        ],
        type: 'PIPELINE',
        jobState: 'JOB_SUCCESS',
        jobNodeState: 'SUCCESS',
        createdAt: null,
      }),
    );
    expect(vertices?.[2]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'b198c3ed00295bad947fc8e68d68bf285c8fbf8a',
        project: 'Multi-Project-Pipeline-B',
        name: 'Node_1',
        state: null,
        nodeState: null,
        access: true,
        parents: [],
        type: 'CROSS_PROJECT_REPO',
        jobState: null,
        jobNodeState: null,
        createdAt: null,
      }),
    );
  });
});
