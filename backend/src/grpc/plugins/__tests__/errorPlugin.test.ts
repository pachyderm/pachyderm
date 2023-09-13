import {GET_VERTICES_QUERY} from '@dash-frontend/queries/GetVerticesQuery';

import {executeQuery} from '@dash-backend/testHelpers';
import {Vertex} from '@graphqlTypes';

describe('errorPlugin', () => {
  it('should transform token expiration errors into authenticated errors', async () => {
    const {data, errors = []} = await executeQuery<{data: Vertex}>(
      GET_VERTICES_QUERY,
      {
        args: {
          projectId: 'Solar-Panel-Data-Sorting',
        },
      },
      {'auth-token': 'expired'},
    );

    expect(data).toBeNull();
    expect(errors).toHaveLength(1);
    expect(errors?.[0].extensions?.code).toBe('UNAUTHENTICATED');
  });
});
