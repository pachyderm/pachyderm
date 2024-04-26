import React from 'react';
import Favicon from 'react-favicon';
import {useHistory} from 'react-router';
import {Route, Switch} from 'react-router-dom';

import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import ClusterRolesModal from '@dash-frontend/components/ClusterRolesModal';
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
import ProjectTab from './components/ProjectTab';
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
    myProjectsCount,
    setMyProjectsCount,
    allProjectsCount,
    setAllProjectsCount,
    searchValue,
    setSearchValue,
    sortButtonText,
    selectedProject,
    setSelectedProject,
    sortDropdown,
  } = useLandingView();
  const browserHistory = useHistory();
  const {
    openModal: openCreateModal,
    closeModal: closeCreateModal,
    isOpen: createIsOpen,
  } = useModal(false);
  const {
    openModal: openRolesModal,
    closeModal: closeRolesModal,
    isOpen: rolesIsOpen,
  } = useModal(false);

  const {
    isAuthActive,
    hasProjectCreate,
    hasClusterAuthSetConfig: hasClusterConfig,
    hasClusterModifyBindings,
  } = useAuthorize({
    permissions: [
      Permission.PROJECT_CREATE,
      Permission.CLUSTER_AUTH_SET_CONFIG,
      Permission.CLUSTER_MODIFY_BINDINGS,
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
            {isAuthActive && (
              <Button onClick={openRolesModal} buttonType="secondary">
                Cluster Roles
              </Button>
            )}
            {hasProjectCreate && (
              <Button onClick={openCreateModal}>Create Project</Button>
            )}
          </TabView.Header>
          <TabView.Body initialActiveTabId="myProjects" showSkeleton={false}>
            <TabView.Body.Header>
              <TabView.Body.Tabs
                placeholder=""
                searchValue={searchValue}
                onSearch={setSearchValue}
                showSearch
              >
                <TabView.Body.Tabs.Tab id="myProjects" count={myProjectsCount}>
                  Your Projects
                </TabView.Body.Tabs.Tab>
                {isAuthActive && (
                  <TabView.Body.Tabs.Tab
                    id="allProjects"
                    count={allProjectsCount}
                  >
                    All Projects
                  </TabView.Body.Tabs.Tab>
                )}
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
            <TabView.Body.Content id="myProjects">
              <ProjectTab
                noProjects={noProjects}
                filteredProjects={filteredProjects}
                selectedProject={selectedProject}
                showOnlyAccessible={true}
                setSelectedProject={setSelectedProject}
                setMyProjectsCount={setMyProjectsCount}
                setAllProjectsCount={setAllProjectsCount}
              />
            </TabView.Body.Content>
            {isAuthActive && (
              <TabView.Body.Content id="allProjects">
                <ProjectTab
                  noProjects={noProjects}
                  filteredProjects={filteredProjects}
                  selectedProject={selectedProject}
                  showOnlyAccessible={false}
                  setSelectedProject={setSelectedProject}
                  setMyProjectsCount={setMyProjectsCount}
                  setAllProjectsCount={setAllProjectsCount}
                />
              </TabView.Body.Content>
            )}
          </TabView.Body>
        </TabView>
      </View>
      <Sidebar>
        {selectedProject && <ProjectPreview project={selectedProject} />}
      </Sidebar>
      {createIsOpen && (
        <CreateProjectModal show={createIsOpen} onHide={closeCreateModal} />
      )}
      {rolesIsOpen && (
        <ClusterRolesModal
          show={rolesIsOpen}
          onHide={closeRolesModal}
          readOnly={!hasClusterModifyBindings}
        />
      )}
    </div>
  );
};

export default LandingRouter;
