import {Group, Circle} from '@pachyderm/components';
import capitalize from 'lodash/capitalize';
import React, {useMemo} from 'react';

import projectStatusAsString from '@dash-frontend/lib/projecStatusAsString';
import {ProjectStatus as ProjectStatusEnum} from '@graphqlTypes';

type ProjectStatusType = {
  status: ProjectStatusEnum;
};

const ProjectStatus: React.FC<ProjectStatusType> = ({status}) => {
  const color = useMemo(() => {
    if (
      ProjectStatusEnum[
        (status as unknown) as keyof typeof ProjectStatusEnum
      ] === ProjectStatusEnum.HEALTHY
    ) {
      return 'green';
    }

    return 'red';
  }, [status]);

  return (
    <Group spacing={8} align="center">
      <Circle color={color} />
      {capitalize(projectStatusAsString(status))}
    </Group>
  );
};

export default ProjectStatus;
