import fs from 'fs';
import path from 'path';

import {sign} from 'jsonwebtoken';

import accounts from '@dash-backend/mock/fixtures/accounts';
import {generateIdTokenForAccount} from '@dash-backend/testHelpers';

import {getAccountFromIdToken} from '../auth';

describe('auth', () => {
  describe('getAccountFromIdToken', () => {
    it('should return the correct account', async () => {
      const idToken = generateIdTokenForAccount(accounts['1']);

      expect(await getAccountFromIdToken(idToken)).toStrictEqual(accounts['1']);
    });

    it('should throw an error for a malformed jwt', async () => {
      await expect(getAccountFromIdToken('fake')).rejects.toThrowError(
        'jwt malformed',
      );
    });

    it('should throw an error for an invalid key id', async () => {
      const idToken = sign(
        {email: 'test@test.com'},
        fs.readFileSync(path.resolve(__dirname, '../../mock/mockPrivate.key')),
        {
          algorithm: 'RS256',
          issuer: process.env.ISSUER_URI,
          subject: '1',
          audience: ['pachd', 'dash'],
          expiresIn: '30 days',
          keyid: 'fake',
        },
      );

      await expect(getAccountFromIdToken(idToken)).rejects.toThrowError(
        'error in secret or public key callback: Could not find matching key',
      );
    });

    it('should throw an error if there is an issue verifying the jwt', async () => {
      const idToken = sign({email: 'test@test.com'}, 'shhh');

      await expect(getAccountFromIdToken(idToken)).rejects.toThrowError();
    });
  });
});
