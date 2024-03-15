import {JourneyObject} from './types';

export class TestIsFinished extends Error {}

export const wait = async (delay: number): Promise<void> =>
  new Promise((resolve) => setTimeout(resolve, delay));

export const iToSeconds = (i: number) => i * 1000;

export const pickRandomJourney = (journeys: JourneyObject[]) => {
  const randomIndex = Math.floor(Math.random() * journeys.length);
  const journey = journeys[randomIndex];
  return journey;
};
