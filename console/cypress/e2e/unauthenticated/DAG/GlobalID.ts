describe('GlobalID', () => {
  before(() => {
    cy.deleteProjectReposAndPipelines('GlobalIdTest');
    cy.multiLineExec(`
      pachctl create project GlobalIdTest
      pachctl config update context --project GlobalIdTest
      `);
    cy.setupProject();
  });
  beforeEach(() => {
    cy.visit('/lineage/GlobalIdTest');
  });

  after(() => {
    cy.deleteProjectReposAndPipelines('GlobalIdTest');
    cy.exec(`pachctl config update context --project default`);
  });

  describe('Advanced Date/Time Filter Using Dates', () => {
    it('should show items we expect in date range', () => {
      //grabbing current date/time values
      const time = new Date();
      //navigating to global ID date/time filter
      cy.findByRole('button', {
        name: /previous jobs/i,
      }).click();
      cy.findByRole('button', {
        name: /filter/i,
      }).click();
      cy.findByTestId('AdvancedDateTimeFilter__startDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByTestId('AdvancedDateTimeFilter__endDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByRole('banner').within(() => {
        cy.findByRole('listitem');
      });
    });

    it('should show no items in past date range', () => {
      //grabbing current date/time values
      const time = new Date();
      //setting date to prior day so no items show
      time.setDate(time.getDate() - 1);
      //navigating to global ID date/time filter
      cy.findByRole('button', {
        name: /previous jobs/i,
      }).click();
      cy.findByRole('button', {
        name: /filter/i,
      }).click();
      cy.findByTestId('AdvancedDateTimeFilter__startDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByTestId('AdvancedDateTimeFilter__endDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByText(/no job sets found in this project\./i);
    });

    it('should show no items in future date range', () => {
      //grabbing current date/time values
      const time = new Date();
      //setting date to prior day so no items show
      time.setDate(time.getDate() + 1);
      //navigating to global ID date/time filter
      cy.findByRole('button', {
        name: /previous jobs/i,
      }).click();
      cy.findByRole('button', {
        name: /filter/i,
      }).click();
      cy.findByTestId('AdvancedDateTimeFilter__startDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByTestId('AdvancedDateTimeFilter__endDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByText(/no job sets found in this project\./i);
    });
  });

  describe('Advanced Date/Time Filter Using Times', () => {
    it('should show items we expect in time range', () => {
      //grabbing current date/time values
      const time = new Date();
      //navigating to global ID date/time filter
      cy.findByRole('button', {
        name: /previous jobs/i,
      }).click();
      cy.findByRole('button', {
        name: /filter/i,
      }).click();
      //set start time to one hour before
      time.setHours(time.getHours() - 1);
      //set date&time to date&time of one hour before
      cy.findByTestId('AdvancedDateTimeFilter__startDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByTestId('AdvancedDateTimeFilter__startTime').type(
        time.toTimeString().split(' ')[0].substring(0, 5),
      );
      //set end time to one hour after
      time.setHours(time.getHours() + 2);
      //set date&time to date&time of one hour after
      cy.findByTestId('AdvancedDateTimeFilter__endDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByTestId('AdvancedDateTimeFilter__endTime').type(
        time.toTimeString().split(' ')[0].substring(0, 5),
      );

      cy.findByRole('banner').within(() => {
        cy.findByRole('listitem');
      });
    });

    it('should show no items in past time range', () => {
      //grabbing current date/time values
      const time = new Date();
      //navigating to global ID date/time filter
      cy.findByRole('button', {
        name: /previous jobs/i,
      }).click();
      cy.findByRole('button', {
        name: /filter/i,
      }).click();
      //set start time to 4 hours before
      time.setHours(time.getHours() - 4);
      //set date&time to date&time of 4 hours before
      cy.findByTestId('AdvancedDateTimeFilter__startDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByTestId('AdvancedDateTimeFilter__startTime').type(
        time.toTimeString().split(' ')[0].substring(0, 5),
      );
      //set end time to 2 hours before
      time.setHours(time.getHours() + 2);
      //set date&time to date&time of 2 hours before
      cy.findByTestId('AdvancedDateTimeFilter__endDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByTestId('AdvancedDateTimeFilter__endTime').type(
        time.toTimeString().split(' ')[0].substring(0, 5),
      );

      cy.findByText(/no job sets found in this project\./i);
    });

    it('should show no items in future time range', () => {
      //grabbing current date/time values
      const time = new Date();
      //navigating to global ID date/time filter
      cy.findByRole('button', {
        name: /previous jobs/i,
      }).click();
      cy.findByRole('button', {
        name: /filter/i,
      }).click();

      //set start time to 2 hours after
      time.setHours(time.getHours() + 2);
      //set date&time to date&time of 2 hours after
      cy.findByTestId('AdvancedDateTimeFilter__startDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByTestId('AdvancedDateTimeFilter__startTime').type(
        time.toTimeString().split(' ')[0].substring(0, 5),
      );
      //set end date&time to date&time of 4 hours after
      time.setHours(time.getHours() + 2);
      cy.findByTestId('AdvancedDateTimeFilter__endDate').type(
        time.toISOString().split('T')[0],
      );
      cy.findByTestId('AdvancedDateTimeFilter__endTime').type(
        time.toTimeString().split(' ')[0].substring(0, 5),
      );

      cy.findByText(/no job sets found in this project\./i);
    });
  });
});
