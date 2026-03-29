// 패키지 선언
package main

// 동시성 패턴: 파이프라인 (Pipeline)
//
// 파이프라인 패턴:
// - 각 단계가 채널로 연결된 고루틴들
// - 이전 단계의 출력이 다음 단계의 입력
// - Unix 파이프(|)의 Go 버전
//
// 단계(Stage): 입력 채널을 받아 처리 후 출력 채널 반환
// 장점: 각 단계가 독립적, 병렬 실행, 메모리 효율적

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// ─────────────────────────────────────────
// 1. 기본 파이프라인: 숫자 처리
// ─────────────────────────────────────────

// generate는 숫자 슬라이스에서 채널 스트림을 생성합니다 (Source 단계).
func generate(done <-chan struct{}, nums ...int) <-chan int {
	out := make(chan int, len(nums))
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}()
	return out
}

// square는 입력 채널의 각 값을 제곱합니다 (Transform 단계).
func square(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-done:
				return
			}
		}
	}()
	return out
}

// filterEven은 짝수만 통과시킵니다 (Filter 단계).
func filterEven(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n%2 == 0 {
				select {
				case out <- n:
				case <-done:
					return
				}
			}
		}
	}()
	return out
}

// sumAll은 채널의 모든 값을 합산합니다 (Sink 단계).
func sumAll(done <-chan struct{}, in <-chan int) int {
	total := 0
	for {
		select {
		case n, ok := <-in:
			if !ok {
				return total
			}
			total += n
		case <-done:
			return total
		}
	}
}

func basicPipeline() {
	fmt.Println("\n--- 1. 기본 파이프라인 ---")
	fmt.Println("  흐름: generate → filterEven → square → sumAll")

	done := make(chan struct{})
	defer close(done)

	// 파이프라인 구성: 1~10 생성 → 짝수 필터 → 제곱 → 합산
	numbers := generate(done, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	evens := filterEven(done, numbers)
	squares := square(done, evens)
	result := sumAll(done, squares)

	fmt.Printf("  1~10 중 짝수의 제곱합: %d\n", result)
	// 2²+4²+6²+8²+10² = 4+16+36+64+100 = 220
}

// ─────────────────────────────────────────
// 2. 제네릭 파이프라인 빌더 (함수형 스타일)
// ─────────────────────────────────────────

// Pipeline은 채널 기반 파이프라인입니다.
type Pipeline[T any] struct {
	done <-chan struct{}
	ch   <-chan T
}

// NewPipeline은 슬라이스에서 파이프라인을 시작합니다.
func NewPipeline[T any](done <-chan struct{}, items ...T) *Pipeline[T] {
	ch := make(chan T, len(items))
	go func() {
		defer close(ch)
		for _, item := range items {
			select {
			case ch <- item:
			case <-done:
				return
			}
		}
	}()
	return &Pipeline[T]{done: done, ch: ch}
}

// MapPipeline은 변환 단계를 추가합니다.
func MapPipeline[T, U any](p *Pipeline[T], f func(T) U) *Pipeline[U] {
	out := make(chan U)
	go func() {
		defer close(out)
		for v := range p.ch {
			select {
			case out <- f(v):
			case <-p.done:
				return
			}
		}
	}()
	return &Pipeline[U]{done: p.done, ch: out}
}

// FilterPipeline은 필터 단계를 추가합니다.
func FilterPipeline[T any](p *Pipeline[T], pred func(T) bool) *Pipeline[T] {
	out := make(chan T)
	go func() {
		defer close(out)
		for v := range p.ch {
			if pred(v) {
				select {
				case out <- v:
				case <-p.done:
					return
				}
			}
		}
	}()
	return &Pipeline[T]{done: p.done, ch: out}
}

// Collect는 파이프라인 결과를 슬라이스로 수집합니다.
func Collect[T any](p *Pipeline[T]) []T {
	var result []T
	for v := range p.ch {
		result = append(result, v)
	}
	return result
}

func genericPipeline() {
	fmt.Println("\n--- 2. 제네릭 파이프라인 ---")

	done := make(chan struct{})
	defer close(done)

	// 숫자 파이프라인: 1~12 → 짝수 → 제곱 → 20 초과 필터
	numPipeline := NewPipeline(done, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)
	evenPipeline := FilterPipeline(numPipeline, func(n int) bool { return n%2 == 0 })
	squaredPipeline := MapPipeline(evenPipeline, func(n int) int { return n * n })
	filteredPipeline := FilterPipeline(squaredPipeline, func(n int) bool { return n > 20 })

	results := Collect(filteredPipeline)
	fmt.Printf("  짝수의 제곱 중 20 초과: %v\n", results)

	// 문자열 파이프라인
	words := []string{"hello", "world", "Go", "pipeline", "concurrent", "programming"}
	done2 := make(chan struct{})
	defer close(done2)
	strPipeline := NewPipeline(done2, words...)
	upperPipeline := MapPipeline(strPipeline, strings.ToUpper)
	longPipeline := FilterPipeline(upperPipeline, func(s string) bool { return len(s) > 4 })

	strResults := Collect(longPipeline)
	fmt.Printf("  5글자 초과 단어 (대문자): %v\n", strResults)
}

// ─────────────────────────────────────────
// 3. 실제 사용 사례: 이미지 처리 파이프라인
// ─────────────────────────────────────────

// Image는 처리할 이미지를 나타냅니다.
type Image struct {
	ID     int
	Name   string
	Width  int
	Height int
}

// ProcessedImage는 처리된 이미지를 나타냅니다.
type ProcessedImage struct {
	Original   Image
	Thumbnail  Image
	Compressed int // 압축률 (%)
	Tags       []string
}

// loadImages는 이미지 목록을 생성합니다 (Source).
func loadImages(ctx context.Context, imageIDs []int) <-chan Image {
	out := make(chan Image)
	go func() {
		defer close(out)
		for _, id := range imageIDs {
			time.Sleep(5 * time.Millisecond) // I/O 시뮬레이션
			img := Image{
				ID:     id,
				Name:   fmt.Sprintf("image_%04d.jpg", id),
				Width:  1920 + id*10,
				Height: 1080 + id*5,
			}
			select {
			case out <- img:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// tagImages는 이미지에 태그를 붙입니다 (Transform, 느린 단계).
func tagImages(ctx context.Context, in <-chan Image) <-chan ProcessedImage {
	out := make(chan ProcessedImage)
	go func() {
		defer close(out)
		for img := range in {
			time.Sleep(10 * time.Millisecond) // AI 태깅 시뮬레이션
			processed := ProcessedImage{
				Original: img,
				Thumbnail: Image{
					ID:     img.ID,
					Name:   fmt.Sprintf("thumb_%04d.jpg", img.ID),
					Width:  img.Width / 10,
					Height: img.Height / 10,
				},
				Compressed: 60 + (img.ID % 20),
				Tags: []string{
					"landscape",
					fmt.Sprintf("resolution_%dx%d", img.Width, img.Height),
				},
			}
			select {
			case out <- processed:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// saveResults는 결과를 저장합니다 (Sink).
func saveResults(ctx context.Context, in <-chan ProcessedImage) int {
	count := 0
	for {
		select {
		case processed, ok := <-in:
			if !ok {
				return count
			}
			fmt.Printf("  저장: %s → 썸네일 %dx%d, 압축률 %d%%, 태그: %v\n",
				processed.Original.Name,
				processed.Thumbnail.Width, processed.Thumbnail.Height,
				processed.Compressed,
				processed.Tags[:1])
			count++
		case <-ctx.Done():
			return count
		}
	}
}

func imageProcessingPipeline() {
	fmt.Println("\n--- 3. 이미지 처리 파이프라인 ---")

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	imageIDs := []int{1, 2, 3, 4, 5}

	start := time.Now()
	images := loadImages(ctx, imageIDs)
	processed := tagImages(ctx, images)
	count := saveResults(ctx, processed)

	fmt.Printf("  처리된 이미지: %d개, 소요 시간: %v\n",
		count, time.Since(start).Round(time.Millisecond))
}

// ─────────────────────────────────────────
// 4. 병렬 파이프라인 단계 (병목 해소)
// ─────────────────────────────────────────

// parallelStage는 느린 단계를 여러 워커로 병렬 실행합니다.
func parallelStage(ctx context.Context, in <-chan int, workers int, process func(int) int) <-chan int {
	out := make(chan int, workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range in {
				result := process(v)
				select {
				case out <- result:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// slowCompute는 느린 계산을 시뮬레이션합니다.
func slowCompute(n int) int {
	time.Sleep(20 * time.Millisecond)
	return int(math.Sqrt(float64(n * n * n)))
}

func parallelPipelineStage() {
	fmt.Println("\n--- 4. 병렬 파이프라인 단계 (병목 해소) ---")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})

	// 직렬 처리 시간 측정
	serialInput := generate(done, 1, 2, 3, 4, 5, 6, 7, 8)
	start := time.Now()
	var serialOut []int
	for v := range serialInput {
		serialOut = append(serialOut, slowCompute(v))
	}
	serialTime := time.Since(start)
	close(done)

	// 병렬 처리 (4개 워커)
	done2 := make(chan struct{})
	defer close(done2)
	parallelInput := generate(done2, 1, 2, 3, 4, 5, 6, 7, 8)
	start = time.Now()
	parallelOut := parallelStage(ctx, parallelInput, 4, slowCompute)
	var parallelResults []int
	for v := range parallelOut {
		parallelResults = append(parallelResults, v)
	}
	parallelTime := time.Since(start)

	fmt.Printf("  직렬 처리 (1워커): %v, 결과 수: %d\n",
		serialTime.Round(time.Millisecond), len(serialOut))
	fmt.Printf("  병렬 처리 (4워커): %v, 결과 수: %d\n",
		parallelTime.Round(time.Millisecond), len(parallelResults))
	fmt.Printf("  속도 향상: %.1fx\n", float64(serialTime)/float64(parallelTime))
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== 동시성 패턴: 파이프라인 (Pipeline) ===")

	basicPipeline()
	genericPipeline()
	imageProcessingPipeline()
	parallelPipelineStage()

	fmt.Println("\n=== 프로그램 정상 종료 ===")
}
