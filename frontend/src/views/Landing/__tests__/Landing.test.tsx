import {mockServer} from '@dash-backend/testHelpers';
import {
  render,
  waitFor,
  within,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import React from 'react';

import {
  withContextProviders,
  click,
  clear,
  type,
} from '@dash-frontend/testHelpers';

import LandingComponent from '../Landing';

describe('Landing', () => {
  const Landing = withContextProviders(() => {
    return <LandingComponent />;
  });

  beforeEach(() => {
    window.localStorage.setItem(
      'pachyderm-console-account-data',
      '{"tutorial_introduction_seen": true}',
    );
  });

  afterEach(() => {
    window.localStorage.removeItem('pachyderm-console-account-data');
  });

  it('should display all projects with only a single tab', async () => {
    render(<Landing />);

    expect(
      await screen.findByRole('heading', {
        name: 'Data-Cleaning-Process',
        level: 5,
      }),
    ).toBeInTheDocument();

    expect(await screen.findAllByRole('row', {})).toHaveLength(11);

    expect(
      await screen.findByRole('tab', {
        name: /projects/i,
      }),
    ).toBeInTheDocument();

    expect(screen.getByRole('tab')).toBeInTheDocument();
  });

  it('should allow users to non-case sensitive search for projects by name', async () => {
    render(<Landing />);

    expect(
      await screen.findByRole('heading', {
        name: 'Data-Cleaning-Process',
        level: 5,
      }),
    ).toBeInTheDocument();
    expect(
      await screen.findByRole('heading', {
        name: 'Solar-Power-Data-Logger-Team-Collab',
        level: 5,
      }),
    ).toBeInTheDocument();

    const searchBox = await screen.findByRole('searchbox');

    await type(searchBox, 'Data-Cleaning-Process');

    await waitFor(() =>
      expect(
        screen.queryByRole('heading', {
          name: 'Solar-Power-Data-Logger-Team-Collab',
          level: 5,
        }),
      ).not.toBeInTheDocument(),
    );
    await screen.findByRole('heading', {
      name: 'Data-Cleaning-Process',
      level: 5,
    });
  });

  it('should display the project names', async () => {
    render(<Landing />);
    expect(
      await screen.findByRole('heading', {
        name: 'Solar-Power-Data-Logger-Team-Collab',
        level: 5,
      }),
    ).toBeInTheDocument();
  });

  it('should display the project status', async () => {
    render(<Landing />);

    expect(await screen.findAllByTestId('ProjectStatus__HEALTHY')).toHaveLength(
      7,
    );
    expect(
      await screen.findAllByTestId('ProjectStatus__UNHEALTHY'),
    ).toHaveLength(4);
  });

  it('should display project descriptions', async () => {
    render(<Landing />);
    expect(
      await screen.findAllByText(
        'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
      ),
    ).toHaveLength(6);
  });

  it('should allow a user to view a project based in default lineage view', async () => {
    render(<Landing />);

    expect(window.location.pathname).not.toBe('/lineage/Data-Cleaning-Process');

    const viewProjectButton = await screen.findByRole('button', {
      name: /View project Data-Cleaning-Process/i,
    });
    await click(viewProjectButton);
    expect(window.location.pathname).toBe('/lineage/Data-Cleaning-Process');
  });

  it('should allow the user to sort by name', async () => {
    render(<Landing />);

    expect(
      await screen.findByRole('heading', {
        name: 'Data-Cleaning-Process',
        level: 5,
      }),
    ).toBeInTheDocument();

    const projectsPanel = screen.getByRole('tabpanel', {
      name: /projects 11/i,
    });
    const projectNamesAZ = within(projectsPanel).getAllByRole('heading', {
      level: 5,
    });

    // Starts in A-Z order
    expect(
      projectNamesAZ.map((projectName) => projectName.textContent),
    ).toEqual([
      'Data-Cleaning-Process',
      'Egress-Examples',
      'Empty-Project',
      'Load-Project',
      'Multi-Project-Pipeline-A',
      'Multi-Project-Pipeline-B',
      'Pipelines-Project',
      'Solar-Panel-Data-Sorting',
      'Solar-Power-Data-Logger-Team-Collab',
      'Solar-Price-Prediction-Modal',
      'Trait-Discovery',
    ]);

    const sortDropdown = await screen.findByRole('button', {
      name: 'Sort by: Name A-Z',
    });
    await click(sortDropdown);
    const nameSort = await screen.findByRole('menuitem', {name: 'Name Z-A'});
    await click(nameSort);

    const projectNamesZA = within(projectsPanel).getAllByRole('heading', {
      level: 5,
    });
    expect(
      projectNamesZA.map((projectName) => projectName.textContent),
    ).toEqual([
      'Trait-Discovery',
      'Solar-Price-Prediction-Modal',
      'Solar-Power-Data-Logger-Team-Collab',
      'Solar-Panel-Data-Sorting',
      'Pipelines-Project',
      'Multi-Project-Pipeline-B',
      'Multi-Project-Pipeline-A',
      'Load-Project',
      'Empty-Project',
      'Egress-Examples',
      'Data-Cleaning-Process',
    ]);
  });

  it('should allow the user to filter projects by status', async () => {
    render(<Landing />);
    const projects = await screen.findByTestId('Landing__view');
    expect(
      await within(projects).findAllByTestId('ProjectStatus__HEALTHY'),
    ).toHaveLength(7);
    expect(
      await within(projects).findAllByTestId('ProjectStatus__UNHEALTHY'),
    ).toHaveLength(4);

    const filterDropdown = await screen.findByRole('button', {
      name: 'Show: All',
    });
    await click(filterDropdown);
    const healthyButton = await screen.findByLabelText('Healthy');
    await click(healthyButton);

    expect(
      await screen.findByRole('button', {name: 'Show: Unhealthy'}),
    ).toBeInTheDocument();

    expect(
      await within(projects).findAllByTestId('ProjectStatus__UNHEALTHY'),
    ).toHaveLength(4);
    expect(
      within(projects).queryByTestId('ProjectStatus__HEALTHY'),
    ).not.toBeInTheDocument();

    const unHealthyButton = await screen.findByLabelText('Unhealthy');
    await click(unHealthyButton);

    expect(
      await screen.findByRole('button', {name: 'Show: None'}),
    ).toBeInTheDocument();

    expect(
      within(projects).queryByTestId('ProjectStatus__HEALTHY'),
    ).not.toBeInTheDocument();
    expect(
      within(projects).queryByTestId('ProjectStatus__UNHEALTHY'),
    ).not.toBeInTheDocument();

    await click(healthyButton);

    expect(
      await screen.findByRole('button', {name: 'Show: Healthy'}),
    ).toBeInTheDocument();

    expect(
      await within(projects).findAllByTestId('ProjectStatus__HEALTHY'),
    ).toHaveLength(7);
    expect(
      within(projects).queryByTestId('ProjectStatus__UNHEALTHY'),
    ).not.toBeInTheDocument();
  });

  it('should display the project details', async () => {
    render(<Landing />);
    await click(await screen.findByText('Solar-Panel-Data-Sorting'));

    expect(await screen.findByText('3/2')).toBeInTheDocument();
    expect(await screen.findByText('3 kB')).toBeInTheDocument();
  });

  it('should not display the project details when the project is empty', async () => {
    render(<Landing />);
    await click(await screen.findByText('Empty-Project'));

    expect(
      await screen.findByText('Create your first repo/pipeline!'),
    ).toBeInTheDocument();
  });

  it('should give users a pachctl command to set their active project', async () => {
    render(<Landing />);

    expect(
      await screen.findByText('Data-Cleaning-Process'),
    ).toBeInTheDocument();

    await click(
      screen.getByRole('button', {
        name: /data-cleaning-process overflow menu/i,
      }),
    );
    await click(
      await screen.findByRole('menuitem', {
        name: /set active project/i,
      }),
    );

    expect(
      await screen.findByText('Set Active Project: "Data-Cleaning-Process"'),
    ).toBeInTheDocument();
    await click(screen.getByRole('button', {name: 'Copy'}));

    expect(window.document.execCommand).toHaveBeenCalledWith('copy');
  });

  describe('Create Project Modal', () => {
    it('should create a project', async () => {
      render(<Landing />);

      const createButton = await screen.findByText('Create Project');
      await click(createButton);

      const modal = screen.getByRole('dialog');
      const nameInput = await within(modal).findByLabelText('Name', {
        exact: false,
      });
      const descriptionInput = await within(modal).findByLabelText(
        'Description (optional)',
        {
          exact: false,
        },
      );

      await type(nameInput, 'New-Project');
      await type(descriptionInput, 'description text');

      expect(mockServer.getState().projects).not.toHaveProperty('New-Project');

      await click(within(modal).getByText('Create'));

      await waitFor(() =>
        expect(mockServer.getState().projects).toHaveProperty('New-Project'),
      );
    });

    it('should error if the project already exists', async () => {
      render(<Landing />);

      const createButton = await screen.findByText('Create Project');
      await click(createButton);

      const modal = screen.getByRole('dialog');
      const nameInput = await within(modal).findByLabelText('Name', {
        exact: false,
      });

      await type(nameInput, 'Empty-Project');

      expect(
        await within(modal).findByText('Project name already in use'),
      ).toBeInTheDocument();
    });

    const validInputs = [
      ['goodproject'],
      ['good-project'],
      ['good_project'],
      ['goodproject1'],
      ['a'.repeat(51)],
    ];
    const invalidInputs = [
      [
        'bad project',
        'Name can only contain alphanumeric characters, underscores, and dashes',
      ],
      [
        'bad!',
        'Name can only contain alphanumeric characters, underscores, and dashes',
      ],
      [
        '_bad',
        'Name can only contain alphanumeric characters, underscores, and dashes',
      ],
      [
        'bad.',
        'Name can only contain alphanumeric characters, underscores, and dashes',
      ],
      ['a'.repeat(52), 'Project name exceeds maximum allowed length'],
    ];
    test.each(validInputs)(
      'should not error with a valid project name (%j)',
      async (input) => {
        render(<Landing />);

        const createButton = await screen.findByText('Create Project');
        await click(createButton);

        const modal = screen.getByRole('dialog');

        const nameInput = await within(modal).findByLabelText('Name', {
          exact: false,
        });

        await type(nameInput, input);

        expect(within(modal).queryByRole('alert')).not.toBeInTheDocument();
      },
    );
    test.each(invalidInputs)(
      'should error with an invalid project name (%j)',
      async (input, assertionText) => {
        render(<Landing />);

        const createButton = await screen.findByText('Create Project');
        await click(createButton);

        const modal = screen.getByRole('dialog');

        const nameInput = await within(modal).findByLabelText('Name', {
          exact: false,
        });

        await type(nameInput, input);

        expect(within(modal).getByText(assertionText)).toBeInTheDocument();
      },
    );
  });

  describe('Update Project Modal', () => {
    it('should update a project description with no text then with text', async () => {
      render(<Landing />);

      // wait for page to populate
      expect(await screen.findByText('Egress-Examples')).toBeInTheDocument();

      // open the modal on the correct row
      expect(screen.queryByRole('dialog')).toBeNull();
      expect(
        screen.queryByRole('menuitem', {
          name: /edit project info/i,
        }),
      ).toBeNull();
      await click(
        screen.getByRole('button', {
          name: /egress-examples overflow menu/i,
        }),
      );
      const menuItem = await screen.findByRole('menuitem', {
        name: /edit project info/i,
      });
      expect(menuItem).toBeVisible();
      await click(menuItem);

      const modal = await screen.findByRole('dialog');
      expect(modal).toBeInTheDocument();

      const descriptionInput = await within(modal).findByRole('textbox', {
        name: /description/i,
        exact: false,
      });
      expect(descriptionInput).toHaveValue(
        'Multiple pipelines outputting to different forms of egress',
      ); // Element should be prepopulated with the existing description.

      await clear(descriptionInput);

      await click(
        within(modal).getByRole('button', {
          name: /confirm changes/i,
        }),
      );

      await waitForElementToBeRemoved(() => screen.queryByRole('dialog'));

      const row = await screen.findByRole('row', {
        name: /egress-examples/i,
        exact: false,
      });

      expect(await within(row).findByText('N/A')).toBeInTheDocument();

      // reopen it to edit and save some data
      expect(
        screen.queryByRole('menuitem', {
          name: /edit project info/i,
        }),
      ).toBeNull();
      await click(
        screen.getByRole('button', {
          name: /egress-examples overflow menu/i,
        }),
      );
      const menuItem2 = await screen.findByRole('menuitem', {
        name: /edit project info/i,
      });
      expect(menuItem2).toBeVisible();
      await click(menuItem2);

      const modal2 = await screen.findByRole('dialog');
      expect(modal2).toBeInTheDocument();

      const descriptionInput2 = await within(modal2).findByRole('textbox', {
        name: /description/i,
        exact: false,
      });
      expect(descriptionInput2).toHaveValue('');

      await clear(descriptionInput2);
      await type(descriptionInput2, 'new desc');

      await click(
        within(modal2).getByRole('button', {
          name: /confirm changes/i,
        }),
      );

      await waitForElementToBeRemoved(() => screen.queryByRole('dialog'));

      const row2 = await screen.findByRole('row', {
        name: /egress-examples/i,
        exact: false,
      });

      expect(await within(row2).findByText('new desc')).toBeInTheDocument();
    }, 20_000);
  });

  describe('Delete Project Modal', () => {
    it('should delete a project', async () => {
      render(<Landing />);
      // wait for page to populate
      expect(await screen.findByText('Egress-Examples')).toBeInTheDocument();
      // open the modal on the correct row
      expect(screen.queryByRole('dialog')).toBeNull();
      expect(
        screen.queryByRole('menuitem', {
          name: /delete project/i,
        }),
      ).toBeNull();
      await click(
        screen.getByRole('button', {
          name: 'Egress-Examples overflow menu', // should be case-sensitive since projects are not case sensitive
        }),
      );
      const menuItem = await screen.findByRole('menuitem', {
        name: /delete project/i,
      });
      expect(menuItem).toBeVisible();
      await click(menuItem);

      const modal = await screen.findByRole('dialog');
      expect(modal).toBeInTheDocument();
      const projectNameInput = await within(modal).findByRole('textbox');
      expect(projectNameInput).toHaveValue('');

      await clear(projectNameInput);

      const confirmButton = within(modal).getByRole('button', {
        name: /delete project/i,
      });

      expect(confirmButton).toBeDisabled();
      await type(projectNameInput, 'egress');
      expect(confirmButton).toBeDisabled();
      await type(projectNameInput, '-example');
      expect(confirmButton).toBeDisabled();
      await type(projectNameInput, 's');
      expect(confirmButton).toBeEnabled();

      await click(confirmButton);
      await waitForElementToBeRemoved(() => screen.queryByRole('dialog'));
      await waitFor(() => {
        expect(
          screen.queryByRole('cell', {
            name: 'Egress-Examples', // should be case-sensitive since projects are not case sensitive
            exact: false,
          }),
        ).not.toBeInTheDocument();
      });
    });
  });
});
