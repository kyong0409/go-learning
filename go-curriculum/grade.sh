#!/bin/bash
# Go 커리큘럼 전체 과제 채점 스크립트
# 사용법: bash grade.sh [phase1|phase2|phase3|phase4|phase5|phase6]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOTAL_PASSED=0
TOTAL_TESTS=0
TOTAL_SCORE=0
TOTAL_MAX=0
FAILED_ASSIGNMENTS=()

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

print_header() {
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}  Go 커리큘럼 과제 채점 시스템 (28개 과제 / 2,780점)${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

grade_assignment() {
    local dir="$1"
    local name="$2"

    if [ ! -d "$dir" ]; then
        return
    fi

    echo -e "${YELLOW}▶ 채점 중: ${name}${NC}"

    local output
    if output=$(cd "$dir" && go test -v -count=1 -timeout=60s 2>&1); then
        local status="PASS"
    else
        local status="FAIL"
    fi

    local grade_line=$(echo "$output" | grep -E "(Passed|통과):" | tail -1)
    local score_line=$(echo "$output" | grep -E "(Score|점수):" | tail -1)

    if [ -n "$grade_line" ] && [ -n "$score_line" ]; then
        local passed=$(echo "$grade_line" | grep -oE '[0-9]+/[0-9]+' | head -1)
        local score=$(echo "$score_line" | grep -oE '[0-9]+/[0-9]+' | head -1)

        local p_num=$(echo "$passed" | cut -d'/' -f1)
        local p_den=$(echo "$passed" | cut -d'/' -f2)
        local s_num=$(echo "$score" | cut -d'/' -f1)
        local s_den=$(echo "$score" | cut -d'/' -f2)

        TOTAL_PASSED=$((TOTAL_PASSED + p_num))
        TOTAL_TESTS=$((TOTAL_TESTS + p_den))
        TOTAL_SCORE=$((TOTAL_SCORE + s_num))
        TOTAL_MAX=$((TOTAL_MAX + s_den))

        if [ "$s_num" -eq "$s_den" ] 2>/dev/null; then
            echo -e "  ${GREEN}✓ ${passed} 통과 | 점수: ${score} (만점!)${NC}"
        elif [ "$s_num" -gt 0 ] 2>/dev/null; then
            echo -e "  ${YELLOW}△ ${passed} 통과 | 점수: ${score}${NC}"
        else
            echo -e "  ${RED}✗ ${passed} 통과 | 점수: ${score}${NC}"
            FAILED_ASSIGNMENTS+=("$name")
        fi
    else
        echo -e "  ${RED}✗ 채점 결과를 파싱할 수 없습니다${NC}"
        FAILED_ASSIGNMENTS+=("$name")
    fi
}

print_header

FILTER="${1:-all}"

# Phase 1
if [ "$FILTER" = "all" ] || [ "$FILTER" = "phase1" ]; then
    echo -e "${BOLD}📘 Phase 1: Go 기초 (3개)${NC}"
    echo "────────────────────────────────────"
    grade_assignment "$SCRIPT_DIR/phase1-basics/assignments/a1-calculator" "P1-A1 계산기 ★☆☆☆☆"
    grade_assignment "$SCRIPT_DIR/phase1-basics/assignments/a2-slice-ops" "P1-A2 슬라이스 ★★☆☆☆"
    grade_assignment "$SCRIPT_DIR/phase1-basics/assignments/a3-word-counter" "P1-A3 단어카운터 ★★½☆☆"
    echo ""
fi

# Phase 2
if [ "$FILTER" = "all" ] || [ "$FILTER" = "phase2" ]; then
    echo -e "${BOLD}📗 Phase 2: 구조체와 인터페이스 (3개)${NC}"
    echo "────────────────────────────────────"
    grade_assignment "$SCRIPT_DIR/phase2-structs-interfaces/assignments/a1-shape-interface" "P2-A1 도형 ★★☆☆☆"
    grade_assignment "$SCRIPT_DIR/phase2-structs-interfaces/assignments/a2-custom-error" "P2-A2 은행에러 ★★★☆☆"
    grade_assignment "$SCRIPT_DIR/phase2-structs-interfaces/assignments/a3-json-config" "P2-A3 JSON설정 ★★★☆☆"
    echo ""
fi

# Phase 3
if [ "$FILTER" = "all" ] || [ "$FILTER" = "phase3" ]; then
    echo -e "${BOLD}📙 Phase 3: 동시성 (6개)${NC}"
    echo "────────────────────────────────────"
    grade_assignment "$SCRIPT_DIR/phase3-concurrency/assignments/a1-pipeline" "P3-A1 파이프라인 ★★★☆☆"
    grade_assignment "$SCRIPT_DIR/phase3-concurrency/assignments/a4-fanout-errgroup" "P3-A4 병렬API ★★★½☆"
    grade_assignment "$SCRIPT_DIR/phase3-concurrency/assignments/a2-worker-pool" "P3-A2 워커풀 ★★★★☆"
    grade_assignment "$SCRIPT_DIR/phase3-concurrency/assignments/a5-rate-limiter" "P3-A5 속도제한 ★★★★½"
    grade_assignment "$SCRIPT_DIR/phase3-concurrency/assignments/a6-graceful-server" "P3-A6 우아한종료 ★★★★½"
    grade_assignment "$SCRIPT_DIR/phase3-concurrency/assignments/a3-chat-server" "P3-A3 채팅서버 ★★★★★"
    echo ""
fi

# Phase 4
if [ "$FILTER" = "all" ] || [ "$FILTER" = "phase4" ]; then
    echo -e "${BOLD}📕 Phase 4: 프로덕션 Go (5개)${NC}"
    echo "────────────────────────────────────"
    grade_assignment "$SCRIPT_DIR/phase4-production/assignments/a1-generic-collection" "P4-A1 제네릭 ★★★☆☆"
    grade_assignment "$SCRIPT_DIR/phase4-production/assignments/a3-rest-api" "P4-A3 REST API ★★★½☆"
    grade_assignment "$SCRIPT_DIR/phase4-production/assignments/a2-middleware-chain" "P4-A2 미들웨어 ★★★★☆"
    grade_assignment "$SCRIPT_DIR/phase4-production/assignments/a4-metrics-collector" "P4-A4 메트릭 ★★★★½"
    grade_assignment "$SCRIPT_DIR/phase4-production/assignments/a5-cache" "P4-A5 캐시 ★★★★½"
    echo ""
fi

# Phase 5
if [ "$FILTER" = "all" ] || [ "$FILTER" = "phase5" ]; then
    echo -e "${BOLD}📓 Phase 5: 고급 시스템 (5개)${NC}"
    echo "────────────────────────────────────"
    grade_assignment "$SCRIPT_DIR/phase5-advanced/assignments/a1-cli-tool" "P5-A1 CLI도구 ★★★☆☆"
    grade_assignment "$SCRIPT_DIR/phase5-advanced/assignments/a3-log-analyzer" "P5-A3 로그분석 ★★★½☆"
    grade_assignment "$SCRIPT_DIR/phase5-advanced/assignments/a2-load-tester" "P5-A2 로드테스터 ★★★★☆"
    grade_assignment "$SCRIPT_DIR/phase5-advanced/assignments/a4-plugin-system" "P5-A4 플러그인 ★★★★½"
    grade_assignment "$SCRIPT_DIR/phase5-advanced/assignments/a5-distributed-lock" "P5-A5 분산잠금 ★★★★★"
    echo ""
fi

# Phase 6
if [ "$FILTER" = "all" ] || [ "$FILTER" = "phase6" ]; then
    echo -e "${BOLD}🔥 Phase 6: K8s 오픈소스 딥다이브 (6개)${NC}"
    echo "────────────────────────────────────"
    grade_assignment "$SCRIPT_DIR/phase6-opensource/assignments/a3-metrics-registry" "P6-A3 메트릭레지스트리 ★★★★☆"
    grade_assignment "$SCRIPT_DIR/phase6-opensource/assignments/a1-controller" "P6-A1 컨트롤러 ★★★★½"
    grade_assignment "$SCRIPT_DIR/phase6-opensource/assignments/a5-watch-system" "P6-A5 감시시스템 ★★★★½"
    grade_assignment "$SCRIPT_DIR/phase6-opensource/assignments/a2-informer" "P6-A2 인포머 ★★★★★"
    grade_assignment "$SCRIPT_DIR/phase6-opensource/assignments/a4-dag-executor" "P6-A4 DAG실행기 ★★★★★"
    grade_assignment "$SCRIPT_DIR/phase6-opensource/assignments/a6-reconciler" "P6-A6 오퍼레이터 ★★★★★"
    echo ""
fi

# 최종 결과
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  최종 채점 결과${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

if [ "$TOTAL_MAX" -gt 0 ]; then
    PERCENT=$((TOTAL_SCORE * 100 / TOTAL_MAX))
    echo -e "  테스트 통과: ${BOLD}${TOTAL_PASSED}/${TOTAL_TESTS}${NC}"
    echo -e "  총 점수:     ${BOLD}${TOTAL_SCORE}/${TOTAL_MAX} (${PERCENT}%)${NC}"
    echo ""

    if [ "$PERCENT" -ge 90 ]; then
        echo -e "  등급: ${GREEN}${BOLD}S  - K8s 컨트리뷰터 준비 완료${NC}"
    elif [ "$PERCENT" -ge 80 ]; then
        echo -e "  등급: ${GREEN}${BOLD}A+ - 프로덕션 Go 개발자${NC}"
    elif [ "$PERCENT" -ge 70 ]; then
        echo -e "  등급: ${GREEN}${BOLD}A  - 숙련 Go 개발자${NC}"
    elif [ "$PERCENT" -ge 60 ]; then
        echo -e "  등급: ${YELLOW}${BOLD}B  - 중급 Go 개발자${NC}"
    elif [ "$PERCENT" -ge 40 ]; then
        echo -e "  등급: ${YELLOW}${BOLD}C  - 기초 완성${NC}"
    else
        echo -e "  등급: ${RED}${BOLD}D  - 계속 화이팅!${NC}"
    fi
else
    echo -e "  ${RED}채점할 과제가 없습니다.${NC}"
fi

if [ ${#FAILED_ASSIGNMENTS[@]} -gt 0 ]; then
    echo ""
    echo -e "  ${RED}미완성 과제:${NC}"
    for a in "${FAILED_ASSIGNMENTS[@]}"; do
        echo -e "    ${RED}• ${a}${NC}"
    done
fi

echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
