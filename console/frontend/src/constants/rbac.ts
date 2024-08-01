export const REPO_ROLES = ['repoReader', 'repoWriter', 'repoOwner'];

export const PROJECT_ROLES = ['projectViewer', 'projectWriter', 'projectOwner'];

export const ALL_PROJECT_ROLES = [...PROJECT_ROLES, ...REPO_ROLES];
export const ALL_CLUSTER_ROLES = [
  'clusterAdmin',
  'projectCreator',
  ...ALL_PROJECT_ROLES,
];

export const CLUSTER_ROLES = [
  'clusterAdmin',
  'projectCreator',
  'oidcAppAdmin',
  'idpAdmin',
  'secretAdmin',
  'identityAdmin',
  'licenseAdmin',
];

export const IGNORED_CLUSTER_ROLES = [
  'oidcAppAdmin',
  'idpAdmin',
  'secretAdmin',
  'identityAdmin',
  'licenseAdmin',
  'debugger',
  'robotUser ',
  'pachdLogReader',
];

export const ALLOWED_PRINCIPAL_TYPES = [
  'user',
  'group',
  'robot',
  'allClusterUsers',
];
