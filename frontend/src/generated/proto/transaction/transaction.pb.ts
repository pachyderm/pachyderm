/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb";
import * as GoogleProtobufEmpty from "../google/protobuf/empty.pb";
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb";
import * as Pfs_v2Pfs from "../pfs/pfs.pb";
import * as Pps_v2Pps from "../pps/pps.pb";
export type DeleteAllRequest = {
    __typename?: "DeleteAllRequest";
};

export type TransactionRequest = {
    __typename?: "TransactionRequest";
    createRepo?: Pfs_v2Pfs.CreateRepoRequest;
    deleteRepo?: Pfs_v2Pfs.DeleteRepoRequest;
    startCommit?: Pfs_v2Pfs.StartCommitRequest;
    finishCommit?: Pfs_v2Pfs.FinishCommitRequest;
    squashCommitSet?: Pfs_v2Pfs.SquashCommitSetRequest;
    createBranch?: Pfs_v2Pfs.CreateBranchRequest;
    deleteBranch?: Pfs_v2Pfs.DeleteBranchRequest;
    updateJobState?: Pps_v2Pps.UpdateJobStateRequest;
    stopJob?: Pps_v2Pps.StopJobRequest;
    createPipelineV2?: Pps_v2Pps.CreatePipelineTransaction;
};

export type TransactionResponse = {
    __typename?: "TransactionResponse";
    commit?: Pfs_v2Pfs.Commit;
};

export type Transaction = {
    __typename?: "Transaction";
    id?: string;
};

export type TransactionInfo = {
    __typename?: "TransactionInfo";
    transaction?: Transaction;
    requests?: TransactionRequest[];
    responses?: TransactionResponse[];
    started?: GoogleProtobufTimestamp.Timestamp;
    version?: string;
};

export type TransactionInfos = {
    __typename?: "TransactionInfos";
    transactionInfo?: TransactionInfo[];
};

export type BatchTransactionRequest = {
    __typename?: "BatchTransactionRequest";
    requests?: TransactionRequest[];
};

export type StartTransactionRequest = {
    __typename?: "StartTransactionRequest";
};

export type InspectTransactionRequest = {
    __typename?: "InspectTransactionRequest";
    transaction?: Transaction;
};

export type DeleteTransactionRequest = {
    __typename?: "DeleteTransactionRequest";
    transaction?: Transaction;
};

export type ListTransactionRequest = {
    __typename?: "ListTransactionRequest";
};

export type FinishTransactionRequest = {
    __typename?: "FinishTransactionRequest";
    transaction?: Transaction;
};

export class API {
    static BatchTransaction(req: BatchTransactionRequest, initReq?: fm.InitReq): Promise<TransactionInfo> {
        return fm.fetchReq<BatchTransactionRequest, TransactionInfo>(`/transaction_v2.API/BatchTransaction`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static StartTransaction(req: StartTransactionRequest, initReq?: fm.InitReq): Promise<Transaction> {
        return fm.fetchReq<StartTransactionRequest, Transaction>(`/transaction_v2.API/StartTransaction`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectTransaction(req: InspectTransactionRequest, initReq?: fm.InitReq): Promise<TransactionInfo> {
        return fm.fetchReq<InspectTransactionRequest, TransactionInfo>(`/transaction_v2.API/InspectTransaction`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteTransaction(req: DeleteTransactionRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<DeleteTransactionRequest, GoogleProtobufEmpty.Empty>(`/transaction_v2.API/DeleteTransaction`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListTransaction(req: ListTransactionRequest, initReq?: fm.InitReq): Promise<TransactionInfos> {
        return fm.fetchReq<ListTransactionRequest, TransactionInfos>(`/transaction_v2.API/ListTransaction`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static FinishTransaction(req: FinishTransactionRequest, initReq?: fm.InitReq): Promise<TransactionInfo> {
        return fm.fetchReq<FinishTransactionRequest, TransactionInfo>(`/transaction_v2.API/FinishTransaction`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteAll(req: DeleteAllRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<DeleteAllRequest, GoogleProtobufEmpty.Empty>(`/transaction_v2.API/DeleteAll`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
}
