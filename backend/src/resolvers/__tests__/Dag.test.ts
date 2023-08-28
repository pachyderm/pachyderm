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
    expect(vertices).toHaveLength(6);
    expect(vertices?.[0]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'Solar-Panel-Data-Sorting_montage_repo',
        name: 'montage',
        state: null,
        access: true,
        parents: ['Solar-Panel-Data-Sorting_montage'],
        type: 'OUTPUT_REPO',
        jobState: null,
        createdAt: 1614136189,
      }),
    );
    expect(vertices?.[1]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'Solar-Panel-Data-Sorting_edges_repo',
        name: 'edges',
        state: null,
        access: true,
        parents: ['Solar-Panel-Data-Sorting_edges'],
        type: 'OUTPUT_REPO',
        jobState: null,
        createdAt: 1614126189,
      }),
    );
    expect(vertices?.[2]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'Solar-Panel-Data-Sorting_images_repo',
        name: 'images',
        state: null,
        access: true,
        parents: [],
        type: 'INPUT_REPO',
        jobState: null,
        createdAt: 1614116189,
      }),
    );
    expect(vertices?.[3]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'Solar-Panel-Data-Sorting_montage',
        name: 'montage',
        state: 'PIPELINE_FAILURE',
        nodeState: 'ERROR',
        access: true,
        parents: [
          'Solar-Panel-Data-Sorting_edges',
          'Solar-Panel-Data-Sorting_images',
        ],
        type: 'PIPELINE',
        jobNodeState: 'RUNNING',
        jobState: 'JOB_CREATED',
        createdAt: null,
      }),
    );
    expect(vertices?.[4]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'Solar-Panel-Data-Sorting_montage_egress',
        name: 'https://egress.com',
        state: null,
        access: true,
        parents: ['Solar-Panel-Data-Sorting_montage_repo'],
        type: 'EGRESS',
        jobState: null,
        createdAt: null,
      }),
    );
    expect(vertices?.[5]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'Solar-Panel-Data-Sorting_edges',
        name: 'edges',
        state: 'PIPELINE_RUNNING',
        nodeState: 'IDLE',
        access: true,
        parents: ['Solar-Panel-Data-Sorting_images'],
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

    const montageRepo = data?.dag.find(
      (vertex) => vertex.id === 'Solar-Panel-Data-Sorting_montage_repo',
    );

    const montagePipeline = data?.dag.find(
      (vertex) => vertex.id === 'Solar-Panel-Data-Sorting_montage',
    );

    expect(montageRepo?.access).toBe(false);
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

          expect(vertices).toHaveLength(4);
          expect(vertices?.[0]).toEqual(
            expect.objectContaining({
              id: 'Solar-Panel-Data-Sorting_edges_repo',
              name: 'edges_repo',
              state: null,
              nodeState: null,
              access: true,
              parents: [],
              type: 'INPUT_REPO',
              jobState: null,
              jobNodeState: null,
              createdAt: null,
            }),
          );
          expect(vertices?.[1]).toEqual(
            expect.objectContaining({
              id: 'Solar-Panel-Data-Sorting_images_repo',
              name: 'images_repo',
              state: null,
              nodeState: null,
              access: true,
              parents: [],
              type: 'INPUT_REPO',
              jobState: null,
              jobNodeState: null,
              createdAt: null,
            }),
          );
          expect(vertices?.[2]).toEqual(
            expect.objectContaining({
              id: 'Solar-Panel-Data-Sorting_montage_repo',
              name: 'montage_repo',
              state: null,
              nodeState: null,
              access: true,
              parents: ['Solar-Panel-Data-Sorting_montage'],
              type: 'OUTPUT_REPO',
              jobState: null,
              jobNodeState: null,
              createdAt: null,
            }),
          );
          expect(vertices?.[3]).toEqual(
            expect.objectContaining({
              id: 'Solar-Panel-Data-Sorting_montage',
              name: 'montage',
              state: null,
              nodeState: null,
              access: true,
              parents: [
                'Solar-Panel-Data-Sorting_edges',
                'Solar-Panel-Data-Sorting_images',
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
    expect(vertices).toHaveLength(4);
    expect(vertices?.[0]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'Multi-Project-Pipeline-A_Node_1_repo',
        name: 'Node_1',
        state: null,
        nodeState: null,
        access: true,
        parents: [],
        type: 'INPUT_REPO',
        jobState: null,
        jobNodeState: null,
        createdAt: 1614126189,
      }),
    );
    expect(vertices?.[1]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'Multi-Project-Pipeline-A_Node_2_repo',
        name: 'Node_2',
        state: null,
        nodeState: null,
        access: true,
        parents: ['Multi-Project-Pipeline-A_Node_2'],
        type: 'OUTPUT_REPO',
        jobState: null,
        jobNodeState: null,
        createdAt: 1614126189,
      }),
    );
    expect(vertices?.[2]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'Multi-Project-Pipeline-A_Node_2',
        name: 'Node_2',
        state: 'PIPELINE_STANDBY',
        nodeState: 'IDLE',
        access: true,
        parents: ['Multi-Project-Pipeline-B_Node_1'],
        type: 'PIPELINE',
        jobState: 'JOB_SUCCESS',
        jobNodeState: 'SUCCESS',
        createdAt: null,
      }),
    );
    expect(vertices?.[3]).toEqual(
      expect.objectContaining({
        __typename: 'Vertex',
        id: 'Multi-Project-Pipeline-B_Node_1',
        name: 'Node',
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
