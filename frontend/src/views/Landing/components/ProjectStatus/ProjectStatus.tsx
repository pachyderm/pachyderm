import capitalize from 'lodash/capitalize';
import React from 'react';

import {ProjectStatus as ProjectStatusEnum} from '@dash-frontend/api/pfs';
import {Group, Circle} from '@pachyderm/components';

type ProjectStatusType = {
  status?: ProjectStatusEnum | null;
};

const color = (status?: ProjectStatusEnum | null) => {
  switch (status) {
    case ProjectStatusEnum.HEALTHY:
      return 'green';
    case ProjectStatusEnum.UNHEALTHY:
      return 'red';
    default:
      return 'gray';
  }
};

const ProjectStatus: React.FC<ProjectStatusType> = ({status}) => {
  return (
    <Group spacing={8} align="center">
      <Circle color={color(status)} />
      <div data-testid={`ProjectStatus__${status}`}>
        {status ? capitalize(status) : 'Pending'}
      </div>
    </Group>
  );
};

export default ProjectStatus;
