# 과제 3: JSON 설정 파일 파서

## 목표

JSON 태그, 중첩 구조체, 유효성 검사, 기본값 설정, 환경변수 오버라이드를 구현합니다.

## 요구사항

### 1. 설정 구조체 정의

```go
type Config struct {
    Server   ServerConfig   `json:"server"`
    Database DatabaseConfig `json:"database"`
    App      AppConfig      `json:"app"`
}

type ServerConfig struct {
    Host    string `json:"host"`
    Port    int    `json:"port"`
    Timeout int    `json:"timeout"` // 초 단위
    TLS     bool   `json:"tls"`
}

type DatabaseConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Name     string `json:"name"`
    User     string `json:"user"`
    Password string `json:"password,omitempty"`
    MaxConns int    `json:"max_connections"`
}

type AppConfig struct {
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Debug       bool     `json:"debug"`
    AllowedHosts []string `json:"allowed_hosts,omitempty"`
}
```

### 2. LoadConfig 함수 구현

```go
// LoadConfig는 JSON 파일을 읽어 Config 구조체로 파싱합니다.
// 파싱 후 기본값을 설정하고 유효성 검사를 수행합니다.
// 환경변수가 설정된 경우 해당 값으로 오버라이드합니다.
func LoadConfig(filename string) (*Config, error)
```

### 3. 기본값 설정

| 필드 | 기본값 |
|------|--------|
| `Server.Host` | `"localhost"` |
| `Server.Port` | `8080` |
| `Server.Timeout` | `30` |
| `Database.Port` | `5432` |
| `Database.MaxConns` | `10` |
| `App.Version` | `"1.0.0"` |

### 4. 환경변수 오버라이드

| 환경변수 | 설정 필드 |
|---------|----------|
| `APP_HOST` | `Config.Server.Host` |
| `APP_PORT` | `Config.Server.Port` |
| `DB_HOST` | `Config.Database.Host` |
| `DB_PORT` | `Config.Database.Port` |
| `DB_PASSWORD` | `Config.Database.Password` |

### 5. 유효성 검사

- `Server.Port`: 1~65535 범위
- `Database.Name`: 비어있으면 에러
- `App.Name`: 비어있으면 에러
- `Database.MaxConns`: 1 이상

## 실행 방법

```bash
go test -v
```

## 채점

```
=== 채점 결과 ===
통과: 18/20
점수: 90/100
```

## 힌트

- `os.ReadFile(filename)`: 파일 읽기
- `json.Unmarshal(data, &config)`: JSON 파싱
- `os.Getenv("VAR_NAME")`: 환경변수 읽기
- `strconv.Atoi(s)`: 문자열 → 정수 변환
- 참고 풀이: `solution/config.go`
