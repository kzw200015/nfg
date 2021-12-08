一个简单的基于`nftables`的端口转发工具

使用方法：

1. 安装nftables
   ```
   apt install -y nftables // Debian/Ubuntu
   yum install -y nftables // CentOS
   ```
2. 打开IPv4转发

   ```
   echo -e "net.ipv4.ip_forward=1" >> /etc/sysctl.conf && sysctl -p
   ```
   可能需要重启生效

2. 新建一个配置文件`nfg.yml`
   ```yml
   rules:
     - srcPort: 8081 #源端口
       dstAddr: 2.2.2.2 #目标地址
       dstPort: 8082 #目标端口
       protocol: both #可选 tcp udp both
     - srcAddr: 3.3.3.3 #源地址，如果不知道填什么就把这一项删了，程序自动获取
       srcPort: 8081-8089 #支持端口段
       dstAddr: 4.4.4.4
       dstPort: 8081-8089
       protocol: tcp
   ```

3. 执行`./nfg -w -s nfg.yml`，监听配置文件的变化实时更新NAT规则，同时每隔一分钟会更新域名所对应的IP
   也可执行`./nfg -g -s nfg.yml`，生成转发规则并输出到终端