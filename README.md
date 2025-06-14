# BE Invoice API

Ứng dụng API quản lý bán hàng, viết bằng [Fiber](https://gofiber.io/) và MongoDB. Dự án cung cấp các chức năng cơ bản như quản lý người dùng, sản phẩm, hóa đơn và thông tin cửa hàng. File `Dockerfile` và `docker-compose.yml` giúp khởi chạy nhanh môi trường phát triển.

## Biến môi trường
Tạo file `.env` ở thư mục gốc dự án (có thể sao chép từ file `env`) và điền các giá trị sau:

```env
MONGO_URL= # Chuỗi kết nối MongoDB (ví dụ: mongodb://user:pass@localhost:27017)
MONGO_NAME=test2 # Tên database
JWT_SECRET=test # Chuỗi ký JWT
PORT=4000 # Cổng API
MINIO_ACCESS_KEY=al8KsxHAbLtfNVsX # Access key MinIO
MINIO_SECRET_KEY=noWZ40KlvEcioZcPhLmMZFcPSkdeuX0K # Secret key MinIO
MINIO_ENDPOINT=image.nghia.myds.me # Host MinIO
MINIO_BUCKET=test # Tên bucket lưu trữ
MINIO_SSL=true # Sử dụng HTTPS (true/false)
```

## Cài đặt
1. Cài đặt [Go](https://go.dev/) >= 1.24 và Docker nếu muốn chạy bằng container.
2. Tạo file `.env` như hướng dẫn trên.
3. Cài đặt phụ thuộc bằng lệnh:
   ```bash
   go mod tidy
   ```
4. Chạy bằng Go:
   ```bash
   go run main.go
   ```
   hoặc khởi động bằng Docker Compose:
   ```bash
   docker-compose up --build
   ```

Nếu muốn build thành file thực thi:
```bash
go build -o backend.exe main.go
```

Docker image có sẵn tại `trungnghia1996/be_invoice:latest`:
```bash
docker run -d -p 4000:4000 --env-file .env trungnghia1996/be_invoice:latest
```

## docker-compose.yml mẫu
Nội dung file `docker-compose.yml` giúp khởi chạy nhanh môi trường phát triển:

```yaml
version: '3.8'

services:
  backend:
    build: .
    container_name: go-fiber-api
    ports:
      - 4000:4000
    environment:
      MONGO_URL: mongodb://admin:cr969bp6x6@mongo:27017
      MONGO_NAME: test
      JWT_SECRET: test
      PORT: 4000
      MINIO_ACCESS_KEY: al8KsxHAbLtfNVsX
      MINIO_SECRET_KEY: noWZ40KlvEcioZcPhLmMZFcPSkdeuX0K
      MINIO_ENDPOINT: image.nghia.myds.me
      MINIO_BUCKET: test
      MINIO_SSL: true
    restart: unless-stopped
    networks:
      - app-network
    depends_on:
      - mongo

  mongo:
    image: mongo:6
    container_name: mongo
    networks:
      - app-network
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: cr969bp6x6
    expose:
      - "27017"
    restart: unless-stopped

networks:
  app-network:
    driver: bridge
```

## Các endpoint chính
Tất cả các endpoint (ngoại trừ `/login` và `/test`) đều yêu cầu header `Authorization: Bearer <token>`.

| METHOD | PATH | Mô tả | Dữ liệu vào (ví dụ) |
|--------|------|-------|----------------------|
|`POST`|`/login`|Đăng nhập|`{"username":"admin","password":"123"}`|
|`GET`|`/test`|Kiểm tra server|-|
|`GET`|`/api/test2`|Kiểm tra token hợp lệ|-|
|`PUT`|`/api/presigned_url`|Lấy URL upload file|`{"key":"logo.png"}`|
|`POST`|`/api/users`|Tạo người dùng|`{"username":"u1","password":"pw","role":"member"}`|
|`GET`|`/api/users?role=member`|Lấy danh sách người dùng|-|
|`PUT`|`/api/users/password`|Đổi mật khẩu|`{"old_password":"a","new_password":"b"}`|
|`PUT`|`/api/users`|Cập nhật người dùng|`{"id":"...","username":"u1","role":"admin"}`|
|`DELETE`|`/api/users?id=1,2`|Xoá người dùng|-|
|`GET`|`/api/products`|Danh sách sản phẩm (phân trang)|-|
|`POST`|`/api/products`|Tạo sản phẩm|`{"name":"sp A","price":10000}`|
|`PUT`|`/api/products`|Cập nhật sản phẩm|`{"id":"...","name":"sp","price":20000}`|
|`DELETE`|`/api/products?id=a,b`|Xoá sản phẩm|-|
|`POST`|`/api/invoices`|Tạo hoá đơn mới|`{"items":[{"productId":"...","name":"Áo","quantity":1,"price":10000}]}`|
|`DELETE`|`/api/invoices?id=a,b`|Xoá hoá đơn|-|
|`GET`|`/api/invoices`|Lọc hoá đơn theo ngày và code|-|
|`PUT`|`/api/invoices`|Cập nhật hoá đơn|`{"id":"...","items":[]}`|
|`GET`|`/api/settings`|Lấy thông tin cửa hàng|-|
|`PUT`|`/api/settings`|Cập nhật thông tin cửa hàng|`{"storeName":"Shop"}`|

Mọi phản hồi đều theo cấu trúc:

```json
{
  "status": "success",
  "message": "...",
  "data": {}
}
```

Ứng dụng mặc định chạy tại `http://localhost:4000`.
