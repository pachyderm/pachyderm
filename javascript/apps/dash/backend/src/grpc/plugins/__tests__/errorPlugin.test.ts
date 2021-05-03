import {executeQuery} from '@dash-backend/testHelpers';
import {GET_DAG_QUERY} from '@dash-frontend/queries/GetDagQuery';
import {Dag} from '@graphqlTypes';

describe('errorPlugin', () => {
  it('should transform token expiration errors into authenticated errors', async () => {
    const {data, errors = []} = await executeQuery<{data: Dag}>(
      GET_DAG_QUERY,
      {
        args: {projectId: '1', nodeWidth: 120, nodeHeight: 60},
      },
      {'auth-token': 'expired'},
    );

    expect(data).toBeNull();
    expect(errors?.length).toBe(1);
    expect(errors?.[0].extensions?.code).toBe('UNAUTHENTICATED');
  });
});
