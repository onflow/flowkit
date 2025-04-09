/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mocks

import (
	"context"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flowkit/v2/tests"
)

const (
	GetAccountFunc                 = "GetAccount"
	SendSignedTransactionFunc      = "SendSignedTransaction"
	GetCollectionFunc              = "GetCollection"
	GetTransactionResultFunc       = "GetTransactionResult"
	GetEventsFunc                  = "GetEvents"
	GetLatestBlockFunc             = "GetLatestBlock"
	GetBlockByHeightFunc           = "GetBlockByHeight"
	GetBlockByIDFunc               = "GetBlockByID"
	ExecuteScriptFunc              = "ExecuteScript"
	GetTransactionFunc             = "GetTransaction"
	GetSystemTransactionFunc       = "GetSystemTransaction"
	GetSystemTransactionResultFunc = "GetSystemTransactionResult"
)

type TestGateway struct {
	Mock                           *Gateway
	SendSignedTransaction          *mock.Call
	GetAccount                     *mock.Call
	GetCollection                  *mock.Call
	GetTransactionResult           *mock.Call
	GetEvents                      *mock.Call
	GetLatestBlock                 *mock.Call
	GetBlockByHeight               *mock.Call
	GetBlockByID                   *mock.Call
	ExecuteScript                  *mock.Call
	GetTransaction                 *mock.Call
	GetTransactionResultsByBlockID *mock.Call
	GetTransactionsByBlockID       *mock.Call
	GetSystemTransaction           *mock.Call
	GetSystemTransactionResult     *mock.Call
	GetLatestProtocolStateSnapshot *mock.Call
	Ping                           *mock.Call
	SecureConnection               *mock.Call
}

func DefaultMockGateway() *TestGateway {
	m := &Gateway{}
	ctxMock := context.Background()
	t := &TestGateway{
		Mock: m,
		SendSignedTransaction: m.On(
			SendSignedTransactionFunc,
			ctxMock,
			mock.AnythingOfType("*flow.Transaction"),
		),
		GetAccount: m.On(
			GetAccountFunc,
			ctxMock,
			mock.AnythingOfType("flow.Address"),
		),
		GetCollection: m.On(
			GetCollectionFunc,
			ctxMock,
			mock.AnythingOfType("flow.Identifier"),
		),
		GetTransactionResult: m.On(
			GetTransactionResultFunc,
			ctxMock,
			mock.AnythingOfType("flow.Identifier"),
			mock.AnythingOfType("bool"),
		),
		GetTransaction: m.On(
			GetTransactionFunc,
			ctxMock,
			mock.AnythingOfType("flow.Identifier"),
		),
		GetEvents: m.On(
			GetEventsFunc,
			ctxMock,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("uint64"),
			mock.AnythingOfType("uint64"),
		),
		ExecuteScript: m.On(
			ExecuteScriptFunc,
			ctxMock,
			mock.AnythingOfType("[]uint8"),
			mock.AnythingOfType("[]cadence.Value"),
		),
		GetBlockByHeight: m.On(GetBlockByHeightFunc, ctxMock, mock.Anything),
		GetBlockByID:     m.On(GetBlockByIDFunc, ctxMock, mock.Anything),
		GetLatestBlock:   m.On(GetLatestBlockFunc, ctxMock),
		GetSystemTransaction: m.On(
			GetSystemTransactionFunc,
			ctxMock,
			mock.AnythingOfType("flow.Identifier"),
		),
		GetSystemTransactionResult: m.On(
			GetSystemTransactionResultFunc,
			ctxMock,
			mock.AnythingOfType("flow.Identifier"),
		),
	}

	// default return values
	t.SendSignedTransaction.Run(func(args mock.Arguments) {
		t.SendSignedTransaction.Return(tests.NewTransaction(), nil)
	})

	t.GetAccount.Run(func(args mock.Arguments) {
		addr := args.Get(1).(flow.Address)
		t.GetAccount.Return(tests.NewAccountWithAddress(addr.String()), nil)
	})

	t.ExecuteScript.Run(func(args mock.Arguments) {
		t.ExecuteScript.Return(cadence.String(""), nil)
	})

	t.GetTransaction.Return(tests.NewTransaction(), nil)
	t.GetCollection.Return(tests.NewCollection(), nil)
	t.GetTransactionResult.Return(tests.NewTransactionResult(nil), nil)
	t.GetEvents.Return([]flow.BlockEvents{}, nil)
	t.GetLatestBlock.Return(tests.NewBlock(), nil)
	t.GetBlockByHeight.Return(tests.NewBlock(), nil)
	t.GetBlockByID.Return(tests.NewBlock(), nil)
	t.GetSystemTransaction.Return(tests.NewTransaction(), nil)
	t.GetSystemTransactionResult.Return(tests.NewTransactionResult(nil), nil)

	return t
}
