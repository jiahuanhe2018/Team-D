���ݵڶ��εķ������и�д��<p>
1.��cmdĿ¼��ִ��go build<p>
2.�Ѿ�������Ǯ��wallet_suncj.dat���� cmdĿ¼��,Ǯ����ַ��1NKyrPEZa5Bb5Hu1uQdSa2t9WnA3XkZNSW,Ǯ����׺��suncj<p>
3.���������� cmd -c chain -s suncj -l 8080 -a 1NKyrPEZa5Bb5Hu1uQdSa2t9WnA3XkZNSW -datadir <cmdĿ¼><p>
  ������ʾ������������ڵ㡣<p>
4.Post json ��ʽ���£�
   {
    
      "From": "1NKyrPEZa5Bb5Hu1uQdSa2t9WnA3XkZNSW",
    
      "To": "1MXBtW5FdMNm15oqbYboHdiKzWm6TgNHj1",
    
      "Value": 100,
    
      "Data": "message"
   
}<p>
�������Ǹ��ڵ�POST������(TxPool)�����Զ�ͬ��������Ľڵ�<p>
5. Blockchain�����Accounts  ���ڽڵ��ڲ��Զ��ؽ���ͬ��<p>
6. ��ҵ��POW��ʽ��ʵ�֣��������ڿ�ɹ��󣬴���ɹ��Ľ��׻��TxPool�Ƴ��������ڸ����ڵ�ͬ��,ͨ��web server��url���Կ��������Լ�state�ı仯��<p>
7. ��win10 ubuntu16.04 ����֤ͨ����<p>