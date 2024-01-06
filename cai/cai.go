package cai
import (
	"fmt"
	"os"
	"math/rand"
)
var clr []int
/*func main() {
	clr = []int{219,171,213,202,220,208,217,183,211,195,223,225,229,85,86,123,153,189,117,105,177,175,204,218}
	Print(false, "000000","uuu")
	Print(true, "ffffff","aaa")
	Print(1, "fpfpfp","uijd\n")
	Print("我的世界")
	Print()
	Print(true,"少女奔腾中……")
	Print(true,"少女祈祷中……")

}*/
func Print(a ...any,) {
	clr = []int{219,171,213,202,220,208,217,183,211,195,223,225,229,85,86,123,153,189,117,105,177,175,204,218}
	if len(a) < 1 {
		fmt.Println()
		return
	}
	format := ""
	if fmt.Sprintf("%T",a[0]) == "bool" {
		for i := 0; i < len(a)-1; i++ {
			format += "%v"
		}
		switch a[0] {
		case false :
			format = "\x1b[38;5;" + fmt.Sprintf("%d",clr[rand.Intn(len(clr))]) + "m" + format +"\x1b[0m"
			fmt.Printf(format,a[1:len(a)]...)
			fmt.Println()
		case true :
			for _, text := range fmt.Sprintf(format,a[1:len(a)]...) {
				fmt.Printf("\x1b[38;5;%dm%c\x1b[0m",clr[rand.Intn(len(clr))],text)

			}
			fmt.Println()
		}
		return
	}
	for i := 0; i < len(a); i++ {
		format += "%v"
	}
	formatret := format + "\n"
	fmt.Printf(formatret,a...)
}


func Strparam(str, info string) string {
	var par string
	for i, param := range os.Args {
		if param == str {
			if i + 2 > len(os.Args) {
				fmt.Println("Usage :")
                              	fmt.Println("\t", str, " ", info,"\n")
                        	os.Exit(0)
                        } else {
                                par = os.Args[i+1]
				args := []string{}
				args1 := os.Args
				os.Args = append(args, os.Args[0:i]...)
				os.Args = append(os.Args,args1[i+2:len(args1)]...)
                                break
			}
                }
        }
	return par
}
