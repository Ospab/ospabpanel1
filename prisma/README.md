# Prisma Integration Notes

This Go project currently uses `database/sql` directly. Prisma is a Node/TypeScript ORM. To leverage Prisma you have two main patterns:

## Option A: Migrations Only via Prisma
- Keep Go runtime data access as-is (or gradually move to an idiomatic Go ORM later).
- Use Prisma schema as the single source of truth for DB structure.
- Run: `npx prisma migrate dev` (needs `DATABASE_URL` env var) for schema changes.
- Pros: Centralized schema modeling, migration management UI (`prisma studio`).
- Cons: Two layers (Prisma + raw SQL) can drift if not disciplined.

## Option B: Add a Node Sidecar Service
- Create a small Node service exposing HTTP/gRPC endpoints that internally uses the Prisma client.
- Go backend calls this service instead of direct DB queries.
- Pros: Full Prisma client power (relations, type safety for that service).
- Cons: Added operational complexity and latency; duplication of auth logic unless refactored.

## Option C: Replace with Go-native Tooling (Recommended for Pure Go)
Consider instead:
- `sqlc` (generates type-safe Go from SQL queries; keeps raw SQL clarity).
- `ent` (schema-as-code, similar ergonomics to Prisma, pure Go).
- `gorm` (popular, less strict, runtime query building).

Given the current codebase already uses raw SQL and custom encryption + auth, Option C (ent or sqlc) may integrate more naturally.

## Environment Variable
Set `DATABASE_URL` to mirror existing .env vars. Example:
```
DATABASE_URL="mysql://root:password@localhost:3306/ospab_panel?charset=utf8mb4&parseTime=True&loc=Local"
```
Adjust credentials as needed.

## Next Steps
1. Install Prisma locally: `npm install prisma --save-dev` and `npx prisma generate`.
2. Create initial migration: `npx prisma migrate dev --name init`.
3. (Optional) Open studio: `npx prisma studio`.

If you confirm the desired integration pattern, we can automate the migration script or introduce a Go-native alternative.
