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

package accounts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flowkit/v2/tests"

	"github.com/onflow/flowkit/v2/config"
)

func Test_KMS_Keys(t *testing.T) {
	confKey := config.AccountKey{
		Type:       config.KeyTypeGoogleKMS,
		Index:      0,
		SigAlgo:    config.DefaultSigAlgo,
		HashAlgo:   config.DefaultHashAlgo,
		ResourceID: "projects/my-project/locations/global/keyRings/flow/cryptoKeys/my-account/cryptoKeyVersions/1",
	}

	kmsKey, err := kmsKeyFromConfig(confKey)
	assert.NoError(t, err)

	_, err = kmsKey.PrivateKey()
	assert.EqualError(t, err, "private key not accessible")
	assert.Equal(t, confKey, kmsKey.ToConfig())
}

func Test_File_key(t *testing.T) {
	confKey := config.AccountKey{
		Type:     config.KeyTypeFile,
		Index:    0,
		SigAlgo:  config.DefaultSigAlgo,
		HashAlgo: config.DefaultHashAlgo,
		Location: "./test.pkey",
	}

	fileKey, err := fileKeyFromConfig(confKey)
	assert.NoError(t, err)

	cKey := fileKey.ToConfig()
	assert.Equal(t, cKey, confKey)

	rw, _ := tests.ReaderWriter()
	key := NewFileKey(confKey.Location, confKey.Index, confKey.SigAlgo, confKey.HashAlgo, rw)
	assert.Equal(t, confKey, key.ToConfig())
}

func Test_BIP44(t *testing.T) {
	confKey := config.AccountKey{
		Type:           config.KeyTypeBip44,
		Index:          0,
		SigAlgo:        config.DefaultSigAlgo,
		HashAlgo:       config.DefaultHashAlgo,
		Mnemonic:       "version field tornado move level pretty inject stereo ten catalog salon swallow",
		DerivationPath: "m/44'/539'/0'/0/0",
	}

	key, err := bip44KeyFromConfig(confKey)
	assert.NoError(t, err)

	const pubKey = "0x2d6daea8b0ba5b1d5935f7846ccdd7e6f9f981e34d3c0a02a927cc79c837eba56c0f9a979195e41143495b72314ffcab60da6b7031060c80dc12f01f7f2096be"
	assert.Equal(t, confKey, key.ToConfig())
	pkey, err := key.PrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, pubKey, (*pkey).PublicKey().String())

	sig, err := key.Signer(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, pubKey, sig.PublicKey().String())
}

func Test_EnvKey(t *testing.T) {
	pk, err := crypto.DecodePrivateKeyHex(config.DefaultSigAlgo, "64cfa38591cf755e84379d78884e5322af0fd2a94cff48569d6578cdd733d455") // TEST KEY DO NOT USE
	assert.NoError(t, err)

	confKey := config.AccountKey{
		Type:       config.KeyTypeHex,
		Index:      0,
		SigAlgo:    config.DefaultSigAlgo,
		HashAlgo:   config.DefaultHashAlgo,
		PrivateKey: pk,
		Env:        "TEST",
	}

	key, err := envKeyFromConfig(confKey)
	assert.NoError(t, err)

	const pubKey = "0x1e585ddefde564eb9d86c606a2cf33996c9434a4f658d7338a7b811e337adf6e38e2ae4a5c7a79751b5bf8b08a90428d0a29aa27e6ddc195099ac1b2deb9519a"
	assert.Equal(t, confKey, key.ToConfig())
	pkey, err := key.PrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, pubKey, (*pkey).PublicKey().String())

	sig, err := key.Signer(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, pubKey, sig.PublicKey().String())
}
