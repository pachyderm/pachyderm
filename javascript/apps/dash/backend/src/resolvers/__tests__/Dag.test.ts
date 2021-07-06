import {
  createSubscriptionClients,
  executeQuery,
  mockServer,
} from '@dash-backend/testHelpers';
import {GET_DAG_QUERY} from '@dash-frontend/queries/GetDagQuery';
import {GET_DAGS_QUERY} from '@dash-frontend/queries/GetDagsQuery';
import {Dag, DagDirection} from '@graphqlTypes';

const doesLinkExistInDag = (
  expectedLink: {source: string; target: string},
  dag: Dag | undefined,
) => {
  if (!dag) {
    return false;
  }

  return dag.links.some((link) => {
    return (
      link.source === expectedLink.source && link.target === expectedLink.target
    );
  });
};

describe('Dag resolver', () => {
  let subscription: ZenObservable.Subscription | null = null;

  afterEach(() => {
    if (subscription) {
      subscription.unsubscribe();
      subscription = null;
    }
  });
  it('should resolve dag data', async () => {
    const {data} = await executeQuery<{dag: Dag}>(GET_DAG_QUERY, {
      args: {
        projectId: '1',
        nodeWidth: 120,
        nodeHeight: 60,
        direction: DagDirection.RIGHT,
      },
    });

    const dag = data?.dag;

    expect(dag?.links.length).toBe(6);
    expect(
      doesLinkExistInDag({source: 'montage', target: 'montage_repo'}, dag),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'edges', target: 'edges_repo'}, dag),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'edges_repo', target: 'montage'}, dag),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'images_repo', target: 'montage'}, dag),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'images_repo', target: 'edges'}, dag),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'montage_repo', target: 'https://egress.com'},
        dag,
      ),
    ).toBe(true);
  });

  it('should correctly render cron inputs', async () => {
    const {data} = await executeQuery<{dag: Dag}>(GET_DAG_QUERY, {
      args: {
        projectId: '3',
        nodeWidth: 120,
        nodeHeight: 60,
        direction: DagDirection.RIGHT,
      },
    });

    const dag = data?.dag;

    expect(dag?.links.length).toBe(2);
    expect(
      doesLinkExistInDag({source: 'cron_repo', target: 'processor'}, dag),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'processor', target: 'processor_repo'}, dag),
    ).toBe(true);
  });

  it('should correctly return access data to a given node', async () => {
    mockServer.setAccount('2');

    const {data} = await executeQuery<{dag: Dag}>(GET_DAG_QUERY, {
      args: {
        projectId: '1',
        nodeWidth: 120,
        nodeHeight: 60,
        direction: DagDirection.RIGHT,
      },
    });

    const montageRepo = data?.dag.nodes.find(
      (node) => node.name === 'montage_repo',
    );

    const montagePipeline = data?.dag.nodes.find(
      (node) => node.name === 'montage',
    );

    expect(montageRepo?.access).toBe(false);
    expect(montagePipeline?.access).toBe(false);
  });

  it('should resolve disconnected components of a dag', (done) => {
    const {observable} = createSubscriptionClients<{
      data: {dags: Dag[]};
    }>(GET_DAGS_QUERY, {
      args: {
        projectId: '2',
        nodeHeight: 60,
        nodeWidth: 120,
        direction: DagDirection.RIGHT,
      },
    });

    subscription = observable.subscribe({
      next: (data) => {
        const dags = data.data.dags;
        expect(dags?.length).toBe(3);

        expect(dags?.[0].links.length).toBe(6);
        expect(
          doesLinkExistInDag(
            {source: 'samples_repo', target: 'likelihoods'},
            dags?.[0],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'reference_repo', target: 'likelihoods'},
            dags?.[0],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'reference_repo', target: 'joint_call'},
            dags?.[0],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'likelihoods', target: 'likelihoods_repo'},
            dags?.[0],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'joint_call', target: 'joint_call_repo'},
            dags?.[0],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'likelihoods_repo', target: 'joint_call'},
            dags?.[0],
          ),
        ).toBe(true);

        expect(dags?.[1].links.length).toBe(2);
        expect(
          doesLinkExistInDag(
            {source: 'training_repo', target: 'models'},
            dags?.[1],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'models', target: 'models_repo'},
            dags?.[1],
          ),
        ).toBe(true);

        expect(dags?.[2].links.length).toBe(14);
        expect(
          doesLinkExistInDag(
            {source: 'raw_data_repo', target: 'split'},
            dags?.[2],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'split', target: 'split_repo'},
            dags?.[2],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'split_repo', target: 'model'},
            dags?.[2],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'model', target: 'model_repo'},
            dags?.[2],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'parameters_repo', target: 'model'},
            dags?.[2],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag({source: 'model_repo', target: 'test'}, dags?.[2]),
        ).toBe(true);
        expect(
          doesLinkExistInDag({source: 'split_repo', target: 'test'}, dags?.[2]),
        ).toBe(true);
        expect(
          doesLinkExistInDag({source: 'test', target: 'test_repo'}, dags?.[2]),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'test_repo', target: 'select'},
            dags?.[2],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'model_repo', target: 'select'},
            dags?.[2],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'select', target: 'select_repo'},
            dags?.[2],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'model_repo', target: 'detect'},
            dags?.[2],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'images_repo', target: 'detect'},
            dags?.[2],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'detect', target: 'detect_repo'},
            dags?.[2],
          ),
        ).toBe(true);
        done();
      },
    });
  });

  it('should send dag id as name of oldest repo', (done) => {
    const {observable} = createSubscriptionClients<{
      data: {dags: Dag[]};
    }>(GET_DAGS_QUERY, {
      args: {
        projectId: '2',
        nodeHeight: 60,
        nodeWidth: 120,
        direction: DagDirection.RIGHT,
      },
    });

    subscription = observable.subscribe({
      next: (data) => {
        const dags = data.data?.dags;

        expect(dags?.[0].id).toBe('samples');
        expect(dags?.[1].id).toBe('training');
        expect(dags?.[2].id).toBe('images');
        done();
      },
    });
  });

  it('should correctly filter sub-dag for jobsets', (done) => {
    const {observable} = createSubscriptionClients<{data: {dags: Dag[]}}>(
      GET_DAGS_QUERY,
      {
        args: {
          projectId: '1',
          nodeHeight: 60,
          nodeWidth: 120,
          direction: DagDirection.RIGHT,
          jobSetId: '33b9af7d5d4343219bc8e02ff44cd55a',
        },
      },
    );

    subscription = observable.subscribe({
      next: (data) => {
        const dags = data.data?.dags;

        expect(dags?.length).toBe(1);

        expect(
          doesLinkExistInDag(
            {source: 'montage', target: 'montage_repo'},
            dags?.[0],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'edges_repo', target: 'montage'},
            dags?.[0],
          ),
        ).toBe(true);
        expect(
          doesLinkExistInDag(
            {source: 'images_repo', target: 'montage'},
            dags?.[0],
          ),
        ).toBe(true);
        done();
      },
    });
  });
});
