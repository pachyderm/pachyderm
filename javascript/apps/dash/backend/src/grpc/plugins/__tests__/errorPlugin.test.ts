import {executeOperation} from '@dash-backend/testHelpers';
import {Dag} from '@graphqlTypes';

describe('errorPlugin', () => {
  it('should transform token expiration errors into authenticated errors', async () => {
    const {data, errors = []} = await executeOperation<{dag: Dag}>(
      'getDag',
      {
        args: {projectId: '1', nodeWidth: 120, nodeHeight: 60},
      },
      {'auth-token': 'expired'},
    );

    expect(data).toBeNull();
    expect(errors.length).toBe(1);
    expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
  });
});
