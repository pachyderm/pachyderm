describe('FileBrowser', () => {
  beforeEach(() => {
    cy.deleteReposAndPipelines();
    cy.visit('/');
  });

  describe('Display and version history', () => {
    beforeEach(() => {
      cy.multiLineExec(
        `pachctl create repo images
        pachctl put file images@master:image1.png -f cypress/fixtures/liberty.png
        pachctl put file images@master:image1.png -f cypress/fixtures/AT-AT.png
        pachctl put file images@test:image1.png -f cypress/fixtures/liberty.png
        pachctl delete file images@test:image1.png        
        pachctl put file images@test:image1.png -f cypress/fixtures/liberty.png
        pachctl put file images@test:image1.png -f cypress/fixtures/AT-AT.png
        pachctl put file images@test:image1.png -f cypress/fixtures/liberty.png
        pachctl put file images@test:image1.png -f cypress/fixtures/AT-AT.png
        `,
      ).visit('/');
    });

    after(() => {
      cy.deleteReposAndPipelines();
    });

    it('should display commits for selected branch', () => {
      cy.visit('/lineage/default/repos/images/branch/master/latest');
      // look at master commits
      cy.findAllByRole('listitem').should('have.length', 2);

      // change branch to test
      cy.findByRole('button', {
        name: /master/i,
      }).click();
      cy.findByRole('menuitem', {
        name: /test/i,
      }).click();

      cy.findAllByTestId('CommitList__listItem').should('have.length', 6);
    });

    it('should display version history for selected file', () => {
      cy.visit('/lineage/default/repos/images/branch/test/latest');
      cy.findByText('image1.png').click();
      cy.findByRole('button', {
        name: 'Load older file versions',
      }).click();

      cy.get(`[aria-label="file metadata"]`).findByText('80.59 kB');
      cy.findByTestId('FileHistory__commitList')
        .children()
        .should(($links) => {
          expect($links).to.have.length(5);
          expect($links[0]).to.contain('Updated');
          expect($links[0]).to.have.property('href');
          expect($links[1]).to.contain('Updated');
          expect($links[1]).to.have.property('href');
          expect($links[2]).to.contain('Updated');
          expect($links[2]).to.have.property('href');
          expect($links[3]).to.contain('Added');
          expect($links[3]).to.have.property('href');
          expect($links[4]).to.contain('Deleted');
          expect($links[4]).to.have.property('href', '');
        });

      cy.findByRole('button', {
        name: 'Load older file versions',
      }).click();
      cy.findByRole('button', {
        name: 'Load older file versions',
      }).should('be.disabled');

      cy.findByTestId('FileHistory__commitList')
        .children()
        .should(($links) => {
          expect($links).to.have.length(6);
          $links[3].click();
        });

      cy.findByRole('dialog').within(() => {
        cy.findByText('58.65 kB');
      });
    });
  });

  describe('Delete', () => {
    beforeEach(() => {
      cy.exec('pachctl create repo images')
        .exec(
          'pachctl put file images@master:image1.png -f cypress/fixtures/liberty.png',
        )
        .exec(
          'pachctl put file images@master:image2.png -f cypress/fixtures/liberty.png',
        )
        .exec(
          'pachctl put file images@master:image3.png -f cypress/fixtures/liberty.png',
        )
        .exec(
          'pachctl put file images@master:image4.png -f cypress/fixtures/liberty.png',
        );
    });

    afterEach(() => {
      cy.deleteReposAndPipelines();
    });

    it('should download and delete multiple files at once', () => {
      cy.visit('/lineage/default/repos/images/branch/master/latest');

      cy.findAllByTestId('FileTableRow__row', {timeout: 60000}).should(
        ($rows) => {
          expect($rows).to.have.length(4);
          expect($rows[0]).to.contain('image1.png');
          expect($rows[1]).to.contain('image2.png');
          expect($rows[2]).to.contain('image3.png');
          expect($rows[3]).to.contain('image4.png');
          $rows[0].click();
          $rows[1].click();
          $rows[3].click();
        },
      );

      cy.findByRole('button', {
        name: /delete selected items/i,
      }).click();

      cy.findByRole('button', {
        name: 'Delete',
      }).click();

      cy.findAllByTestId('FileTableRow__row', {timeout: 60000}).should(
        ($rows) => {
          expect($rows[0]).to.contain('image3.png');
        },
      );
    });

    it('should download multiple files', () => {
      cy.on('window:before:load', (win) => {
        cy.stub(win, 'open').callsFake(cy.stub().as('open'));
      });

      cy.visit('/lineage/default/repos/images/branch/master/latest');

      cy.findAllByTestId('FileTableRow__row', {timeout: 60000}).should(
        ($rows) => {
          expect($rows).to.have.length(4);
          expect($rows[0]).to.contain('image1.png');
          expect($rows[1]).to.contain('image2.png');
          expect($rows[2]).to.contain('image3.png');
          expect($rows[3]).to.contain('image4.png');
          $rows[0].click();
          $rows[1].click();
          $rows[3].click();
        },
      );

      cy.findByRole('button', {
        name: /download selected items/i,
      }).click();

      cy.get('@open').should(
        'have.been.calledOnceWithExactly',
        '/proxyForward/archive/ASi1L_0gZXUBACQCZGVmYXVsdC9pbWFnZXNAbWFzdGVyOjEucG5nADI0LnBuZwMAYBRFsUKG2QQ.zip',
      );
    });
  });
});
