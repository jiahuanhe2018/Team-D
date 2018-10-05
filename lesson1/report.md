# 公链开发第一课作业

## 一、需要对ring signature，zk-snark进行比较，ring signatrue decoy的数量在多少的时候更消耗时间空间，zk-snark在使用上占用多少空间，计算时间相比哪个更快

### 1.  ring-signature

用的是Golang版本的ring-signature进行编译的：[https://github.com/t-bast/ring-signatures](https://github.com/t-bast/ring-signatures)

编译后在命令行下运行：

用ring-signatures generate分别生成三对公私钥

* A:

> Public key: BOOFA0Yd5uwW8UHaCLGxr6Qb4I1FCjWaLl/JXNrkrCXVLLGie+bDTPIvDyjZGUcESlXQIyL7e/t0EFPypMafdJJmT29vCSORcHe2bzZvdkW9nubu88claBdyxUa7umrkFw==
>
> Private key: NB7/MbcMIdsxweknsVnKgIIjBSkMllTwL+KsOVkdVvBS2zITwZDN1J5zN0xdYvw1

- B:

> Public key: BG5mAUib9Y/ffsIGVPs2npexJUdMHRKVHE+VaajYmwYl9dJT9exhCYl7jXFcBgx2MyTeXsEzIcNYa0Zf4+pqby/ixNO8qTJi3r8TW5e88eu/XyXGQzkn9XoLtIleoZm6Uw==
>
> Private key: KdQpEoPS3hlGMSY/RIdzctzsLL8DJdgZ2Q+NSlbcYJmmzgzrVCRGbw5G+zPb6ikt

* C:

> Public key: BCwLkbHX0NpjUpApPnzsJMqCBi/4yow9xO9NRQGB3rvn43K2yeSzBnspNljClO5hXdyj036Mo3qdF+fTIBWwRSIkxI1ZsRsYwedOUMDD1vgKXRb36QsHyGiptf6mu0CLtw==
>
> Private key: jnqCO9+eYinabViR+j/em+Vwl9m1HAx0O8D6mSSaax8+9kDeGdttGxVfIg6fn+J5

签名，sign的用法示例：

```
USAGE:
   Alice has private key "Pr1v4T3k3y", public key "4l1c3" and wants to sign the message "hello!".
   She wants to use Bob and Carol's public keys to form a ring.
   Bob's public key is "b0b" and Carol's public key is "c4r0l".
   Alice can form the ring [c4r0l, 4l1c3, b0b] and hide herself in that ring with the following command:
   ring-signatures sign --message "hello!" --private-key 4l1c3 --ring-index 1 --ring c4r0l --ring 4l1c3 --ring b0b
```

> ring-signatures sign --message "hello world" --private-key NB7/MbcMIdsxweknsVnKgIIjBSkMllTwL+KsOVkdVvBS2zITwZDN1J5zN0xdYvw1 --ring-index 1  --ring "BG5mAUib9Y/ffsIGVPs2npexJUdMHRKVHE+VaajYmwYl9dJT9exhCYl7jXFcBgx2MyTeXsEzIcNYa0Zf4+pqby/ixNO8qTJi3r8TW5e88eu/XyXGQzkn9XoLtIleoZm6Uw\=\=" --ring "BOOFA0Yd5uwW8UHaCLGxr6Qb4I1FCjWaLl/JXNrkrCXVLLGie+bDTPIvDyjZGUcESlXQIyL7e/t0EFPypMafdJJmT29vCSORcHe2bzZvdkW9nubu88claBdyxUa7umrkFw\=\=" --ring "BCwLkbHX0NpjUpApPnzsJMqCBi/4yow9xO9NRQGB3rvn43K2yeSzBnspNljClO5hXdyj036Mo3qdF+fTIBWwRSIkxI1ZsRsYwedOUMDD1vgKXRb36QsHyGiptf6mu0CLtw\=\="

返回数据：

> Signing message...
>
> eyJSIjpbIkJHNW1BVWliOVkvZmZzSUdWUHMybnBleEpVZE1IUktWSEUrVmFhalltd1lsOWRKVDlleGhDWWw3alhGY0JneDJNeVRlWHNFekljTllhMFpmNCtwcWJ5L2l4Tk84cVRKaTNyOFRXNWU4OGV1L1h5WEdRemtuOVhvTHRJbGVvWm02VXc9PSIsIkJPT0ZBMFlkNXV3VzhVSGFDTEd4cjZRYjRJMUZDaldhTGwvSlhOcmtyQ1hWTExHaWUrYkRUUEl2RHlqWkdVY0VTbFhRSXlMN2UvdDBFRlB5cE1hZmRKSm1UMjl2Q1NPUmNIZTJielp2ZGtXOW51YnU4OGNsYUJkeXhVYTd1bXJrRnc9PSIsIkJDd0xrYkhYME5walVwQXBQbnpzSk1xQ0JpLzR5b3c5eE85TlJRR0IzcnZuNDNLMnllU3pCbnNwTmxqQ2xPNWhYZHlqMDM2TW8zcWRGK2ZUSUJXd1JTSWt4STFac1JzWXdlZE9VTUREMXZnS1hSYjM2UXNIeUdpcHRmNm11MENMdHc9PSJdLCJTIjpbIk5zR21PWUJhUXZPblJ2ZnhSampkYVg5VWJ0amxQQ3NwcG54QXJsdjZTQ2NCT2FmT2lyMnlFa291a2lzZTFLTTMiLCJsZUZrYk9zYlZURFF0bWg1QjZkbFpuNGZseUM3QlJUUFkvTW84MFJWdWpUeDFnejlYYW94MlZpbjE2ZUFsYXd2IiwiSlhhRmxrMDdHS0FVNk1XM3BtL0hhbGFDTUJrMnBjaVJYa1dRT1NBM2t4dFhaVnpodUdIV2x1cy85aFdoMTFROSJdLCJFIjoiYmJqR3ZKWlpzTnlGZUx3K1JTT0hJcnhTcDU0OXJTUGVKNDRQMkJ5dmNPQT0ifQ==

对签名进行验证：

>ring-signatures verify --message "hello world" --signature "eyJSIjpbIkJHNW1BVWliOVkvZmZzSUdWUHMybnBleEpVZE1IUktWSEUrVmFhalltd1lsOWRKVDlleGhDWWw3alhGY0JneDJNeVRlWHNFekljTllhMFpmNCtwcWJ5L2l4Tk84cVRKaTNyOFRXNWU4OGV1L1h5WEdRemtuOVhvTHRJbGVvWm02VXc9PSIsIkJPT0ZBMFlkNXV3VzhVSGFDTEd4cjZRYjRJMUZDaldhTGwvSlhOcmtyQ1hWTExHaWUrYkRUUEl2RHlqWkdVY0VTbFhRSXlMN2UvdDBFRlB5cE1hZmRKSm1UMjl2Q1NPUmNIZTJielp2ZGtXOW51YnU4OGNsYUJkeXhVYTd1bXJrRnc9PSIsIkJDd0xrYkhYME5walVwQXBQbnpzSk1xQ0JpLzR5b3c5eE85TlJRR0IzcnZuNDNLMnllU3pCbnNwTmxqQ2xPNWhYZHlqMDM2TW8zcWRGK2ZUSUJXd1JTSWt4STFac1JzWXdlZE9VTUREMXZnS1hSYjM2UXNIeUdpcHRmNm11MENMdHc9PSJdLCJTIjpbIk5zR21PWUJhUXZPblJ2ZnhSampkYVg5VWJ0amxQQ3NwcG54QXJsdjZTQ2NCT2FmT2lyMnlFa291a2lzZTFLTTMiLCJsZUZrYk9zYlZURFF0bWg1QjZkbFpuNGZseUM3QlJUUFkvTW84MFJWdWpUeDFnejlYYW94MlZpbjE2ZUFsYXd2IiwiSlhhRmxrMDdHS0FVNk1XM3BtL0hhbGFDTUJrMnBjaVJYa1dRT1NBM2t4dFhaVnpodUdIV2x1cy85aFdoMTFROSJdLCJFIjoiYmJqR3ZKWlpzTnlGZUx3K1JTT0hJcnhTcDU0OXJTUGVKNDRQMkJ5dmNPQT0ifQ=="

返回数据：

> iSignature is valid.

经过几次尝试，得出这个结果，需要注意的点：

> 设置ring-index需要是当前签名者在所有人中的位置要和后面输入公钥的顺序相对应
>
> 还有输入私钥的时候**不**加双引号
>
> 公钥加双引号

签名数量的测试还没有来的及做

### 2. zk-snark的签名方式是C++的，我之前没有使用过C++进行开发，环境没有搭建成功，代码都没有运行，暂时无法运行测试效果。

---

## 二、将bitcoin、ethereum、monaro、zcash、EOS的交易、相关交易属性、块大小以及填入多少交易写在report中



| 公链项目     | 交易属性 | 块大小                   | 交易数量    |
| -------- | ---- | --------------------- | ------- |
| bitcoin  | UTXO | 1M                    | 大约3000笔 |
| ethereum | 账户   | 根据gas消耗限制，目前约为670万gas | 大约200笔  |
| monaro   | 账户   | 不限定，根据过去的块大小相应的增大或减小  | 不限定     |
| zcash    | 账户   | 1M                    |         |
| EOS      | 账户   | 有限制，但未查到具体的值          |         |

