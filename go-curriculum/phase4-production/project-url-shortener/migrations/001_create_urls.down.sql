-- migrations/001_create_urls.down.sql
-- URL 단축 서비스 초기 스키마 롤백
-- 실행: migrate -database $DATABASE_URL -path migrations down 1

-- 트리거 제거
DROP TRIGGER IF EXISTS short_urls_updated_at ON short_urls;

-- 함수 제거
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 인덱스 제거 (테이블 삭제 시 자동으로 제거되지만 명시적으로 작성)
DROP INDEX IF EXISTS idx_short_urls_created_at;
DROP INDEX IF EXISTS idx_short_urls_expires_at;
DROP INDEX IF EXISTS idx_short_urls_short_code;

-- 테이블 제거
DROP TABLE IF EXISTS short_urls;
