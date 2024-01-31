import classnames from 'classnames';
import React from 'react';

import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import BrandedDocLink from '@dash-frontend/components/BrandedDocLink';
import EmptyState from '@dash-frontend/components/EmptyState';
import {UUID_WITHOUT_DASHES_REGEX} from '@dash-frontend/constants/pachCore';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {NodeState} from '@dash-frontend/lib/types';
import {
  Icon,
  Input,
  Form,
  Button,
  ChevronDownSVG,
  StatusCheckmarkSVG,
  StatusWarningSVG,
  StatusDotsSVG,
  NavigationHistorySVG,
  CloseSVG,
} from '@pachyderm/components';

import {SearchResultItem} from '../../../ProjectHeader/components/GlobalSearch/components/SearchResultsDropdown/components';

import {AdvancedDateTimeFilter} from './components/AdvancedDateTimeFilter';
import {SimpleDateFilters} from './components/SimpleDateFilters';
import styles from './GlobalFilter.module.css';
import useGlobalFilter from './hooks/useGlobalFilter';

const GlobalFilter: React.FC = () => {
  const {
    containerRef,
    formCtx,
    dropdownOpen,
    setDropdownOpen,
    loading,
    globalIdFilter,
    handleClick,
    jobsCount,
    filteredJobs,
    chips,
    clearFilters,
    showClearButton,
    toggleDateTimePicker,
    dateTimePickerOpen,
    dateTimeFilterValue,
    buttonText,
    removeFilter,
  } = useGlobalFilter();

  // TODO: does jobsets poll? It clearly triggers when the DAG renders. That means it can be stale data. Refetch when you open the dropdown? Disable when dropdown is not open?
  return (
    <div className={styles.base} ref={containerRef}>
      <Button
        buttonType="tertiary"
        onClick={() => setDropdownOpen(!dropdownOpen)}
        IconSVG={!globalIdFilter ? NavigationHistorySVG : undefined}
      >
        {globalIdFilter && <div className={styles.dot} />}
        {!globalIdFilter ? (
          'Previous Jobs'
        ) : (
          <div className={styles.buttonContent}>
            <Icon small color="white">
              <NavigationHistorySVG />
            </Icon>
            {buttonText}
            <Icon small color="white">
              <ChevronDownSVG />
            </Icon>
          </div>
        )}
      </Button>
      {dropdownOpen && (
        <Form formContext={formCtx} className={classnames(styles.dropdownBase)}>
          <div className={styles.inputWrapper}>
            <div className={styles.input}>
              <Input
                className={styles.input}
                placeholder="Copy and paste your Global ID"
                type="text"
                id="globalId"
                name="globalId"
                spellCheck={false}
                validationOptions={{
                  pattern: {
                    value: UUID_WITHOUT_DASHES_REGEX,
                    message: 'Not a valid Global ID',
                  },
                }}
                disabled={loading}
                autoFocus={true}
                clearable
              />
            </div>
            {globalIdFilter && (
              <Button
                className={styles.clearFilterButton}
                onClick={removeFilter}
                IconSVG={CloseSVG}
                buttonType="ghost"
                iconPosition="end"
                color="black"
              >
                Clear Filter
              </Button>
            )}
          </div>
          {!dateTimePickerOpen && (
            <SimpleDateFilters
              chips={chips}
              clearFilters={clearFilters}
              showClearButton={showClearButton}
              toggleDateTimePicker={toggleDateTimePicker}
            />
          )}
          {dateTimePickerOpen && (
            <AdvancedDateTimeFilter
              toggleDateTimePicker={toggleDateTimePicker}
              dateTimeFilterValue={dateTimeFilterValue}
            />
          )}
          <div className={styles.infoText}>
            Global IDs provide a historical snapshot of the DAG when a job had
            run. You will see all the pipelines and subjobs that were processed
            for a specific job.{' '}
            <BrandedDocLink
              pathWithoutDomain="concepts/advanced-concepts/globalid/"
              className={styles.link}
            >
              Learn more about Global IDs.
            </BrandedDocLink>
          </div>
          {jobsCount === 0 ? (
            <EmptyState
              className={styles.emptyState}
              title=""
              message="No job sets found in this project."
            />
          ) : jobsCount !== 0 && filteredJobs && filteredJobs.length === 0 ? (
            <EmptyState
              className={styles.emptyState}
              title=""
              message="No job sets found with the given filters."
            />
          ) : (
            <ul className={styles.listContainer}>
              {filteredJobs?.map((job) => {
                const isFailedSubjob =
                  restJobStateToNodeState(job.state) === NodeState.ERROR;

                const isRunningSubjob =
                  restJobStateToNodeState(job.state) === NodeState.RUNNING;

                return (
                  <li key={job.job?.id} className={styles.listItem}>
                    <SearchResultItem
                      key={job.job?.id}
                      searchValue=""
                      title=""
                      onClick={() => handleClick(job.job?.id || '')}
                    >
                      <span className={styles.datetimeWrapper}>
                        {isFailedSubjob ? (
                          <Icon small color="red" aria-label="Failing status">
                            <StatusWarningSVG data-testid="GlobalFilter__status_failing" />
                          </Icon>
                        ) : isRunningSubjob ? (
                          <Icon small color="green" aria-label="Running status">
                            <StatusDotsSVG data-testid="GlobalFilter__status_running" />
                          </Icon>
                        ) : (
                          <Icon small color="green" aria-label="Passing status">
                            <StatusCheckmarkSVG data-testid="GlobalFilter__status_passing" />
                          </Icon>
                        )}

                        {job?.created
                          ? getStandardDateFromISOString(job.created)
                          : '-'}
                      </span>
                      <span className={styles.id}>{job.job?.id}</span>
                    </SearchResultItem>
                  </li>
                );
              })}
            </ul>
          )}
        </Form>
      )}
    </div>
  );
};

export default GlobalFilter;
