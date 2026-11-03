# TVTN: Tìm việc, tìm người

![Status: in development](https://img.shields.io/badge/status-in%20development-yellow)
[![Test](https://github.com/baronrectangula/tvtn/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/baronrectangula/tvtn/actions/workflows/test.yaml)
[![Lint](https://github.com/baronrectangula/tvtn/actions/workflows/lint.yaml/badge.svg?branch=main)](https://github.com/baronrectangula/tvtn/actions/workflows/lint.yaml)
[![License: AGPL-3.0-only](https://img.shields.io/badge/license-AGPL--3.0--only-blue)](LICENSE)

Tiếng Việt | [English](README.en.md)

## Mục lục

- [Setup và chạy](#setup-và-chạy)
  - [Docker Compose](#docker-compose)
  - [Chạy trực tiếp](#chạy-trực-tiếp)
- [Cấu hình](#cấu-hình)
- [Kiểm tra](#kiểm-tra)
- [Git hooks](#git-hooks)
  - [Commit message](#commit-message)
  - [Xác nhận thay đổi cấu hình](#xác-nhận-thay-đổi-cấu-hình)
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

## Git hooks

Chạy `make hooks` một lần sau khi clone repo.

### Commit message

CI bỏ qua merge commit khi kiểm tra định dạng commit message và footer cấu hình, kể cả chỉnh sửa chỉ phát sinh khi giải quyết xung đột merge. Các commit thông thường vẫn phải tuân theo Conventional Commits và có footer cấu hình khi cần.

Hook và CI kiểm tra Conventional Commit theo dạng `type(scope): mô tả`, với `type` là `feat`, `chore`,
`fix`, `test`, `doc`, `refactor` hoặc `perf` (chữ thường). `scope` không bắt buộc. Body và footer
cách tiêu đề một dòng trống. Commit có dấu `!` phải có footer giải thích thay đổi
không tương thích; cũng có thể dùng footer này mà không thêm `!`:

```text
feat(api)!: đổi định dạng phản hồi

BREAKING CHANGE: Trường user_id được đổi thành id.
```

Có thể dùng `BREAKING-CHANGE` thay cho `BREAKING CHANGE`.

### Xác nhận thay đổi cấu hình

Khi commit thay đổi đã stage trong
`config.toml` hoặc `.env.sample`, kiểm tra lại nội dung rồi thêm một trong hai footer
sau một dòng trống:

```text
Config-Change: Tôi xác nhận cấu hình và thông tin đăng nhập này chỉ dùng để phát triển trên máy local.
```

Footer phải khớp chính xác, có phân biệt chữ hoa và chữ thường. Hook này nhắc người
commit kiểm tra thông tin local; không tự phát hiện secret và có thể bị bỏ qua bằng
`--no-verify`.

CI chạy lint, test và kiểm tra footer khi push lên `main` hoặc có PR từ nhánh trong cùng repo.
Commit mới trên PR tự chạy lại các check; push nhánh tính năng không chạy thêm một lượt riêng.
Các job, kể cả thông báo, đều bỏ qua PR từ fork và bản repo được fork.
PR kiểm tra các commit chưa có trên nhánh đích; push kiểm tra các commit mới trong lần push đó.
Lần push đầu của một nhánh hoặc sau khi viết lại lịch sử chỉ kiểm tra các commit chưa có trên nhánh mặc định.

Khi commit đổi cấu hình mà thiếu footer xác nhận hợp lệ, bot comment trên PR
(hoặc commit nếu không có PR), nhắc tên tác giả PR, chủ repo cá nhân và các admin
(mỗi người chỉ một lần), kèm link
từng commit vi phạm và link check. Lỗi chạy CI không gửi thông báo này. Workflow thông báo
phải có trên nhánh mặc định; thông báo cá nhân còn tùy cài đặt GitHub của người nhận.

## Giấy phép

TVTN được phát hành theo [AGPL-3.0-only](LICENSE). Các dependency trong `vendor/` giữ giấy phép riêng.
