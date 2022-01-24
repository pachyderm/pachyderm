import {render} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {withContextProviders} from 'testHelpers';

import {UserSVG, WorkspacesSVG} from '../../Svg';
import SideNav from '../SideNav';

describe('SideNav', () => {
  const renderTestBed = () => {
    const TestBed = withContextProviders(() => (
      <BrowserRouter>
        <SideNav breakpoint={200}>
          <SideNav.SideNavList>
            <SideNav.SideNavLink
              IconSVG={WorkspacesSVG}
              to={`/workspaces`}
              tooltipContent={'Workspaces'}
            >
              Workspaces
            </SideNav.SideNavLink>
            <SideNav.SideNavLink
              IconSVG={UserSVG}
              to={`/members`}
              tooltipContent={'Sync'}
            >
              Members
            </SideNav.SideNavLink>
          </SideNav.SideNavList>
        </SideNav>
      </BrowserRouter>
    ));

    return render(<TestBed />);
  };

  it('should allow nav item selection and show the selection', async () => {
    const {findByText} = renderTestBed();

    const members = await findByText('Members');

    userEvent.click(members);

    expect(window.location.pathname).toBe('/members');
    expect(members.closest('a')).toHaveAttribute('aria-current', 'page');
  });

  it('should be able to be minimized and expanded', async () => {
    const {findByTestId, queryByText} = renderTestBed();

    let workspaces = queryByText('Workspaces');
    const expandAndCollapse = await findByTestId('SideNav__toggle');
    expect(workspaces).toBeVisible();

    userEvent.click(expandAndCollapse);
    workspaces = queryByText('Workspaces');
    expect(workspaces).toBeNull();

    userEvent.click(expandAndCollapse);
    workspaces = queryByText('Workspaces');
    expect(workspaces).toBeVisible();
  });
});
