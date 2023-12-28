package main
import (
	"os"
	"io"
	"fmt"
	// "path"
	"sort"
//	"flag"
	"embed"	
	"bufio"
	// "net"
	"strings"
	"archive/tar"
	"compress/gzip"
	"math/rand"
	"path/filepath"
	"syscall"
	"os/signal"
	"net/http"
	"github.com/ulikunitz/xz"
)

//go:embed icon/*
var content embed.FS

var homePath string//分享目录
var tarPath string//tar储存目录
var randomNumbr string//随机数
var pathNameArgs []string//定义路径最后一级名称的数组
var pathNameArgsFix []string//定义路径最后一级名称重复加前缀的数组
var pathOrign []string//定义原始输入路径的数组

var clr = [...]int{219,171,213,202,220,208,217,183,211,195,223,225,229,85,86,123,153,189,117,105,177,175,204,218}//这啥呀(_^_)_这是

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
	const charset = "abcdefABCDEF0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func rm(filePath string) error {//移除函数，用于停止后移除共享文件夹
	return os.RemoveAll(filePath)
}

func tarGzFiles(tarType, outputFile string, files []string) error {
	tarFile, _ := os.Create(outputFile)
	defer tarFile.Close()
	var tarWriter *tar.Writer
	switch tarType {
	case "gz" ://判断是否需要使用gzip压缩
		Writer := gzip.NewWriter(tarFile)
		defer Writer.Close()
		tarWriter = tar.NewWriter(Writer)
	case "xz" ://判断是否需要使用xzip压缩
		Writer, _ := xz.NewWriter(tarFile)
		defer Writer.Close()
		tarWriter = tar.NewWriter(Writer)
	case "t" :
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

func textColoful(str string) {//我真的不想看见白色得文本出现喵～
	for _, text := range str {
		fmt.Printf("\x1b[38;5;%dm%c",clr[rand.Intn(len(clr))],text)
	}
}
func cPrint(str string) {
	fmt.Printf("\x1b[38;5;%dm",clr[rand.Intn(len(clr))])
	fmt.Println(str)
}

func main() {
	port := 5050
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {//通过管道获取参数
		scanner := bufio.NewScanner(os.Stdin)
		var inputLines []string
		for scanner.Scan() {
			inputLines = append(inputLines, scanner.Text())
		
		}
		os.Args = append(os.Args,inputLines...)
	}
//	flag.IntVar(&port, "p", 5050, "listen port")//端口参数获取
//	flag.Parse()//更新参数变量
	for i, param := range os.Args {
		args := []string{}
		argsCopy := os.Args
		switch param {
		case "-p", "--port" :
			if i + 2 > len(os.Args) {
				cPrint("Usage:")
				cPrint("-p int\vlisten port (default 5050)")
				os.Exit(0)
			} else {
				portStr := os.Args[i+1]
				//	var port int
				_, err := fmt.Sscanf(portStr, "%d", &port)//转成数字是为了比较的……
				if err != nil {
					cPrint("喂！你有在输入端口吗(\"▔□▔)/\n错误的端口(´･_･`)")
					os.Exit(0)
				}
				os.Args = append(args, os.Args[0:i]...)
				os.Args = append(os.Args,argsCopy[i+2:len(argsCopy)]...)
				break
			}
		case "-h", "--help" :
			cPrint("Usage: shrd [PATH]... [-p <PORT>]")
			cPrint("\t-p, --port 似乎需要一个整数的样子(笑)")
			cPrint("\t-h, --help 可莉不知道哦(●´ϖ`●)")
			textColoful("少女乞讨中");fmt.Println("...")
			os.Exit(0)

		}
	}


	textColoful("Http Server\n")
	randomNumbr = ("____" + randomString(12))//共享文件夹名称
	homePath = fmt.Sprintf("%s/%s",os.Getenv("HOME"),randomNumbr)//共享文件夹绝对路径
	os.Mkdir(homePath, os.ModePerm)//创建共享文件夹
	tarPath = fmt.Sprintf("%s/%s",homePath,randomNumbr)//共享
	os.Mkdir(tarPath, os.ModePerm)
		
	// 打印接收到的数据

	pathArgs := make([]string, 2)//定义存储路径的数组；长度为2用于没有输入参数时定义默认值


	if (len(os.Args) == 1/*数组第0个元素默认是启动程序*/) || (port != 5050 && len(os.Args) == 3/*加了port参数-p之后长度变成3*/) {
		os.Args[0] = ""
		pathArgs[1] = "."//定义默认路径为当前目录
	} else if os.Args[1] == "-p"/*使用了端口值并且使用了路径时*/ {
		os.Args[0] = ""
		pathArgs = os.Args
		pathArgs[1] = randomString(12)//将-p参数替换为随机字符一次，以防将当前目录下与-p同名的文件夹给分享出去……
		pathArgs[2] = randomString(12)//将端口这一参数替换为随机的字符串，以防将目录下与端口名称一样的目录给分享出去……
	} else {
		os.Args[0] = ""
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
		fmt.Println("\n\x1b[38;5;211m这就是你输入的路径?(_^_)_")
		rm(homePath)//清除分享目录
		os.Exit(1)
	}

	for i := 1 ; i <= len(pathOrign) - 1/*因为数组第0个元素是程序名称，所有原来数组长度为输入的路径+1,如果路径个数为1时，只需要循环一此，以此类推*/; i++ {
		fromPath := pathOrign[i]//读取需要链接的原始路径
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

		fmt.Printf("\n\x1b[38;5;85m\u2605\x1b[38;5;159mfrom\t<--\x1b[38;5;%dm  %s\n",clr[rand.Intn(len(clr))],pathOrign[i])
		downloadAddr := fmt.Sprintf("localhost:%d/%s"/*,ipArgs[0]*/,port,pathNameArgs[i-1])//根据自己的网络修改吧～
		fmt.Printf("\x1b[38;5;85m\u2605\x1b[38;5;159mto\t-->\x1b[38;5;%dm  %s\n",clr[rand.Intn(len(clr))],downloadAddr)

		fromPath := pathOrign[i]//链接原路径获取
		toPath := fmt.Sprintf("%s/%s",homePath,pathNameArgsFix[i-1])//链接目录路径
		err := os.Symlink(fromPath, toPath)//创建链接
		if err != nil {
			fmt.Println(err)
			rm(homePath)
			os.Exit(1)
		//	panic(err)
		}
	}

	sigChan := make(chan os.Signal, 1)//用于退出时清理共享文件夹;监听ctrl+c
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	go func() {//退出时执行的
		<-sigChan
		if rm(homePath) != nil {//移除共享文件夹

		fmt.Printf("\n\x1b[38;5;211m似乎并没有删除什么 (@_@)")
		} else {
			fmt.Println("\n\x1b[38;5;85m似乎删掉了什么（￣▽￣）",randomNumbr)
			}
		os.Exit(0)
	}()
	http.HandleFunc("/", handler)//挂载根目录
	http.HandleFunc("/favicon.ico", serveFavicon)
	address := fmt.Sprintf(":%d", port)//http监听地址:端口
	fmt.Printf("\n\x1b[38;5;85mHTTP%s @ %s (・o・)\n\n\033[0m", address, homePath)
	fmt.Printf("\x1b[38;5;%dm",clr[rand.Intn(len(clr))])
	err := http.ListenAndServe(address, nil)//开启http服务
	if err != nil {
		if rm(homePath) != nil {//移除共享文件夹
		} else {
			if port < 1024 {
				errstr := fmt.Sprintf("%v",err)
				if (errstr[len(errstr)-2:]) == "ed" {
					fmt.Println("\x1b[38;5;211m你什么身份呀(╯°口°)╯(┴—┴这个端口也是你能用的！")
					fmt.Println("似乎删掉了什么 (゜o゜)",randomNumbr)
					os.Exit(1)
				} else {
					fmt.Println("\x1b[38;5;211m该端口无了 ≡≡ﾍ( ´Д`)ﾉ")
					fmt.Println("似乎删掉了什么 (゜o゜)",randomNumbr)
					os.Exit(1)
				}
			} else {

				fmt.Println("\x1b[38;5;211m该端口无了 ≡≡ﾍ( ´Д`)ﾉ")
				fmt.Println("似乎删掉了什么 (゜o゜)",randomNumbr)
				os.Exit(1)
			}
		}
	}
	select {}

}

func handler(w http.ResponseWriter, r *http.Request) {
	path := (homePath + r.URL.Path) // 获取添加到tar的目录的信息
	if r.URL.Query().Get("m") == "gz" || r.URL.Query().Get("m") =="xz" || r.URL.Query().Get("m") == "t" || r.URL.Query().Get("m") == "tgz" {//重定向tar.gz压缩包
		fmt.Printf("\x1b[38;5;%dm请求路径 (:\t%s\n", clr[rand.Intn(len(clr))],r.URL.Path)//打印请求
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
			//	fmt.Println("\x1b[38;5;85m",pname,"\033[0m")
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
		var tarType string
		switch r.URL.Query().Get("m") {//判断格式
			case "gz" :
				tarType = "gz"
				fileName = pathName + ".tar.gz"
			case "xz" :
				tarType = "xz"
				fileName = pathName + ".tar.xz"
			case "t" :
				tarType = "t"
				fileName = pathName + ".tar"
		}
		tarTo = tarPath + "/" + fileName
		err := tarGzFiles(tarType, tarTo, tarFrom)
		if err != nil {
			fmt.Println("Error:", err)
			return
		} else {
			fmt.Println("\x1b[38;5;85m已就绪 (>ω<)\t" + fileName + "\033[0m")
		}
		fileURL := "/" + randomNumbr + "/" + fileName//设置重定向的链接
		w.Header().Set("Location", fileURL)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
		w.WriteHeader(http.StatusFound) // 使用302状态码进行重定向

	} else {//主要成分！？
///*
		fileInfo, err := os.Lstat(path)
			if err != nil {
				fmt.Printf("\x1b[38;5;%dm请求路径 ):\t%s\n", clr[rand.Intn(len(clr))],r.URL.Path)//打印请求
				fmt.Fprintf(w,"No Found Funky(_^_)_")
				fmt.Println(err)
				return
			} else {
				fmt.Printf("\x1b[38;5;%dm请求路径 (:\t%s\n", clr[rand.Intn(len(clr))],r.URL.Path)//打印请求
			}
		if fileInfo.IsDir() {
htmlcode := `<meta charset="UTF-8">
<script>
coloR = ["#82b1ff","#ff89cf","#FB7299","#6dc781"];
var rdmc = ("rgb(" + ~~(256 * Math.random()) + "," + ~~(256 * Math.random()) + "," + ~~(256 * Math.random()) + ")");coloR.push(rdmc);
document.documentElement.style.setProperty('--cl',(coloR[~~(Math.random()*coloR.length)]));
</script>
<style type="text/css">
body {
	background-color: #000000ff;
	color:white;
	-webkit-text-stroke: 1.2px var(--cl);
}
a {
	color: white;
	font-size: 120%;
	text-decoration: none;
}
a:link {
	color: white;
	-webkit-text-stroke: 1.2px var(--cl);
}
a:hover {
	color: white;
	-webkit-text-stroke: 1.2px var(--cl);
}
a:visited {
	color: white;
	-webkit-text-stroke: 1.2px var(--cl);
}
a:active {
	color: white;
	-webkit-text-stroke: 1.2px var(--cl);
}
h1 {
	color : var(--cl);
}
</style>
`
			var dirLink []string
			var fileLink []string
			var dirName []string
			var fileName []string
			Link, _ := os.ReadDir(path)
			for _, Link := range Link {
				switch fmt.Sprintln(Link)[:1] {//判断路径类型
					case "d" :
						dirLink = append(dirLink,r.URL.Path + Link.Name())
						dirName = append(dirName,Link.Name())
					case "-" :
						fileLink = append(fileLink,r.URL.Path + Link.Name())
						fileName = append(fileName,Link.Name())
					case "L" :
						var realLink string
						absPath, _ := filepath.EvalSymlinks(path + Link.Name())
						realLink, _ = filepath.Abs(absPath)
						fileInfo, _ := os.Lstat(realLink)
						switch fmt.Sprintln(fileInfo.Mode())[:1] {
							case "d" :
								dirLink = append(dirLink,r.URL.Path + Link.Name())
								dirName = append(dirName,Link.Name())
							case "-" :
								fileLink = append(fileLink,r.URL.Path + Link.Name())
								fileName = append(fileName,Link.Name())
//							case "L" :
//								这里大概不会有软链接了的……
						}
				}
			}
			sort.Slice(fileName, func(i, j int) bool {
				suffixI := getFileSuffix(fileName[i])
				suffixJ := getFileSuffix(fileName[j])
				_Rs := false
				if r.URL.Query().Get("s") =="type" {
					_Rs = strings.Compare(suffixI, suffixJ) < 0
				} else if r.URL.Query().Get("rs") =="type" {
					_Rs = strings.Compare(suffixI, suffixJ) > 0
				}
					return _Rs
			})
			if r.Header.Get("User-Agent")[:1] == "c" {//判断是否是curl请求
				for i, _ := range dirLink {//打印目录列表
					w.Write([]byte(fmt.Sprintf("\x1b[38;5;%dm",clr[rand.Intn(len(clr))]) + dirName[i] + "/\n"))
				}
				for i, _ := range fileLink {//打印文件列表
					w.Write([]byte(fmt.Sprintf("\x1b[38;5;%dm",clr[rand.Intn(len(clr))]) + fileName[i] + "\n"))
				}

			} else {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(htmlcode))
				w.Write([]byte(r.URL.Path + "<pre><ul><a href=\"../\">../</a></ul>"))
				for i, dir := range dirLink {
					w.Write([]byte("<ul><a href='" + dir + "/'>" + dirName[i] + "/</a><br /></ul>"))
				}
				for i, file := range fileLink {
					w.Write([]byte("<ul><a href='" + file + "'>" + fileName[i] + "</a><br /></ul>"))
				}
				w.Write([]byte("</pre>"))
			}

		} else {
			http.ServeFile(w, r, path)
		}
//*/
		/*
		//你要用来做html网页服务器可以删掉上面的
		//fileServer := http.FileServer(http.Dir(homePath))
		//fileServer.ServeHTTP(w, r)
		*/
	}
}

func getFileSuffix(fileName string) string {// getFileSuffix 函数用于获取文件后缀
	lastDotIndex := strings.LastIndex(fileName, ".")// 使用strings.LastIndex获取最后一个点的位置
	if lastDotIndex == -1 {
		return ""
	}
	return strings.ToLower(fileName[lastDotIndex+1:])
}

func serveFavicon(w http.ResponseWriter, r *http.Request) {//网页图标
	faviconPath := "icon/favicon.png"
	file, _ := content.Open(faviconPath)
	defer file.Close()
	io.Copy(w, file)
}
