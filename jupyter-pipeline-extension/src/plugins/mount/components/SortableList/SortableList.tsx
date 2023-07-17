import React, {useCallback, useState} from 'react';
import {
  Repo,
  Mount,
  ListMountsResponse,
  Project,
  ProjectInfo,
} from '../../types';
import {caretUpIcon, caretDownIcon} from '@jupyterlab/ui-components';
import {useSort, stringComparator} from '../../../../utils/hooks/useSort';
import ListMount from './ListMount';
import ListUnmount from './ListUnmount';

type SortableListProps = {
  items: Array<Mount | Repo>;
  open: (path: string) => void;
  updateData: (data: ListMountsResponse) => void;
  mountedItems: Mount[];
  type: string;
  projects: ProjectInfo[];
};

const nameComparator = {
  name: 'Name',
  func: stringComparator,
  accessor: (item: Mount | Repo) =>
    'name' in item ? item.name : `${item.project}_${item.repo}`,
};

const SortableList: React.FC<SortableListProps> = ({
  open,
  items,
  updateData,
  mountedItems,
  type,
  projects,
}) => {
  const allProjectsLabel = 'All projects';
  const projectsList: string[] = projects.map(
    (project: ProjectInfo): string => project.project.name,
  );
  const [selectedProject, setSelectedProject] =
    useState<string>(allProjectsLabel);

  const onChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedProject(e.target.value);
  };

  const {sortedData, setComparator, reversed} = useSort<Mount | Repo>({
    data: items,
    initialSort: nameComparator,
    initialDirection: 1,
  });

  const nameClick = useCallback(() => {
    setComparator(nameComparator);
  }, [setComparator]);

  return (
    <div className="pachyderm-mount-sortableList">
      {type === 'unmounted' && (
        <div className="pachyderm-mount-sortableList-projectFilter">
          <div className="pachyderm-mount-projectFilter-headerItem">
            <span>Project: </span>
            <select
              name="project"
              value={selectedProject}
              className="pachyderm-mount-project-list-select"
              onChange={onChange}
              data-testid="ProjectList__select"
            >
              <option key={allProjectsLabel} value={allProjectsLabel}>
                {allProjectsLabel}
              </option>
              {projectsList.map((project) => (
                <option key={project} value={project}>
                  {project}
                </option>
              ))}
            </select>
          </div>
        </div>
      )}
      <div className="pachyderm-mount-sortableList-header">
        <div
          className="pachyderm-mount-sortableList-headerItem"
          onClick={nameClick}
        >
          <span>Name</span>
          <span>
            {reversed ? <caretDownIcon.react /> : <caretUpIcon.react />}
          </span>
        </div>
      </div>
      <ul className="pachyderm-mount-sortableList-content">
        {sortedData &&
          sortedData.map((item: Mount | Repo) =>
            'name' in item ? (
              <ListMount
                item={item}
                key={item.name}
                open={open}
                updateData={updateData}
              />
            ) : (
              (selectedProject === item.project ||
                selectedProject === allProjectsLabel) && (
                <ListUnmount
                  item={item}
                  key={`${item.project}_${item.repo}`}
                  updateData={updateData}
                  mountedItems={mountedItems}
                />
              )
            ),
          )}
      </ul>
    </div>
  );
};

export default SortableList;
