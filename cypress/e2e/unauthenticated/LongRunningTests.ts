describe('Long Running Tests', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.exec(`pachctl update defaults --project default <<EOF
    {
      "createPipelineRequest": {
        "datumTries": 1,
      }
    }
    EOF 
    `);
    cy.setupProject('error-opencv-with-one-success');
    cy.visit('/lineage/default');

    //images repo
    cy.findByRole('button', {
      name: 'GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65 repo',
      timeout: 30_000,
    }).should('exist');

    // open job
    cy.findByRole('button', {name: /previous jobs/i}).click();
    cy.findAllByTestId('GlobalFilter__status_passing', {
      timeout: 240_000,
    }).should('have.length', 1);
    cy.findAllByTestId('GlobalFilter__status_failing', {
      timeout: 240_000,
    }).should('have.length', 2);
  });

  after(() => {
    cy.deleteReposAndPipelines();
    cy.exec(`yes | pachctl delete project default`);
    cy.exec(`pachctl update defaults --project default <<EOF
    {
      "createPipelineRequest": {}
    }
    EOF 
    `);
    cy.exec(`pachctl create project default`);
  });

  describe('Download Canvas', () => {
    after(() => {
      cy.task('deleteDownloadedFile', 'default.svg');
    });

    beforeEach(() => {
      cy.visit('/');
      cy.findByRole('heading', {name: 'default'});
      cy.task('deleteDownloadedFile', 'default.svg');
    });

    it("should download a Project's canvas correctly", () => {
      /*
    The app seems to require a full loading of `/`. Otherwise it will redirect
    when visiting '/project/1/pipelines/montage'. This was happening when
    Cypress had to do a full page load to skip the tutorial.

    Since that is temporarily disabled, we can get the same behavior by just
    waiting for some page text to render. I think this is happening because the
    app needs to load an auth token and put it in local storage.
    */
      cy.findByRole('heading', {name: 'default'});
      cy.visit('/lineage/default');

      cy.findByRole('button', {
        name: 'Open DAG controls menu',
      }).click();
      cy.findByText('Download Canvas').click();

      cy.waitUntil(() =>
        cy.task('readDownloadedFileMaybe', `default.svg`),
      ).then((svgContent: unknown) => {
        const parser = new DOMParser();
        const svgDoc = parser.parseFromString(
          svgContent as string,
          'image/svg+xml',
        );
        const textElements = svgDoc.querySelectorAll('text');
        expect(textElements.length).to.be.greaterThan(0);

        const svgTextArray = Array.from(textElements).map(
          (el) => el.textContent,
        );
        svgTextArray.pop(); // Last line contains downloaded datetime that we don't need to assert on

        expect(svgTextArray).to.deep.equal([
          'images',
          'Upload Files',
          'montage',
          'Output',
          'Pipeline',
          'Running',
          'Subjob',
          'Unrunnable',
          '1',
          's3://test',
          'Egress',
          'edges',
          'Output',
          'Pipeline',
          'Running',
          'Subjob',
          'Failure',
          '1',
        ]);
      });
    });
  });

  describe('JobSets', () => {
    beforeEach(() => {
      cy.visit('/lineage/default');
    });

    it('should allow a user to select a subset of jobsets to inspect subjobs', () => {
      cy.findByText('Jobs').click();
      cy.findByLabelText('expand filters').click();
      cy.findAllByText('Failed').filter(':visible').click();
      cy.findAllByTestId('RunsList__row').should('have.length', 2);

      cy.findAllByTestId('RunsList__row').eq(0).click();
      cy.findAllByTestId('RunsList__row').eq(1).click();
      cy.findByText('Detailed info for 2 jobs');
      cy.findByText('Subjobs').click();

      cy.findAllByTestId('JobsList__row').should('have.length', 2);

      cy.findAllByTestId('JobsList__row')
        .eq(1)
        .within(() => cy.findByLabelText('Failed datums').click());

      cy.findByTestId('Filter__FAILEDChip').should('exist');
      cy.findByTestId('SidePanel__closeLeft').click();
      cy.findAllByTestId('LogRow__base')
        .first()
        .parent()
        .parent()
        .scrollTo('bottom');
      cy.findByText(
        /AttributeError: 'NoneType' object has no attribute 'shape'/,
      ).should('be.visible');
    });
  });

  describe('GlobalFilters', () => {
    beforeEach(() => {
      cy.visit('/lineage/default');
    });

    it('simple date filter and jq filters work correctly', () => {
      // open job
      cy.findByRole('button', {name: /previous jobs/i}).click();
      cy.findAllByTestId('GlobalFilter__status_passing').should(
        'have.length',
        1,
      );
      cy.findAllByTestId('GlobalFilter__status_failing').should(
        'have.length',
        2,
      );

      // the simple date filter works
      cy.findByText(/\(2\) last hour/i).click();
      cy.findAllByTestId('SearchResultItem__container').should(
        'have.length',
        2,
      );
      cy.findByRole('button', {name: 'Clear'}).click();
      cy.findAllByTestId('SearchResultItem__container').should(
        'have.length',
        3,
      );

      // ensure the right items are shown in the global id filter on the dag
      cy.findAllByRole('listitem').eq(0).click();

      //edges pipeline
      cy.findByRole('button', {
        name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 pipeline',
      }).should('not.exist');
      //images repo
      cy.findByRole('button', {
        name: 'GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65 repo',
      }).should('exist');
      //edges repo
      cy.findByRole('button', {
        name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 repo',
      }).should('exist');
      //montage pipeline
      cy.findByRole('button', {
        name: 'GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6 pipeline',
      }).should('exist');

      // ensure the right items are shown when you click a global ID
      cy.findAllByRole('listitem').eq(1).click();

      //montage pipeline
      cy.findByRole('button', {
        name: 'GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6 pipeline',
      }).should('not.exist');
      //images repo
      cy.findByRole('button', {
        name: 'GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65 repo',
      }).should('exist');
      //montage repo
      cy.findByRole('button', {
        name: 'GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6 repo',
      }).should('not.exist');
      //edges pipeline
      cy.findByRole('button', {
        name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 pipeline',
      }).should('exist');
    });

    it('should render a sub-dag with a globalId filter', () => {
      cy.findByText('Jobs').click();
      cy.findByLabelText('expand filters').click();
      cy.findAllByText('Failed').filter(':visible').click();

      cy.findByRole('rowgroup', {
        name: 'runs list',
      }).within(() => {
        cy.findAllByRole('row').should('have.length', 2);
      });

      cy.findAllByTestId('RunsList__row')
        .eq(1)
        .findByTestId('DropdownButton__button')
        .click();
      cy.findByRole('menuitem', {
        name: /apply global id and view in dag/i,
      }).click();

      // images repo
      cy.findByRole('button', {
        name: 'GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65 repo',
      }).should('exist');
      // images -> edges
      cy.get(
        'path[id="8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65_d0e1e9a51269508c3f11c0e64c721c3ea6204838"]',
      ).should('exist');
      // images -> montage
      cy.get(
        'path[id="8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6"]',
      ).should('not.exist');
      //edges node
      cy.findByRole('button', {
        name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 pipeline',
      }).should('exist');
      // edges -> montage
      cy.get(
        'path[id="d0e1e9a51269508c3f11c0e64c721c3ea6204838_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6"]',
      ).should('not.exist');
      //montage node
      cy.findByRole('button', {
        name: 'GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6 pipeline',
      }).should('not.exist');

      cy.findByTestId('Node__state-ERROR').should('exist');
    });
  });
});
