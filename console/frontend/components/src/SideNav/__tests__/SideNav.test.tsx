import {render, screen} from '@testing-library/react';
import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {click} from '@dash-frontend/testHelpers';

import {SupportSVG, DirectionsSVG, AddCircleSVG} from '../../Svg';
import SideNav from '../SideNav';

describe('SideNav', () => {
  const renderTestbed = (clickFunc?: () => void) => {
    render(
      <BrowserRouter>
        <SideNav breakpoint={200}>
          <SideNav.SideNavList>
            <SideNav.SideNavItem
              IconSVG={AddCircleSVG}
              onClick={clickFunc}
              tooltipContent={'Create'}
            >
              Create
            </SideNav.SideNavItem>
            <SideNav.SideNavItem
              IconSVG={DirectionsSVG}
              to={`/workspaces`}
              tooltipContent={'Workspaces'}
            >
              Workspaces
            </SideNav.SideNavItem>
            <SideNav.SideNavItem
              IconSVG={SupportSVG}
              to={`/members`}
              tooltipContent={'Sync'}
            >
              Members
            </SideNav.SideNavItem>
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

  it('should allow click functions for nav item selection', async () => {
    const clickFunc = jest.fn();

    renderTestbed(clickFunc);

    const create = await screen.findByText('Create');

    expect(clickFunc).toHaveBeenCalledTimes(0);

    await click(create);

    expect(clickFunc).toHaveBeenCalledTimes(1);
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
