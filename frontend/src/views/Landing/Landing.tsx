import React from 'react';
import Favicon from 'react-favicon';
import {useHistory} from 'react-router';
import {Route, Switch} from 'react-router-dom';

import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import EmptyState from '@dash-frontend/components/EmptyState';
import Sidebar from '@dash-frontend/components/Sidebar';
import {TabView} from '@dash-frontend/components/TabView';
import View from '@dash-frontend/components/View';
import {
  useAuthorize,
  Permission,
  ResourceType,
} from '@dash-frontend/hooks/useAuthorize';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';
import {
  Group,
  DefaultDropdown,
  useModal,
  Button,
  useNotificationBanner,
} from '@pachyderm/components';

import {CLUSTER_CONFIG} from '../Project/constants/projectPaths';
import {clusterConfigRoute} from '../Project/utils/routes';

import ClusterConfig from './components/ClusterConfig';
import CreateProjectModal from './components/CreateProjectModal';
import LandingHeader from './components/LandingHeader';
import ProjectPreview from './components/ProjectPreview';
import ProjectRow from './components/ProjectRow';
import {useLandingView} from './hooks/useLandingView';
import styles from './Landing.module.css';

const LandingRouter = () => {
  const {loading: loadingEnterprise, enterpriseActive} = useEnterpriseActive();
  const {add} = useNotificationBanner();

  return (
    <>
      <BrandedTitle title="Projects" />
      {/* Safari will not progamatically update its icon. So we serve HPE
    favicon from index.html and set the pachyderm icon when we know we
    are not in enterprise. */}
      {!loadingEnterprise && !enterpriseActive && (
        <Favicon url="/img/pachyderm.ico" />
      )}
      <LandingHeader />
      <Switch>
        <Route path="/" exact component={Landing} />
        <Route
          path={CLUSTER_CONFIG}
          exact
          component={() => <ClusterConfig triggerNotification={add} />}
        />
      </Switch>
    </>
  );
};

export const Landing: React.FC = () => {
  const {
    filterFormCtx,
    filterStatus,
    handleSortSelect,
    noProjects,
    filteredProjects,
    showOnlyAccessible,
    projectCount,
    searchValue,
    setSearchValue,
    sortButtonText,
    selectedProject,
    setSelectedProject,
    sortDropdown,
    viewButtonText,
    handleViewSelect,
    viewDropdown,
  } = useLandingView();
  const browserHistory = useHistory();
  const {openModal, closeModal, isOpen} = useModal(false);

  const {
    isAuthActive,
    hasProjectCreate,
    hasClusterAuthSetConfig: hasClusterConfig,
  } = useAuthorize({
    permissions: [
      Permission.PROJECT_CREATE,
      Permission.CLUSTER_AUTH_SET_CONFIG,
    ],
    resource: {type: ResourceType.CLUSTER, name: ''},
  });

  return (
    <div className={styles.base}>
      <View data-testid="Landing__view">
        <TabView errorMessage="Error loading projects">
          <TabView.Header heading="Projects">
            {hasClusterConfig && (
              <Button
                onClick={() => browserHistory.push(clusterConfigRoute)}
                buttonType="secondary"
              >
                Cluster Defaults
              </Button>
            )}
            {hasProjectCreate && (
              <Button onClick={openModal}>Create Project</Button>
            )}
          </TabView.Header>
          <TabView.Body initialActiveTabId={'Projects'} showSkeleton={false}>
            <TabView.Body.Header>
              <TabView.Body.Tabs
                placeholder=""
                searchValue={searchValue}
                onSearch={setSearchValue}
                showSearch
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
                  items={sortDropdown}
                >
                  Sort by: {sortButtonText}
                </DefaultDropdown>

                {isAuthActive && (
                  <DefaultDropdown
                    storeSelected
                    initialSelectId="Your Projects"
                    onSelect={handleViewSelect}
                    items={viewDropdown}
                  >
                    View: {viewButtonText}
                  </DefaultDropdown>
                )}
                <TabView.Body.Dropdown
                  formCtx={filterFormCtx}
                  buttonText={filterStatus}
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
              <Group className={styles.projectsList} spacing={16} vertical>
                {noProjects ? (
                  <EmptyState
                    title="No projects exist."
                    message="Create a project to get started."
                  />
                ) : filteredProjects.length === 0 ? (
                  <EmptyState
                    title="No projects match your current filters."
                    message="Try adjusting or resetting your filters to see more projects."
                  />
                ) : (
                  filteredProjects.map((project) => (
                    <ProjectRow
                      showOnlyAccessible={showOnlyAccessible}
                      multiProject={true}
                      project={project}
                      key={project?.project?.name}
                      setSelectedProject={() => setSelectedProject(project)}
                      isSelected={
                        project?.project?.name ===
                        selectedProject?.project?.name
                      }
                    />
                  ))
                )}
              </Group>
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
  );
};

export default LandingRouter;
