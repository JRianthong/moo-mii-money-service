# moo-mii-money-service

ระบบบันทึกรายรับ-รายจ่ายผ่าน LINE เขียนด้วย Go + Fiber + GORM + Supabase โดยวางโครงแบบ Practical DDD/CQRS

## Architecture

```text
cmd/api                         # composition root
internal/domain                 # aggregate, value objects, domain rules
internal/application/commands   # write use cases
internal/application/queries    # read use cases
internal/application/ports      # command/query ports owned by application layer
internal/infrastructure         # LINE and GORM/Postgres adapters
internal/interfaces/http        # Fiber routes/webhook boundary
migrations                      # SQL for Supabase
```

Dependency direction:

```text
interfaces/infrastructure -> application -> domain
```

## Domain Model

`Transaction` is the aggregate root for the first practical slice of the product. It is created through domain behavior:

```text
RecordIncome(...)
RecordExpense(...)
```

The domain avoids primitive obsession for important concepts:

```text
UserID
Money
Category
Note
SourceRef
OccurredAt
TransactionID
```

`SourceRef` stores the source system and LINE message id. The database has a unique index on `(source_system, source_message_id)` so LINE webhook retries do not create duplicate transactions.

`Category` belongs to a transaction type. Expense categories cannot be used for income transactions, and income categories cannot be used for expense transactions.

The `transactions` table has RLS enabled. The current service is designed to write/read through the Go backend using a server-side database connection, so no public `anon` or `authenticated` table policies are granted by default.

Default expense categories:

```text
อาหาร, เครื่องดื่ม, เดินทาง, ช้อปปิ้ง, ที่อยู่อาศัย, ค่าสาธารณูปโภค, สุขภาพ, การศึกษา, บันเทิง, ท่องเที่ยว, ครอบครัว, หนี้สิน, อื่น ๆ
```

Default income categories:

```text
เงินเดือน, ฟรีแลนซ์, ธุรกิจ, การลงทุน, โบนัส, ของขวัญ, เงินคืน, อื่น ๆ
```

## CQRS Shape

Commands write through `TransactionWriter`.

Queries read through `TransactionReader`.

They can still use the same Supabase database at this stage, but the application layer already treats writes and reads as separate contracts.

## LINE Commands

```text
จ่าย 120 ข้าวกลางวัน
รับ 2500 ฟรีแลนซ์
จ่าย 120 #เดินทาง grab
-45 กาแฟ
+3000 เงินเดือน
สรุป
ล่าสุด
หมวดหมู่
วิธีใช้
```

## Setup

1. Use the configured Supabase project: `https://iabdtjpltxsnuxmikssk.supabase.co`.
2. Apply the SQL files in `migrations/` in order. This project has already applied them through Supabase MCP.
3. Copy `.env.example` to `.env` or export the same variables.
4. Replace `<database-password>` in `DATABASE_URL` with the Supabase database password.
5. Set the LINE Messaging API webhook URL to:

```text
https://your-domain.example/webhooks/line
```

## Run

```bash
go mod tidy
go run ./cmd/api
```

## Deploy to Vercel

This project supports Vercel Go Functions through:

```text
api/healthz.go
api/webhooks/line.go
vercel.json
```

Set these environment variables in Vercel Project Settings:

```text
DATABASE_URL
LINE_CHANNEL_SECRET
LINE_CHANNEL_ACCESS_TOKEN
DEFAULT_CURRENCY
```

`DATABASE_URL` should use the Supabase database connection string with `sslmode=require`.

After deployment, use one of these URLs:

```text
https://<your-vercel-domain>/healthz
https://<your-vercel-domain>/webhooks/line
```

Set the LINE Messaging API webhook URL to:

```text
https://<your-vercel-domain>/webhooks/line
```

## Environment

```text
PORT=8080
SUPABASE_URL=https://iabdtjpltxsnuxmikssk.supabase.co
SUPABASE_PUBLISHABLE_KEY=sb_publishable_FBclZ4_E6KUq2E1I13am1w_KExvlpHN
DATABASE_URL=postgres://postgres:<database-password>@db.iabdtjpltxsnuxmikssk.supabase.co:5432/postgres?sslmode=require
LINE_CHANNEL_SECRET=...
LINE_CHANNEL_ACCESS_TOKEN=...
DEFAULT_CURRENCY=THB
```
