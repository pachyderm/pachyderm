import {mockServer, executeOperation} from '@dash-backend/testHelpers';

describe('Auth resolver', () => {
  describe('exchangeCode', () => {
    const operationName = 'exchangeCode';
    const variables = {code: 'xyz'};

    it('should exchange an auth code for a pach token', async () => {
      const {data: exchangeCode} = await executeOperation<{exchangeCode: string}>(
        operationName,
        variables,
      );

      expect(exchangeCode).toBeTruthy();
    });

    it('should return an error if there is a problem exchanging the id token', async () => {
      mockServer.setHasInvalidIdToken(true);

      const {data, errors = []} = await executeOperation<{exchangeCode: string}>(
        operationName,
        variables,
      );

      expect(data).toBeNull();
      expect(errors.length).toBe(1);
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });

    it('should return an error if there is an issue with the IDP', async () => {
      mockServer.setTokenError(true);

      const {data, errors = []} = await executeOperation<{exchangeCode: string}>(
        operationName,
        variables,
      );

      expect(data).toBeNull();
      expect(errors.length).toBe(1);
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });
  });
});
