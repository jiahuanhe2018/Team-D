#!/bin/bash
pubKey=()
privatekey="GMFtw/DFtlnPWoXeUIECqU/Tb0Cx4g4HDndC1krDx7NOW+WtYsf7lLS/51FCj7qi"
signPubKey="BGqnxeYJy0y6o+Y0MkOnrMz7+Q99DHre7jvFKy25tM4n/ziAiA7idL8JiXY/Qt9u72rfAwipgl86YqBcESraw2OHZ/Ho6Sy/TYnQZtcbgkotWm+15WrT4jgm1JbRWDAo9Q=="
msg="11111111111222"

for ((i=1; i<=$1; i ++))
do
  pub=` ring-signatures g | awk -F"Public key:" '$2!=""{print $2}'`
  pubKey[i-1]="--ring "${pub}
done

cmd=(gtime -v  ring-signatures sign -m \"${msg}\" --private-key ${privatekey}
--ring-index 0 --ring ${signPubKey}  ${pubKey[@]}
);

res=""

for i in ${cmd[@]}
do
   res=${res}" "$i
done

echo $res | sh  > qq

echo ========= sign finish ============
echo ==========sign result ==========
cat qq

echo ==========verify =============
signStr=`cat qq |awk 'NR==2{print $0}'| sed 's/[\t]$//g' `
echo "#"${signStr}"#"

verifyCmd=(gtime -v ring-signatures verify -m \"${msg}\" -s ${signStr})

echo ${verifyCmd[@]} | sh

echo =========verify end=======

#{res}`


#echo ${pubKey[@]}
