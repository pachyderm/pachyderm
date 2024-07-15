/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

export enum State {
    UNKNOWN = "UNKNOWN",
    RUNNING = "RUNNING",
    SUCCESS = "SUCCESS",
    FAILURE = "FAILURE",
    CLAIMED = "CLAIMED",
}

export type Group = {
    __typename?: "Group";
    namespace?: string;
    group?: string;
};

export type TaskInfo = {
    __typename?: "TaskInfo";
    id?: string;
    group?: Group;
    state?: State;
    reason?: string;
    inputType?: string;
    inputData?: string;
};

export type ListTaskRequest = {
    __typename?: "ListTaskRequest";
    group?: Group;
};
