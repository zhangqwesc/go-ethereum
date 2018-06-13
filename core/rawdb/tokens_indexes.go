package rawdb

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	logger "github.com/ethereum/go-ethereum/log"
)

//TransferFilter is topic[0] -> event Transfer(address,address,uint256)
var TransferFilter = common.HexToHash("ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

func encodeTokenKey(address common.Hash, tokenAddress common.Address, time []byte, hash common.Hash, direction byte) []byte {
	data := make([]byte, 0, 82) //prefix(1) + 20 + 20 + 8 + 32 + 1 = 82
	data = append(data, erc20TokenPrefix...)
	data = append(data, address.Bytes()[12:]...)
	data = append(data, tokenAddress.Bytes()...)
	data = append(data, time...)
	data = append(data, hash.Bytes()...)
	data = append(data, direction)
	return data
}

func decodeTokenKey(data []byte) (address, tokenAddress common.Address, time uint64, hash common.Hash, direction byte) {
	if len(data) != 82 {
		logger.Crit("TokenKey length isn't 82")
	}
	address.SetBytes(data[1:21])
	tokenAddress.SetBytes(data[21:41])
	time = binary.BigEndian.Uint64(data[41:49])
	hash.SetBytes(data[49:81])
	direction = data[81]
	return
}

func encodeTokenValue(address common.Hash, value []byte) []byte {
	data := make([]byte, 0, 20+len(value)) //20 + len(value)
	data = append(data, address.Bytes()[12:]...)
	data = append(data, value...)
	return data
}

func decodeTokenValue(data []byte) (address common.Address, value big.Int) {
	address.SetBytes(data[:20])
	value.SetBytes(data[20:])
	return
}

func encodeOwnedKey(address common.Hash, tokenAddress common.Address) []byte {
	data := make([]byte, 0, 41) //prefix(1) + 20 + 20
	data = append(data, erc20OwnerPrefix...)
	data = append(data, address.Bytes()[12:]...)
	data = append(data, tokenAddress.Bytes()...)
	return data
}

func decodeOwnedKey(data []byte) (address, tokenAddress common.Address) {
	address.SetBytes(data[1:21])
	tokenAddress.SetBytes(data[21:41])
	return
}

//WriteTokenTransfer stores all token transfer, filter by event 'Transfer(address, address, uint256)'
func WriteTokenTransfer(db DatabaseWriter, hcDb ethdb.Database, logs []*types.Log) {
	cache := make(map[common.Hash][]byte) //cache the timestamp of the block to avoid checking the database every time
	for _, log := range logs {
		if log.Topics[0] != TransferFilter {
			continue
		}
		time, ok := cache[log.BlockHash]
		if !ok {
			ts := ReadHeader(hcDb, log.BlockHash, log.BlockNumber).Time.Uint64()
			time = make([]byte, 8)
			binary.BigEndian.PutUint64(time, ^ts)
			cache[log.BlockHash] = time
		}
		if err := db.Put(encodeTokenKey(log.Topics[1], log.Address, time, log.TxHash, 1), encodeTokenValue(log.Topics[2], log.Data)); err != nil {
			logger.Crit("Failed to store token transfer for from")
		}
		if err := db.Put(encodeTokenKey(log.Topics[2], log.Address, time, log.TxHash, 0), encodeTokenValue(log.Topics[1], log.Data)); err != nil {
			logger.Crit("Failed to store token transfer for to")
		}
		if err := db.Put(encodeOwnedKey(log.Topics[1], log.Address), nil); err != nil {
			logger.Crit("Failed to store token owned for from")
		}
		if err := db.Put(encodeOwnedKey(log.Topics[2], log.Address), nil); err != nil {
			logger.Crit("Failed to store token owned for to")
		}
	}
}

//ReadTokenTransfer read token transfer information
func ReadTokenTransfer(ldb *ethdb.LDBDatabase, address, tokenAddress *common.Address, start, end int) (list []RPCTokenTransferEntry, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("ReadTokenTransfer get a fatal error: %v", e)
			return
		}
	}()

	if end >= 0 && start >= end {
		list = []RPCTokenTransferEntry{}
		return
	}
	if start < 0 {
		list = []RPCTokenTransferEntry{}
		return
	}

	prefix := make([]byte, 0, 41) //prefix(1) + 20 + 20
	prefix = append(prefix, erc20TokenPrefix...)
	prefix = append(prefix, address.Bytes()...)
	prefix = append(prefix, tokenAddress.Bytes()...)

	it := ldb.NewIteratorWithPrefix(prefix)

	for it.Next() {
		addr, _, time, hash, direction := decodeTokenKey(it.Key())
		addr2, value := decodeTokenValue(it.Value())
		if direction > 0 {
			list = append(list, RPCTokenTransferEntry{addr, addr2, &value, hash, time})
		} else {
			list = append(list, RPCTokenTransferEntry{addr2, addr, &value, hash, time})
		}
	}

	if len(list) == 0 {
		list = []RPCTokenTransferEntry{}
		return
	}

	if start >= len(list) {
		list = []RPCTokenTransferEntry{}
		return
	}
	if end > len(list) || end < 0 {
		end = len(list)
	}
	list = list[start:end]
	return
}

//DeleteTokenTransfer del token transfer information
func DeleteTokenTransfer(db DatabaseDeleter, logs []*types.Log, time uint64) {
	timeBytes := make([]byte, 0, 8)
	binary.BigEndian.PutUint64(timeBytes, ^time)
	for _, log := range logs {
		if err := db.Delete(encodeTokenKey(log.Topics[1], log.Address, timeBytes, log.TxHash, 1)); err != nil {
			logger.Crit("Failed to delete token transfer for from")
		}
		if err := db.Delete(encodeTokenKey(log.Topics[2], log.Address, timeBytes, log.TxHash, 0)); err != nil {
			logger.Crit("Failed to delete token transfer for from")
		}
	}
}
