##说明
项目目录在cmd

集成了pow与pos，靠命令行来区分，原有的代码保留<p>
###原有的：<p>
go run main.go -c chain -l 10000 -secio<p>
###pow共识:<p>
go run main.go -c chainpow -l 10000 -secio<p>
###pos共识:<p>
go run main.go -c chainpos -l 10000 -secio<p>


###注意问题:<p>
1.对输入没有进行有效性验证，输入必须为数字<p>
2.对一些边界条件没有做判断，可能会出问题<p>
3.挖矿的时候，会几个实例同时进行，最先挖出的去竞争上链，会出现数据竞争的问题。<p>
4.对识发现与同步机制还是有点问题，但不妨碍演示，后续改进。<p>
5.在win10 ubuntu16.04 上验证通过。




