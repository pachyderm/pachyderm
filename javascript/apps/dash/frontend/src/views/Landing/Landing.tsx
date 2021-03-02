import {Group, TableView} from '@pachyderm/components';
import React, {useState} from 'react';
import {useForm} from 'react-hook-form';

import {useProjects} from 'hooks/useProjects';

import ProjectRow from './components/ProjectRow/ProjectRow';
import styles from './Landing.module.css';

const Landing: React.FC = () => {
  const [searchValue, setSearchValue] = useState('');
  const formCtx = useForm();
  const {projects} = useProjects();

  return (
    <div className={styles.wrapper}>
      <TableView title="Projects" errorMessage="Error loading projects">
        <TableView.Header heading="Projects" headerButtonHidden />
        <TableView.Body initialActiveTabId="Public" showSkeleton={false}>
          <TableView.Body.Header>
            <TableView.Body.Tabs
              placeholder=""
              searchValue={searchValue}
              onSearch={setSearchValue}
              showSearch
            >
              <TableView.Body.Tabs.Tab id="Public" count={projects.length}>
                Public
              </TableView.Body.Tabs.Tab>
              <TableView.Body.Tabs.Tab id="Private" count={0}>
                Private
              </TableView.Body.Tabs.Tab>
              <TableView.Body.Tabs.Tab id="Playground" count={0}>
                Private
              </TableView.Body.Tabs.Tab>
            </TableView.Body.Tabs>
            <Group spacing={32}>
              <TableView.Body.Dropdown
                formCtx={formCtx}
                buttonText="Sort by: Created On"
              />
              <TableView.Body.Dropdown
                formCtx={formCtx}
                buttonText=" Filter by: None"
              />
            </Group>
          </TableView.Body.Header>
          <TableView.Body.Content id="Public">
            <table className={styles.table}>
              <tbody>
                {projects.map((project) => (
                  <ProjectRow project={project} key={project.id} />
                ))}
              </tbody>
            </table>
          </TableView.Body.Content>
          <TableView.Body.Content id="Private" />
          <TableView.Body.Content id="Playground" />
        </TableView.Body>
      </TableView>
    </div>
  );
};

export default Landing;
