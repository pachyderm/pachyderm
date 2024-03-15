import {Page} from 'playwright';

export type Journey = (page: Page) => Promise<void>;
export type JourneyObject = {name: string; journey: Journey};
