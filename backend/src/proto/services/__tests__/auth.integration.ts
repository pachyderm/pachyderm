import {ResourceType} from '@graphqlTypes';

import {
  activateEnterprise,
  deactivateAuthEnterprise,
  deactivateEnterprise,
} from './utils';

let token = '';

describe('services/auth', () => {
  describe('environment', () => {
    it('there is an existing enterprise key on your system', () => {
      expect(process.env.PACHYDERM_ENTERPRISE_KEY).not.toBeFalsy();
    });
  });

  describe('activate', () => {
    afterEach(async () => {
      const pachClient = await deactivateEnterprise(token);
      pachClient.auth().deactivate();
    });

    it('enables auth', async () => {
      const pachClient = await activateEnterprise();
      const auth = pachClient.auth();

      const res = await auth.activate();
      token = res?.pachToken;
      expect(token).toBeTruthy();
    });
  });

  describe('deactivate', () => {
    afterEach(async () => {
      await deactivateEnterprise();
    });
    it('deactivates auth', async () => {
      const pachClient = await activateEnterprise(token);
      const auth = pachClient.auth();

      const authRes = await auth.activate();
      token = authRes?.pachToken;
      expect(token).toBeTruthy();
      pachClient.setAuthnToken(token);

      const res = await pachClient.auth().deactivate();
      expect(res).toStrictEqual({});
    });

    it('errors if auth is not active', async () => {
      const pachClient = await activateEnterprise(token);

      await expect(pachClient.auth().deactivate()).rejects.toThrow(
        '12 UNIMPLEMENTED: the auth service is not activated',
      );
    });
  });

  describe('getRoleBinding', () => {
    afterEach(async () => {
      const pachClient = await deactivateAuthEnterprise(token);

      const pps = pachClient.pps();
      const pfs = pachClient.pfs();
      await pps.deleteAll();
      await pfs.deleteAll();
    });

    it('gives correct role binding data for repo', async () => {
      const pachClient = await activateEnterprise();
      const auth = pachClient.auth();

      const authRes = await auth.activate();
      token = authRes?.pachToken;
      expect(token).toBeTruthy();

      pachClient.setAuthnToken(token);

      const pfs = pachClient.pfs();
      await pfs.createProject({name: 'test'});
      await pfs.createRepo({projectId: 'test', repo: {name: 'images'}});

      const res = await pachClient
        .auth()
        .getRoleBinding({resource: {type: 2, name: 'test/images'}});

      expect(res?.binding?.entriesMap).toStrictEqual([
        ['pach:root', {rolesMap: [['repoOwner', true]]}],
      ]);
    });
  });

  describe('getPermissions', () => {
    afterEach(async () => {
      const pachClient = await deactivateAuthEnterprise(token);

      const pps = pachClient.pps();
      const pfs = pachClient.pfs();
      await pps.deleteAll();
      await pfs.deleteAll();
    });

    it('gives correct role and permission data for repo', async () => {
      const pachClient = await activateEnterprise();
      const auth = pachClient.auth();

      const authRes = await auth.activate();
      token = authRes?.pachToken;
      expect(token).toBeTruthy();

      pachClient.setAuthnToken(token);

      const pfs = pachClient.pfs();
      await pfs.createProject({name: 'test'});
      await pfs.createRepo({projectId: 'test', repo: {name: 'images'}});

      // using int type
      const intRes = await pachClient
        .auth()
        .getPermissions({resource: {type: 2, name: 'test/images'}});

      expect(intRes.rolesList).toStrictEqual(['clusterAdmin', 'repoOwner']);

      // using enum type
      const enumRes = await pachClient.auth().getPermissions({
        resource: {type: ResourceType.REPO, name: 'test/images'},
      });

      expect(enumRes.rolesList).toStrictEqual(['clusterAdmin', 'repoOwner']);
    });
  });
});
