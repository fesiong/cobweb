package cobweb

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetLogNumber(name string) int64 {
	filePath := fmt.Sprintf("%sdata/%s.num", ExecPath, name)
	result, err := os.ReadFile(filePath)
	if err != nil {
		return 0
	}

	num, err := strconv.ParseInt(strings.TrimSpace(string(result)), 10, 64)
	if err != nil {
		return 0
	}

	return num
}

func StoreLogNumber(name string, num int64) error {
	filePath := fmt.Sprintf("%sdata/%s.num", ExecPath, name)

	data := fmt.Sprintf("%d", num)

	err := os.WriteFile(filePath, []byte(data), 0755)

	return err
}
