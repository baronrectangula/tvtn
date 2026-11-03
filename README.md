# TVTN: Tìm việc, tìm người

Tiếng Việt | [English](README.en.md)

## Mục lục

- [Setup và chạy](#setup-và-chạy)
  - [Docker Compose](#docker-compose)
  - [Chạy trực tiếp](#chạy-trực-tiếp)
- [Cấu hình](#cấu-hình)
- [Kiểm tra](#kiểm-tra)
- [Giấy phép](#giấy-phép)

## Setup và chạy

Chạy các lệnh dưới đây từ thư mục gốc dự án. Cấu hình và thông tin đăng nhập
local đã có sẵn. Nếu cần thay đổi, xem phần [Cấu hình](#cấu-hình).

### Docker Compose

Khuyến khích dùng cách này để chạy ứng dụng cùng các service hạ tầng.
Cần cài Docker và Docker Compose; có thêm Make thì dùng lệnh rút gọn:

```sh
docker compose --profile rest-server up -d --build
# Hoặc:
make compose profile=rest-server
```

REST server chạy ở `http://127.0.0.1:8080`. Để xem log hoặc dừng stack:

```sh
docker compose logs -f rest-server
docker compose --profile "*" down   # Giữ lại dữ liệu
```

### Chạy trực tiếp

Cần Go 1.27+. Docker Compose có sẵn cấu hình cho các service hạ tầng;
khởi động và đợi các service sẵn sàng bằng:

```sh
docker compose up -d --wait
```

Sau đó, dùng Bash/Zsh nạp thông tin đăng nhập rồi chạy server:

```sh
set -a
source .env.sample
set +a
go run ./cmd/rest-server -config config.toml -addr 127.0.0.1:8080
```

Nếu có Make, có thể thay lệnh `go run` bằng `make run`. Để chọn chương trình
hoặc cổng khác:

```sh
make run cmd=rest-server args="-addr 127.0.0.1:9090"
```

## Cấu hình

Ứng dụng dùng chung [config.toml](config.toml) cho cả hai cách chạy.

| Muốn đổi | Chỉnh ở đâu |
| --- | --- |
| Kết nối PostgreSQL | `host`, `port`, `database`, `ssl_mode` trong mục `[postgres]` của TOML |
| Host PostgreSQL theo môi trường | Biến `POSTGRES_HOST` ghi đè TOML; Compose đặt sẵn `postgres`, chạy trực tiếp mặc định dùng `127.0.0.1` |
| Tài khoản và mật khẩu | Biến `POSTGRES_USER`, `POSTGRES_PASSWORD`; ứng dụng không đọc hai giá trị này từ TOML |
| Địa chỉ HTTP | Cờ `-addr` khi chạy trực tiếp; mục `ports` của service `rest-server` khi dùng Compose |

Muốn dùng thông tin đăng nhập riêng, chép file mẫu rồi sửa các giá trị trong `.env`:

```sh
cp .env.sample .env
```

Compose tự đọc `.env`, ưu tiên biến đã export trong shell. Khi chạy trực tiếp,
thay `source .env.sample` bằng `source .env` trong hướng dẫn trên.
Git đã bỏ qua `.env`; giữ `.env.sample` làm mẫu local.

Đổi cấu hình xong thì chạy lại ứng dụng. Với Compose, chạy lại lệnh khởi động stack.
Nếu đổi database, sửa cả `database` trong TOML và `POSTGRES_DB` trong Compose.
Nếu PostgreSQL đã có dữ liệu, cần đổi database hoặc tài khoản, mật khẩu trong
PostgreSQL cho khớp; đổi biến môi trường không tự cập nhật dữ liệu đã khởi tạo.

## Kiểm tra

Server có hai endpoint: `/livez` kiểm tra HTTP, `/readyz` kiểm tra kết nối PostgreSQL.

```sh
curl --fail http://127.0.0.1:8080/livez
curl --fail http://127.0.0.1:8080/readyz
```

Chạy test bằng Go 1.27+:

```sh
go test ./...
# Hoặc:
make test
```

## Giấy phép

TVTN được phát hành theo [AGPL-3.0-only](LICENSE). Các dependency trong `vendor/` giữ giấy phép riêng.
