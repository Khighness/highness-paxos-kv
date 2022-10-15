package toolkit

import "os"

// @Author KHighness
// @Update 2022-10-15

func IsDeployLocal() bool {
	return os.Getenv("env") == ""
}
