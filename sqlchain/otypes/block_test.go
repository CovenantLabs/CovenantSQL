/*
 * Copyright 2018 The CovenantSQL Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package otypes

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"

	"github.com/CovenantSQL/CovenantSQL/crypto/asymmetric"
	"github.com/CovenantSQL/CovenantSQL/crypto/hash"
	"github.com/CovenantSQL/CovenantSQL/utils"
)

func TestSignAndVerify(t *testing.T) {
	block, err := createRandomBlock(genesisHash, true)

	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	if err = block.Verify(); err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	block.SignedHeader.BlockHash[0]++

	if err = block.Verify(); err != ErrHashVerification {
		t.Fatalf("unexpected error: %v", err)
	}

	h := &hash.Hash{}
	block.PushAckedQuery(h)

	if err = block.Verify(); err != ErrMerkleRootVerification {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeaderMarshalUnmarshaler(t *testing.T) {
	block, err := createRandomBlock(genesisHash, false)

	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	origin := &block.SignedHeader.Header
	enc, err := utils.EncodeMsgPack(origin)

	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	dec := &Header{}
	if err = utils.DecodeMsgPack(enc.Bytes(), dec); err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	bts1, err := origin.MarshalHash()
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	bts2, err := dec.MarshalHash()
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	if !bytes.Equal(bts1, bts2) {
		t.Fatal("hash not stable")
	}

	if !reflect.DeepEqual(origin, dec) {
		t.Fatalf("values don't match:\n\tv1 = %+v\n\tv2 = %+v", origin, dec)
	}
}

func TestSignedHeaderMarshaleUnmarshaler(t *testing.T) {
	block, err := createRandomBlock(genesisHash, true)

	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	origin := &block.SignedHeader
	enc, err := utils.EncodeMsgPack(origin)

	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	dec := &SignedHeader{}

	if err = utils.DecodeMsgPack(enc.Bytes(), dec); err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	bts1, err := origin.MarshalHash()
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	bts2, err := dec.MarshalHash()
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	if !bytes.Equal(bts1, bts2) {
		t.Fatal("hash not stable")
	}

	if !reflect.DeepEqual(origin.Header, dec.Header) {
		t.Fatalf("values don't match:\n\tv1 = %+v\n\tv2 = %+v", origin.Header, dec.Header)
	}

	if err = origin.Verify(); err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	if err = dec.Verify(); err != nil {
		t.Fatalf("error occurred: %v", err)
	}
}

func TestBlockMarshalUnmarshaler(t *testing.T) {
	origin, err := createRandomBlock(genesisHash, false)
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}
	origin2, err := createRandomBlock(genesisHash, false)
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	blocks := make(Blocks, 0, 2)
	blocks = append(blocks, origin)
	blocks = append(blocks, origin2)
	blocks = append(blocks, nil)

	blocks2 := make(Blocks, 0, 2)
	blocks2 = append(blocks2, origin)
	blocks2 = append(blocks2, origin2)
	blocks2 = append(blocks2, nil)

	bts1, err := blocks.MarshalHash()
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	bts2, err := blocks2.MarshalHash()
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	if !bytes.Equal(bts1, bts2) {
		t.Fatal("hash not stable")
	}

	enc, err := utils.EncodeMsgPack(origin)

	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	dec := &Block{}

	if err = utils.DecodeMsgPack(enc.Bytes(), dec); err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	bts1, err = origin.MarshalHash()
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	bts2, err = dec.MarshalHash()
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	if !bytes.Equal(bts1, bts2) {
		t.Fatal("hash not stable")
	}

	if !reflect.DeepEqual(origin, dec) {
		t.Fatalf("values don't match:\n\tv1 = %+v\n\tv2 = %+v", origin, dec)
	}
}

func TestGenesis(t *testing.T) {
	genesis, err := createRandomBlock(genesisHash, true)

	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	if err = genesis.VerifyAsGenesis(); err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	if err = genesis.SignedHeader.VerifyAsGenesis(); err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	// Test non-genesis block
	genesis, err = createRandomBlock(genesisHash, false)

	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	if err = genesis.VerifyAsGenesis(); err != nil {
		t.Logf("Error occurred as expected: %v", err)
	} else {
		t.Fatal("unexpected result: returned nil while expecting an error")
	}

	if err = genesis.SignedHeader.VerifyAsGenesis(); err != nil {
		t.Logf("Error occurred as expected: %v", err)
	} else {
		t.Fatal("unexpected result: returned nil while expecting an error")
	}

	// Test altered public key block
	genesis, err = createRandomBlock(genesisHash, true)

	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	_, pub, err := asymmetric.GenSecp256k1KeyPair()

	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	genesis.SignedHeader.Signee = pub

	if err = genesis.VerifyAsGenesis(); err != nil {
		t.Logf("Error occurred as expected: %v", err)
	} else {
		t.Fatal("unexpected result: returned nil while expecting an error")
	}

	if err = genesis.SignedHeader.VerifyAsGenesis(); err != nil {
		t.Logf("Error occurred as expected: %v", err)
	} else {
		t.Fatal("unexpected result: returned nil while expecting an error")
	}

	// Test altered signature
	genesis, err = createRandomBlock(genesisHash, true)

	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	genesis.SignedHeader.Signature.R.Add(genesis.SignedHeader.Signature.R, big.NewInt(int64(1)))
	genesis.SignedHeader.Signature.S.Add(genesis.SignedHeader.Signature.S, big.NewInt(int64(1)))

	if err = genesis.VerifyAsGenesis(); err != nil {
		t.Logf("Error occurred as expected: %v", err)
	} else {
		t.Fatalf("unexpected error: %v", err)
	}

	if err = genesis.SignedHeader.VerifyAsGenesis(); err != nil {
		t.Logf("Error occurred as expected: %v", err)
	} else {
		t.Fatal("unexpected result: returned nil while expecting an error")
	}
}
