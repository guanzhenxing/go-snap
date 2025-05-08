package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Go-Snap 配置模块示例程序")
	fmt.Println("============================")
	fmt.Println("可用示例:")
	fmt.Println("1. 基本配置示例")
	fmt.Println("2. 环境变量示例")
	fmt.Println("3. 配置监听示例")
	fmt.Println("4. 配置验证示例")
	fmt.Println("0. 运行所有示例")
	fmt.Println("============================")

	// 获取用户选择
	var choice string
	if len(os.Args) > 1 {
		choice = os.Args[1]
	} else {
		fmt.Print("请选择要运行的示例 (0-4): ")
		fmt.Scanln(&choice)
	}

	// 清理选择输入
	choice = strings.TrimSpace(choice)

	// 运行选中的示例
	switch choice {
	case "0":
		fmt.Println("\n运行所有示例...")
		runExample(1)
		runExample(2)
		runExample(3)
		runExample(4)
	case "1", "2", "3", "4":
		number, _ := strconv.Atoi(choice)
		runExample(number)
	default:
		fmt.Println("无效的选择，请输入0-4之间的数字")
	}
}

// runExample 运行指定编号的示例
func runExample(number int) {
	fmt.Printf("\n\n========== 运行示例 #%d ==========\n\n", number)

	switch number {
	case 1:
		RunBasicExample()
	case 2:
		RunEnvExample()
	case 3:
		RunWatcherExample()
	case 4:
		RunValidatorExample()
	}

	fmt.Printf("\n\n========== 示例 #%d 运行完成 ==========\n", number)
}
