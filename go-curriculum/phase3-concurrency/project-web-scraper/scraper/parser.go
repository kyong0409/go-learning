// 패키지 scraper
package scraper

import (
	"bytes"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// ParseHTML은 HTML 바이트에서 페이지 제목과 링크 목록을 추출합니다.
// baseURL은 상대 경로를 절대 경로로 변환하는 데 사용됩니다.
func ParseHTML(body []byte, baseURL string) (title string, links []string) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", nil
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return "", nil
	}

	seen := make(map[string]bool) // 중복 링크 제거

	// HTML 트리를 재귀적으로 순회
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				// <title> 텍스트 추출
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					title = strings.TrimSpace(n.FirstChild.Data)
				}

			case "a":
				// <a href="..."> 링크 추출
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						link := resolveURL(base, attr.Val)
						if link != "" && !seen[link] {
							seen[link] = true
							links = append(links, link)
						}
						break
					}
				}
			}
		}

		// 자식 노드 순회
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return title, links
}

// resolveURL은 상대 URL을 절대 URL로 변환합니다.
// 유효하지 않거나 스크레이핑 대상이 아닌 URL은 빈 문자열을 반환합니다.
func resolveURL(base *url.URL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}

	// 스킵할 스킴
	lower := strings.ToLower(href)
	for _, skip := range []string{"javascript:", "mailto:", "tel:", "#", "data:"} {
		if strings.HasPrefix(lower, skip) {
			return ""
		}
	}

	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(ref)

	// HTTP/HTTPS만 허용
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}

	// 쿼리 파라미터와 프래그먼트 제거 (선택적 단순화)
	resolved.Fragment = ""

	return resolved.String()
}

// ExtractText는 HTML에서 순수 텍스트를 추출합니다.
func ExtractText(body []byte) string {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return ""
	}

	var sb strings.Builder

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		// script, style 태그 내용은 건너뜀
		if n.Type == html.ElementNode {
			if n.Data == "script" || n.Data == "style" {
				return
			}
		}

		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				sb.WriteString(text)
				sb.WriteByte(' ')
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return strings.TrimSpace(sb.String())
}
