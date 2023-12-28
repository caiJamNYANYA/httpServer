# 介绍
**花花绿绿的？命令行文件分享软件（￣▽￣）**
做这个软件主要为了在局域网内快速的分享文件到其他设备(有公网的话也可以远程分享)，原理是在一个文件夹内创建符号链接文件，然后在这个文件夹启动http静态文件服务，所以…………
还有……chatgpt真好用，快夸我只会cv编程！
# 使用 （软件不支持windows！！）
```bash
##获取##
git clone https://github.com/caiJamNYANYA/httpServer.git
cd httpServer
go build

```
```bash
##分享##
./shrd # 分享当前目录
./shrd -p 8080 # 分享当前目录监听在8080端口
./shrd /home/yuki.n/ # 分享一个目录
./shrd /etc/fstab # 分享一个文件
./shrd * # 分享当前目录下所有文件、文件夹
find path | ./shrd #使用管道
./shrd ~/data /bin download/ /usr/local/ 1234.mp3 -p 15532 # 分享多个目录、文件并监听在15532端口
``` 
```bash
##可以在终端快速下载文件和打包成tar或者压缩过的tar包的文件夹##
wget http://ip:port/file # 下载文件
wget http://ip:port/path/(dir/ || file)\?m=t # (?m=t || ?m=gz || ?m=xz)下在打包或者打包并压缩文件夹……文件也是可以的
curl http://ip:port/path/ # 获取文件列表
curl http://ip:port/path/\?s=type #文件类型后缀排序,?rs=type倒序
```
编译完成建议将程序添加到bin目录或给程序所在的目录添加到PATH

**使用愉快**
