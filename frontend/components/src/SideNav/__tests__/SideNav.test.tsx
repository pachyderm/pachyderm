import {render, screen} from '@testing-library/react';
import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {click} from '@dash-frontend/testHelpers';

import {SupportSVG, DirectionsSVG} from '../../Svg';
import SideNav from '../SideNav';

describe('SideNav', () => {
  const renderTestbed = () => {
    render(
      <BrowserRouter>
        <SideNav breakpoint={200}>
          <SideNav.SideNavList>
            <SideNav.SideNavLink
              IconSVG={DirectionsSVG}
              to={`/workspaces`}
              tooltipContent={'Workspaces'}
            >
              Workspaces
            </SideNav.SideNavLink>
            <SideNav.SideNavLink
              IconSVG={SupportSVG}
              to={`/members`}
              tooltipContent={'Sync'}
            >
              Members
            </SideNav.SideNavLink>
          </SideNav.SideNavList>
        </SideNav>
      </BrowserRouter>,
    );
  };

  it('should allow nav item selection and show the selection', async () => {
    renderTestbed();

    const members = await screen.findByText('Members');

    await click(members);

    expect(window.location.pathname).toBe('/members');
    expect(members.closest('a')).toHaveAttribute('aria-current', 'page');
  });

  it('should be able to be minimized and expanded', async () => {
    renderTestbed();

    let workspaces = screen.queryByText('Workspaces');
    const expandAndCollapse = await screen.findByTestId('SideNav__toggle');
    expect(workspaces).toBeVisible();

    await click(expandAndCollapse);
    workspaces = screen.queryByText('Workspaces');
    expect(workspaces).not.toBeInTheDocument();

    await click(expandAndCollapse);
    workspaces = screen.queryByText('Workspaces');
    expect(workspaces).toBeVisible();
  });
});
