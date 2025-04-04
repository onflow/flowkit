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

package gateway

import (
	"context"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

//go:generate  mockery --name=Gateway

// Gateway describes blockchain access interface
type Gateway interface {
	GetAccount(context.Context, flow.Address) (*flow.Account, error)
	SendSignedTransaction(context.Context, *flow.Transaction) (*flow.Transaction, error)
	GetTransaction(context.Context, flow.Identifier) (*flow.Transaction, error)
	GetTransactionResultsByBlockID(ctx context.Context, blockID flow.Identifier) ([]*flow.TransactionResult, error)
	GetTransactionResult(context.Context, flow.Identifier, bool) (*flow.TransactionResult, error)
	GetTransactionsByBlockID(context.Context, flow.Identifier) ([]*flow.Transaction, error)
	GetSystemTransaction(ctx context.Context, blockID flow.Identifier) (*flow.Transaction, error)
	GetSystemTransactionResult(ctx context.Context, blockID flow.Identifier) (*flow.TransactionResult, error)
	ExecuteScript(context.Context, []byte, []cadence.Value) (cadence.Value, error)
	ExecuteScriptAtHeight(context.Context, []byte, []cadence.Value, uint64) (cadence.Value, error)
	ExecuteScriptAtID(context.Context, []byte, []cadence.Value, flow.Identifier) (cadence.Value, error)
	GetLatestBlock(context.Context) (*flow.Block, error)
	GetBlockByHeight(context.Context, uint64) (*flow.Block, error)
	GetBlockByID(context.Context, flow.Identifier) (*flow.Block, error)
	GetEvents(context.Context, string, uint64, uint64) ([]flow.BlockEvents, error)
	GetCollection(context.Context, flow.Identifier) (*flow.Collection, error)
	GetLatestProtocolStateSnapshot(context.Context) ([]byte, error)
	Ping() error
	WaitServer(context.Context) error
	SecureConnection() bool
}
