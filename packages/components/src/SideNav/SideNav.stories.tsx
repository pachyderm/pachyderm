import React from 'react';
import {MemoryRouter} from 'react-router-dom';

import {
  SettingsSVG,
  SupportSVG,
  DirectionsSVG,
  RepoSVG,
  PipelineSVG,
  JobsSVG,
} from '../Svg';

import {SideNav} from './';

export default {title: 'SideNav'};

export const DarkMode = () => (
  <MemoryRouter initialEntries={['/workspaces']}>
    <div
      style={{
        display: 'flex',
        position: 'relative',
        flex: '1 0 auto',
        height: '100%',
      }}
    >
      <SideNav breakpoint={1200}>
        <SideNav.SideNavList noPadding>
          <SideNav.SideNavItem>
            <SideNav.SideNavLink
              IconSVG={DirectionsSVG}
              to="workspaces"
              tooltipContent="Workspaces"
            >
              Workspaces
            </SideNav.SideNavLink>
          </SideNav.SideNavItem>
          <SideNav.SideNavItem>
            <SideNav.SideNavLink
              IconSVG={SettingsSVG}
              to="settings"
              tooltipContent="Settings"
            >
              Settings
            </SideNav.SideNavLink>
          </SideNav.SideNavItem>
          <SideNav.SideNavItem>
            <SideNav.SideNavLink
              IconSVG={SupportSVG}
              to="support"
              tooltipContent="Support"
              disabled
            >
              Support
            </SideNav.SideNavLink>
          </SideNav.SideNavItem>
        </SideNav.SideNavList>
      </SideNav>
    </div>
  </MemoryRouter>
);

export const LightMode = () => (
  <MemoryRouter initialEntries={['/repos']}>
    <SideNav breakpoint={1200} styleMode="light">
      <SideNav.SideNavList>
        <SideNav.SideNavButton
          IconSVG={DirectionsSVG}
          to="/"
          tooltipContent="Switch View"
        >
          View Lineage
        </SideNav.SideNavButton>
      </SideNav.SideNavList>
      <SideNav.SideNavList>
        <SideNav.SideNavItem>
          <SideNav.SideNavLink
            IconSVG={RepoSVG}
            to="repos"
            tooltipContent="Repos"
            styleMode="light"
            showIconWhenExpanded
          >
            Repositories
          </SideNav.SideNavLink>
        </SideNav.SideNavItem>
        <SideNav.SideNavItem>
          <SideNav.SideNavLink
            IconSVG={PipelineSVG}
            to="pipelines"
            tooltipContent="Pipelines"
            styleMode="light"
            showIconWhenExpanded
          >
            Pipelines
          </SideNav.SideNavLink>
        </SideNav.SideNavItem>
      </SideNav.SideNavList>
      <SideNav.SideNavList>
        <SideNav.SideNavItem>
          <SideNav.SideNavLink
            IconSVG={JobsSVG}
            to="settings"
            tooltipContent="Settings"
            styleMode="light"
            showIconWhenExpanded
          >
            Jobs
          </SideNav.SideNavLink>
        </SideNav.SideNavItem>
        <SideNav.SideNavItem>
          <SideNav.SideNavLink
            IconSVG={SupportSVG}
            to="support"
            tooltipContent="Support"
            styleMode="light"
            disabled
            showIconWhenExpanded
          >
            Support
          </SideNav.SideNavLink>
        </SideNav.SideNavItem>
      </SideNav.SideNavList>
    </SideNav>
  </MemoryRouter>
);
