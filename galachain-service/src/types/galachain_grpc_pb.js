// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// proto/galachain.proto
'use strict';
var grpc = require('@grpc/grpc-js');
var galachain_pb = require('./galachain_pb.js');

function serialize_galachain_BalanceRequest(arg) {
  if (!(arg instanceof galachain_pb.BalanceRequest)) {
    throw new Error('Expected argument of type galachain.BalanceRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_BalanceRequest(buffer_arg) {
  return galachain_pb.BalanceRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_galachain_BalanceResponse(arg) {
  if (!(arg instanceof galachain_pb.BalanceResponse)) {
    throw new Error('Expected argument of type galachain.BalanceResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_BalanceResponse(buffer_arg) {
  return galachain_pb.BalanceResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_galachain_GetPriceRequest(arg) {
  if (!(arg instanceof galachain_pb.GetPriceRequest)) {
    throw new Error('Expected argument of type galachain.GetPriceRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_GetPriceRequest(buffer_arg) {
  return galachain_pb.GetPriceRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_galachain_HealthCheckRequest(arg) {
  if (!(arg instanceof galachain_pb.HealthCheckRequest)) {
    throw new Error('Expected argument of type galachain.HealthCheckRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_HealthCheckRequest(buffer_arg) {
  return galachain_pb.HealthCheckRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_galachain_HealthCheckResponse(arg) {
  if (!(arg instanceof galachain_pb.HealthCheckResponse)) {
    throw new Error('Expected argument of type galachain.HealthCheckResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_HealthCheckResponse(buffer_arg) {
  return galachain_pb.HealthCheckResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_galachain_PriceResponse(arg) {
  if (!(arg instanceof galachain_pb.PriceResponse)) {
    throw new Error('Expected argument of type galachain.PriceResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_PriceResponse(buffer_arg) {
  return galachain_pb.PriceResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_galachain_PriceUpdate(arg) {
  if (!(arg instanceof galachain_pb.PriceUpdate)) {
    throw new Error('Expected argument of type galachain.PriceUpdate');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_PriceUpdate(buffer_arg) {
  return galachain_pb.PriceUpdate.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_galachain_SwapRequest(arg) {
  if (!(arg instanceof galachain_pb.SwapRequest)) {
    throw new Error('Expected argument of type galachain.SwapRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_SwapRequest(buffer_arg) {
  return galachain_pb.SwapRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_galachain_SwapResponse(arg) {
  if (!(arg instanceof galachain_pb.SwapResponse)) {
    throw new Error('Expected argument of type galachain.SwapResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_SwapResponse(buffer_arg) {
  return galachain_pb.SwapResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_galachain_WalletSessionRequest(arg) {
  if (!(arg instanceof galachain_pb.WalletSessionRequest)) {
    throw new Error('Expected argument of type galachain.WalletSessionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_WalletSessionRequest(buffer_arg) {
  return galachain_pb.WalletSessionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_galachain_WalletSessionResponse(arg) {
  if (!(arg instanceof galachain_pb.WalletSessionResponse)) {
    throw new Error('Expected argument of type galachain.WalletSessionResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_WalletSessionResponse(buffer_arg) {
  return galachain_pb.WalletSessionResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_galachain_WatchPricesRequest(arg) {
  if (!(arg instanceof galachain_pb.WatchPricesRequest)) {
    throw new Error('Expected argument of type galachain.WatchPricesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_galachain_WatchPricesRequest(buffer_arg) {
  return galachain_pb.WatchPricesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


// GalaChain service for managing GalaChain operations
var GalaChainServiceService = exports.GalaChainServiceService = {
  // Get current price for a trading pair
getPrice: {
    path: '/galachain.GalaChainService/GetPrice',
    requestStream: false,
    responseStream: false,
    requestType: galachain_pb.GetPriceRequest,
    responseType: galachain_pb.PriceResponse,
    requestSerialize: serialize_galachain_GetPriceRequest,
    requestDeserialize: deserialize_galachain_GetPriceRequest,
    responseSerialize: serialize_galachain_PriceResponse,
    responseDeserialize: deserialize_galachain_PriceResponse,
  },
  // Get wallet balance
getBalance: {
    path: '/galachain.GalaChainService/GetBalance',
    requestStream: false,
    responseStream: false,
    requestType: galachain_pb.BalanceRequest,
    responseType: galachain_pb.BalanceResponse,
    requestSerialize: serialize_galachain_BalanceRequest,
    requestDeserialize: deserialize_galachain_BalanceRequest,
    responseSerialize: serialize_galachain_BalanceResponse,
    responseDeserialize: deserialize_galachain_BalanceResponse,
  },
  // Execute a token swap
executeSwap: {
    path: '/galachain.GalaChainService/ExecuteSwap',
    requestStream: false,
    responseStream: false,
    requestType: galachain_pb.SwapRequest,
    responseType: galachain_pb.SwapResponse,
    requestSerialize: serialize_galachain_SwapRequest,
    requestDeserialize: deserialize_galachain_SwapRequest,
    responseSerialize: serialize_galachain_SwapResponse,
    responseDeserialize: deserialize_galachain_SwapResponse,
  },
  // Watch price updates (streaming)
watchPrices: {
    path: '/galachain.GalaChainService/WatchPrices',
    requestStream: false,
    responseStream: true,
    requestType: galachain_pb.WatchPricesRequest,
    responseType: galachain_pb.PriceUpdate,
    requestSerialize: serialize_galachain_WatchPricesRequest,
    requestDeserialize: deserialize_galachain_WatchPricesRequest,
    responseSerialize: serialize_galachain_PriceUpdate,
    responseDeserialize: deserialize_galachain_PriceUpdate,
  },
  // Create wallet connection session
createWalletSession: {
    path: '/galachain.GalaChainService/CreateWalletSession',
    requestStream: false,
    responseStream: false,
    requestType: galachain_pb.WalletSessionRequest,
    responseType: galachain_pb.WalletSessionResponse,
    requestSerialize: serialize_galachain_WalletSessionRequest,
    requestDeserialize: deserialize_galachain_WalletSessionRequest,
    responseSerialize: serialize_galachain_WalletSessionResponse,
    responseDeserialize: deserialize_galachain_WalletSessionResponse,
  },
  // Health check
healthCheck: {
    path: '/galachain.GalaChainService/HealthCheck',
    requestStream: false,
    responseStream: false,
    requestType: galachain_pb.HealthCheckRequest,
    responseType: galachain_pb.HealthCheckResponse,
    requestSerialize: serialize_galachain_HealthCheckRequest,
    requestDeserialize: deserialize_galachain_HealthCheckRequest,
    responseSerialize: serialize_galachain_HealthCheckResponse,
    responseDeserialize: deserialize_galachain_HealthCheckResponse,
  },
};

exports.GalaChainServiceClient = grpc.makeGenericClientConstructor(GalaChainServiceService, 'GalaChainService');
