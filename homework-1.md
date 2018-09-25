一、 环签名时间，基本为O(N)线性增长  
ring size: 1, cost 0.006836  
ring size: 2, gen:0.019718 sign: 0.024721 verify: 0.018060  
ring size: 2, cost 0.062518   
ring size: 4, gen:0.032470 sign: 0.053889 verify: 0.037212  
ring size: 4, cost 0.123580   
ring size: 8, gen:0.045644 sign: 0.065100 verify: 0.065768  
ring size: 8, cost 0.176520   
ring size: 16, gen:0.072690 sign: 0.137627 verify: 0.138022  
ring size: 16, cost 0.348347   
ring size: 32, gen:0.149773 sign: 0.258290 verify: 0.268936  
ring size: 32, cost 0.677007   
ring size: 64, gen:0.338758 sign: 0.549643 verify: 0.577867  
ring size: 64, cost 1.466276   
ring size: 128, gen:0.613416 sign: 1.043730 verify: 1.217091  
ring size: 128, cost 2.874246   



二、 各币种区块大小

名称|交易相关属性 |块大小 |交易数量 
----|----|----|---
bitcoin | in,out,锁定时间 |1M|3000
ethereum|TxHash,TxReceipt Status,Block Height,TimeStamp,From,To,Value,Gas Limit,Gas Used By Transaction,Gas Price,Actual Tx Cost/Fee,Nonce & {Position},Input Data|动态调整|
monero| |动态调整|
zcash| |2M|
eos| |1M（动态）|