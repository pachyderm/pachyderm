import {Permission, ResourceType} from '@graphqlTypes';

import {
  activateEnterprise,
  deactivateEnterprise,
  activateAuth,
  getTestPachClient,
} from './utils';

describe('services/auth', () => {
  describe('environment', () => {
    it('there is an existing enterprise key on your system', () => {
      expect(process.env.PACHYDERM_ENTERPRISE_KEY).not.toBeFalsy();
    });
  });

  describe('activate', () => {
    afterEach(async () => {
      const pachClient = await deactivateEnterprise();
      pachClient.auth().deactivate();
    });

    it('enables auth', async () => {
      const pachClient = await activateEnterprise();

      const res = await activateAuth(pachClient);
      expect(res?.pachToken).toBeTruthy();
    });
  });

  describe('deactivate', () => {
    afterEach(async () => {
      await deactivateEnterprise();
    });
    it('deactivates auth', async () => {
      const pachClient = await activateEnterprise();
      await activateAuth(pachClient);

      const res = await pachClient.auth().deactivate();
      expect(res).toStrictEqual({});
    });

    it('errors if auth is not active', async () => {
      const pachClient = await activateEnterprise();

      await expect(pachClient.auth().deactivate()).rejects.toThrow(
        '12 UNIMPLEMENTED: the auth service is not activated',
      );
    });
  });

  describe('Roles & Permissions', () => {
    afterEach(async () => {
      const pachClient = await deactivateEnterprise({deactivateAuth: true});

      const pps = pachClient.pps();
      const pfs = pachClient.pfs();
      await pps.deleteAll();
      await pfs.deleteAll();
    });

    it('getRoleBinding gives correct role binding data for repo', async () => {
      const pachClient = await activateEnterprise();
      await activateAuth(pachClient);

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

    it('getPermissions gives correct role and permission data for repo', async () => {
      const pachClient = await activateEnterprise();
      await activateAuth(pachClient);

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

    it('modifyRoles updates role bindings for a specific resource', async () => {
      const pachClient = await activateEnterprise();
      await activateAuth(pachClient);

      const pfs = pachClient.pfs();
      await pfs.createProject({name: 'test'});
      await pfs.createRepo({projectId: 'test', repo: {name: 'images'}});

      await pachClient.auth().modifyRoleBinding({
        resource: {type: 2, name: 'test/images'},
        principal: 'pach:root',
        rolesList: ['repoWriter'],
      });

      const res = await pachClient
        .auth()
        .getRoleBinding({resource: {type: 2, name: 'test/images'}});

      expect(res?.binding?.entriesMap).toStrictEqual([
        ['pach:root', {rolesMap: [['repoWriter', true]]}],
      ]);
    });
  });

  describe('authorize', () => {
    afterEach(async () => {
      const pachClient = await deactivateEnterprise({deactivateAuth: true});

      const pps = pachClient.pps();
      const pfs = pachClient.pfs();
      await pps.deleteAll();
      await pfs.deleteAll();
    });

    it('correctly returns project auth checks', async () => {
      const pachClient = await activateEnterprise();
      await activateAuth(pachClient);

      const pfs = pachClient.pfs();
      await pfs.createProject({name: 'test'});
      await pfs.createRepo({projectId: 'test', repo: {name: 'images'}});

      let auth = pachClient.auth();
      expect(
        await auth.authorize({
          permissionsList: [Permission.PROJECT_MODIFY_BINDINGS],
          resource: {type: ResourceType.PROJECT, name: 'test'},
        }),
      ).toEqual(
        expect.objectContaining({
          authorized: true,
          satisfiedList: [404],
          missingList: [],
          principal: 'pach:root',
        }),
      );
      expect(
        await auth.authorize({
          permissionsList: [Permission.REPO_DELETE],
          resource: {type: ResourceType.REPO, name: 'test/images'},
        }),
      ).toEqual(
        expect.objectContaining({
          authorized: true,
          satisfiedList: [203],
          missingList: [],
          principal: 'pach:root',
        }),
      );
      expect(
        await auth.authorize({
          permissionsList: [Permission.REPO_WRITE],
          resource: {type: ResourceType.REPO, name: 'test/images'},
        }),
      ).toEqual(
        expect.objectContaining({
          authorized: true,
          satisfiedList: [201],
          missingList: [],
          principal: 'pach:root',
        }),
      );
      expect(
        await auth.authorize({
          permissionsList: [Permission.REPO_MODIFY_BINDINGS],
          resource: {type: ResourceType.REPO, name: 'test/images'},
        }),
      ).toEqual(
        expect.objectContaining({
          authorized: true,
          satisfiedList: [202],
          missingList: [],
          principal: 'pach:root',
        }),
      );

      const {token} = await auth.getRobotToken({
        robot: 'testbot',
        ttl: 111111,
      });
      const robotPachClient = getTestPachClient(token);
      auth = robotPachClient.auth();
      expect(
        await auth.authorize({
          permissionsList: [Permission.PROJECT_MODIFY_BINDINGS],
          resource: {type: ResourceType.PROJECT, name: 'test'},
        }),
      ).toEqual(
        expect.objectContaining({
          authorized: false,
          satisfiedList: [],
          missingList: [404],
          principal: 'robot:testbot',
        }),
      );
      expect(
        await auth.authorize({
          permissionsList: [Permission.REPO_DELETE],
          resource: {type: ResourceType.REPO, name: 'test/images'},
        }),
      ).toEqual(
        expect.objectContaining({
          authorized: false,
          satisfiedList: [],
          missingList: [203],
          principal: 'robot:testbot',
        }),
      );
      expect(
        await auth.authorize({
          permissionsList: [Permission.REPO_WRITE],
          resource: {type: ResourceType.REPO, name: 'test/images'},
        }),
      ).toEqual(
        expect.objectContaining({
          authorized: false,
          satisfiedList: [],
          missingList: [201],
          principal: 'robot:testbot',
        }),
      );
      expect(
        await auth.authorize({
          permissionsList: [Permission.REPO_MODIFY_BINDINGS],
          resource: {type: ResourceType.REPO, name: 'test/images'},
        }),
      ).toEqual(
        expect.objectContaining({
          authorized: false,
          satisfiedList: [],
          missingList: [202],
          principal: 'robot:testbot',
        }),
      );
    });
  });
});
