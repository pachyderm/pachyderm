describe('Repos', () => {
  before(() => {
    cy.deleteReposAndPipelines();
  });

  beforeEach(() => {
    cy.visit('/');
    cy.findAllByText(/^View(\sProject)*$/)
      .eq(0)
      .click();
    cy.findByRole('button', {name: 'Create', timeout: 12000}).click();
    cy.findByRole('menuitem', {
      name: 'Input Repository',
      timeout: 12000,
    }).click();
    cy.findByLabelText('Name', {exact: false, timeout: 12000}).type('TestRepo');
    cy.findByRole('dialog').within(() => cy.findByText('Create').click());
    cy.exec('pachctl create branch TestRepo@test');
  });

  afterEach(() => {
    cy.deleteReposAndPipelines();
  });

  it('should allow a user to create a repo', () => {
    cy.findByRole('button', {name: 'Create', timeout: 12000}).click();
    cy.findByRole('menuitem', {
      name: 'Input Repository',
      timeout: 12000,
    }).click();

    cy.findByLabelText('Name', {exact: false, timeout: 12000}).clear();
    cy.findByLabelText('Name', {exact: false, timeout: 12000}).type('NewRepo');
    cy.findByLabelText('Description', {exact: false}).type(
      'New repo description',
    );
    cy.findByRole('dialog').within(() => cy.findByText('Create').click());

    cy.findByText('NewRepo').click();
    cy.findByText('New repo description', {timeout: 15000});
  });

  it('should allow a user to cancel an upload', () => {
    cy.findByText('TestRepo', {timeout: 12000}).click();

    cy.findByRole('button', {name: 'Repo Actions'}).click();
    cy.findByRole('menuitem', {name: 'Upload Files'}).click();
    cy.findByText('For large file uploads via CTL');

    cy.fixture('fruit.png', null).as('file');
    cy.waitUntil(() =>
      cy.findByLabelText('Attach Files').should('not.be.disabled'),
    );
    cy.findByLabelText('Attach Files').selectFile(
      [
        {
          contents: '@file',
          fileName: 'fruit.png',
        },
      ],
      {force: true},
    );

    cy.findByText('fruit.png').should('exist');
    cy.findByLabelText('Upload Selected Files').click();
    cy.findByTestId('FileCard__cancel').click({force: true}); // file card can sometimes be above the fold
    cy.findByText('fruit.png').should('not.exist');
  });

  it('should upload a file for the specified path, branch, message, and files', () => {
    cy.findByText('TestRepo', {timeout: 12000}).click();
    cy.findByRole('button', {name: 'Repo Actions'}).click();
    cy.findByRole('menuitem', {name: 'Upload Files'}).click();

    cy.fixture('cat.png', null).as('file1');

    cy.waitUntil(() =>
      cy.findByLabelText('Attach Files').should('not.be.disabled'),
    );
    cy.findByLabelText('Attach Files').selectFile(
      [
        {
          contents: '@file1',
          fileName: 'cat.png',
        },
      ],
      {force: true},
    );

    cy.waitUntil(() =>
      cy.findByLabelText('Upload Selected Files').should('not.be.disabled'),
    );

    cy.findByRole('textbox', {name: 'File Path'}).type('/image_store');
    cy.findByRole('textbox', {name: 'Commit Message'}).type('initial images');
    cy.findByRole('combobox').click();
    cy.findByRole('option', {name: 'test'}).click();

    cy.findByLabelText('Upload Selected Files').click();
    cy.findByLabelText('Commit Selected Files').click();

    // Needs to wait for commit polling to update
    cy.visit('/lineage/default/repos/TestRepo');
    cy.findByText('263.07 kB', {timeout: 30000}).should('exist');
    cy.findByText('New').should('exist');
    cy.findByText('Parent Commit').should('exist');

    cy.findByRole('link', {name: 'Previous Commits'}).click();

    cy.findAllByText('initial images').should('exist');
    cy.findAllByText('image_store').should('have.length', 2).should('exist');
  });

  it('should allow a user to upload images and not append files, and view differences between commits', () => {
    cy.findByText('TestRepo', {timeout: 12000}).click();
    cy.findByRole('button', {name: 'Repo Actions'}).click();
    cy.findByRole('menuitem', {name: 'Upload Files'}).click();
    cy.findByText('For large file uploads via CTL');

    cy.fixture('fruit.png', null).as('file1');
    cy.fixture('cat.png', null).as('file2');

    cy.waitUntil(() =>
      cy.findByLabelText('Attach Files').should('not.be.disabled'),
    );
    cy.findByLabelText('Attach Files').selectFile(
      [
        {
          contents: '@file1',
          fileName: 'fruit.png',
        },
        {
          contents: '@file2',
          fileName: 'cat.png',
        },
      ],
      {force: true},
    );

    cy.waitUntil(() =>
      cy.findByLabelText('Upload Selected Files').should('not.be.disabled'),
    );
    cy.findByLabelText('Upload Selected Files').click();
    cy.findByLabelText('Commit Selected Files').click();

    // Needs to wait for commit polling to update
    cy.visit('/lineage/default/repos/TestRepo');
    cy.findByText('667.61 kB', {timeout: 30000});
    cy.findByText('2');
    cy.findByText('New');

    cy.findByRole('button', {name: 'Repo Actions'}).click();
    cy.findByRole('menuitem', {name: 'Upload Files'}).click();
    cy.findByText('For large file uploads via CTL');

    cy.waitUntil(() =>
      cy.findByLabelText('Attach Files').should('not.be.disabled'),
    );
    cy.findByLabelText('Attach Files').selectFile(
      [
        {
          contents: '@file1',
          fileName: 'car.png',
        },
      ],
      {force: true},
    );

    cy.waitUntil(() =>
      cy.findByLabelText('Upload Selected Files').should('be.visible').click(),
    );
    cy.findByLabelText('Commit Selected Files').click();

    cy.visit('/lineage/default/repos/TestRepo');
    cy.findByText('1.08 MB', {timeout: 30000});
  });

  it('should allow a user to delete a repo', () => {
    cy.findByText('TestRepo', {timeout: 12000}).click();

    cy.findByRole('button', {name: 'Repo Actions'}).click();
    cy.findByRole('menuitem', {name: 'Delete Repo'}).click();

    cy.findByText('Delete').click();
    cy.findByText('TestRepo').should('not.exist');
  });

  it('should allow a user to update repo metadata', () => {
    cy.findByText('TestRepo', {timeout: 12000}).click();
  
    cy.findByRole('tab', {name: /user metadata/i}).click();
    cy.findAllByRole('button', {name: /edit/i}).eq(0).click();
    cy.findByRole('heading', {name: /edit repo metadata/i});

    cy.findByRole('button', {name: /add new/i}).click();

    cy.findAllByPlaceholderText('key').eq(0).type('newKey');
    cy.findAllByPlaceholderText('value').eq(0).type('newValue');
    cy.findAllByPlaceholderText('key').eq(1).type('deleteKey');
    cy.findAllByPlaceholderText('value').eq(1).type('deleteValue');

    cy.findByRole('button', {name: /apply metadata/i}).click();

    cy.findByRole('heading', {name: /edit repo metadata/i}).should(
      'not.exist',
    );
    cy.findByText('newKey').should('exist');
    cy.findByText('newValue').should('exist');
    cy.findByText('deleteKey').should('exist');
    cy.findByText('deleteValue').should('exist');

    cy.findAllByRole('button', {name: /edit/i}).eq(0).click();
    cy.findByRole('heading', {name: /edit repo metadata/i});
    cy.findByRole('button', {name: /delete metadata row 0/i}).click();
    cy.findByRole('button', {name: /apply metadata/i}).click();

    cy.findByText('deleteKey').should('not.exist');
    cy.findByText('deleteValue').should('not.exist');
  });

  it('should allow a user to update commit metadata', () => {
    cy.findByText('TestRepo', {timeout: 12000}).click();
  
    cy.findByRole('tab', {name: /user metadata/i}).click();
    cy.findAllByRole('button', {name: /edit/i}).eq(1).click();
    cy.findByRole('heading', {name: /edit commit metadata/i});

    cy.findAllByPlaceholderText('key').eq(0).type('commitKey');
    cy.findAllByPlaceholderText('value').eq(0).type('commitValue');

    cy.findByRole('button', {name: /apply metadata/i}).click();
    
    cy.findByRole('heading', {name: /edit commit metadata/i}).should('not.exist');

    cy.findByText('commitKey').should('exist');
    cy.findByText('commitValue').should('exist');
  });

  it('should allow a user to select a repo from the list view to inspect commits', () => {
    cy.setupProject().visit('/');
    cy.findAllByText(/^View(\sProject)*$/)
      .eq(0)
      .click();
    cy.findByText('Repositories').click();
    cy.findAllByTestId('RepoListRow__row', {timeout: 30000}).should(
      'have.length',
      3,
    );

    cy.findByText('images').click();
    cy.findByText('Detailed info for images');
    cy.findByText('Commits').click();

    cy.findAllByTestId('CommitsList__row').should('have.length', 1);

    cy.findAllByTestId('CommitsList__row')
      .first()
      .within(() => cy.findByTestId('DropdownButton__button').click());
    cy.findByText('Inspect commit').click();

    cy.findByText('Commit files for');
    cy.findAllByText('cat.png').last();
    cy.findByText('Added');

    // table filters are kept on close
    cy.findByRole('button', {name: 'Close'}).click();
    cy.findAllByTestId('CommitsList__row').should('have.length', 1);
  });
});
