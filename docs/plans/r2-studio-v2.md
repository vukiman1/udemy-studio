# Plan: studio-v2 — học mọi nơi qua Cloudflare R2

Ngày lập: 2026-11-25 · Nhánh tích hợp: `feature/r2` (tách từ `develop`)

## 1. Mục tiêu

```
MAC: studio (hiện tại)        MAC: localv2                  R2 (riêng tư)               VPS: vpsv2
tải · dịch · nén       ──▶    "Đưa lên R2" 1 khoá   ──PUT──▶  studio.json + courses/…  ◀──GET── dựng thư viện
                              (≤ 9,5 GB)                                              ◀─ link ký ─ trình duyệt xem video
khoá học xong ──▶ đóng gói ra Google Drive / ổ cứng (phase sau)
```

- Mỗi lúc chỉ **một khoá** nằm trên R2 (gói miễn phí 10 GB lưu trữ). Đưa khoá mới lên thì khoá cũ bị xoá.
- **vpsv2** chỉ đọc: học, ghi tiến độ và ghi chú trên VPS. Không tải, không dịch, không nén.
- Kết nối R2 (S3 API URL + Access Key ID + Secret) được **dán qua giao diện**, nên đổi bucket hay đổi tài khoản Cloudflare chỉ cần dán bộ mới.

## 2. Nguyên tắc an toàn (không đụng bản đang chạy)

| Hạng mục | Bản hiện tại (giữ nguyên) | Bản v2 |
|---|---|---|
| Nhánh | `develop` → `main` | `feature/r2`; mọi PR R2 vào nhánh này |
| Image | `ghcr.io/vukiman1/udemy-studio:latest` (CI trên `main`) | tag riêng `:v2-dev`, **không bao giờ** `latest` |
| Compose | `~/udemy-studio/compose.yaml` | `compose.v2-local.yaml` (Mac), `compose.vps.yaml` (VPS); tên project riêng |
| Cổng | 8765 | localv2: 8766 |
| Dữ liệu | `~/udemy-studio/data` | `~/udemy-studio-v2/data` (Mac), `data/` riêng trên VPS |
| Thư viện | đọc–ghi | localv2 gắn **`:ro`** (chỉ đọc); vpsv2 không gắn thư viện |
| Bucket | — | `studio-dev` để thử; bucket thật chỉ dùng sau phase 5 |
| Token R2 | — | localv2: **Object Read & Write**; vpsv2: **Object Read only**; cả hai chỉ giới hạn vào 1 bucket |
| Tên miền | `studio.kimandev.site` | `learn-dev.kimandev.site`, đặt sau Cloudflare Access |

- **Bẫy CI cần tránh:** workflow `docker-image.yml` có `workflow_dispatch` và luôn đẩy tag `latest`. **Không chạy tay workflow này trên `feature/r2`.** Image v2 build bằng workflow riêng hoặc build tay (xem phase 0).
- `develop` luôn sẵn sàng phát hành. Thỉnh thoảng gộp `develop` vào `feature/r2` để không lệch xa.
- Chỉ khi luồng localv2 ↔ vpsv2 chạy trọn vẹn (phase 4 đạt) mới gộp `feature/r2` → `develop` → `main`.

## 3. Thiết kế

### 3.1 Vai trò mỗi bản (biến môi trường, cố định theo nơi triển khai)

- `UDEMY_STUDIO_ROLE=library` (mặc định): như hiện nay; có thêm nút "Đưa lên R2".
- `UDEMY_STUDIO_ROLE=cloud`: thư viện đọc từ R2; ẩn Tải, Dịch, Nén, Đưa lên R2.

### 3.2 Kết nối R2 (dán qua giao diện, trang "Lưu trữ đám mây")

- **S3 API URL**: `https://<account-id>.r2.cloudflarestorage.com/<bucket>`. App tự tách account id và tên bucket.
- **Access Key ID** và **Secret Access Key**: mã hoá bằng `sealedbox` như khoá AI; chỉ hiện gợi ý dạng `abcd••••`.
- Nút **Kiểm tra kết nối**:
  1. Đọc thử bucket.
  2. Ghi thử một object nhỏ `.studio-probe` rồi xoá ngay, để biết token **chỉ đọc** hay **đọc–ghi**.
  3. Báo nếu quyền không hợp vai trò, ví dụ bản `library` mà token chỉ đọc thì không đẩy được.

### 3.3 Cấu trúc bucket

```
studio.json                       mục lục; ghi SAU CÙNG khi đẩy xong
courses/<course-dir>/…            bản sao nguyên thư mục khoá (video, phụ đề, bài đọc, tài liệu, cover.jpg, .course.json)
```

`studio.json`:
```json
{
  "format": 1,
  "publishedAt": "2026-11-25T10:00:00Z",
  "course": { "dir": "fundamentals-…", "title": "Fundamentals of …", "sizeBytes": 6400000000 },
  "files": [{ "path": "01 - Introduction/001 Welcome.mp4", "sizeBytes": 12345, "etag": "…" }]
}
```

- vpsv2 chỉ đọc **một object** để dựng thư viện, không phải liệt kê bucket (liệt kê là thao tác Class A, đắt hơn).
- Đổi khoá: xoá khoá cũ → tải khoá mới → ghi `studio.json`. vpsv2 không bao giờ thấy khoá đang tải dở.

### 3.4 Phục vụ trên vpsv2

- **Video** (`/media/{id}`): trả 302 sang **link có chữ ký** (SigV4, hết hạn sau 6 giờ). Trình duyệt tải thẳng từ R2, không qua VPS. Thẻ `<video>` khác tên miền không cần CORS.
- **Phụ đề, bài đọc (chèn theme), tài liệu, ảnh bìa:** vpsv2 đọc từ R2 rồi trả về; các file này nhỏ.
- **Tiến độ, ghi chú:** lưu ở `data/` của VPS. Mã bài học lấy theo Udemy (manifest), nên sau này gộp tiến độ về Mac được (ngoài phạm vi plan này).

### 3.5 Đẩy khoá từ localv2

- Tác vụ chạy nền, có tiến độ và nút huỷ, theo mẫu Dịch/Nén.
- Kiểm tra trước: tổng dung lượng ≤ **9,5 GB**; quá thì báo và gợi ý nén.
- Tải lên theo từng mảnh (multipart) cho file lớn. Chạy lại thì **bỏ qua file đã có cùng kích thước và etag** (tải tiếp khi bị ngắt).
- Khoá tải bằng bản Python cũ (chưa có manifest đầy đủ): cảnh báo nên tải lại qua app trước, để mã bài học ổn định.

### 3.6 Thư viện ngoài (cần duyệt)

- **aws-sdk-go-v2** (`service/s3`, `feature/s3/manager`): S3 client, multipart, presign. Cloudflare có hướng dẫn chính thức cho nó.
- Kiểm thử tích hợp: **SeaweedFS** (server tương thích S3) chạy trong Docker trên máy dev qua `task test:s3`, để test không tốn quota R2 và không cần mạng. MinIO đã ngừng phát hành image công khai nên không dùng được.

## 4. Các phase

Mỗi phase là một hoặc vài PR vào `feature/r2`. **Chỉ sang phase sau khi tiêu chí đạt của phase trước đã đạt.**

### Phase 0: Hạ tầng thử nghiệm (chưa có code tính năng)

- [ ] Nhánh `feature/r2` (đã tạo).
- [ ] `compose.v2-local.yaml`: project `udemy-studio-v2`, cổng `127.0.0.1:8766`, `data` riêng, thư viện gắn `:ro`, `ROLE=library`, image `:v2-dev`.
- [ ] `compose.vps.yaml`: `ROLE=cloud`, không gắn thư viện, có `cloudflared` (profile `tunnel`).
- [ ] Image `:v2-dev`. Chọn một trong hai:
  - (a) workflow riêng `docker-image-v2.yml`: chạy khi push `feature/r2`, đẩy tag `v2-dev`, concurrency group riêng (**sửa CI, cần duyệt**);
  - (b) build tay bằng `docker buildx --platform linux/amd64,linux/arm64`.
- [ ] Kim An tạo bucket `studio-dev`, 2 token (đọc–ghi và chỉ đọc), cài Docker trên VPS, tạo tunnel cho `learn-dev.kimandev.site` kèm Access.
- **Đạt khi:** localv2 và vpsv2 đều chạy (chưa có R2); bản hiện tại ở cổng 8765 không bị ảnh hưởng.

### Phase 1: Kết nối R2

- Backend: domain `objectstore` (cổng S3 dạng interface), adapter R2/S3, store cấu hình mã hoá, API `GET/PUT/DELETE /api/storage`, `POST /api/storage/check`.
- Web: trang "Lưu trữ đám mây" (theo thiết kế trang Cài đặt AI), mục điều hướng mới.
- Test: tách S3 URL (URL hợp lệ, sai, thiếu bucket); khoá được mã hoá; kiểm tra quyền đọc / đọc–ghi bằng SeaweedFS.
- **Đạt khi:** dán bộ thông tin `studio-dev` vào cả hai bản → "Kiểm tra" báo đúng quyền (localv2: đọc–ghi; vpsv2: chỉ đọc).

### Phase 2: Đẩy khoá từ localv2

- Tác vụ "Đưa lên R2": kiểm tra dung lượng, xoá khoá cũ, multipart, tải tiếp khi bị ngắt, ghi `studio.json` sau cùng, huỷ giữa chừng.
- Web: nút trên thẻ khoá, hộp thoại kế hoạch/tiến độ/kết quả (mẫu hộp thoại Nén), nhãn "Đang trên R2" ở thẻ khoá.
- Test (SeaweedFS):
  - đẩy trọn một khoá nhỏ;
  - chạy lại thì không tải lại gì;
  - huỷ giữa chừng thì `studio.json` cũ còn nguyên;
  - vượt ngưỡng thì từ chối;
  - đổi khoá thì khoá cũ bị xoá.
- **Đạt khi:** đẩy một khoá thử (vài bài) lên `studio-dev`; nhìn trong dashboard R2 thấy đúng cấu trúc.

### Phase 3: vpsv2 đọc từ R2

- Thư viện dựng từ `studio.json`; `fs.FS` trên R2 cho file nhỏ; `/media` chuyển sang link ký; `ROLE=cloud` ẩn các nút ghi.
- Test: dựng thư viện từ mục lục mẫu; link ký đúng định dạng và hạn; bản `cloud` trả 404/403 cho API ghi.
- **Đạt khi:** trên VPS mở `learn-dev`: thấy khoá, xem video (tua được), phụ đề song ngữ, bài đọc, ghi chú, tiến độ lưu được.

### Phase 4: Hoàn thiện và thử trọn luồng

- Trình phát tự xin link mới khi link hết hạn (tạm dừng lâu rồi phát lại).
- Thử trọn luồng trên điện thoại và máy khác mạng: đẩy khoá A → học trên vpsv2 → đẩy khoá B (A bị xoá) → vpsv2 tự thấy B.
- Đo thao tác Class A/B một ngày học bình thường, so với hạn mức miễn phí.
- **Đạt khi:** checklist thử tay đạt trên Mac, điện thoại và VPS.

### Phase 5: Phát hành

- Gộp `develop` vào `feature/r2`, chạy lại toàn bộ test → PR `feature/r2` → `develop` → `main`.
- Image chính `latest` có cả hai vai trò. Chuyển vpsv2 sang `latest`; tạo bucket thật nếu muốn.
- Cập nhật README: hướng dẫn tạo bucket, token, compose VPS.

### Phase 6 (sau): Đóng gói khoá học xong

- Lệnh hoặc nút "Đóng gói": nén thư mục khoá kèm tiến độ/ghi chú, xuất ra ổ cứng hoặc đẩy lên Google Drive (qua rclone).

## 5. Rủi ro và cách xử lý

| Rủi ro | Cách xử lý |
|---|---|
| Vượt 10 GB, bị tính phí | Ngưỡng 9,5 GB; xoá khoá cũ trước khi đẩy; chỉ dùng lưu trữ Standard (gói miễn phí không áp dụng cho Infrequent Access) |
| Lộ khoá trên VPS | Token chỉ đọc, giới hạn 1 bucket; khoá mã hoá trong DB; app sau Cloudflare Access |
| Link ký hết hạn giữa chừng | Hạn 6 giờ; trình phát tự xin link mới (phase 4) |
| Đẩy dở làm hỏng bản đang học | `studio.json` ghi sau cùng; huỷ hoặc lỗi thì mục lục cũ giữ nguyên |
| Lỡ đụng image hoặc thư viện thật | Tag `v2-dev`; thư viện `:ro`; không chạy tay `docker-image.yml` trên `feature/r2` |
| Tiến độ hai nơi lệch nhau | Chấp nhận trong plan này; manifest có mã bài Udemy nên gộp được về sau |

## 6. Quyết định đã chốt (2026-11-25)

1. Dùng **aws-sdk-go-v2**; test tích hợp dùng **SeaweedFS** (ban đầu chốt MinIO, đổi ngày 2026-11-26 vì MinIO ngừng phát hành image).
2. Image `:v2-dev` do **workflow CI riêng** `docker-image-v2.yml` build: chạy khi push `feature/r2`, concurrency group riêng, chỉ đẩy tag `v2-dev`.
3. VPS: 1 CPU, 2 GB RAM, trống 5 GB, có Docker. Đủ cho vpsv2: video nằm trên R2 và đi thẳng tới trình duyệt; VPS chỉ phục vụ giao diện, phụ đề, tiến độ.
4. Ngưỡng **9,5 GB**; link ký hết hạn sau **6 giờ**.

## 7. Ba cách triển khai, chung một image

| Trường hợp | Video ở đâu | Vai trò |
|---|---|---|
| 1. Mac (hiện tại) | Ổ Mac | `library` |
| 2. Mac + R2 + VPS nhỏ (plan này) | R2 | Mac: `library` (có nút "Đưa lên R2"); VPS: `cloud` |
| 3. VPS đủ lớn, không cần R2 | Ổ VPS | `library`, **y như trường hợp 1** (compose cũ, gắn thư viện, tunnel, Access) |

- Trường hợp 3 đã chạy được với code hiện tại. Plan này phải giữ ổ đĩa là nguồn mặc định: vai trò `cloud` chỉ là một nguồn thêm vào, nút "Đưa lên R2" chỉ hiện khi đã cấu hình R2.
- Trường hợp 3 trên VPS: tải từ Udemy bằng IP trung tâm dữ liệu dễ bị chặn hơn (khi đó tải ở Mac rồi `rsync` sang); nén H.265 trên 1 CPU rất chậm, nên nén ở Mac trước.
