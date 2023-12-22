package main
import (
	"os"
	"io"
	"fmt"
	"flag"
//	"net"
//	"strings"
	"archive/tar"
	"compress/gzip"
	"math/rand"
	"path/filepath"
	"syscall"
	"os/signal"
	"net/http"
)
var homePath string//分享目录
var tarPath string//tar储存目录
var randomNumbr string//随机数
var pathNameArgs []string//定义路径最后一级名称的数组
var pathNameArgsFix []string//定义路径最后一级名称重复加前缀的数组
var pathOrign []string//定义原始输入路径的数组


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

func tarGzFiles(gz bool, outputFile string, files []string) error {
	tarFile, _ := os.Create(outputFile)
	defer tarFile.Close()
	var tarWriter *tar.Writer
	if gz {//判断是否需要压缩gzip
		gzWriter := gzip.NewWriter(tarFile)
		defer gzWriter.Close()
		tarWriter = tar.NewWriter(gzWriter)
	} else {
		tarWriter = tar.NewWriter(tarFile)
	}
	defer tarWriter.Close()
	for _, file := range files {
		filepath.Walk(file, func(path string, info os.FileInfo, err error) error {
			header, _ := tar.FileInfoHeader(info, info.Name())
			relPath, _ := filepath.Rel(filepath.Dir(file), path)
			header.Name = relPath
			tarWriter.WriteHeader(header); 
			if !info.Mode().IsRegular() {
				return nil
			}
			fileToWrite, _ := os.Open(path)
			defer fileToWrite.Close()
			io.Copy(tarWriter, fileToWrite)
			return nil
		})
	}
	return nil
}

func main() {
	var port int
	flag.IntVar(&port, "p", 5050, "listen port")//端口参数获取
	flag.Parse()//更新参数变量
	
	fmt.Printf("\u2606Http Server\u2606\n")

	randomNumbr = fmt.Sprintf("____%s",randomString(12))//共享文件夹名称
	homePath = fmt.Sprintf("%s/%s",os.Getenv("HOME"),randomNumbr)//共享文件夹绝对路径
	os.Mkdir(homePath, os.ModePerm)//创建共享文件夹
	tarPath = fmt.Sprintf("%s/%s",homePath,randomNumbr)//共享
	os.Mkdir(tarPath, os.ModePerm)

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
		absDir, _ := filepath.Abs(value)
		_, err := os.Stat(absDir)//检查绝对路径是否有效，只输出有效路径
		if err != nil {
			//该路径不存在
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
		dir, file := filepath.Split(fromPath)//判断路径类型；避免bug出现
		var pathName string
		if dir == "/" && file == "" {//unix类系统的/作为路径分隔符是唯一不能用作文件名的字符，所以直接以/为名称给根目录创建链接会出错
			pathName = "R0OT"//更改根目录的目标链接名
		} else {
			pathName = file//目标链接文件名
			}	
		pathNameArgs = append(pathNameArgs, pathName)//添加文件名到数组
	}

	pathNameArgsFix = argsFix(pathNameArgs)//将重名的链接名称添加前缀…………为什么不能放在for里面喵

	for i := 1 ; i <= len(pathOrign) - 1; i++ {
		 /*var ipArgs []string
		 addrs, _ := net.InterfaceAddrs()//获取本地ip地址
		 for _, addr := range addrs {
			 if strings.HasPrefix(addr.String(), "192.168.1"/*匹配前缀，我家的ip网段是一个192.168.1.x请自行修改*//*) {
				 ipArgs = append(ipArgs, strings.Split(addr.String(),"/")[0])//添加到数组
			 }
		 }
		 //需要导入net包和strings包	
		 */

		fmt.Println("\n\033[32m\u2605\033[34mfrom\t<--\x1b[1;0m",pathOrign[(i)])
		downloadAddr := fmt.Sprintf("localhost:%d/%s"/*,ipArgs[0]*/,port,pathNameArgs[(i - 1)])//根据自己的网络修改吧～
		fmt.Println("\033[32m\u2605\033[34mto\t-->\x1b[1;0m",downloadAddr)
		
		fromPath := pathOrign[(i)]//链接原路径获取
		toPath := fmt.Sprintf("%s/%s",homePath,pathNameArgsFix[(i - 1 )])//链接目录路径
		err := os.Symlink(fromPath, toPath)//创建链接
		if err != nil {
			rm(homePath)
			panic(err)
		}
	}
	
	sigChan := make(chan os.Signal, 1)//用于退出时清理共享文件夹;监听ctrl+c
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
	http.HandleFunc("/", handler)//挂载根目录
	address := fmt.Sprintf(":%d", port)//http监听地址:端口
	fmt.Printf("\n\033[34mHTTP%s @ %s\n\n\033[0m", address, homePath)
	err := http.ListenAndServe(address, nil)//开启http服务
	if err != nil {
		if rm(homePath) != nil {//移除共享文件夹
		} else {
			fmt.Printf("\n\x1b[1;31mPort Error")
			fmt.Println("\nRemove done!",randomNumbr)
			os.Exit(1)
		}
	}
	select {}
}


func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\033[35m请求路径:\t", r.URL.Path, "\n\033[0m")//打印请求
	path := (homePath + r.URL.Path) // 获取添加到tar的目录的信息
	if r.URL.Query().Get("m") == "gz" || r.URL.Query().Get("m") == "t" || r.URL.Query().Get("m") == "tgz" {//重定向tar.gz压缩包
		var linkPath string//获取软链接路径到原始路径使用的变量
		var folderPath string
		var tarFrom []string//tar文件创建的输入
		Dir, File := filepath.Split(path)//判断路径是否有/符号因为“os.Readlink”函数匹配的路径末尾不能有/
		if Dir == homePath + "/" && File == "" {//下载根目录
			for _, pname := range pathNameArgsFix {
				Path := homePath + "/" + pname
				folderPath ,_ := os.Readlink(Path)//软链接里面的文件夹能正常被打包，这里获取软链接文件的对应的原始路径
				linkPath = folderPath
				tarFrom = append(tarFrom, linkPath)
			//	fmt.Println("\033[32m",pname,"\033[0m")
			//	fmt.Println("A")
			}
		} else if File == "" {//下载单个文件夹，如果请求的末尾有斜杠将不会输出File变量,这个是给目录准备的，所以打包下载文件时就不要手贱加/了
			linkPath = Dir[:len(Dir)-1]//将/删掉
			folderPath ,_ = os.Readlink(linkPath)
			if folderPath == "" { 
				linkPath = linkPath
			} else {
				linkPath = folderPath
			}
			tarFrom = append(tarFrom, linkPath)
		//	fmt.Println("B")

		} else {//下载单个文件夹
			folderPath ,_ := os.Readlink(path)
			if folderPath == "" {
				linkPath = path
			} else {
			linkPath = folderPath
			}
			tarFrom = append(tarFrom, linkPath)
		//	fmt.Println("C")
		}
	//	fmt.Println("LinkPath",linkPath,"\tFolderPath",folderPath)

		_, pathName := filepath.Split(linkPath)
		if Dir == homePath + "/" && File == "" {
			pathName = "Nyanya"//跟上面一样，删掉这里虽然不会出错，但是文件变成只有扩展名的就很奇怪
		} else {
			pathName = pathName
		}
		var tarTo string//tar文件目标路径
		var fileName string//302重定向文件名
		var gz bool//判断是否要使用gz压缩
		if r.URL.Query().Get("m") == "t" {
			fileName = pathName + ".tar"
		} else if r.URL.Query().Get("m") == "gz" {
			fileName = pathName + ".tar.gz"
			gz = true
		} else if r.URL.Query().Get("m") == "tgz" { //如果你要觉得tar.gz扩展名不好看……
			fileName = pathName + ".tgz"
			gz = true
		}
			tarTo = tarPath + "/" + fileName
			err := tarGzFiles(gz, tarTo, tarFrom)
			if err != nil {
				fmt.Println("Error:", err)
				return
			} else {
				fmt.Println("\x1b[1;32m归档已就绪\t", tarTo,"\n\033[0m")
			}
		fileURL := "/" + randomNumbr + "/" + fileName//设置重定向的链接
		w.Header().Set("Location", fileURL)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
		w.WriteHeader(http.StatusFound) // 使用302状态码进行重定向

	} else {//静态文件列表
		fileServer := http.FileServer(http.Dir(homePath))
		fileServer.ServeHTTP(w, r)
	}

}
