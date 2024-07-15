import {Input} from '@dash-frontend/api/pps';

export const checkCronInputs = (input?: Input) => {
  if (!input) return false;
  if (input.cron) return true;

  input.join?.forEach((input) => {
    checkCronInputs(input);
  });
  input.group?.forEach((input) => {
    checkCronInputs(input);
  });
  input.cross?.forEach((input) => {
    checkCronInputs(input);
  });
  input.union?.forEach((input) => {
    checkCronInputs(input);
  });

  return false;
};
