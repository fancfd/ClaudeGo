BINARY  := monitor
PORT    := 8080

.PHONY: build run run-file run-web test tidy clean

## build: 编译生成可执行文件
build:
	go build -o $(BINARY) .

## run: 控制台模式运行
run: build
	./$(BINARY) --console

## run-file: 控制台 + 文件双输出
run-file: build
	./$(BINARY) --console --file metrics.jsonl

## run-web: 控制台 + Web 服务器（默认端口 $(PORT)）
run-web: build
	./$(BINARY) --console --web --port $(PORT)

## test: 运行所有单元测试
test:
	go test ./...

## tidy: 整理 go.mod / go.sum 依赖
tidy:
	go mod tidy

## clean: 删除编译产物和日志文件
clean:
	rm -f $(BINARY) metrics.jsonl

## help: 列出所有可用目标
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
