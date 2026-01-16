// package: galachain
// file: galachain.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as galachain_pb from "./galachain_pb";

interface IGalaChainServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    getPrice: IGalaChainServiceService_IGetPrice;
    getBalance: IGalaChainServiceService_IGetBalance;
    executeSwap: IGalaChainServiceService_IExecuteSwap;
    watchPrices: IGalaChainServiceService_IWatchPrices;
    createWalletSession: IGalaChainServiceService_ICreateWalletSession;
    healthCheck: IGalaChainServiceService_IHealthCheck;
}

interface IGalaChainServiceService_IGetPrice extends grpc.MethodDefinition<galachain_pb.GetPriceRequest, galachain_pb.PriceResponse> {
    path: "/galachain.GalaChainService/GetPrice";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<galachain_pb.GetPriceRequest>;
    requestDeserialize: grpc.deserialize<galachain_pb.GetPriceRequest>;
    responseSerialize: grpc.serialize<galachain_pb.PriceResponse>;
    responseDeserialize: grpc.deserialize<galachain_pb.PriceResponse>;
}
interface IGalaChainServiceService_IGetBalance extends grpc.MethodDefinition<galachain_pb.BalanceRequest, galachain_pb.BalanceResponse> {
    path: "/galachain.GalaChainService/GetBalance";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<galachain_pb.BalanceRequest>;
    requestDeserialize: grpc.deserialize<galachain_pb.BalanceRequest>;
    responseSerialize: grpc.serialize<galachain_pb.BalanceResponse>;
    responseDeserialize: grpc.deserialize<galachain_pb.BalanceResponse>;
}
interface IGalaChainServiceService_IExecuteSwap extends grpc.MethodDefinition<galachain_pb.SwapRequest, galachain_pb.SwapResponse> {
    path: "/galachain.GalaChainService/ExecuteSwap";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<galachain_pb.SwapRequest>;
    requestDeserialize: grpc.deserialize<galachain_pb.SwapRequest>;
    responseSerialize: grpc.serialize<galachain_pb.SwapResponse>;
    responseDeserialize: grpc.deserialize<galachain_pb.SwapResponse>;
}
interface IGalaChainServiceService_IWatchPrices extends grpc.MethodDefinition<galachain_pb.WatchPricesRequest, galachain_pb.PriceUpdate> {
    path: "/galachain.GalaChainService/WatchPrices";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<galachain_pb.WatchPricesRequest>;
    requestDeserialize: grpc.deserialize<galachain_pb.WatchPricesRequest>;
    responseSerialize: grpc.serialize<galachain_pb.PriceUpdate>;
    responseDeserialize: grpc.deserialize<galachain_pb.PriceUpdate>;
}
interface IGalaChainServiceService_ICreateWalletSession extends grpc.MethodDefinition<galachain_pb.WalletSessionRequest, galachain_pb.WalletSessionResponse> {
    path: "/galachain.GalaChainService/CreateWalletSession";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<galachain_pb.WalletSessionRequest>;
    requestDeserialize: grpc.deserialize<galachain_pb.WalletSessionRequest>;
    responseSerialize: grpc.serialize<galachain_pb.WalletSessionResponse>;
    responseDeserialize: grpc.deserialize<galachain_pb.WalletSessionResponse>;
}
interface IGalaChainServiceService_IHealthCheck extends grpc.MethodDefinition<galachain_pb.HealthCheckRequest, galachain_pb.HealthCheckResponse> {
    path: "/galachain.GalaChainService/HealthCheck";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<galachain_pb.HealthCheckRequest>;
    requestDeserialize: grpc.deserialize<galachain_pb.HealthCheckRequest>;
    responseSerialize: grpc.serialize<galachain_pb.HealthCheckResponse>;
    responseDeserialize: grpc.deserialize<galachain_pb.HealthCheckResponse>;
}

export const GalaChainServiceService: IGalaChainServiceService;

export interface IGalaChainServiceServer extends grpc.UntypedServiceImplementation {
    getPrice: grpc.handleUnaryCall<galachain_pb.GetPriceRequest, galachain_pb.PriceResponse>;
    getBalance: grpc.handleUnaryCall<galachain_pb.BalanceRequest, galachain_pb.BalanceResponse>;
    executeSwap: grpc.handleUnaryCall<galachain_pb.SwapRequest, galachain_pb.SwapResponse>;
    watchPrices: grpc.handleServerStreamingCall<galachain_pb.WatchPricesRequest, galachain_pb.PriceUpdate>;
    createWalletSession: grpc.handleUnaryCall<galachain_pb.WalletSessionRequest, galachain_pb.WalletSessionResponse>;
    healthCheck: grpc.handleUnaryCall<galachain_pb.HealthCheckRequest, galachain_pb.HealthCheckResponse>;
}

export interface IGalaChainServiceClient {
    getPrice(request: galachain_pb.GetPriceRequest, callback: (error: grpc.ServiceError | null, response: galachain_pb.PriceResponse) => void): grpc.ClientUnaryCall;
    getPrice(request: galachain_pb.GetPriceRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: galachain_pb.PriceResponse) => void): grpc.ClientUnaryCall;
    getPrice(request: galachain_pb.GetPriceRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: galachain_pb.PriceResponse) => void): grpc.ClientUnaryCall;
    getBalance(request: galachain_pb.BalanceRequest, callback: (error: grpc.ServiceError | null, response: galachain_pb.BalanceResponse) => void): grpc.ClientUnaryCall;
    getBalance(request: galachain_pb.BalanceRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: galachain_pb.BalanceResponse) => void): grpc.ClientUnaryCall;
    getBalance(request: galachain_pb.BalanceRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: galachain_pb.BalanceResponse) => void): grpc.ClientUnaryCall;
    executeSwap(request: galachain_pb.SwapRequest, callback: (error: grpc.ServiceError | null, response: galachain_pb.SwapResponse) => void): grpc.ClientUnaryCall;
    executeSwap(request: galachain_pb.SwapRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: galachain_pb.SwapResponse) => void): grpc.ClientUnaryCall;
    executeSwap(request: galachain_pb.SwapRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: galachain_pb.SwapResponse) => void): grpc.ClientUnaryCall;
    watchPrices(request: galachain_pb.WatchPricesRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<galachain_pb.PriceUpdate>;
    watchPrices(request: galachain_pb.WatchPricesRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<galachain_pb.PriceUpdate>;
    createWalletSession(request: galachain_pb.WalletSessionRequest, callback: (error: grpc.ServiceError | null, response: galachain_pb.WalletSessionResponse) => void): grpc.ClientUnaryCall;
    createWalletSession(request: galachain_pb.WalletSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: galachain_pb.WalletSessionResponse) => void): grpc.ClientUnaryCall;
    createWalletSession(request: galachain_pb.WalletSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: galachain_pb.WalletSessionResponse) => void): grpc.ClientUnaryCall;
    healthCheck(request: galachain_pb.HealthCheckRequest, callback: (error: grpc.ServiceError | null, response: galachain_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
    healthCheck(request: galachain_pb.HealthCheckRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: galachain_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
    healthCheck(request: galachain_pb.HealthCheckRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: galachain_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
}

export class GalaChainServiceClient extends grpc.Client implements IGalaChainServiceClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public getPrice(request: galachain_pb.GetPriceRequest, callback: (error: grpc.ServiceError | null, response: galachain_pb.PriceResponse) => void): grpc.ClientUnaryCall;
    public getPrice(request: galachain_pb.GetPriceRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: galachain_pb.PriceResponse) => void): grpc.ClientUnaryCall;
    public getPrice(request: galachain_pb.GetPriceRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: galachain_pb.PriceResponse) => void): grpc.ClientUnaryCall;
    public getBalance(request: galachain_pb.BalanceRequest, callback: (error: grpc.ServiceError | null, response: galachain_pb.BalanceResponse) => void): grpc.ClientUnaryCall;
    public getBalance(request: galachain_pb.BalanceRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: galachain_pb.BalanceResponse) => void): grpc.ClientUnaryCall;
    public getBalance(request: galachain_pb.BalanceRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: galachain_pb.BalanceResponse) => void): grpc.ClientUnaryCall;
    public executeSwap(request: galachain_pb.SwapRequest, callback: (error: grpc.ServiceError | null, response: galachain_pb.SwapResponse) => void): grpc.ClientUnaryCall;
    public executeSwap(request: galachain_pb.SwapRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: galachain_pb.SwapResponse) => void): grpc.ClientUnaryCall;
    public executeSwap(request: galachain_pb.SwapRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: galachain_pb.SwapResponse) => void): grpc.ClientUnaryCall;
    public watchPrices(request: galachain_pb.WatchPricesRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<galachain_pb.PriceUpdate>;
    public watchPrices(request: galachain_pb.WatchPricesRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<galachain_pb.PriceUpdate>;
    public createWalletSession(request: galachain_pb.WalletSessionRequest, callback: (error: grpc.ServiceError | null, response: galachain_pb.WalletSessionResponse) => void): grpc.ClientUnaryCall;
    public createWalletSession(request: galachain_pb.WalletSessionRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: galachain_pb.WalletSessionResponse) => void): grpc.ClientUnaryCall;
    public createWalletSession(request: galachain_pb.WalletSessionRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: galachain_pb.WalletSessionResponse) => void): grpc.ClientUnaryCall;
    public healthCheck(request: galachain_pb.HealthCheckRequest, callback: (error: grpc.ServiceError | null, response: galachain_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
    public healthCheck(request: galachain_pb.HealthCheckRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: galachain_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
    public healthCheck(request: galachain_pb.HealthCheckRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: galachain_pb.HealthCheckResponse) => void): grpc.ClientUnaryCall;
}
