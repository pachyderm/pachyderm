import { group, sleep } from "k6";
import { Counter } from 'k6/metrics';

const DURATION = 30;
const POLL_INTERVAL = 3;

import { authenticate, Authentication, createGraphqlClient, Operation } from "../utils";

export const options = {
  vus: 20,
  duration: `${DURATION}s`,
};

// ran once for all vus, results passed to default export function
export const setup = () => {
  return authenticate();
};

export default function (auth: Authentication) {
  if (!auth || !auth.pachToken || !auth.idToken) {
    sleep(DURATION);
    throw `not authenticated`;
  }

  const { query, pollRequest } = createGraphqlClient(auth);
  const pages = [
    () => {
      group("landing", () => {
        query(Operation.getAccount);
        query(Operation.projects);

        pollRequest(()=>{
          query(Operation.project);
        }, POLL_INTERVAL, DURATION);
      });
    },
    () => {
      group("lineage", () => {
        query(Operation.getAccount);
        query(Operation.projects);
        query(Operation.project);
        query(Operation.repos);

        pollRequest(()=>{
          query(Operation.jobSets);
        }, POLL_INTERVAL, DURATION);
      });
    },
    () => {
      group("repo", () => {
        query(Operation.getAccount);
        query(Operation.project);
        query(Operation.repos);
        query(Operation.getDag);

        pollRequest(()=>{
          query(Operation.repo);
          query(Operation.jobSets);
          query(Operation.getCommits);
        }, POLL_INTERVAL, DURATION);
      });
    },
    () => {
      group("pipeline", () => {
        query(Operation.getAccount);
        query(Operation.project);
        query(Operation.repos);
        query(Operation.getDag);
        query(Operation.pipeline);

        pollRequest(()=>{
          query(Operation.jobs);
          query(Operation.jobSets);
        }, POLL_INTERVAL, DURATION)
      });
    },
    () => {
      group("files", () => {
        query(Operation.getAccount);
        query(Operation.project);
        query(Operation.repos);
        query(Operation.getDag);
        query(Operation.getFiles);

        pollRequest(()=>{
          query(Operation.repo);
          query(Operation.jobSets);
          query(Operation.getCommits);
        }, POLL_INTERVAL, DURATION);
      });
    },
  ];

  const pageIndex = Math.floor(Math.random() * pages.length);
  pages[pageIndex]();
}
