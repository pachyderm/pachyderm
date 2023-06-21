export const REPO_ROLES = ['repoReader', 'repoWriter', 'repoOwner'];

export const PROJECT_ROLES = ['projectViewer', 'projectWriter', 'projectOwner'];

export const ALL_ROLES = [...PROJECT_ROLES, ...REPO_ROLES];

export const CLUSTER_ROLES = [
  'clusterAdmin',
  'oidcAppAdmin',
  'idpAdmin',
  'secretAdmin',
  'identityAdmin',
  'licenseAdmin',
];

export const IGNORED_CLUSTER_ROLES = [
  'projectCreator',
  'oidcAppAdmin',
  'idpAdmin',
  'secretAdmin',
  'identityAdmin',
  'licenseAdmin',
  'debugger',
  'robotUser ',
  'pachdLogReader',
];
