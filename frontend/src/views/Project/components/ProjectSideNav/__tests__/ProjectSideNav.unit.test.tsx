import {render, screen, within} from '@testing-library/react';
import React from 'react';

import {withContextProviders, click, type} from '@dash-frontend/testHelpers';

import ProjectSideNavComponent from '..';

describe('project sidenav', () => {
  const ProjectSideNav = withContextProviders(ProjectSideNavComponent);

  describe('create repo modal', () => {
    const validInputs = [
      ['good'],
      ['good-'],
      ['good_'],
      ['good1'],
      ['_good'],
      ['a'.repeat(63)],
    ];
    const invalidInputs = [
      [
        'bad repo',
        'Repo name can only contain alphanumeric characters, underscores, and dashes',
      ],
      [
        'bad!',
        'Repo name can only contain alphanumeric characters, underscores, and dashes',
      ],
      [
        'bad.',
        'Repo name can only contain alphanumeric characters, underscores, and dashes',
      ],
      ['a'.repeat(64), 'Repo name exceeds maximum allowed length'],
    ];
    test.each(validInputs)(
      'should not error with a valid repo name (%j)',
      async (input) => {
        window.history.replaceState('', '', '/project/Empty-Project');

        render(<ProjectSideNav />);

        await click(
          screen.getByRole('button', {
            name: /create/i,
          }),
        );
        await click(
          screen.getByRole('menuitem', {
            name: /input repository/i,
          }),
        );

        const modal = screen.getByRole('dialog');
        const nameInput = await within(modal).findByLabelText('Name', {
          exact: false,
        });

        await type(nameInput, input);

        expect(within(modal).queryByRole('alert')).not.toBeInTheDocument();
      },
    );
    test.each(invalidInputs)(
      'should error with an invalid repo name (%j)',
      async (input, assertionText) => {
        window.history.replaceState('', '', '/project/Empty-Project');

        render(<ProjectSideNav />);

        await click(
          screen.getByRole('button', {
            name: /create/i,
          }),
        );
        await click(
          screen.getByRole('menuitem', {
            name: /input repository/i,
          }),
        );

        const modal = screen.getByRole('dialog');
        const nameInput = await within(modal).findByLabelText('Name', {
          exact: false,
        });

        await type(nameInput, input);

        expect(within(modal).getByText(assertionText)).toBeInTheDocument();
      },
    );
  });
});
