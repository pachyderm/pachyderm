import {SortableItem, stringComparator, useSort} from '@pachyderm/components';
import capitalize from 'lodash/capitalize';
import every from 'lodash/every';
import reduce from 'lodash/reduce';
import {useCallback, useEffect, useMemo, useState} from 'react';
import {useForm} from 'react-hook-form';

import {useProjects} from '@dash-frontend/hooks/useProjects';
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
  const [selectedProject, setSelectedProject] = useState<Project>();
  const [projectsLoaded, setProjectsLoaded] = useState(false);

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

  const HEALTHY = watch('HEALTHY', true);
  const UNHEALTHY = watch('UNHEALTHY', true);

  const filters = useMemo(
    () => ({
      HEALTHY,
      UNHEALTHY,
    }),
    [HEALTHY, UNHEALTHY],
  );

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
      return filters[project.status] && project.name.includes(searchValue);
    });
  }, [filters, searchValue, sortedProjects]);

  useEffect(() => {
    if (!loading && !projectsLoaded) {
      setSelectedProject(filteredProjects[0]);
      setProjectsLoaded(true);
    }
  }, [loading, filteredProjects, projectsLoaded]);

  return {
    filterStatus,
    filterFormCtx,
    handleSortSelect,
    loading,
    multiProject: projects.length > 1,
    projects: filteredProjects,
    projectCount: projects.length,
    searchValue,
    setSearchValue,
    sortButtonText,
    selectedProject,
    setSelectedProject,
  };
};
