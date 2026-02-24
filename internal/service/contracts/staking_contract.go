package contracts

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type StakingContract struct {
	EventSignatures map[common.Hash]string
	EventNames      map[string]common.Hash
	IgnoredEvents   map[string]bool
}

func NewStakingContract() *StakingContract {
	sc := &StakingContract{
		EventSignatures: make(map[common.Hash]string),
		EventNames:      make(map[string]common.Hash),
		IgnoredEvents:   make(map[string]bool),
	}

	// 注册“关注”的事件签名
	events := map[string]string{
		"Deposit":              "Deposit(address,uint256,uint256)",
		"RequestUnstake":       "RequestUnstake(address,uint256,uint256)",
		"Claim":                "Claim(address,uint256,uint256)",
		"Withdraw":             "Withdraw(address,uint256,uint256,uint256)",
		"SetZeroToken":         "SetZeroToken(address)",
		"PauseWithdraw":        "PauseWithdraw()",
		"UnpauseWithdraw":      "UnpauseWithdraw()",
		"PauseClaim":           "PauseClaim()",
		"UnpauseClaim":         "UnpauseClaim()",
		"SetStartBlock":        "SetStartBlock(uint256)",
		"SetEndBlock":          "SetEndBlock(uint256)",
		"SetZeroTokenPerBlock": "SetZeroTokenPerBlock(uint256)",
		"AddPool":              "AddPool(uint256,address,uint256,uint256,uint256,uint256)",
		"UpdatePoolInfo":       "UpdatePoolInfo(uint256,uint256,uint256)",
		"SetPoolWeight":        "SetPoolWeight(uint256,uint256,uint256)",
		"UpdatePool":           "UpdatePool(uint256,uint256,uint256)",
	}

	for eventName, signature := range events {
		hash := crypto.Keccak256Hash([]byte(signature))
		sc.EventSignatures[hash] = eventName
		sc.EventNames[eventName] = hash
	}

	ignoredEvents := []string{
		"SetZeroToken",
	}
	for _, eventName := range ignoredEvents {
		sc.IgnoredEvents[eventName] = true
	}

	return sc
}

// GetEventName 根据哈希获取事件名称
func (sc *StakingContract) GetEventName(eventHash common.Hash) (string, bool) {
	name, exists := sc.EventSignatures[eventHash]
	return name, exists
}

// GetEventSignature 获取事件签名哈希
func (sc *StakingContract) GetEventSignature(eventName string) (common.Hash, bool) {
	hash, exists := sc.EventNames[eventName]
	return hash, exists
}

// IsIgnoredEvent 检查是否是不关注的事件
func (sc *StakingContract) IsIgnoredEvent(eventName string) bool {
	return sc.IgnoredEvents[eventName]
}
