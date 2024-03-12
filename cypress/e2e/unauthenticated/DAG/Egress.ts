describe('Egress', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.multiLineExec(`
    pachctl create repo data
    echo '{"pipeline": {"name": "db"},  "egress": {"sqlDatabase": {"url": "db://url","secret": {"name": "name", "key": "key"}}} ,"transform": {"cmd": ["sh"],"stdin": [""]},"input": {"pfs": {"glob": "/*","repo": "data"}}}'  | pachctl create pipeline
    echo '{"pipeline": {"name": "storage"},  "egress": {"objectStorage": {"url": "os://url"}} ,"transform": {"cmd": ["sh"],"stdin": [""]},"input": {"pfs": {"glob": "/*","repo": "data"}}}'  | pachctl create pipeline
    echo '{"pipeline": {"name": "s3"},"s3Out": true,  "egress": {"objectStorage": {"url": "os://url"}} ,"transform": {"cmd": ["sh"],"stdin": [""]},"input": {"pfs": {"glob": "/*","repo": "data"}}}'  | pachctl create pipeline
  `);
  });
  beforeEach(() => {
    cy.visit('/lineage/default');
  });
  after(() => {
    cy.deleteReposAndPipelines();
  });

  it('should show egress node to DB', () => {
    cy.get('g[id="GROUP_9df9d3dd4577573624e30e51042165fe56af7954"]')
      .should('exist')
      .findByText('Egress');

    cy.findByRole('button', {
      name: 'GROUP_9df9d3dd4577573624e30e51042165fe56af7954 egress url',
      timeout: 10000,
    })
      .should('exist')
      .findByText('db://url');

    cy.get('g[id="GROUP_9df9d3dd4577573624e30e51042165fe56af7954_to"]')
      .should('exist')
      .findByText('Egresses to DB');
  });

  it('should show egress node to storage', () => {
    cy.get('g[id="GROUP_fb938dcb6e541dc9b68c5dd82b180d6cd6f5c0fe"]')
      .should('exist')
      .findByText('Egress');

    cy.findByRole('button', {
      name: 'GROUP_fb938dcb6e541dc9b68c5dd82b180d6cd6f5c0fe egress url',
      timeout: 10000,
    })
      .should('exist')
      .findByText('os://url');

    cy.get('g[id="GROUP_fb938dcb6e541dc9b68c5dd82b180d6cd6f5c0fe_to"]')
      .should('exist')
      .findByText('Egresses to storage');
  });

  it('should show egress node to S3', () => {
    cy.get('g[id="GROUP_a32a4c56f315caa78900fffc273b4709df5d815b"]')
      .should('exist')
      .findByText('Egress');

    cy.findByRole('button', {
      name: 'GROUP_a32a4c56f315caa78900fffc273b4709df5d815b egress url',
      timeout: 10000,
    })
      .should('exist')
      .findByText('s3://s3');

    cy.get('g[id="GROUP_a32a4c56f315caa78900fffc273b4709df5d815b_to"]')
      .should('exist')
      .findByText('Egresses to S3');
  });
});
