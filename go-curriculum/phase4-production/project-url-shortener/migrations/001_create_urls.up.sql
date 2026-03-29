-- migrations/001_create_urls.up.sql
-- URL 단축 서비스 초기 스키마 생성
-- 실행: migrate -database $DATABASE_URL -path migrations up

-- UUID 확장 활성화 (선택적, 대신 serial 사용)
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- short_urls 테이블: 단축 URL 정보를 저장합니다.
CREATE TABLE IF NOT EXISTS short_urls (
    id           BIGSERIAL PRIMARY KEY,
    original_url TEXT        NOT NULL,
    short_code   VARCHAR(20) NOT NULL,
    click_count  BIGINT      NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NULL,

    -- 단축 코드는 전역적으로 유일해야 합니다.
    CONSTRAINT short_urls_short_code_unique UNIQUE (short_code),

    -- URL 길이 제한 (일반적으로 2048자 이내)
    CONSTRAINT short_urls_original_url_length CHECK (LENGTH(original_url) <= 2048),

    -- 단축 코드 형식 검증 (영문자, 숫자, 하이픈, 밑줄)
    CONSTRAINT short_urls_short_code_format CHECK (short_code ~ '^[a-zA-Z0-9_-]+$')
);

-- 단축 코드 인덱스: 리다이렉트 시 빠른 조회를 위해 필수
CREATE INDEX IF NOT EXISTS idx_short_urls_short_code ON short_urls (short_code);

-- 만료 시각 인덱스: 만료된 URL 정리 작업에 사용
CREATE INDEX IF NOT EXISTS idx_short_urls_expires_at ON short_urls (expires_at)
    WHERE expires_at IS NOT NULL;

-- 생성 시각 인덱스: 목록 조회 시 최신순 정렬에 사용
CREATE INDEX IF NOT EXISTS idx_short_urls_created_at ON short_urls (created_at DESC);

-- updated_at 자동 갱신 함수
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- updated_at 자동 갱신 트리거
CREATE TRIGGER short_urls_updated_at
    BEFORE UPDATE ON short_urls
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 샘플 데이터 (개발/테스트용)
-- INSERT INTO short_urls (original_url, short_code) VALUES
--     ('https://golang.org', 'golang'),
--     ('https://github.com', 'github');
