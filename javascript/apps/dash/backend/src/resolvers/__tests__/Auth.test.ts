import accounts from '@dash-backend/mock/fixtures/accounts';
import {
  mockServer,
  executeOperation,
  createOperation,
} from '@dash-backend/testHelpers';
import {Account, Tokens} from '@graphqlTypes';

describe('Auth resolver', () => {
  describe('exchangeCode', () => {
    const operationName = 'exchangeCode';
    const variables = {code: 'xyz'};

    it('should exchange an auth code for a pach token', async () => {
      const {data} = await executeOperation<{
        exchangeCode: Tokens;
      }>(operationName, variables);

      expect(data?.exchangeCode.pachToken).toBeTruthy();
    });

    it('should return an error if there is a problem exchanging the id token', async () => {
      mockServer.setHasInvalidIdToken(true);

      const {data, errors = []} = await executeOperation<{
        exchangeCode: Tokens;
      }>(operationName, variables);

      expect(data).toBeNull();
      expect(errors.length).toBe(1);
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });

    it('should return an error if there is an issue with the IDP', async () => {
      mockServer.setTokenError(true);

      const {data, errors = []} = await executeOperation<{
        exchangeCode: Tokens;
      }>(operationName, variables);

      expect(data).toBeNull();
      expect(errors.length).toBe(1);
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });
  });
  // TODO: Use generated queries once available
  describe('account', () => {
    it("should return the user's account info", async () => {
      const {data} = await createOperation<{account: Account}>(`
        query getAccount {
          account {
            id
            email
          }
        }
      `);

      expect(data?.account).toStrictEqual(accounts['1']);
    });

    it('should return an error if the user is not authenticated', async () => {
      const {data, errors = []} = await createOperation<{account: Account}>(
        `
        query getAccount {
          account {
            id
            email
          }
        }
      `,
        {},
        {'id-token': ''},
      );

      expect(data).toBeNull();
      expect(errors.length).toBe(1);
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });
  });
});
