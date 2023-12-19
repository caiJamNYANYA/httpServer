# 介绍
该程序的原理是在一个文件夹内创建符号链接文件，并分享该文件夹，所以…………
### 此程序仅使用于unix类系统，不支持windows！！

```bash
##获取程序##
git clone https://githubfast.com/caiJamNYANYA/httpServer.git
cd httpServer
go build

##示例##
./shrd # 分享当前目录
./shrd -p 8080 # 分享当前目录监听在8080端口
./shrd /home/yuki.n # 分享一个目录
./shrd /etc/fstab # 分享一个文件
./shrd ** # 分享当前目录下所有文件、文件夹到网页根目录（支持统配符）
./shrd /bin -p 15532 ~/data download/ /usr/local/ 1234.mp3 # 分享多个目录、文件并监听在15532端口
```
编译完成您可以将程序添加到bin目录或给程序所在的目录添加到PATH
<center>**使用愉快**</center>
