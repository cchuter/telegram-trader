// package: galachain
// file: galachain.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class GetPriceRequest extends jspb.Message { 
    getPair(): string;
    setPair(value: string): GetPriceRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetPriceRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetPriceRequest): GetPriceRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetPriceRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetPriceRequest;
    static deserializeBinaryFromReader(message: GetPriceRequest, reader: jspb.BinaryReader): GetPriceRequest;
}

export namespace GetPriceRequest {
    export type AsObject = {
        pair: string,
    }
}

export class PriceResponse extends jspb.Message { 
    getPair(): string;
    setPair(value: string): PriceResponse;
    getPrice(): string;
    setPrice(value: string): PriceResponse;
    getTimestamp(): number;
    setTimestamp(value: number): PriceResponse;
    getBid(): string;
    setBid(value: string): PriceResponse;
    getAsk(): string;
    setAsk(value: string): PriceResponse;
    getVolume24h(): string;
    setVolume24h(value: string): PriceResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PriceResponse.AsObject;
    static toObject(includeInstance: boolean, msg: PriceResponse): PriceResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PriceResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PriceResponse;
    static deserializeBinaryFromReader(message: PriceResponse, reader: jspb.BinaryReader): PriceResponse;
}

export namespace PriceResponse {
    export type AsObject = {
        pair: string,
        price: string,
        timestamp: number,
        bid: string,
        ask: string,
        volume24h: string,
    }
}

export class BalanceRequest extends jspb.Message { 
    getUserId(): number;
    setUserId(value: number): BalanceRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BalanceRequest.AsObject;
    static toObject(includeInstance: boolean, msg: BalanceRequest): BalanceRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BalanceRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BalanceRequest;
    static deserializeBinaryFromReader(message: BalanceRequest, reader: jspb.BinaryReader): BalanceRequest;
}

export namespace BalanceRequest {
    export type AsObject = {
        userId: number,
    }
}

export class BalanceResponse extends jspb.Message { 
    clearBalancesList(): void;
    getBalancesList(): Array<TokenBalance>;
    setBalancesList(value: Array<TokenBalance>): BalanceResponse;
    addBalances(value?: TokenBalance, index?: number): TokenBalance;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BalanceResponse.AsObject;
    static toObject(includeInstance: boolean, msg: BalanceResponse): BalanceResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BalanceResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BalanceResponse;
    static deserializeBinaryFromReader(message: BalanceResponse, reader: jspb.BinaryReader): BalanceResponse;
}

export namespace BalanceResponse {
    export type AsObject = {
        balancesList: Array<TokenBalance.AsObject>,
    }
}

export class TokenBalance extends jspb.Message { 
    getToken(): string;
    setToken(value: string): TokenBalance;
    getBalance(): string;
    setBalance(value: string): TokenBalance;
    getUsdValue(): string;
    setUsdValue(value: string): TokenBalance;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TokenBalance.AsObject;
    static toObject(includeInstance: boolean, msg: TokenBalance): TokenBalance.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TokenBalance, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TokenBalance;
    static deserializeBinaryFromReader(message: TokenBalance, reader: jspb.BinaryReader): TokenBalance;
}

export namespace TokenBalance {
    export type AsObject = {
        token: string,
        balance: string,
        usdValue: string,
    }
}

export class SwapRequest extends jspb.Message { 
    getUserId(): number;
    setUserId(value: number): SwapRequest;
    getFromToken(): string;
    setFromToken(value: string): SwapRequest;
    getToToken(): string;
    setToToken(value: string): SwapRequest;
    getAmount(): string;
    setAmount(value: string): SwapRequest;
    getSlippageBps(): number;
    setSlippageBps(value: number): SwapRequest;
    getFeeTier(): number;
    setFeeTier(value: number): SwapRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SwapRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SwapRequest): SwapRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SwapRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SwapRequest;
    static deserializeBinaryFromReader(message: SwapRequest, reader: jspb.BinaryReader): SwapRequest;
}

export namespace SwapRequest {
    export type AsObject = {
        userId: number,
        fromToken: string,
        toToken: string,
        amount: string,
        slippageBps: number,
        feeTier: number,
    }
}

export class SwapResponse extends jspb.Message { 
    getTxHash(): string;
    setTxHash(value: string): SwapResponse;
    getAmountIn(): string;
    setAmountIn(value: string): SwapResponse;
    getAmountOut(): string;
    setAmountOut(value: string): SwapResponse;
    getFee(): string;
    setFee(value: string): SwapResponse;
    getStatus(): string;
    setStatus(value: string): SwapResponse;
    getErrorMessage(): string;
    setErrorMessage(value: string): SwapResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SwapResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SwapResponse): SwapResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SwapResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SwapResponse;
    static deserializeBinaryFromReader(message: SwapResponse, reader: jspb.BinaryReader): SwapResponse;
}

export namespace SwapResponse {
    export type AsObject = {
        txHash: string,
        amountIn: string,
        amountOut: string,
        fee: string,
        status: string,
        errorMessage: string,
    }
}

export class WatchPricesRequest extends jspb.Message { 
    clearPairsList(): void;
    getPairsList(): Array<string>;
    setPairsList(value: Array<string>): WatchPricesRequest;
    addPairs(value: string, index?: number): string;
    getIntervalMs(): number;
    setIntervalMs(value: number): WatchPricesRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WatchPricesRequest.AsObject;
    static toObject(includeInstance: boolean, msg: WatchPricesRequest): WatchPricesRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WatchPricesRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WatchPricesRequest;
    static deserializeBinaryFromReader(message: WatchPricesRequest, reader: jspb.BinaryReader): WatchPricesRequest;
}

export namespace WatchPricesRequest {
    export type AsObject = {
        pairsList: Array<string>,
        intervalMs: number,
    }
}

export class PriceUpdate extends jspb.Message { 
    getPair(): string;
    setPair(value: string): PriceUpdate;
    getPrice(): string;
    setPrice(value: string): PriceUpdate;
    getTimestamp(): number;
    setTimestamp(value: number): PriceUpdate;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PriceUpdate.AsObject;
    static toObject(includeInstance: boolean, msg: PriceUpdate): PriceUpdate.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PriceUpdate, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PriceUpdate;
    static deserializeBinaryFromReader(message: PriceUpdate, reader: jspb.BinaryReader): PriceUpdate;
}

export namespace PriceUpdate {
    export type AsObject = {
        pair: string,
        price: string,
        timestamp: number,
    }
}

export class WalletSessionRequest extends jspb.Message { 
    getUserId(): number;
    setUserId(value: number): WalletSessionRequest;
    getMethod(): string;
    setMethod(value: string): WalletSessionRequest;

    hasPrivateKey(): boolean;
    clearPrivateKey(): void;
    getPrivateKey(): string | undefined;
    setPrivateKey(value: string): WalletSessionRequest;

    hasPublicKey(): boolean;
    clearPublicKey(): void;
    getPublicKey(): string | undefined;
    setPublicKey(value: string): WalletSessionRequest;

    hasAddress(): boolean;
    clearAddress(): void;
    getAddress(): string | undefined;
    setAddress(value: string): WalletSessionRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WalletSessionRequest.AsObject;
    static toObject(includeInstance: boolean, msg: WalletSessionRequest): WalletSessionRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WalletSessionRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WalletSessionRequest;
    static deserializeBinaryFromReader(message: WalletSessionRequest, reader: jspb.BinaryReader): WalletSessionRequest;
}

export namespace WalletSessionRequest {
    export type AsObject = {
        userId: number,
        method: string,
        privateKey?: string,
        publicKey?: string,
        address?: string,
    }
}

export class WalletSessionResponse extends jspb.Message { 
    getSessionId(): string;
    setSessionId(value: string): WalletSessionResponse;
    getMethod(): string;
    setMethod(value: string): WalletSessionResponse;

    hasQrCodeUri(): boolean;
    clearQrCodeUri(): void;
    getQrCodeUri(): string | undefined;
    setQrCodeUri(value: string): WalletSessionResponse;

    hasDeepLink(): boolean;
    clearDeepLink(): void;
    getDeepLink(): string | undefined;
    setDeepLink(value: string): WalletSessionResponse;

    hasAddress(): boolean;
    clearAddress(): void;
    getAddress(): string | undefined;
    setAddress(value: string): WalletSessionResponse;
    getSuccess(): boolean;
    setSuccess(value: boolean): WalletSessionResponse;
    getErrorMessage(): string;
    setErrorMessage(value: string): WalletSessionResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WalletSessionResponse.AsObject;
    static toObject(includeInstance: boolean, msg: WalletSessionResponse): WalletSessionResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WalletSessionResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WalletSessionResponse;
    static deserializeBinaryFromReader(message: WalletSessionResponse, reader: jspb.BinaryReader): WalletSessionResponse;
}

export namespace WalletSessionResponse {
    export type AsObject = {
        sessionId: string,
        method: string,
        qrCodeUri?: string,
        deepLink?: string,
        address?: string,
        success: boolean,
        errorMessage: string,
    }
}

export class HealthCheckRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): HealthCheckRequest.AsObject;
    static toObject(includeInstance: boolean, msg: HealthCheckRequest): HealthCheckRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: HealthCheckRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): HealthCheckRequest;
    static deserializeBinaryFromReader(message: HealthCheckRequest, reader: jspb.BinaryReader): HealthCheckRequest;
}

export namespace HealthCheckRequest {
    export type AsObject = {
    }
}

export class HealthCheckResponse extends jspb.Message { 
    getStatus(): string;
    setStatus(value: string): HealthCheckResponse;

    getDependenciesMap(): jspb.Map<string, string>;
    clearDependenciesMap(): void;
    getUptimeSeconds(): number;
    setUptimeSeconds(value: number): HealthCheckResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): HealthCheckResponse.AsObject;
    static toObject(includeInstance: boolean, msg: HealthCheckResponse): HealthCheckResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: HealthCheckResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): HealthCheckResponse;
    static deserializeBinaryFromReader(message: HealthCheckResponse, reader: jspb.BinaryReader): HealthCheckResponse;
}

export namespace HealthCheckResponse {
    export type AsObject = {
        status: string,

        dependenciesMap: Array<[string, string]>,
        uptimeSeconds: number,
    }
}

export enum ErrorCode {
    UNKNOWN = 0,
    INVALID_REQUEST = 1,
    WALLET_NOT_CONNECTED = 2,
    INSUFFICIENT_BALANCE = 3,
    SLIPPAGE_EXCEEDED = 4,
    RATE_LIMIT_EXCEEDED = 5,
    NETWORK_ERROR = 6,
    TRANSACTION_FAILED = 7,
}
