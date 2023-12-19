```bash
go build # 构建程序
./shrd # 运行程序

##示例##
./shrd -p 8080 # 分享当前目录监听在8080端口
./shrd / # 分享根目录
./shrd ./** # 分享当前目录下所有文件、文件夹到网页根目录
./shrd /bin ~/data download /usr/local/
```
