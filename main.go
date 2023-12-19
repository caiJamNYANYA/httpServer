package main
import (
	"os"
	"fmt"
	"flag"
	"net"
	"strings"
	"math/rand"
	"path/filepath"
	"syscall"
	"os/signal"
	"net/http"
)

func argsFix(arr []string) []string {//为聪明文件名加前缀，以防重名然后出现bug
	countMap := make(map[string]int)
	for i, v := range arr {
		count := countMap[v]
		countMap[v]++
		if count > 0 {
			arr[i] = fmt.Sprintf("%d_%s", count,v)
		}
	}
	return arr
}

func randomString(length int) string {//随机字符，，
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func rm(filePath string) error {//移除函数，用于停止后移除共享文件夹
	return os.RemoveAll(filePath)
}

func main() {
	var port int
	flag.IntVar(&port, "p", 5050, "listen port")//端口参数获取
	flag.Parse()//更新参数变量
	
	fmt.Printf("\u2606Http Server\u2606\n")
	randomNumbr := fmt.Sprintf("____%s",randomString(12))//共享文件夹名称
	homePath := fmt.Sprintf("%s/%s",os.Getenv("HOME"),randomNumbr)//共享文件夹绝对路径
	os.Mkdir(homePath, os.ModePerm)//创建共享文件夹
	
	var pathOrign []string//定义原始输入路径的数组
	var pathNameArgs []string//定义路径最后一级名称的数组
	pathArgs := make([]string, 2)//定义存储路径的数组；长度为2用于没有输入参数时定义默认值

	if (len(os.Args) == 1/*数组第0个元素默认是启动程序*/) || (os.Args[1] == "-p" && len(os.Args) == 3/*加了port参数-p之后长度变成3*/) {
		copy(pathArgs, os.Args)//获取原来数组
		pathArgs[1] = "."//定义默认路径为当前目录
		} else if os.Args[1] == "-p"/*使用了端口值并且使用了路径时*/ {
			copy(pathArgs, os.Args)
			pathArgs = os.Args
			pathArgs[1] = randomString(12)//将-p参数替换为随机字符一次，以防将当前目录下与-p同名的文件夹给分享出去……
			pathArgs[2] = randomString(12)//将端口这一参数替换为随机的字符串，以防将目录下与端口名称一样的目录给分享出去……
			} else {
				pathArgs = os.Args//只使用了路径参数时
				}

	for _, value := range pathArgs {//读取数组获取绝对路径
		// var absDir string
		absDir, err := filepath.Abs(value)
		fmt.Sprintln(err)//没有会报错没有使用该变量

		fileInfo, err := os.Stat(absDir)//检查绝对路径是否有效，只输出有效路径
		if err != nil {//该路径不存在时输出
		fmt.Sprintln("无法获取文件信息:", err, fileInfo)
		} else {
		 pathOrign = append(pathOrign,absDir)//将有效的路径存储进pathOrign数组
		}
	}

	if (len(pathOrign)) == 1 {//如果没有一个正确路径时退出程序，如果想使用空目录并手动添加文件可以删掉这段……
		fmt.Println("\n\x1b[1;31mNo file input")
		rm(homePath)//清除分享目录
		os.Exit(1)
		}

	for i := 1 ; i <= len(pathOrign) - 1/*因为数组第0个元素是程序名称，所有原来数组长度为输入的路径+1,如果路径个数为1时，只需要循环一此，以此类推*/; i++ {
		fromPath := pathOrign[(i)]//读取需要链接的原始路径
		file := ""
		dir, file := filepath.Split(fromPath)//判断路径类型；避免bug出现
		var pathName string
		if dir == "/" && file == "" {//unix类系统的/是非法文件名，所以直接给根目录创建链接会出错
			pathName = "R0OT"//更改根目录的目标链接名
		} else {
			pathName = file//目标链接文件名
			}
		if pathName == "index.html" {//替换index.html,分防止浏览器使用该文件而显示不了文件列表
			pathName = "index.html_file"
			}
		
		pathNameArgs = append(pathNameArgs, pathName)//添加文件名到数组
	}

	resultArray := argsFix(pathNameArgs)//将重名的链接名称添加前缀…………为什么不能放在for里面喵

	for i := 1 ; i <= len(pathOrign) - 1; i++ {
		var ipArgs []string
		addrs, _ := net.InterfaceAddrs()//获取本地ip地址
		for _, addr := range addrs {
			if strings.HasPrefix(addr.String(), "192.168.1"/*匹配前缀，这是我家的ip网段*/) {
				ipArgs = append(ipArgs, strings.Split(addr.String(),"/")[0])//添加到数组
			}
		}

		fmt.Println("\n\033[32m\u2605\033[34mfrom\t<--\x1b[1;0m",pathOrign[(i)])
		downloadAddr := fmt.Sprintf("%s:%d/%s",ipArgs[0],port,pathNameArgs[(i - 1)])
		fmt.Println("\033[32m\u2605\033[34mto\t-->\x1b[1;0m",downloadAddr)
		
		fromPath := pathOrign[(i)]//链接原路径获取
		toPath := fmt.Sprintf("%s/%s",homePath,resultArray[(i - 1 )])//链接目录路径
		err := os.Symlink(fromPath, toPath)//创建链接
		if err != nil {
			panic(err)
		}
	}
	
	sigChan := make(chan os.Signal, 1)//用于退出时清理共享文件夹监听ctrl+c
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	go func() {//退出时执行的
		<-sigChan
		if rm(homePath) != nil {//移除共享文件夹

		fmt.Printf("\n\x1b[1;31mRemove RRROR")
		} else {
			fmt.Println("\n\x1b[1;32mRemove done!",randomNumbr)
			}
		os.Exit(0)
	}()
	// fileServer := http.FileServer(http.Dir(homePath))
	// fmt.Println(fileServer)
	http.Handle("/", http.StripPrefix("/"/*http挂载目录*/, http.FileServer(http.Dir(homePath))/*http监听根目录*/))
	address := fmt.Sprintf(":%d", port)//http监听地址:端口
	fmt.Printf("\n\033[34mHTTP%s @ %s", address, homePath)
	sttr := http.ListenAndServe(address, nil)//开启http服务
	if sttr != nil {
		if rm(homePath) != nil {//移除共享文件夹
			} else {
				fmt.Printf("\n\x1b[1;31mPort Error")
				fmt.Println("\nRemove done!",randomNumbr)
				os.Exit(1)
			}
	}
	select {}
}