rm -rf mocks/mock_*.go

bin/mockgen -source internal/utils/time.go -destination mocks/mock_time.go -package mocks
