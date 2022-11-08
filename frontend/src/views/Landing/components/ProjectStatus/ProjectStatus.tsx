import {ProjectStatus as ProjectStatusEnum} from '@graphqlTypes';
import capitalize from 'lodash/capitalize';
import React, {useMemo} from 'react';

import {Group, Circle} from '@pachyderm/components';

type ProjectStatusType = {
  status: ProjectStatusEnum;
};

const ProjectStatus: React.FC<ProjectStatusType> = ({status}) => {
  const color = useMemo(() => {
    if (status === ProjectStatusEnum.HEALTHY) {
      return 'green';
    }

    return 'red';
  }, [status]);

  return (
    <Group spacing={8} align="center">
      <Circle color={color} />
      <div data-testid={`ProjectStatus__${status}`}>{capitalize(status)}</div>
    </Group>
  );
};

export default ProjectStatus;
