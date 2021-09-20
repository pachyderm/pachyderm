import {GET_DAG_QUERY} from '@dash-frontend/queries/GetDagQuery';
import {GET_DAGS_QUERY} from '@dash-frontend/queries/GetDagsQuery';

import {
  createSubscriptionClients,
  executeQuery,
  mockServer,
} from '@dash-backend/testHelpers';
import {Vertex} from '@graphqlTypes';

describe('Dag resolver', () => {
  let subscription: ZenObservable.Subscription | null = null;

  afterEach(() => {
    if (subscription) {
      subscription.unsubscribe();
      subscription = null;
    }
  });
  it('should resolve dag data', async () => {
    const {data} = await executeQuery<{dag: Vertex[]}>(GET_DAG_QUERY, {
      args: {
        projectId: '1',
      },
    });

    const vertices = data?.dag;
    expect(vertices?.length).toBe(6);
    expect(vertices?.[0].name).toEqual('montage_repo');
    expect(vertices?.[0].parents).toEqual(['montage']);
    expect(vertices?.[1].name).toEqual('edges_repo');
    expect(vertices?.[1].parents).toEqual(['edges']);
    expect(vertices?.[2].name).toEqual('images_repo');
    expect(vertices?.[2].parents).toEqual([]);
    expect(vertices?.[3].name).toEqual('montage');
    expect(vertices?.[3].parents).toEqual(['edges', 'images']);
    expect(vertices?.[4].name).toEqual('https://egress.com');
    expect(vertices?.[4].parents).toEqual(['montage_repo']);
    expect(vertices?.[5].name).toEqual('edges');
    expect(vertices?.[5].parents).toEqual(['images']);
  });

  it('should correctly return access data to a given node', async () => {
    mockServer.setAccount('2');

    const {data} = await executeQuery<{dag: Vertex[]}>(GET_DAG_QUERY, {
      args: {
        projectId: '1',
      },
    });

    const montageRepo = data?.dag.find(
      (vertex) => vertex.name === 'montage_repo',
    );

    const montagePipeline = data?.dag.find(
      (vertex) => vertex.name === 'montage',
    );

    expect(montageRepo?.access).toBe(false);
    expect(montagePipeline?.access).toBe(false);
  });

  it('should correctly filter sub-dag for jobsets', () => {
    const {observable} = createSubscriptionClients<{data: {dags: Vertex[]}}>(
      GET_DAGS_QUERY,
      {
        args: {
          projectId: '1',
          jobSetId: '33b9af7d5d4343219bc8e02ff44cd55a',
        },
      },
    );

    return new Promise<void>((resolve) => {
      subscription = observable.subscribe({
        next: (data) => {
          const vertices = data.data?.dags;

          expect(vertices?.length).toBe(4);
          expect(vertices?.[0].name).toEqual('edges_repo');
          expect(vertices?.[0].parents).toEqual([]);
          expect(vertices?.[1].name).toEqual('images_repo');
          expect(vertices?.[1].parents).toEqual([]);
          expect(vertices?.[2].name).toEqual('montage_repo');
          expect(vertices?.[2].parents).toEqual(['montage']);
          expect(vertices?.[3].name).toEqual('montage');
          expect(vertices?.[3].parents).toEqual(['edges', 'images']);

          resolve();
        },
      });
    });
  });
});
