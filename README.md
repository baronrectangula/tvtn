# TVTN: Tìm việc, tìm người

## Yêu cầu

- Go 1.27 trở lên.
- PostgreSQL 18, cài trực tiếp hoặc chạy bằng Docker hay công cụ khác.
- Docker Compose nếu dùng cấu hình Docker có sẵn; Make nếu muốn dùng lệnh có sẵn.

Chạy các lệnh bên dưới từ thư mục gốc dự án. Nếu tự chuẩn bị PostgreSQL, tạo user
và database rồi cập nhật `config.toml`, `POSTGRES_USER` và `POSTGRES_PASSWORD`
cho phù hợp.

## Chuẩn bị PostgreSQL

[docker-compose.yaml](docker-compose.yaml) có sẵn PostgreSQL cho môi trường local:

```sh
docker compose up -d --wait
# Hoặc:
make compose
```

## Chạy server

Nạp thông tin đăng nhập mẫu trong Bash hoặc Zsh, rồi chạy server:

```sh
set -a
source .env.sample
set +a
go run ./cmd/rest-server -config config.toml -addr 127.0.0.1:8080
```

Ứng dụng đọc thông tin đăng nhập từ biến môi trường, không tự nạp file `.env`.
Sau khi nạp biến môi trường, có thể thay lệnh `go run` bằng:

```sh
make run
# Chọn chương trình và truyền thêm cờ:
make run cmd=rest-server args="-addr 127.0.0.1:9090"
```

Server có `/livez` và `/readyz` để kiểm tra trạng thái. Nhấn Ctrl+C để dừng.

## Chạy test

```sh
go test ./...
# Hoặc:
make test
```
