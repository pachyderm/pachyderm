import {SortableItem, stringComparator, useSort} from '@pachyderm/components';
import capitalize from 'lodash/capitalize';
import every from 'lodash/every';
import reduce from 'lodash/reduce';
import {useCallback, useMemo, useState} from 'react';
import {useForm} from 'react-hook-form';

import {useProjects} from '@dash-frontend/hooks/useProjects';
import projectStatusAsString from '@dash-frontend/lib/projecStatusAsString';
import {Project} from '@graphqlTypes';

type sortOptions = {
  [key: string]: SortableItem<Project>;
};

type statusFormType = {
  [key: string]: boolean;
};

const nameComparator = {
  name: 'Name',
  func: stringComparator,
  accessor: (project: Project) => project.name,
};

const dateComparator = {
  name: 'Created On',
  func: (a: number, b: number) => b - a,
  accessor: (project: Project) => project.createdAt,
};

const sortOptions: sortOptions = {
  'Created On': dateComparator,
  'Name A-Z': nameComparator,
};

export const useLandingView = () => {
  const {projects, loading} = useProjects();
  const {sortedData: sortedProjects, setComparator} = useSort({
    data: projects,
    initialSort: dateComparator,
  });

  const [searchValue, setSearchValue] = useState('');
  const [sortButtonText, setSortButtonText] = useState('Created On');

  const handleSortSelect = useCallback(
    (id: string) => {
      setComparator(sortOptions[id]);
      setSortButtonText(id);
    },
    [setComparator, setSortButtonText],
  );

  const filterFormCtx = useForm<statusFormType>({
    defaultValues: {
      HEALTHY: true,
      UNHEALTHY: true,
    },
  });

  const {watch} = filterFormCtx;

  const filters = watch(['HEALTHY', 'UNHEALTHY'], {
    HEALTHY: true,
    UNHEALTHY: true,
  });

  const filterStatus = useMemo(() => {
    if (every(filters, (filter) => !!filter)) {
      return 'Show: All';
    }
    const statusString = reduce(
      filters,
      (acc, val, key) => {
        if (val) {
          const status = capitalize(key);
          return acc ? `${acc}, ${status}` : status;
        }
        return acc;
      },
      '',
    );

    return `Show: ${statusString ? statusString : 'None'}`;
  }, [filters]);

  const filteredProjects = useMemo(() => {
    return sortedProjects.filter((project) => {
      return (
        filters[projectStatusAsString(project.status)] &&
        project.name.includes(searchValue)
      );
    });
  }, [filters, searchValue, sortedProjects]);

  return {
    filterStatus,
    filterFormCtx,
    handleSortSelect,
    loading,
    projects: filteredProjects,
    projectCount: projects.length,
    searchValue,
    setSearchValue,
    sortButtonText,
  };
};
