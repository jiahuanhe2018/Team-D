根据第二课的范例进行改写。<p>
1.在cmd目录下执行go build<p>
2.已经生成了钱包wallet_suncj.dat，在 cmd目录下,钱包地址是1NKyrPEZa5Bb5Hu1uQdSa2t9WnA3XkZNSW,钱包后缀是suncj<p>
3.启动命令行 cmd -c chain -s suncj -l 8080 -a 1NKyrPEZa5Bb5Hu1uQdSa2t9WnA3XkZNSW -datadir <cmd目录><p>
  根据提示可以启动多个节点。<p>
4.Post json 格式如下：
   {
    
      "From": "1NKyrPEZa5Bb5Hu1uQdSa2t9WnA3XkZNSW",
    
      "To": "1MXBtW5FdMNm15oqbYboHdiKzWm6TgNHj1",
    
      "Value": 100,
    
      "Data": "message"
   
}<p>
不论在那个节点POST，交易(TxPool)都会自动同步到其余的节点<p>
5. Blockchain里面的Accounts  会在节点内部自动重建并同步<p>
6. 作业以POW方式来实现，所以在挖矿成功后，打包成功的交易会从TxPool移除，并会在各个节点同步,通过web server的url可以看到交易以及state的变化。<p>
7. 在win10 ubuntu16.04 上验证通过。<p>