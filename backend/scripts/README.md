# Database Maintenance Scripts

Scripts untuk membersihkan dan memelihara database, khususnya untuk refresh tokens.

## 📋 Daftar Script

### 1. cleanup_old_tokens.sql
Script SQL untuk membersihkan token-token lama secara manual.

**Fungsi:**
- Revoke token yang sudah expired
- Batasi jumlah token aktif per user (max 3)
- Hapus token yang sudah di-revoke >30 hari

**Cara Pakai:**
```bash
# Menggunakan psql
PGPASSWORD=mydevelopment psql -h localhost -p 3479 -U postgres -d erp_db -f scripts/cleanup_old_tokens.sql

# Atau untuk test (tidak commit)
PGPASSWORD=mydevelopment psql -h localhost -p 3479 -U postgres -d erp_db -c "BEGIN; \i scripts/cleanup_old_tokens.sql; ROLLBACK;"
```

### 2. cmd/maintenance/main.go
Program Go untuk maintenance otomatis yang bisa dijadikan cron job.

**Fungsi:**
- Sama seperti SQL script tapi dalam bentuk Go program
- Bisa dijadwalkan dengan cron
- Memberikan report statistik

**Cara Pakai:**
```bash
# Run manual
go run cmd/maintenance/main.go

# Atau build dulu
go build -o bin/maintenance cmd/maintenance/main.go
./bin/maintenance

# Jadwalkan sebagai cron job (setiap hari jam 2 pagi)
# Tambahkan ke crontab:
0 2 * * * cd /path/to/backend && /usr/local/go/bin/go run cmd/maintenance/main.go >> /var/log/erp-maintenance.log 2>&1
```

## 🔧 Perubahan Code

### auth_handler.go
**Ditambahkan:**
- Logging detail saat menerima refresh token dari cookie
- Logging saat set cookie baru
- Logging konfigurasi cookie (domain, secure, samesite)

**Berguna untuk:**
- Debug masalah cookie tidak ter-update
- Track request dari client mana yang bermasalah
- Verifikasi konfigurasi cookie

### auth_service.go
**Ditambahkan:**
1. **Logging di RefreshToken():**
   - Log token hash yang dicari
   - Cek apakah token di-revoke atau tidak ada
   - Log timestamp created/revoked untuk analisis

2. **Token Limit di storeRefreshToken():**
   - Otomatis revoke token lama jika >3 token aktif
   - Hanya keep 2 newest + 1 baru = 3 total
   - Log setiap revocation untuk audit

**Dampak:**
- Mencegah akumulasi token (dari 11 token → max 3)
- Otomatis cleanup saat login/refresh
- Lebih aman dan efisien

## 📊 Monitoring

### Cek Status Token
```sql
-- Lihat jumlah token per user
SELECT
    user_id,
    COUNT(*) FILTER (WHERE is_revoked = false) as active_tokens,
    COUNT(*) FILTER (WHERE is_revoked = true) as revoked_tokens,
    MAX(created_at) as latest_token
FROM refresh_tokens
GROUP BY user_id
ORDER BY active_tokens DESC;

-- Lihat token yang expired tapi belum di-revoke
SELECT COUNT(*) as expired_active_tokens
FROM refresh_tokens
WHERE is_revoked = false AND expires_at < NOW();

-- Lihat token yang baru dibuat (24 jam terakhir)
SELECT
    user_id,
    LEFT(token_hash, 16) as token_preview,
    created_at,
    is_revoked
FROM refresh_tokens
WHERE created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;
```

### Lihat Logs
```bash
# Terminal backend akan menampilkan:
# 🔍 = Debug info
# ✅ = Success
# 🚨 = Error
# ⚠️  = Warning
# 🍪 = Cookie operations
# 📊 = Statistics
# 🗑️  = Deletion/cleanup

# Contoh log yang bagus:
# 🔍 DEBUG [RefreshToken]: Cookie token preview: abc123...
# 🔍 DEBUG [RefreshToken]: Searching for token hash: abc123...
# ✅ DEBUG [RefreshToken]: Token found - Created: 2025-12-19, Expires: 2025-12-26
# 🔍 DEBUG [RefreshToken]: Revoking old token (hash: abc123...)
# ✅ DEBUG [RefreshToken]: Old token revoked (1 rows affected)
# 🍪 DEBUG [RefreshToken]: Setting new cookie with token: def456...
# ✅ DEBUG [StoreToken]: New token stored (expires: 2025-12-26)

# Contoh log bermasalah:
# 🚨 DEBUG [RefreshToken]: Token found but REVOKED at 2025-12-19 22:36:01
# ⚠️  WARNING: User has 11 active tokens (limit: 3), revoking oldest ones
```

## 🐛 Troubleshooting

### Problem: Browser masih pakai token lama
**Solusi:**
1. Cek log di terminal backend
2. Pastikan ada log "Setting new cookie"
3. Cek browser DevTools → Application → Cookies
4. Pastikan cookie `refresh_token` ter-update setelah refresh

### Problem: User punya banyak token aktif
**Solusi:**
1. Run maintenance script: `go run cmd/maintenance/main.go`
2. Atau run SQL cleanup: `psql ... -f scripts/cleanup_old_tokens.sql`
3. Setelah update code, masalah tidak akan terulang

### Problem: Token di-revoke tapi masih dipakai
**Cek log:**
```
🚨 DEBUG [RefreshToken]: Token found but REVOKED at ...
```
**Penyebab:** Cookie browser tidak ter-update.

**Solusi:**
1. Clear browser cookies untuk situs ini
2. Login ulang
3. Monitor log untuk memastikan cookie ter-update

## ⏰ Recommended Schedule

### Cron Jobs
```cron
# Maintenance harian (jam 2 pagi)
0 2 * * * cd /path/to/backend && go run cmd/maintenance/main.go

# Atau seminggu sekali (Minggu jam 3 pagi)
0 3 * * 0 cd /path/to/backend && go run cmd/maintenance/main.go

# Production: gunakan binary yang sudah di-build
0 2 * * * /opt/erp/bin/maintenance >> /var/log/erp-maintenance.log 2>&1
```

## 📈 Metrics to Track

Pantau metrics berikut untuk kesehatan sistem:

1. **Average tokens per user**: Seharusnya ≤2.5
2. **Expired but not revoked**: Seharusnya 0
3. **Revoked tokens >30 days**: Cleanup secara berkala
4. **Token creation rate**: Normal: 1-3 per user per hari

## 🔐 Security Notes

- **Token hash disimpan**, bukan raw token (secure)
- **httpOnly cookie** mencegah XSS attacks
- **Token rotation** setiap refresh untuk security
- **Max 3 tokens** mencegah session accumulation
- **30-day cleanup** untuk audit trail balance

## 📞 Support

Jika ada masalah:
1. Cek logs di terminal backend
2. Run `go run cmd/maintenance/main.go` untuk report
3. Query database dengan SQL di atas untuk detail
