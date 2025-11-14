# Order Service -- Full Commerce Layer Documentation

## 1. Thành phần cần thêm

  -----------------------------------------------------------------------
  Thành phần cần thêm        Mục đích                  Bắt buộc / Khuyến
                                                       nghị
  -------------------------- ------------------------- ------------------
  **orders**                 Lưu đơn hàng trước thanh  **Bắt buộc**
                             toán                      

  **order_items**            Lưu các khóa học trong    **Bắt buộc**
                             đơn hàng                  

  **payment_intents /        Lưu trạng thái thanh toán **Bắt buộc**
  payments**                 từ Stripe                 

  **Stripe webhook logs**    Idempotency & xử lý       **Bắt buộc**
                             webhook                   

  **entitlement service /    Ghi nhận enroll khi thanh **Bắt buộc**
  enrollment trigger**       toán thành công           

  **discounts / coupons /    Giảm giá                  Khuyến nghị
  promotions**                                         

  **invoices / receipts**    Hóa đơn                   Khuyến nghị

  **refund requests**        Hoàn tiền khóa học        Khuyến nghị

  **audit logs**             Chống gian lận            Khuyến nghị

  **fraud checks**           Kiểm tra risk trước khi   Optional (quan
                             thanh toán                trọng khi scale)
  -----------------------------------------------------------------------

## 2. Vì sao cần commerce layer?

Hệ thống hiện tại chỉ có phần học tập (courses, enrollments...).\
Để bán khóa học, cần một lớp thương mại gồm các chức năng:

-   Tạo order (1 hoặc nhiều khóa học)
-   Tạo Stripe PaymentIntent
-   Xử lý webhook để xác nhận thanh toán
-   Kích hoạt enroll khi thanh toán thành công
-   Quản lý hoàn tiền
-   Audit logs để điều tra lỗi/gian lận

------------------------------------------------------------------------

# 3. Chi tiết từng bảng

## 3.1 Bảng `orders`

Order là bước trung gian **trước khi thanh toán**.

### Fields

-   id (uuid, pk)\
-   user_id\
-   total_amount\
-   currency\
-   status:
    -   created\
    -   pending_payment\
    -   paid\
    -   failed\
    -   cancelled\
-   payment_intent_id (Stripe)\
-   created_at\
-   updated_at\
-   expires_at (chống spam đơn hàng)

### Lý do cần:

-   Giữ trạng thái xuyên suốt quá trình thanh toán.
-   Không rely vào course → enroll trực tiếp (dễ lỗi, không scale).

------------------------------------------------------------------------

## 3.2 Bảng `order_items`

Đơn hàng có thể chứa nhiều khóa học.

### Fields

-   id\
-   order_id\
-   course_id\
-   price_snapshot\
-   quantity\
-   created_at

### Lý do cần:

-   Snapshot giá tại thời điểm mua.
-   Không bị ảnh hưởng khi instructor đổi giá sau này.

------------------------------------------------------------------------

## 3.3 Bảng `payments`

Stripe tạo PaymentIntent → backend phải lưu.

### Fields

-   id\
-   order_id\
-   stripe_payment_intent_id\
-   amount\
-   currency\
-   status\
-   raw_webhook (jsonb)

### Lý do:

Stripe webhook có thể đến trễ hoặc sai thứ tự → cần idempotency.

------------------------------------------------------------------------

## 3.4 Bảng `webhook_events`

Đảm bảo mỗi event từ Stripe chỉ xử lý đúng 1 lần.

### Fields

-   id\
-   stripe_event_id (unique)\
-   type\
-   payload jsonb\
-   processed boolean\
-   processed_at\
-   created_at

------------------------------------------------------------------------

## 3.5 Cấp quyền học (Enrollment Trigger)

Flow:

    payment_intent.succeeded → order-service publish "order.paid"
    → enrollment-service xử lý và thêm vào course_enrollments

Order-service **không trực tiếp** insert vào enrollments để:

-   tách domain
-   dễ mở rộng (subscription, bundle)
-   dễ audit và xử lý async

------------------------------------------------------------------------

## 3.6 Coupons

### coupons

-   id\
-   code\
-   percent_off / amount_off\
-   expires_at\
-   max_usage\
-   per_user_limit\
-   applicable_course_ids

### coupon_redemptions

-   id\
-   coupon_id\
-   user_id\
-   order_id\
-   redeemed_at

------------------------------------------------------------------------

## 3.7 Invoices

Mặc dù Stripe tạo invoice, backend vẫn nên lưu record nội bộ.

### invoices

-   id\
-   order_id\
-   user_id\
-   total_amount\
-   issued_at\
-   pdf_url\
-   stripe_charge_id

------------------------------------------------------------------------

## 3.8 Refunds

### refund_requests

-   id\
-   order_id\
-   user_id\
-   reason\
-   status\
-   stripe_refund_id

Flow:

    User request refund →
      order-service gửi Stripe refund API →
        Stripe webhook refund.succeeded →
          revoke enrollment (tùy policy)

------------------------------------------------------------------------

## 3.9 Fraud Detection

### fraud_logs

-   id\
-   user_id\
-   order_id\
-   risk_level\
-   details (IP, device, anomaly...)

Dùng để bắt: - nhiều charge rồi refund - card testing - abuse coupon

------------------------------------------------------------------------

## 3.10 Audit Logs

### audit_logs

-   actor_id\
-   action\
-   resource_type\
-   resource_id\
-   before json\
-   after json\
-   created_at

Audit cực quan trọng để kiểm tra: - thay đổi trạng thái order - xử lý
webhook lỗi

------------------------------------------------------------------------

# 4. API cần có trong Order-Service

## Public API

### `POST /orders`

Tạo order + order_items.

### `POST /orders/:id/pay`

Tạo Stripe PaymentIntent.

### `GET /orders/:id`

Xem chi tiết đơn hàng.

### `POST /stripe/webhook`

Nhận webhook từ Stripe.

------------------------------------------------------------------------

## Internal Events (Kafka/NATS)

-   order.created\
-   order.pending_payment\
-   order.paid\
-   order.failed\
-   refund.requested\
-   refund.completed

------------------------------------------------------------------------

# 5. Full logic từ order → payment → enroll

    User chọn khóa học
        ↓
    Order-service → tạo order + order_items
        ↓
    Order-service → tạo Stripe PaymentIntent
        ↓
    Frontend redirect hoặc sử dụng PaymentElement
        ↓
    Stripe tính tiền
        ↓
    Stripe gửi webhook payment_intent.succeeded
        ↓
    Order-service update order → paid
        ↓
    Order-service publish event → "order.paid"
        ↓
    Enrollment-service tạo course_enrollments
        ↓
    Notification-service gửi email thành công

------------------------------------------------------------------------

# 6. Kết luận

Order-service là lớp thương mại trung gian bảo vệ toàn bộ hệ thống học
tập.\
Design này đảm bảo:

-   chống double payment
-   chống xử lý webhook lặp
-   RESILIENT trước lỗi Stripe
-   Dễ mở rộng subscription hoặc bundle
-   100% production-grade theo tiêu chuẩn của Udemy, Coursera,
    Skillshare.
