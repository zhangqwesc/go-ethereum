// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package rawdb

import (
	"encoding/binary"
	"math/big"

	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

// ReadTxLookupEntry retrieves the positional metadata associated with a transaction
// hash to allow retrieving the transaction or receipt by hash.
func ReadTxLookupEntry(db DatabaseReader, hash common.Hash) (common.Hash, uint64, uint64) {
	data, _ := db.Get(append(txLookupPrefix, hash.Bytes()...))
	if len(data) == 0 {
		return common.Hash{}, 0, 0
	}
	var entry TxLookupEntry
	if err := rlp.DecodeBytes(data, &entry); err != nil {
		log.Error("Invalid transaction lookup entry RLP", "hash", hash, "err", err)
		return common.Hash{}, 0, 0
	}
	return entry.BlockHash, entry.BlockIndex, entry.Index
}

// WriteTxLookupEntries stores a positional metadata for every transaction from
// a block, enabling hash based transaction and receipt lookups.
func WriteTxLookupEntries(db DatabaseWriter, block *types.Block) {
	for i, tx := range block.Transactions() {
		entry := TxLookupEntry{
			BlockHash:  block.Hash(),
			BlockIndex: block.NumberU64(),
			Index:      uint64(i),
		}
		data, err := rlp.EncodeToBytes(entry)
		if err != nil {
			log.Crit("Failed to encode transaction lookup entry", "err", err)
		}
		if err := db.Put(append(txLookupPrefix, tx.Hash().Bytes()...), data); err != nil {
			log.Crit("Failed to store transaction lookup entry", "err", err)
		}
	}
}

// DeleteTxLookupEntry removes all transaction data associated with a hash.
func DeleteTxLookupEntry(db DatabaseDeleter, hash common.Hash) {
	db.Delete(append(txLookupPrefix, hash.Bytes()...))
}

func encodeAddrTxsKey(address common.Address, timestamp *big.Int, hash common.Hash, direction byte) []byte {
	data := make([]byte, 0, 65) //addrTxsPrefix(4) + address(20) + timestamp(8) + hash(32) + direction(1) = 65
	data = append(data, addrTxsPrefix...)
	data = append(data, address.Bytes()...)
	data = append(data, timestamp.Bytes()...)
	data = append(data, hash.Bytes()...)
	data = append(data, direction)
	return data
}

func decodeAddrTxsKey(data []byte) (address common.Address, timestamp *big.Int, hash common.Hash, direction byte) {
	address.SetBytes(data[4:24])
	timestamp.SetBytes(data[24:32])
	hash.SetBytes(data[32:64])
	direction = data[64]
	return
}

// ReadAddrTxs return all transactions that address send or receive
func ReadAddrTxs(db DatabaseReader, address common.Address) {

}

// WriteAddrTxs stores all address's transations
func WriteAddrTxs(config *params.ChainConfig, db DatabaseWriter, block *types.Block, receipts types.Receipts) {
	time := block.Time()
	blockHash := block.Hash()
	blockNumber := block.Number()
	signer := types.MakeSigner(config, block.Number())
	for i, tx := range block.Transactions() {
		receipt := receipts[i]
		from, _ := types.Sender(signer, tx)
		to := tx.To()
		value := tx.Value()
		gasPrice := tx.GasPrice()
		gasUsed := receipt.GasUsed
		hash := tx.Hash()
		contractAddr := receipt.ContractAddress
		status := receipt.Status

		entry := AddrTxEntry{from, *to, value, gasPrice, gasUsed, blockHash, blockNumber, contractAddr, status}

		putValue, err := rlp.EncodeToBytes(entry)
		if err != nil {
			log.Crit("Failed to encode AddrTxEntry")
		}

		//indexes from address
		if err := db.Put(encodeAddrTxsKey(from, time, hash, byte(1)), putValue); err != nil {
			log.Crit("Failed to store AddrTxEntry for from")
		}
		//indexes to address
		if err := db.Put(encodeAddrTxsKey(*to, time, hash, byte(0)), putValue); err != nil {
			log.Crit("Failed to store AddrTxEntry for to")
		}
	}
	log.Warn("WriteAddrTxs success", "BlockNumber", block.NumberU64())
}

// DeleteAddrTxs removes all transaction
func DeleteAddrTxs(config *params.ChainConfig, db DatabaseDeleter, block *types.Block) {
	// time := block.Time()
	// signer := types.MakeSigner(config, block.Number())

}

// ReadTransaction retrieves a specific transaction from the database, along with
// its added positional metadata.
func ReadTransaction(db DatabaseReader, hash common.Hash) (*types.Transaction, common.Hash, uint64, uint64) {
	blockHash, blockNumber, txIndex := ReadTxLookupEntry(db, hash)
	if blockHash == (common.Hash{}) {
		return nil, common.Hash{}, 0, 0
	}
	body := ReadBody(db, blockHash, blockNumber)
	if body == nil || len(body.Transactions) <= int(txIndex) {
		log.Error("Transaction referenced missing", "number", blockNumber, "hash", blockHash, "index", txIndex)
		return nil, common.Hash{}, 0, 0
	}
	return body.Transactions[txIndex], blockHash, blockNumber, txIndex
}

// ReadReceipt retrieves a specific transaction receipt from the database, along with
// its added positional metadata.
func ReadReceipt(db DatabaseReader, hash common.Hash) (*types.Receipt, common.Hash, uint64, uint64) {
	blockHash, blockNumber, receiptIndex := ReadTxLookupEntry(db, hash)
	if blockHash == (common.Hash{}) {
		return nil, common.Hash{}, 0, 0
	}
	receipts := ReadReceipts(db, blockHash, blockNumber)
	if len(receipts) <= int(receiptIndex) {
		log.Error("Receipt refereced missing", "number", blockNumber, "hash", blockHash, "index", receiptIndex)
		return nil, common.Hash{}, 0, 0
	}
	return receipts[receiptIndex], blockHash, blockNumber, receiptIndex
}

// ReadBloomBits retrieves the compressed bloom bit vector belonging to the given
// section and bit index from the.
func ReadBloomBits(db DatabaseReader, bit uint, section uint64, head common.Hash) ([]byte, error) {
	key := append(append(bloomBitsPrefix, make([]byte, 10)...), head.Bytes()...)

	binary.BigEndian.PutUint16(key[1:], uint16(bit))
	binary.BigEndian.PutUint64(key[3:], section)

	return db.Get(key)
}

// WriteBloomBits stores the compressed bloom bits vector belonging to the given
// section and bit index.
func WriteBloomBits(db DatabaseWriter, bit uint, section uint64, head common.Hash, bits []byte) {
	key := append(append(bloomBitsPrefix, make([]byte, 10)...), head.Bytes()...)

	binary.BigEndian.PutUint16(key[1:], uint16(bit))
	binary.BigEndian.PutUint64(key[3:], section)

	if err := db.Put(key, bits); err != nil {
		log.Crit("Failed to store bloom bits", "err", err)
	}
}
