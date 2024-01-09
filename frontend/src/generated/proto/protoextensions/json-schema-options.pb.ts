/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/
export type FieldOptions = {
    __typename?: "FieldOptions";
    ignore?: boolean;
    required?: boolean;
    minLength?: number;
    maxLength?: number;
    pattern?: string;
};

export type FileOptions = {
    __typename?: "FileOptions";
    ignore?: boolean;
    extension?: string;
};

export type MessageOptions = {
    __typename?: "MessageOptions";
    ignore?: boolean;
    allFieldsRequired?: boolean;
    allowNullValues?: boolean;
    disallowAdditionalProperties?: boolean;
    enumsAsConstants?: boolean;
};

export type EnumOptions = {
    __typename?: "EnumOptions";
    enumsAsConstants?: boolean;
    enumsAsStringsOnly?: boolean;
    enumsTrimPrefix?: boolean;
    ignore?: boolean;
};
