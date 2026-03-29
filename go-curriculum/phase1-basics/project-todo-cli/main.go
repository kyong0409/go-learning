// main.go: Todo CLI 앱 진입점
//
// 사용법:
//   go run main.go add "할 일 제목"   - 새 항목 추가
//   go run main.go list               - 전체 목록 출력
//   go run main.go list --done        - 완료 항목만 출력
//   go run main.go list --pending     - 미완료 항목만 출력
//   go run main.go done <id>          - 항목 완료 처리
//   go run main.go delete <id>        - 항목 삭제
//   go run main.go update <id> "새 제목" - 제목 수정
//   go run main.go clear-done         - 완료 항목 모두 삭제
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"go-curriculum/phase1/todo-cli/todo"
)

// 데이터 파일 경로
const dataFile = "todos.json"

func main() {
	// 명령줄 인자가 없으면 도움말 출력
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	// Todo 목록 불러오기
	list, err := todo.Load(dataFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "데이터 로드 실패: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	// 명령 처리
	switch command {
	case "add":
		cmdAdd(list, args)
	case "list", "ls":
		cmdList(list, args)
	case "done", "complete":
		cmdDone(list, args)
	case "delete", "del", "rm":
		cmdDelete(list, args)
	case "update":
		cmdUpdate(list, args)
	case "clear-done":
		cmdClearDone(list)
	case "help", "-h", "--help":
		printHelp()
		return
	default:
		fmt.Fprintf(os.Stderr, "알 수 없는 명령: %q\n", command)
		printHelp()
		os.Exit(1)
	}

	// 변경사항 저장 (add/done/delete/update/clear-done)
	if command != "list" && command != "ls" && command != "help" {
		if err := list.Save(dataFile); err != nil {
			fmt.Fprintf(os.Stderr, "데이터 저장 실패: %v\n", err)
			os.Exit(1)
		}
	}
}

// ─────────────────────────────────────────
// 명령 처리 함수들
// ─────────────────────────────────────────

// cmdAdd: 새 할 일 추가
func cmdAdd(list *todo.TodoList, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "오류: 제목을 입력하세요")
		fmt.Fprintln(os.Stderr, "사용법: todo add \"할 일 제목\"")
		os.Exit(1)
	}

	title := strings.Join(args, " ")
	item, err := list.Add(title)
	if err != nil {
		fmt.Fprintf(os.Stderr, "추가 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("추가됨: [%d] %s\n", item.ID, item.Title)
}

// cmdList: 목록 출력
func cmdList(list *todo.TodoList, args []string) {
	var filter *bool
	var filterLabel string

	// 플래그 파싱
	for _, arg := range args {
		switch arg {
		case "--done":
			v := true
			filter = &v
			filterLabel = " (완료)"
		case "--pending":
			v := false
			filter = &v
			filterLabel = " (미완료)"
		}
	}

	items := list.List(filter)
	total, done, pending := list.Count()

	fmt.Printf("=== Todo 목록%s ===\n", filterLabel)

	if len(items) == 0 {
		fmt.Println("(항목이 없습니다)")
	} else {
		for _, item := range items {
			status := "[ ]"
			if item.Done {
				status = "[x]"
			}
			createdStr := item.CreatedAt.Format("2006-01-02")
			fmt.Printf("  %s [%3d] %-40s %s\n",
				status, item.ID, item.Title, createdStr)
		}
	}

	fmt.Printf("\n전체: %d | 완료: %d | 미완료: %d\n", total, done, pending)
}

// cmdDone: 완료 처리
func cmdDone(list *todo.TodoList, args []string) {
	id, ok := parseID(args, "done")
	if !ok {
		return
	}

	if err := list.Complete(id); err != nil {
		fmt.Fprintf(os.Stderr, "완료 처리 실패: %v\n", err)
		os.Exit(1)
	}

	item, _ := list.Get(id)
	fmt.Printf("완료됨: [%d] %s\n", item.ID, item.Title)
}

// cmdDelete: 삭제
func cmdDelete(list *todo.TodoList, args []string) {
	id, ok := parseID(args, "delete")
	if !ok {
		return
	}

	item, err := list.Get(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "삭제 실패: %v\n", err)
		os.Exit(1)
	}
	title := item.Title

	if err := list.Delete(id); err != nil {
		fmt.Fprintf(os.Stderr, "삭제 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("삭제됨: [%d] %s\n", id, title)
}

// cmdUpdate: 제목 수정
func cmdUpdate(list *todo.TodoList, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "오류: ID와 새 제목을 입력하세요")
		fmt.Fprintln(os.Stderr, "사용법: todo update <id> \"새 제목\"")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "유효하지 않은 ID: %q\n", args[0])
		os.Exit(1)
	}

	newTitle := strings.Join(args[1:], " ")
	if err := list.Update(id, newTitle); err != nil {
		fmt.Fprintf(os.Stderr, "수정 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("수정됨: [%d] %s\n", id, newTitle)
}

// cmdClearDone: 완료 항목 모두 삭제
func cmdClearDone(list *todo.TodoList) {
	removed := list.ClearDone()
	if removed == 0 {
		fmt.Println("완료된 항목이 없습니다.")
	} else {
		fmt.Printf("완료 항목 %d개 삭제됨\n", removed)
	}
}

// ─────────────────────────────────────────
// 헬퍼 함수
// ─────────────────────────────────────────

// parseID: args에서 정수 ID를 파싱합니다.
func parseID(args []string, cmd string) (int, bool) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "오류: ID를 입력하세요\n")
		fmt.Fprintf(os.Stderr, "사용법: todo %s <id>\n", cmd)
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "유효하지 않은 ID: %q\n", args[0])
		os.Exit(1)
	}

	return id, true
}

// printHelp: 도움말 출력
func printHelp() {
	fmt.Println(`Todo CLI - 간단한 할 일 관리 도구

사용법:
  go run main.go <명령> [인자...]

명령:
  add <제목>              새 할 일 추가
  list [--done|--pending] 목록 출력 (필터 옵션 가능)
  ls   [--done|--pending] list의 별칭
  done <id>               항목 완료 처리
  complete <id>           done의 별칭
  delete <id>             항목 삭제
  del <id>                delete의 별칭
  update <id> <새 제목>   제목 수정
  clear-done              완료된 항목 모두 삭제
  help                    이 도움말 출력

예시:
  go run main.go add "Go 언어 공부하기"
  go run main.go list
  go run main.go done 1
  go run main.go delete 2
  go run main.go update 3 "Go 언어 마스터하기"
  go run main.go list --done
  go run main.go clear-done`)
}
