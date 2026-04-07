package cache

import "fmt"

type CacheError struct {
	Code    string
	Message string
	Err     error
}

func (e *CacheError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("cache error %s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("cache error %s: %s", e.Code, e.Message)
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

var (
	ErrCacheMiss             = &CacheError{Code: "ERR_CACHE_MISS", Message: "cache miss"}
	ErrCacheExpired          = &CacheError{Code: "ERR_CACHE_EXPIRED", Message: "cache expired"}
	ErrCacheChecksumMismatch = &CacheError{Code: "ERR_CACHE_CHECKSUM_MISMATCH", Message: "cache checksum mismatch"}
	ErrCacheWriteFailed      = &CacheError{Code: "ERR_CACHE_WRITE_FAILED", Message: "cache write failed"}
	ErrCacheDiskSpaceLow     = &CacheError{Code: "ERR_CACHE_DISK_SPACE_LOW", Message: "cache disk space low"}
	ErrSourceUnavailable     = &CacheError{Code: "ERR_SOURCE_UNAVAILABLE", Message: "source unavailable"}
)

