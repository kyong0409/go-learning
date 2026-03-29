// main.go
// Cobra CLI 도구의 진입점
//
// Cobra는 Go에서 가장 널리 쓰이는 CLI 프레임워크입니다.
// Kubernetes(kubectl), Hugo, GitHub CLI 등이 Cobra를 사용합니다.
//
// 실행 예시:
//   go run main.go --help
//   go run main.go greet --name 홍길동
//   go run main.go version
//   go run main.go config show
package main

import "github.com/curriculum/cobra-example/cmd"

func main() {
	// 모든 로직은 cmd 패키지에 위임합니다.
	// main.go는 최대한 얇게 유지하는 것이 관례입니다.
	cmd.Execute()
}
