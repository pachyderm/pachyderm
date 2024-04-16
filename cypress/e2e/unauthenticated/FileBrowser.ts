describe('FileBrowser', () => {
  beforeEach(() => {
    cy.deleteReposAndPipelines();
    cy.visit('/');
  });

  describe('File search', () => {
    beforeEach(() => {
      cy.multiLineExec(
        `pachctl create repo images
        pachctl put file images@master:one.png -f cypress/fixtures/liberty.png
        pachctl put file images@master:a/b/c/one/one.png -f cypress/fixtures/liberty.png
        pachctl put file images@master:two.png -f cypress/fixtures/liberty.png
        `,
      ).visit('/');
    });

    after(() => {
      cy.deleteReposAndPipelines();
    });

    it('should search for a file', () => {
      cy.visit('/lineage/default/repos/images/latest');

      cy.findByRole('textbox').type('one');
      cy.findByText('/one.png');
      cy.findByText('/a/b/c/one/');
      cy.findByText('/a/b/c/one/one.png');
    });
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
      cy.visit('/lineage/default/repos/images/latest');

      cy.findByRole('button', {
        name: /viewing all commits/i,
      }).click();
      cy.findByRole('menuitem', {
        name: /master/i,
      }).click();

      // look at master commits
      cy.findByTestId('CommitSelect__button').click();
      cy.findAllByRole('menuitem').should('have.length', 2);

      // change branch to test
      cy.findByRole('button', {
        name: /master/i,
      }).click();
      cy.findByRole('menuitem', {
        name: /test/i,
      }).click();

      cy.findByTestId('CommitSelect__button').click();
      cy.findAllByRole('menuitem').should('have.length', 6);
    });

    it('should display version history for selected file', () => {
      cy.visit('/lineage/default/repos/images/latest');

      cy.findByRole('button', {
        name: /viewing all commits/i,
      }).click();
      cy.findByRole('menuitem', {
        name: /test/i,
      }).click();

      cy.findAllByText('image1.png').last().click();
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
        .exec('echo "hi" | pachctl put file images@master:1.txt -f -')
        .exec('echo "hi" | pachctl put file images@master:2.txt -f -')
        .exec('echo "hi" | pachctl put file images@master:3.txt -f -')
        .exec('echo "hi" | pachctl put file images@master:4.txt -f -');
    });

    afterEach(() => {
      cy.deleteReposAndPipelines();
    });

    it('should download and delete multiple files at once', () => {
      cy.visit('/lineage/default/repos/images/latest');

      cy.findByRole('button', {
        name: /viewing all commits/i,
      }).click();
      cy.findByRole('menuitem', {
        name: /master/i,
      }).click();

      cy.findAllByTestId('FileTableRow__row', {timeout: 60000}).should(
        ($rows) => {
          expect($rows).to.have.length(4);
          expect($rows[0]).to.contain('1.txt');
          expect($rows[1]).to.contain('2.txt');
          expect($rows[2]).to.contain('3.txt');
          expect($rows[3]).to.contain('4.txt');
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
          expect($rows[0]).to.contain('3.txt');
        },
      );
    });

    it('should download multiple files', () => {
      // Use this to debug cypress events to figure out why this is failing in CI.
      // localStorage.debug = 'cypress:*';

      cy.on('window:before:load', (win) => {
        cy.stub(win, 'open').callsFake(cy.stub().as('open'));
      });

      cy.visit('/lineage/default/repos/images/latest');

      cy.findByRole('button', {
        name: /viewing all commits/i,
      }).click();
      cy.findByRole('menuitem', {
        name: /master/i,
      }).click();

      cy.findAllByTestId('FileTableRow__row', {timeout: 60000}).should(
        ($rows) => {
          expect($rows).to.have.length(4);
          expect($rows[0]).to.contain('1.txt');
          expect($rows[1]).to.contain('2.txt');
          expect($rows[2]).to.contain('3.txt');
          expect($rows[3]).to.contain('4.txt');
          $rows[0].click();
          $rows[1].click();
          $rows[3].click();
        },
      );

      cy.findByRole('button', {
        name: /download selected items/i,
      }).click();

      cy.get('@open').should('have.been.calledOnce');

      // Reference: FRON-1311
      // This code will assert that the zip has all of the expected files.
      // cy.waitUntil(() => cy.task('checkZipFiles'), {timeout: 60_000}).then(
      //   (files) => {
      //     expect(files).to.have.length(3);
      //     expect((files as unknown as string[])[0]).to.match(
      //       /default\/images\/[a-zA-Z0-9_]*\/1.txt/i,
      //     );
      //     expect((files as unknown as string[])[1]).to.match(
      //       /default\/images\/[a-zA-Z0-9_]*\/2.txt/i,
      //     );
      //     expect((files as unknown as string[])[2]).to.match(
      //       /default\/images\/[a-zA-Z0-9_]*\/4.txt/i,
      //     );
      //   },
      // );
    });
  });
});
