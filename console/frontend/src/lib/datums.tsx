import capitalize from 'lodash/capitalize';
import React from 'react';

import {DatumState} from '@dash-frontend/api/pps';
import {
  ArrowCircleRightSVG,
  Icon,
  StatusCheckmarkSVG,
  StatusDotsSVG,
  StatusSkipSVG,
  StatusUnknownSVG,
  StatusWarningSVG,
} from '@pachyderm/components';

export const getDatumStateIcon = (state?: DatumState) => {
  const IconSVG = getDatumStateSVG(state);
  return (
    <Icon color={getDatumStateColor(state) ?? undefined}>
      {IconSVG && <IconSVG />}
    </Icon>
  );
};

export const getDatumStateSVG = (state?: DatumState) => {
  switch (state) {
    case DatumState.FAILED:
      return StatusWarningSVG;
    case DatumState.RECOVERED:
      return ArrowCircleRightSVG;
    case DatumState.SKIPPED:
      return StatusSkipSVG;
    case DatumState.STARTING:
      return StatusDotsSVG;
    case DatumState.SUCCESS:
      return StatusCheckmarkSVG;
    case DatumState.UNKNOWN:
      return StatusUnknownSVG;
    default:
      return null;
  }
};

export const getDatumStateColor = (state?: DatumState) => {
  switch (state) {
    case DatumState.FAILED:
      return 'red';
    case DatumState.RECOVERED:
    case DatumState.SKIPPED:
    case DatumState.UNKNOWN:
      return 'black';
    case DatumState.STARTING:
    case DatumState.SUCCESS:
      return 'green';
    default:
      return null;
  }
};

export const readableDatumState = (datumState: DatumState | string) => {
  return capitalize(datumState);
};
