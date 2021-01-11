# grun
## go 版本的ssh并发执行命令  
###  feature:  
  1、支持ssh key验证和用户名、密码验证  
  2、支持管道和here document传递ip列表  
  3、支持定义常用命令别名  
  4、支持本地编写脚本，远端执行  
  5、支持sudo到root执行命令  
  /etc/sudoers关闭 Defaults    requiretty  
  /etc/sudoers开启NOPASSWD  user     ALL=(ALL)       NOPASSWD: ALL  
  6、支持分发文件，以及提权分发  
  7、ip列表中逗号、分号、空格自动分隔为多个ip  
  例如，下面的ip列表都可以支持  
  echo "10.10.10.1 10.10.10.2 10.10.10.3 10.10.10.4 10.10.10.5"|./grun id  
  echo "10.10.10.1,10.10.10.2,10.10.10.3,10.10.10.4,10.10.10.5"|./grun id  
  echo "10.10.10.1|10.10.10.2|10.10.10.3|10.10.10.4|10.10.10.5"|./grun id  

  echo "  
  10.10.10.1  
  10.10.10.2  
  10.10.10.3  
  10.10.10.4  
  10.10.10.5  
  "|./grun id  

./grun id <<EOF  
10.10.10.1  
10.10.10.2  
10.10.10.3  
10.10.10.4  
10.10.10.5  
EOF  

./grun id <<EOF  
10.10.10.1,10.10.10.2,10.10.10.3,10.10.10.4,10.10.10.5  
EOF  

### 常用参数  
  -b    if run cmd as root      
        #类似ansible -b,提权到root运行  

  -c    only copy local file to remote machine's some directory[can config]     
        #分发文件到目标机器，可以-b -c 一起使用，进行提权分发   

  -f    set concurrent num (default 300)    
        #设置运行并发数  

  -n    print result without new line between ip and result   
        #输出格式上ip和返回结果中不添加新行  
        #例如正常输出为  
        10.0.0.159  
        kvm-10-0-0-159  
        10.0.0.161  
        kvm-10-0-0-161  
        10.0.0.158  
        kvm-10-0-0-158  
        10.0.0.162  
        kvm-10-0-0-162  
        10.0.0.160  
        kvm-10-0-0-160  
        10.0.0.163  
        kvm-10-0-0-163  
        加上-n后为如下格式，便于后续awk等命令的操作  
        10.0.0.160 kvm-10-0-0-160  
        10.0.0.161 kvm-10-0-0-161  
        10.0.0.159 kvm-10-0-0-159  
        10.0.0.163 kvm-10-0-0-163  
        10.0.0.158 kvm-10-0-0-158  
        10.0.0.162 kvm-10-0-0-162  

  -nb   close backup when copy  
        #分发文件的时候是否备份，默认是要备份  
  -nc   close color print  
        #关闭输出格式上的颜色输出  

  -r    copy script file to remote and run  
        #本地编写的脚本，用上-r参数后，先分发到目标机器，然后再目标机器执行  
        #例如：当前目录有个脚本文件，可以直接按如下方式执行  
        #echo "10.10.10.1 10.10.10.2 10.10.10.3 10.10.10.4 10.10.10.5"|.grun -b -r sys_init.sh  
  -t    set ssh connect time out, unit is second (default 2)  

  -v    open debug mode  

### cfg配置文件寻找顺序如下，找到第一个后停止    
当前目录  
家目录  
/etc/  
### 编译 
go build -o grun *.go  
sudo ln -sf `pwd`/grun /usr/bin/grun 