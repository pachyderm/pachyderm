import capitalize from 'lodash/capitalize';
import every from 'lodash/every';
import reduce from 'lodash/reduce';
import {useCallback, useEffect, useMemo, useState} from 'react';
import {useForm} from 'react-hook-form';

import {ProjectInfo} from '@dash-frontend/api/pfs';
import {useProjects} from '@dash-frontend/hooks/useProjects';
import {useGetProjectStatus} from '@dash-frontend/hooks/useProjectStatus';
import {getUnixSecondsFromISOString} from '@dash-frontend/lib/dateTime';
import {
  SortableItem,
  useSort,
  stringComparator,
  numberComparator,
} from '@pachyderm/components';

type sortOptionsType = {
  [key: string]: SortableItem<ProjectInfo>;
};

type statusFormType = {
  [key: string]: boolean;
};

const sortOptions: sortOptionsType = {
  Newest: {
    name: 'Newest',
    reverse: true,
    func: numberComparator,
    accessor: (project: ProjectInfo) =>
      getUnixSecondsFromISOString(project.createdAt),
  },
  Oldest: {
    name: 'Oldest',
    func: numberComparator,
    accessor: (project: ProjectInfo) =>
      getUnixSecondsFromISOString(project.createdAt),
  },
  'Name A-Z': {
    name: 'Name A-Z',
    func: stringComparator,
    accessor: (project: ProjectInfo) =>
      project.project?.name?.toLowerCase() || '',
  },
  'Name Z-A': {
    name: 'Name Z-A',
    reverse: true,
    func: stringComparator,
    accessor: (project: ProjectInfo) =>
      project.project?.name?.toLowerCase() || '',
  },
};

export const useLandingView = () => {
  const {projects, loading} = useProjects();
  const getProjectStatus = useGetProjectStatus();
  const {
    sortedData: sortedProjects,
    setComparator,
    comparatorName,
  } = useSort({
    data: projects,
    initialSort: sortOptions['Name A-Z'],
  });

  const [searchValue, setSearchValue] = useState('');
  const [sortButtonText, setSortButtonText] = useState('Name A-Z');
  const [selectedProject, setSelectedProject] = useState<ProjectInfo>();
  const [projectsLoaded, setProjectsLoaded] = useState(false);

  const handleSortSelect = useCallback(
    (id: string) => {
      if (id !== comparatorName) {
        setComparator(sortOptions[id]);
        setSortButtonText(id);
      }
    },
    [comparatorName, setComparator],
  );

  const sortDropdown = useMemo(
    () =>
      Object.values(sortOptions).map((option) => ({
        id: option.name,
        content: option.name,
      })),
    [],
  );

  type ViewType = 'Your Projects' | 'All Projects';
  const [viewButtonText, setViewButtonText] =
    useState<ViewType>('Your Projects');

  const handleViewSelect = useCallback(
    (id: string) => {
      if (id !== viewButtonText) {
        setViewButtonText(id as ViewType);
      }
    },
    [viewButtonText],
  );

  const viewDropdown = [
    {content: 'Your Projects', id: 'Your Projects'},
    {content: 'All Projects', id: 'All Projects'},
  ];

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
      const projectStatus = getProjectStatus(project?.project?.name);
      const projectStatusIsPending = !projectStatus;
      const isNotFilteredOutByStatus = projectStatus && filters[projectStatus];

      const projectName = project.project?.name?.toLowerCase() || '';
      const isSearchingForProject = projectName.includes(
        searchValue.toLowerCase(),
      );
      return (
        isSearchingForProject &&
        (projectStatusIsPending || isNotFilteredOutByStatus)
      );
    });
  }, [searchValue, sortedProjects, getProjectStatus, filters]);

  useEffect(() => {
    if (!loading && !projectsLoaded) {
      setSelectedProject(filteredProjects[0]);
      setProjectsLoaded(true);
    }
  }, [loading, filteredProjects, projectsLoaded]);

  return {
    filterFormCtx,
    filterStatus,
    handleSortSelect,
    noProjects: projects.length === 0,
    showOnlyAccessible: viewButtonText === 'Your Projects',
    filteredProjects,
    projectCount: filteredProjects.length + '/' + projects.length,
    searchValue,
    setSearchValue,
    sortButtonText,
    selectedProject,
    setSelectedProject,
    sortDropdown,
    viewButtonText,
    handleViewSelect,
    viewDropdown,
  };
};
