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
	"fmt"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/flow-emulator/adapters"
	"github.com/onflow/flow-emulator/emulator"
	"github.com/pkg/errors"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/status"
)

type EmulatorKey struct {
	PublicKey crypto.PublicKey
	SigAlgo   crypto.SignatureAlgorithm
	HashAlgo  crypto.HashAlgorithm
}

type EmulatorGateway struct {
	emulator        *emulator.Blockchain
	adapter         *adapters.SDKAdapter
	accessAdapter   *adapters.AccessAdapter
	logger          *zerolog.Logger
	emulatorOptions []emulator.Option
}

func UnwrapStatusError(err error) error {
	return errors.New(status.Convert(err).Message())
}

func NewEmulatorGateway(key *EmulatorKey) *EmulatorGateway {
	return NewEmulatorGatewayWithOpts(key)
}

func NewEmulatorGatewayWithOpts(key *EmulatorKey, opts ...func(*EmulatorGateway)) *EmulatorGateway {
	noopLogger := zerolog.Nop()
	gateway := &EmulatorGateway{
		logger:          &noopLogger,
		emulatorOptions: []emulator.Option{},
	}
	for _, opt := range opts {
		opt(gateway)
	}

	gateway.emulator = newEmulator(key, gateway.emulatorOptions...)
	gateway.adapter = adapters.NewSDKAdapter(gateway.logger, gateway.emulator)
	gateway.accessAdapter = adapters.NewAccessAdapter(gateway.logger, gateway.emulator)
	gateway.emulator.EnableAutoMine()
	return gateway
}

func WithLogger(logger *zerolog.Logger) func(g *EmulatorGateway) {
	return func(g *EmulatorGateway) {
		g.logger = logger
	}
}

func WithEmulatorOptions(options ...emulator.Option) func(g *EmulatorGateway) {
	return func(g *EmulatorGateway) {
		g.emulatorOptions = append(g.emulatorOptions, options...)
	}
}

func newEmulator(key *EmulatorKey, emulatorOptions ...emulator.Option) *emulator.Blockchain {
	var opts []emulator.Option

	if key != nil {
		opts = append(opts, emulator.WithServicePublicKey(key.PublicKey, key.SigAlgo, key.HashAlgo))
	}

	opts = append(opts, emulatorOptions...)

	b, err := emulator.New(opts...)
	if err != nil {
		panic(err)
	}

	return b
}

func (g *EmulatorGateway) GetAccount(ctx context.Context, address flow.Address) (*flow.Account, error) {
	account, err := g.adapter.GetAccount(ctx, address)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return account, nil
}

func (g *EmulatorGateway) SendSignedTransaction(ctx context.Context, tx *flow.Transaction) (*flow.Transaction, error) {
	err := g.adapter.SendTransaction(ctx, *tx)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return tx, nil
}

func (g *EmulatorGateway) GetTransactionResult(ctx context.Context, ID flow.Identifier, _ bool) (*flow.TransactionResult, error) {
	result, err := g.adapter.GetTransactionResult(ctx, ID)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return result, nil
}

func (g *EmulatorGateway) GetTransaction(ctx context.Context, id flow.Identifier) (*flow.Transaction, error) {
	transaction, err := g.adapter.GetTransaction(ctx, id)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return transaction, nil
}

func (g *EmulatorGateway) GetTransactionResultsByBlockID(ctx context.Context, id flow.Identifier) ([]*flow.TransactionResult, error) {
	txr, err := g.adapter.GetTransactionResultsByBlockID(ctx, id)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return txr, nil
}

func (g *EmulatorGateway) GetTransactionsByBlockID(ctx context.Context, id flow.Identifier) ([]*flow.Transaction, error) {
	txr, err := g.adapter.GetTransactionsByBlockID(ctx, id)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return txr, nil
}

func (g *EmulatorGateway) GetSystemTransaction(ctx context.Context, blockID flow.Identifier) (*flow.Transaction, error) {
	return nil, errors.New("GetSystemTransaction is not implemented")
}

func (g *EmulatorGateway) GetSystemTransactionResult(ctx context.Context, blockID flow.Identifier) (*flow.TransactionResult, error) {
	return nil, errors.New("GetSystemTransactionResult is not implemented")
}

func (g *EmulatorGateway) Ping() error {
	ctx := context.Background()
	err := g.adapter.Ping(ctx)
	if err != nil {
		return UnwrapStatusError(err)
	}
	return nil
}

func (g *EmulatorGateway) WaitServer(ctx context.Context) error {
	return nil
}

type scriptQuery struct {
	id     flow.Identifier
	height uint64
	latest bool
}

func (g *EmulatorGateway) executeScriptQuery(
	ctx context.Context,
	script []byte,
	arguments []cadence.Value,
	query scriptQuery,
) (cadence.Value, error) {
	args, err := cadenceValuesToMessages(arguments)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	var result []byte
	if query.id != flow.EmptyID {
		result, err = g.adapter.ExecuteScriptAtBlockID(ctx, query.id, script, args)
	} else if query.height > 0 {
		result, err = g.adapter.ExecuteScriptAtBlockHeight(ctx, query.height, script, args)
	} else {
		result, err = g.adapter.ExecuteScriptAtLatestBlock(ctx, script, args)
	}

	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	value, err := messageToCadenceValue(result)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	return value, nil
}

func (g *EmulatorGateway) ExecuteScript(
	ctx context.Context,
	script []byte,
	arguments []cadence.Value,
) (cadence.Value, error) {
	return g.executeScriptQuery(ctx, script, arguments, scriptQuery{latest: true})
}

func (g *EmulatorGateway) ExecuteScriptAtHeight(
	ctx context.Context,
	script []byte,
	arguments []cadence.Value,
	height uint64,
) (cadence.Value, error) {
	return g.executeScriptQuery(ctx, script, arguments, scriptQuery{height: height})
}

func (g *EmulatorGateway) ExecuteScriptAtID(
	ctx context.Context,
	script []byte,
	arguments []cadence.Value,
	id flow.Identifier,
) (cadence.Value, error) {
	return g.executeScriptQuery(ctx, script, arguments, scriptQuery{id: id})
}

func (g *EmulatorGateway) GetLatestBlock(ctx context.Context, isSealed bool) (*flow.Block, error) {
	block, _, err := g.adapter.GetLatestBlock(ctx, isSealed)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	return block, nil
}

func cadenceValuesToMessages(values []cadence.Value) ([][]byte, error) {
	msgs := make([][]byte, len(values))
	for i, val := range values {
		msg, err := jsoncdc.Encode(val)
		if err != nil {
			return nil, fmt.Errorf("convert: %w", err)
		}
		msgs[i] = msg
	}
	return msgs, nil
}

func messageToCadenceValue(m []byte) (cadence.Value, error) {
	v, err := jsoncdc.Decode(nil, m)
	if err != nil {
		return nil, fmt.Errorf("convert: %w", err)
	}

	return v, nil
}

func (g *EmulatorGateway) GetEvents(
	ctx context.Context,
	eventType string,
	startHeight uint64,
	endHeight uint64,
) ([]flow.BlockEvents, error) {
	events := make([]flow.BlockEvents, 0)

	for height := startHeight; height <= endHeight; height++ {
		events = append(events, g.getBlockEvent(ctx, height, eventType))
	}

	return events, nil
}

func (g *EmulatorGateway) getBlockEvent(ctx context.Context, height uint64, eventType string) flow.BlockEvents {
	events, _ := g.adapter.GetEventsForHeightRange(ctx, eventType, height, height)
	return *events[0]
}

func (g *EmulatorGateway) GetCollection(ctx context.Context, id flow.Identifier) (*flow.Collection, error) {
	collection, err := g.adapter.GetCollectionByID(ctx, id)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return collection, nil
}

func (g *EmulatorGateway) GetBlockByID(ctx context.Context, id flow.Identifier) (*flow.Block, error) {
	block, _, err := g.adapter.GetBlockByID(ctx, id)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return block, nil
}

func (g *EmulatorGateway) GetBlockByHeight(ctx context.Context, height uint64) (*flow.Block, error) {
	block, _, err := g.adapter.GetBlockByHeight(ctx, height)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return block, nil
}

func (g *EmulatorGateway) GetLatestProtocolStateSnapshot(ctx context.Context) ([]byte, error) {
	snapshot, err := g.adapter.GetLatestProtocolStateSnapshot(ctx)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return snapshot, nil
}

// SecureConnection placeholder func to complete gateway interface implementation
func (g *EmulatorGateway) SecureConnection() bool {
	return false
}

func (g *EmulatorGateway) CoverageReport() *runtime.CoverageReport {
	return g.emulator.CoverageReport()
}

func (g *EmulatorGateway) RollbackToBlockHeight(height uint64) error {
	return g.emulator.RollbackToBlockHeight(height)
}
