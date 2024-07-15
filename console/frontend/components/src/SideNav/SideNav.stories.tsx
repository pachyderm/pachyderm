import React from 'react';
import {MemoryRouter} from 'react-router-dom';

import {
  AddCircleSVG,
  SupportSVG,
  DirectionsSVG,
  RepoSVG,
  PipelineSVG,
  JobsSVG,
} from '../Svg';

import {SideNav} from './';

export default {title: 'SideNav'};

export const Default = () => (
  <MemoryRouter initialEntries={['/repos']}>
    <SideNav breakpoint={1200}>
      <SideNav.SideNavList>
        <SideNav.SideNavItem
          IconSVG={DirectionsSVG}
          to="/dag"
          tooltipContent="DAG"
          showIconWhenExpanded
        >
          DAG
        </SideNav.SideNavItem>
        <SideNav.SideNavItem
          IconSVG={AddCircleSVG}
          onClick={() => alert('Repo Creation')}
          tooltipContent="Create New Repo"
          showIconWhenExpanded
        >
          Create Repo
        </SideNav.SideNavItem>
      </SideNav.SideNavList>
      <SideNav.SideNavList label="Lists">
        <SideNav.SideNavItem
          IconSVG={JobsSVG}
          to="/jobs"
          tooltipContent="Jobs"
          showIconWhenExpanded
        >
          Jobs
        </SideNav.SideNavItem>
        <SideNav.SideNavItem
          IconSVG={PipelineSVG}
          to="/pipelines"
          tooltipContent="Pipelines"
          showIconWhenExpanded
        >
          Pipelines
        </SideNav.SideNavItem>
        <SideNav.SideNavItem
          IconSVG={RepoSVG}
          to="/repos"
          tooltipContent="Repos"
          showIconWhenExpanded
        >
          Repositories
        </SideNav.SideNavItem>
      </SideNav.SideNavList>
      <SideNav.SideNavList label="Settings" border>
        <SideNav.SideNavItem
          IconSVG={SupportSVG}
          to="support"
          tooltipContent="Support"
          disabled
          showIconWhenExpanded
        >
          Support
        </SideNav.SideNavItem>
      </SideNav.SideNavList>
    </SideNav>
  </MemoryRouter>
);
