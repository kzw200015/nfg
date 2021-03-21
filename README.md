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

2. 新建一个配置文件`nat.conf`

   ```
   8081,22,example.com
   8082,12345,example.com,192.168.1.6
   ```

   格式为：  
   `本地端口,远程端口,远程地址[,本地地址]`  
   本地地址是可选的，如果不知道填什么就不填  
   同时支持端口段转发，只要把端口换成端口段，如`8081-8089`

3. 执行`./nfg -c nat.conf`

程序将会监听配置文件的变化实时更新NAT规则，同时每隔一分钟会更新域名所对应的IP