参考了github上的开源项目，但是还是没有写完。

增加了数据库leveldb,VM,状态树，交易池,P2P通讯。

未最终写完，会继续调试，把课上的一些公链优化合并进来。

 


# 共识
共识模块设计成可插拔的模式。 可自行替换共识算法，以适应不同需求的场景。

共识算法有以下几种选择——

## 1. Solo
Solo是一个最小可用的共识引擎。在该机制下，全网只有一个出块节点，其余节点 默认相信该出块节点是诚实的 ，接收它所提案的区块并做验证。

在该共识机制下，区块链系统退化为中心化的分布式账本系统。 

Solo采用惰性出块的方式。当出块节点的交易池接收到的交易达到某一预设上限，或者新交易驻留时间超过某一阈值时，则会提案新的区块。

其他节点在接收到区块后，执行如下流程：
1. 验证区块头是否合法；若合法，执行下一流程，同时广播给其他节点。
2. 执行区块中的交易，记录下交易收据`receipts`和当前状态`state_root`。
3. 比对交易数据的Hash和状态root是否与区块头所提供的数据相符。若相符，则提交区块到数据库。

## 2. VRF + BFT
VRF(Verifiable Random Function)

未完成

## 3. POW

一个常规的工作量证明共识算法 
### 出块间隔
全网大约每30秒生成一个新区块。

### 难度调整
工作量证明的算法实际为暴力计算难题。基于公式`Hash(data,nonce) < target`，挖矿节点遍历所有的nonce情况，找出一个满足该不等式的nonce，并为区块签名。

`target`的计算公式为`target = MAXIMUM_TARGET / difficulty`。`MAXIMUN_TARGET`为一个预设的最大上限(2^(256-32))；`diffculty`为难度，保存在区块的`consensus info`字段。

每经过20160个区块（大约一周），进行一次难度调整。难度调整公式为`new_target = curr_target * 实际20160个块的出块时间 / 理论20160个块的出块时间（1周）`


#  交易

## 1 .通过jsonrpc接口来模拟交易
   http://ip:8081/

 handlers.GetBlockHandler 

 handlers.GetBlockHashHandler 

 handlers.GetHeaderHandler 
		
 handlers.GetTxHandler 

 handlers.GetReceiptHandler 

 handlers.SendTxHandler 