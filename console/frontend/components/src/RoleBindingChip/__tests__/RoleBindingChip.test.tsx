import {render, screen} from '@testing-library/react';
import React from 'react';

import {click} from '@dash-frontend/testHelpers';

import {RoleBindingChip} from '../RoleBindingChip';

describe('RoleBindingChip', () => {
  it('should call the delete role function on an x click', async () => {
    const deleteRoleSpy = jest.fn();

    render(<RoleBindingChip roleName="repoOwner" deleteRole={deleteRoleSpy} />);

    await click(screen.getByLabelText('delete role repoOwner'));

    expect(deleteRoleSpy).toHaveBeenCalledTimes(1);
  });

  it('should not be interactive if locked is true', async () => {
    const deleteRoleSpy = jest.fn();

    render(
      <RoleBindingChip
        roleName="repoOwner"
        locked
        deleteRole={deleteRoleSpy}
      />,
    );

    expect(
      screen.queryByLabelText('delete role repoOwner'),
    ).not.toBeInTheDocument();
    expect(screen.getByLabelText('role repoOwner locked')).toBeInTheDocument();
  });

  it('should not be interactive if readOnly true', async () => {
    const deleteRoleSpy = jest.fn();

    render(
      <RoleBindingChip
        roleName="repoOwner"
        readOnly
        deleteRole={deleteRoleSpy}
      />,
    );

    expect(
      screen.queryByLabelText('delete role repoOwner'),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByLabelText('locked role repoOwner'),
    ).not.toBeInTheDocument();
  });
});
