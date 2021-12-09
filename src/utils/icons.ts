import {LabIcon} from '@jupyterlab/ui-components';

import fileSvg from '../../style/icons/file.svg';
import mountLogoSvg from '../../style/icons/mount-logo.svg';
import repoSvg from '../../style/icons/repo.svg';

export const fileIcon = new LabIcon({
  name: 'jupyterlab-pachyderm:file',
  svgstr: fileSvg,
});

export const mountLogoIcon = new LabIcon({
  name: 'jupyterlab-pachyderm:mount-logo',
  svgstr: mountLogoSvg,
});

export const repoIcon = new LabIcon({
  name: 'jupyterlab-pachyderm:repo',
  svgstr: repoSvg,
});
