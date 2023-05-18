describe('FileBrowser', () => {
  before(() => {
    cy.exec('jq -r .pachReleaseCommit version.json')
      .then(() => {
        cy.exec('pachctl create repo images')
          .exec(
            'pachctl put file images@master:image1.png -f http://imgur.com/46Q8nDz.png',
          )
          .exec(
            'pachctl put file images@master:image1.png -f http://imgur.com/8MN9Kg0.png',
          )
          .exec(
            'pachctl put file images@test:image1.png -f http://imgur.com/46Q8nDz.png',
          )
          .exec('pachctl delete file images@test:image1.png')
          .exec(
            'pachctl put file images@test:image1.png -f http://imgur.com/46Q8nDz.png',
          )
          .exec(
            'pachctl put file images@test:image1.png -f http://imgur.com/8MN9Kg0.png',
          );
      })
      .visit('/');
  });
  beforeEach(() => {
    cy.visit('/');
  });
  after(() => {
    cy.deleteReposAndPipelines();
  });

  it('should display commits for selected branch', () => {
    cy.visit('/lineage/default/repos/images/branch/master/latest');
    cy.findAllByTestId('CommitList__listItem').should('have.length', 2);
    cy.findAllByTestId('DropdownButton__button').eq(1).click();
    cy.findByText('test').click();
    cy.findAllByTestId('CommitList__listItem').should('have.length', 4);
  });

  it('should display version history for selected file', () => {
    cy.visit('/lineage/default/repos/images/branch/test/latest');
    cy.findByText('image1.png').click();
    cy.findByRole('button', {
      name: 'Load older file versions',
    }).click();
    cy.findByRole('button', {
      name: 'Load older file versions',
    }).should('be.disabled');

    cy.get(`[aria-label="file metadata"]`).findByText('80.59 kB');
    cy.findByTestId('FileHistory__commitList')
      .children()
      .should(($links) => {
        expect($links).to.have.length(4);
        expect($links[0]).to.contain('Updated');
        expect($links[0]).to.have.property('href');
        expect($links[1]).to.contain('Added');
        expect($links[1]).to.have.property('href');
        expect($links[2]).to.contain('Deleted');
        expect($links[2]).to.have.property('href', '');
        expect($links[3]).to.contain('Added');
        expect($links[3]).to.have.property('href');
        $links[3].click();
      });

    cy.findByText('58.65 kB');
  });
});
