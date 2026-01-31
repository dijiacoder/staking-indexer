package scanner

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const contractABI = `[
  {
    "anonymous": false,
    "inputs": [
      { "indexed": true, "name": "user", "type": "address" },
      { "indexed": true, "name": "poolId", "type": "uint256" },
      { "indexed": false, "name": "amount", "type": "uint256" }
    ],
    "name": "Deposit",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      { "indexed": true, "name": "user", "type": "address" },
      { "indexed": true, "name": "poolId", "type": "uint256" },
      { "indexed": false, "name": "amount", "type": "uint256" }
    ],
    "name": "RequestUnstake",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      { "indexed": true, "name": "user", "type": "address" },
      { "indexed": true, "name": "poolId", "type": "uint256" },
      { "indexed": false, "name": "amount", "type": "uint256" },
      { "indexed": true, "name": "blockNumber", "type": "uint256" }
    ],
    "name": "Withdraw",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      { "indexed": true, "name": "user", "type": "address" },
      { "indexed": true, "name": "poolId", "type": "uint256" },
      { "indexed": false, "name": "MetaNodeReward", "type": "uint256" }
    ],
    "name": "Claim",
    "type": "event"
  }
]`

type DecodedEvent struct {
	Name        string
	User        common.Address
	PoolID      *big.Int
	Amount      *big.Int
	Reward      *big.Int
	BlockNumber *big.Int
}

type EventDecoder struct {
	abi abi.ABI
}

func NewEventDecoder() (*EventDecoder, error) {
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, err
	}
	return &EventDecoder{abi: parsedABI}, nil
}

func (d *EventDecoder) DecodeLog(log types.Log) (*DecodedEvent, error) {
	if len(log.Topics) == 0 {
		return nil, fmt.Errorf("no topics in log")
	}

	event, err := d.abi.EventByID(log.Topics[0])
	if err != nil {
		return nil, err
	}

	decoded := &DecodedEvent{Name: event.Name}
	
	// Decode indexed parameters
	indexed := make(map[string]interface{})
	var indexedArgs abi.Arguments
	for _, arg := range event.Inputs {
		if arg.Indexed {
			indexedArgs = append(indexedArgs, arg)
		}
	}
	if err := abi.ParseTopicsIntoMap(indexed, indexedArgs, log.Topics[1:]); err != nil {
		return nil, err
	}
	
	// Map indexed fields
	if val, ok := indexed["user"]; ok {
		decoded.User = val.(common.Address)
	}
	if val, ok := indexed["poolId"]; ok {
		decoded.PoolID = val.(*big.Int)
	}
	if val, ok := indexed["blockNumber"]; ok {
		decoded.BlockNumber = val.(*big.Int)
	}

	// Decode non-indexed parameters
	if len(log.Data) > 0 {
		nonIndexed := make(map[string]interface{})
		var nonIndexedArgs abi.Arguments
		for _, arg := range event.Inputs {
			if !arg.Indexed {
				nonIndexedArgs = append(nonIndexedArgs, arg)
			}
		}
		if err := nonIndexedArgs.UnpackIntoMap(nonIndexed, log.Data); err != nil {
			return nil, err
		}
		
		if val, ok := nonIndexed["amount"]; ok {
			decoded.Amount = val.(*big.Int)
		}
		if val, ok := nonIndexed["MetaNodeReward"]; ok {
			decoded.Reward = val.(*big.Int)
		}
	}

	return decoded, nil
}
