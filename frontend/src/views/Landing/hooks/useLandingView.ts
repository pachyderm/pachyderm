import capitalize from 'lodash/capitalize';
import every from 'lodash/every';
import reduce from 'lodash/reduce';
import {useCallback, useEffect, useMemo, useState} from 'react';
import {useForm} from 'react-hook-form';

import {ProjectInfo} from '@dash-frontend/api/pfs';
import {usePipelineSummaries} from '@dash-frontend/hooks/usePipelineSummary';
import {useProjects} from '@dash-frontend/hooks/useProjects';
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
  const allProjectNames = projects.map((p) => p.project?.name);
  const {pipelineSummaries} = usePipelineSummaries(allProjectNames);
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
  const [selectedProject, setSelectedProject] = useState<string>();
  const [projectsLoaded, setProjectsLoaded] = useState(false);
  const [myProjectsCount, setMyProjectsCount] = useState(0);
  const [allProjectsCount, setAllProjectsCount] = useState(0);

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
      const pipelineSummary = pipelineSummaries?.find(
        (summary) => summary.summary.project?.name === project?.project?.name,
      );
      const projectStatus = pipelineSummary?.projectStatus;
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
  }, [sortedProjects, pipelineSummaries, filters, searchValue]);

  useEffect(() => {
    if (!loading && !projectsLoaded) {
      setSelectedProject(filteredProjects[0].project?.name);
      setProjectsLoaded(true);
    }
  }, [loading, filteredProjects, projectsLoaded]);

  const selectedProjectInfo = useMemo(() => {
    return projects.find(
      (project) => project.project?.name === selectedProject,
    );
  }, [projects, selectedProject]);

  return {
    filterFormCtx,
    filterStatus,
    handleSortSelect,
    noProjects: projects.length === 0,
    filteredProjects,
    myProjectsCount,
    setMyProjectsCount,
    allProjectsCount,
    setAllProjectsCount,
    allProjectNames,
    searchValue,
    setSearchValue,
    sortButtonText,
    selectedProject,
    selectedProjectInfo,
    setSelectedProject,
    sortDropdown,
  };
};
