import React from 'react';

import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import Sidebar from '@dash-frontend/components/Sidebar';
import {TabView} from '@dash-frontend/components/TabView';
import View from '@dash-frontend/components/View';
import {Group, DefaultDropdown, useModal} from '@pachyderm/components';

import CreateProjectModal from './components/CreateProjectModal';
import LandingHeader from './components/LandingHeader';
import LandingSkeleton from './components/LandingSkeleton';
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
    sortDropdown,
  } = useLandingView();
  const {openModal, closeModal, isOpen} = useModal(false);

  if (loading) return <LandingSkeleton />;

  return (
    <>
      <BrandedTitle title="Projects" />
      <LandingHeader projects={projects} />
      {/* Tutorial is temporarily disabled because of "Project" Console Support */}
      <div className={styles.base}>
        <View data-testid="Landing__view">
          <TabView errorMessage="Error loading projects">
            <TabView.Header
              heading="Projects"
              headerButtonText="Create Project"
              headerButtonAction={openModal}
            />
            <TabView.Body initialActiveTabId={'Projects'} showSkeleton={false}>
              <TabView.Body.Header>
                <TabView.Body.Tabs
                  placeholder=""
                  searchValue={searchValue}
                  onSearch={setSearchValue}
                  showSearch={multiProject}
                >
                  <TabView.Body.Tabs.Tab id="Projects" count={projectCount}>
                    Projects
                  </TabView.Body.Tabs.Tab>
                </TabView.Body.Tabs>

                <Group spacing={32}>
                  <DefaultDropdown
                    storeSelected
                    initialSelectId="Name A-Z"
                    onSelect={handleSortSelect}
                    buttonOpts={{disabled: !multiProject}}
                    items={sortDropdown}
                  >
                    Sort by: {sortButtonText}
                  </DefaultDropdown>

                  <TabView.Body.Dropdown
                    formCtx={filterFormCtx}
                    buttonText={filterStatus}
                    disabled={!multiProject}
                  >
                    <TabView.Body.Dropdown.Item
                      name="HEALTHY"
                      id="HEALTHY"
                      label="Healthy"
                    />
                    <TabView.Body.Dropdown.Item
                      name="UNHEALTHY"
                      id="UNHEALTHY"
                      label="Unhealthy"
                    />
                  </TabView.Body.Dropdown>
                </Group>
              </TabView.Body.Header>
              <TabView.Body.Content id={'Projects'}>
                <table className={styles.table}>
                  <tbody>
                    {projects.map((project) => (
                      <ProjectRow
                        multiProject={multiProject}
                        project={project}
                        key={project.id}
                        setSelectedProject={() => setSelectedProject(project)}
                        isSelected={project.id === selectedProject?.id}
                      />
                    ))}
                  </tbody>
                </table>
              </TabView.Body.Content>
              <TabView.Body.Content id="Personal" />
              <TabView.Body.Content id="Playground" />
            </TabView.Body>
          </TabView>
        </View>
        <Sidebar>
          {selectedProject && <ProjectPreview project={selectedProject} />}
        </Sidebar>
        {isOpen && <CreateProjectModal show={isOpen} onHide={closeModal} />}
      </div>
    </>
  );
};

export default Landing;
