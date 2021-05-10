import {Group, TableView, DefaultDropdown} from '@pachyderm/components';
import React from 'react';

import Sidebar from '@dash-frontend/components/Sidebar';
import View from '@dash-frontend/components/View';

import LandingHeader from './components/LandingHeader';
import ProjectPreview from './components/ProjectPreview';
import ProjectRow from './components/ProjectRow';
import {useLandingView} from './hooks/useLandingView';
import styles from './Landing.module.css';

const Landing: React.FC = () => {
  const {
    filterFormCtx,
    filterStatus,
    handleSortSelect,
    loading,
    multiProject,
    projects,
    projectCount,
    searchValue,
    setSearchValue,
    sortButtonText,
    selectedProject,
    setSelectedProject,
  } = useLandingView();

  if (loading) return null; // TODO loading view

  return (
    <>
      <LandingHeader />
      <div className={styles.base}>
        <View>
          <TableView title="Projects" errorMessage="Error loading projects">
            <TableView.Header heading="Projects" headerButtonHidden />
            <TableView.Body
              initialActiveTabId={multiProject ? 'Team' : 'All'}
              showSkeleton={false}
            >
              <TableView.Body.Header>
                <TableView.Body.Tabs
                  placeholder=""
                  searchValue={searchValue}
                  onSearch={setSearchValue}
                  showSearch
                >
                  {multiProject ? (
                    <>
                      <TableView.Body.Tabs.Tab id="Team" count={projectCount}>
                        Team
                      </TableView.Body.Tabs.Tab>
                      <TableView.Body.Tabs.Tab id="Personal" count={0}>
                        Personal
                      </TableView.Body.Tabs.Tab>
                      <TableView.Body.Tabs.Tab id="Playground" count={0}>
                        Playground
                      </TableView.Body.Tabs.Tab>
                    </>
                  ) : (
                    <TableView.Body.Tabs.Tab id="All" count={projectCount}>
                      All
                    </TableView.Body.Tabs.Tab>
                  )}
                </TableView.Body.Tabs>
                <Group spacing={32}>
                  <DefaultDropdown
                    storeSelected
                    initialSelectId="Created On"
                    onSelect={handleSortSelect}
                    items={[
                      {id: 'Created On', content: 'Created On'},
                      {id: 'Name A-Z', content: 'Name A-Z'},
                    ]}
                  >
                    Sort by: {sortButtonText}
                  </DefaultDropdown>

                  <TableView.Body.Dropdown
                    formCtx={filterFormCtx}
                    buttonText={filterStatus}
                  >
                    <TableView.Body.Dropdown.Item
                      name="HEALTHY"
                      id="HEALTHY"
                      label="Healthy"
                    />
                    <TableView.Body.Dropdown.Item
                      name="UNHEALTHY"
                      id="UNHEALTHY"
                      label="Unhealthy"
                    />
                  </TableView.Body.Dropdown>
                </Group>
              </TableView.Body.Header>
              <TableView.Body.Content id={multiProject ? 'Team' : 'All'}>
                <table className={styles.table}>
                  <tbody>
                    {projects.map((project) => (
                      <ProjectRow
                        project={project}
                        key={project.id}
                        setSelectedProject={() => setSelectedProject(project)}
                        isSelected={project.id === selectedProject?.id}
                      />
                    ))}
                  </tbody>
                </table>
              </TableView.Body.Content>
              <TableView.Body.Content id="Personal" />
              <TableView.Body.Content id="Playground" />
            </TableView.Body>
          </TableView>
        </View>
        <Sidebar>
          {selectedProject && <ProjectPreview project={selectedProject} />}
        </Sidebar>
      </div>
    </>
  );
};

export default Landing;
