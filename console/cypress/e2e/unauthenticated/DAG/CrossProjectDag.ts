describe('Cross Project Dag', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.multiLineExec(
      `pachctl create project p1
      pachctl create repo disconnected
      pachctl create repo repo-1 --project p1
      echo '{"pipeline": { "name": "pipeline-1", "project": { "name": "default" } },"input": {"pfs": { "glob": "/", "repo": "repo-1", "project": "p1" }},"transform":{"cmd":["sh"],"stdin": ["sleep 0"]}}' | pachctl update pipeline
      `,
    );
  });
  beforeEach(() => {
    cy.visit('/lineage/default');
  });
  after(() => {
    cy.deleteReposAndPipelines();
  });

  it('should render cross project dag node', () => {
    //project: p1 name: repo-1
    cy.findByRole('button', {
      name: 'GROUP_32e1641f4678e352d24f83e58629272fc925d3bd connected project',
      timeout: 10000,
    }).should('exist');

    //project: p1 name: repo-1
    cy.findByRole('button', {
      name: 'GROUP_32e1641f4678e352d24f83e58629272fc925d3bd connected repo',
      timeout: 10000,
    }).should('exist');

    //project: default name: pipeline-1
    cy.findByRole('button', {
      name: 'GROUP_ee46d663bb60604343cc1e26f01b4b1198820079 pipeline',
      timeout: 10000,
    }).should('exist');

    //project: default name: disconnected
    cy.findByRole('button', {
      name: 'GROUP_2e848120b1d55fa48d8e7d788af68e8cb0c80eae repo',
      timeout: 10000,
    }).should('exist');
  });

  it('should render a sub-dag with a globalId filter', () => {
    cy.findByText('Jobs', {timeout: 10000}).click();

    cy.findByRole('rowgroup', {
      name: 'runs list',
      timeout: 60000,
    }).within(() => {
      cy.findAllByRole('row').should('have.length', 1);
    });

    cy.findByTestId('RunsList__row')
      .findByTestId('DropdownButton__button')
      .click();
    cy.findByRole('menuitem', {
      name: /apply global id and view in dag/i,
    }).click();

    //project: p1 name: repo-1
    cy.findByRole('button', {
      name: 'GROUP_32e1641f4678e352d24f83e58629272fc925d3bd connected project',
      timeout: 10000,
    }).should('exist');

    //project: p1 name: repo-1
    cy.findByRole('button', {
      name: 'GROUP_32e1641f4678e352d24f83e58629272fc925d3bd connected repo',
      timeout: 10000,
    }).should('exist');

    //project: default name: pipeline-1
    cy.findByRole('button', {
      name: 'GROUP_ee46d663bb60604343cc1e26f01b4b1198820079 pipeline',
      timeout: 10000,
    }).should('exist');

    //project: default name: disconnected
    cy.findByRole('button', {
      name: 'GROUP_2e848120b1d55fa48d8e7d788af68e8cb0c80eae repo',
      timeout: 10000,
    }).should('not.exist');
  });

  it('should update the url correctly when selecting the connected project', () => {
    //project: p1 name: repo-1
    cy.findByRole('button', {
      name: 'GROUP_32e1641f4678e352d24f83e58629272fc925d3bd connected project',
      timeout: 10000,
    }).click();

    cy.url().should('contain', '/lineage/p1');
  });

  it('should update the url correctly when selecting the connected repo', () => {
    //project: p1 name: repo-1
    cy.findByRole('button', {
      name: 'GROUP_32e1641f4678e352d24f83e58629272fc925d3bd connected repo',
      timeout: 10000,
    }).click();

    cy.url().should('contain', '/lineage/p1/repos/repo-1');
  });
});
