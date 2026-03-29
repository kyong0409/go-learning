// main.go
// deploy-tool: kubectl과 유사한 배포 관리 CLI 도구
//
// 실행 예시:
//   deploy-tool deploy --app myapp --image nginx:latest
//   deploy-tool status myapp
//   deploy-tool logs myapp
//   deploy-tool config set registry docker.io
package main

import "github.com/curriculum/deploy-tool/cmd"

func main() {
	cmd.Execute()
}
