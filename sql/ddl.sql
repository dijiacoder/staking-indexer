-- 基于 MySQL 8.x

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ================================
-- 1. 链上合约扫描游标表
-- ================================
CREATE TABLE chain_scan_cursor (
       id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
       chain_id BIGINT NOT NULL COMMENT '链ID（以太坊:1, BSC:56, Polygon:137）',
       contract_address VARCHAR(42) NOT NULL COMMENT '合约地址',
       contract_name VARCHAR(64) NOT NULL COMMENT '合约名称',
       last_scanned_block BIGINT NOT NULL COMMENT '最近已扫描区块（可能未确认）',
       last_confirmed_block BIGINT NOT NULL COMMENT '最近已确认区块高度',
       confirmation_blocks INT NOT NULL DEFAULT 12 COMMENT '确认区块数',
       scan_status TINYINT NOT NULL DEFAULT 1 COMMENT '扫描状态：1-正常 2-回滚中 3-暂停',
       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
       updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
       UNIQUE KEY uk_chain_contract (chain_id, contract_address),
       KEY idx_chain_scan (chain_id, last_scanned_block)
) ENGINE=InnoDB COMMENT='链上合约扫描游标';

-- ================================
-- 2. 区块头表（Reorg检测用）
-- ================================
CREATE TABLE chain_blocks (
      id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
      chain_id BIGINT NOT NULL COMMENT '链ID',
      block_number BIGINT NOT NULL COMMENT '区块高度',
      block_hash VARCHAR(66) NOT NULL COMMENT '区块Hash',
      parent_hash VARCHAR(66) NOT NULL COMMENT '父区块Hash',
      is_confirmed TINYINT NOT NULL DEFAULT 0 COMMENT '是否已确认',
      created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
      UNIQUE KEY uk_chain_block (chain_id, block_number),
      KEY idx_block_hash (block_hash)
) ENGINE=InnoDB COMMENT='区块头缓存表';

-- ================================
-- 3. Staking Pool 定义表
-- （来自 MetaNodeStake.sol）
-- ================================
CREATE TABLE staking_pools (
       id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
       chain_id BIGINT NOT NULL COMMENT '链ID',
       contract_address VARCHAR(42) NOT NULL COMMENT 'Staking合约地址',
       pool_id BIGINT NOT NULL COMMENT 'Pool ID（合约内定义）',
       stake_token VARCHAR(42) NOT NULL COMMENT '质押Token地址（ETH为0x0）',
       reward_token VARCHAR(42) NOT NULL COMMENT '奖励Token地址',
       start_block BIGINT NOT NULL COMMENT '开始区块',
       end_block BIGINT NOT NULL COMMENT '结束区块',
       reward_per_block DECIMAL(38,0) NOT NULL COMMENT '每区块奖励',
       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
       UNIQUE KEY uk_pool (chain_id, contract_address, pool_id)
) ENGINE=InnoDB COMMENT='Staking池定义表';

-- ================================
-- 4. 用户质押聚合状态表
-- ================================
CREATE TABLE staking_user_positions (
        id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
        chain_id BIGINT NOT NULL COMMENT '链ID',
        contract_address VARCHAR(42) NOT NULL COMMENT '合约地址',
        pool_id BIGINT NOT NULL COMMENT 'Pool ID',
        user_address VARCHAR(42) NOT NULL COMMENT '用户地址',
        staked_amount DECIMAL(38,0) NOT NULL DEFAULT 0 COMMENT '当前质押数量',
        reward_debt DECIMAL(38,0) NOT NULL DEFAULT 0 COMMENT '奖励债务',
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
        UNIQUE KEY uk_user_pool (chain_id, contract_address, pool_id, user_address)
) ENGINE=InnoDB COMMENT='用户质押实时状态';

-- ================================
-- 5. Staking 事件表（最终确认事件）
-- ================================
CREATE TABLE staking_events (
        id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
        chain_id BIGINT NOT NULL COMMENT '链ID',
        contract_address VARCHAR(42) NOT NULL COMMENT '合约地址',
        pool_id BIGINT NOT NULL COMMENT 'Pool ID',
        event_type VARCHAR(16) NOT NULL COMMENT '事件类型：Deposit / Withdraw / Claim',
        user_address VARCHAR(42) NOT NULL COMMENT '用户地址',
        amount DECIMAL(38,0) NOT NULL COMMENT '数量（wei）',
        block_number BIGINT NOT NULL COMMENT '区块高度',
        tx_hash VARCHAR(66) NOT NULL COMMENT '交易Hash',
        log_index INT NOT NULL COMMENT '日志索引',
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
        UNIQUE KEY uk_tx_log (tx_hash, log_index),
        KEY idx_user (chain_id, user_address),
        KEY idx_pool_block (chain_id, pool_id, block_number)
) ENGINE=InnoDB COMMENT='Staking事件表';

SET FOREIGN_KEY_CHECKS = 1;
