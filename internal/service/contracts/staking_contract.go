package contracts

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type StakingContract struct {
	// 事件签名映射表
	EventSignatures map[common.Hash]string
	// 用于快速查找的反向映射
	EventNames map[string]common.Hash
}

func NewStakingContract() *StakingContract {
	sc := &StakingContract{
		EventSignatures: make(map[common.Hash]string),
		EventNames:      make(map[string]common.Hash),
	}

	// 注册所有事件签名
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
		"AddPool":              "AddPool(address,uint256,uint256,uint256,uint256)",
		"UpdatePoolInfo":       "UpdatePoolInfo(uint256,uint256,uint256)",
		"SetPoolWeight":        "SetPoolWeight(uint256,uint256,uint256)",
		"UpdatePool":           "UpdatePool(uint256,uint256,uint256)",
	}

	for eventName, signature := range events {
		hash := crypto.Keccak256Hash([]byte(signature))
		sc.EventSignatures[hash] = eventName
		sc.EventNames[eventName] = hash
	}

	return sc
}

// IsTrackedEvent 检查是否是我们关注的事件
func (sc *StakingContract) IsTrackedEvent(eventHash common.Hash) bool {
	_, exists := sc.EventSignatures[eventHash]
	return exists
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

// IsUserInteractionEvent 检查是否是需要存储到数据库的用户交互事件
func (sc *StakingContract) IsUserInteractionEvent(eventName string) bool {
	switch eventName {
	case "Deposit", "RequestUnstake", "Claim", "Withdraw", "AddPool":
		return true
	default:
		return false
	}
}