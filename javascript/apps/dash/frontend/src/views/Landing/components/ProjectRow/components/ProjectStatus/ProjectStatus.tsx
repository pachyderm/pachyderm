import {Group, Circle} from '@pachyderm/components';
import React, {useMemo} from 'react';

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
      {`${status.toString().charAt(0)}${status
        .toString()
        .substring(1)
        .toLowerCase()}`}
    </Group>
  );
};

export default ProjectStatus;
