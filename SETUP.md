# Revio — Pre-Run Setup Checklist

Everything you need to configure before starting the project. Work through each section in order.

---

## 1. Prerequisites — Install These First

| Tool | Minimum Version | Install |
|------|----------------|---------|
| Docker Desktop | 24.x | https://www.docker.com/products/docker-desktop/ |
| Go | 1.21 | https://go.dev/dl/ |
| Node.js | 20.x | https://nodejs.org/ |
| Git | any | https://git-scm.com/ |
| psql (optional) | 16.x | bundled with PostgreSQL |

Verify with:

```powershell
docker --version
go version
node --version
npm --version
```

---

## 2. GitHub OAuth App — Required for Login

You need a GitHub OAuth App to enable "Continue with GitHub" login and to sync pull requests.

### Steps

1. Go to: https://github.com/settings/developers
2. Click **OAuth Apps** > **New OAuth App**
3. Fill in:

   | Field | Value |
   |-------|-------|
   | Application name | Revio (or any name) |
   | Homepage URL | `http://localhost:3000` |
   | Authorization callback URL | `http://localhost:8080/api/auth/github/callback` |

4. Click **Register application**
5. On the next page, copy the **Client ID** — you will need it
6. Click **Generate a new client secret** and copy the value immediately — it is only shown once

> If you plan to deploy to a real domain, update the callback URL to match your production domain.

---

## 3. Environment Files — Fill These In

### 3.1 Root `.env` (used by Docker Compose)

Copy the example and edit it:

```powershell
Copy-Item .env.example .env
```

Open `.env` and set every value:

```env
# ─── Database ────────────────────────────────────────────────
DB_USER=revio                        # Keep as-is or change
DB_PASSWORD=revio_secret             # CHANGE THIS — use any strong password
DB_NAME=revio                        # Keep as-is or change

# ─── Backend ─────────────────────────────────────────────────
JWT_SECRET=                          # REQUIRED — paste a random 32+ character string
                                     # Generate one: openssl rand -hex 32
                                     # Or use: https://generate-secret.vercel.app/32

# ─── GitHub OAuth ─────────────────────────────────────────────
GITHUB_CLIENT_ID=                    # REQUIRED — paste the Client ID from step 2
GITHUB_CLIENT_SECRET=                # REQUIRED — paste the Client Secret from step 2

# ─── These can stay as defaults for local dev ─────────────────
GITHUB_REDIRECT_URL=http://localhost:8080/api/auth/github/callback
FRONTEND_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:8080
PORT=8080
ENV=development
DB_HOST=postgres                     # Use "localhost" if running backend outside Docker
DB_PORT=5432
DB_SSLMODE=disable
```

### 3.2 Backend `.env` (used when running Go directly without Docker)

```powershell
Copy-Item .env.example backend\.env
```

Then in `backend\.env`, change `DB_HOST` from `postgres` to `localhost`:

```env
DB_HOST=localhost      # ← Change this when running outside Docker
```

### 3.3 Frontend `.env.local` (used when running Next.js directly without Docker)

```powershell
Copy-Item frontend\.env.local.example frontend\.env.local
```

The default value is correct for local development:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## 4. Generate a JWT Secret

You must set `JWT_SECRET` to a random string of at least 32 characters. Never use a short or guessable string.

**Option A — PowerShell:**

```powershell
-join ((65..90) + (97..122) + (48..57) | Get-Random -Count 48 | ForEach-Object { [char]$_ })
```

**Option B — OpenSSL (if installed):**

```bash
openssl rand -hex 32
```

**Option C — Online generator:**

https://generate-secret.vercel.app/64

Paste the result into `.env` as the value of `JWT_SECRET`.

---

## 5. Database Setup

The database is PostgreSQL 16 running in Docker. The migration file is at:

```
backend/migrations/001_initial.sql
```

It creates all tables, indexes, triggers, and a `pr_metrics` view automatically when Docker starts (the file is mounted into the PostgreSQL init directory).

**You do not need to run the migration manually** — Docker handles it on first start.

If you need to run it manually against a local PostgreSQL:

```bash
psql -h localhost -U revio -d revio -f backend/migrations/001_initial.sql
```

### Tables created

| Table | Purpose |
|-------|---------|
| `users` | Registered accounts (email/password and GitHub OAuth) |
| `repositories` | GitHub repos connected per user |
| `pull_requests` | Synced PR data from GitHub |
| `reviews` | PR reviews and their states |
| `notifications` | In-app notifications |
| `sync_jobs` | Background sync job tracking |

### View created

| View | Purpose |
|------|---------|
| `pr_metrics` | Pre-computed merge time and review time per PR |

---

## 6. GitHub Webhook (Optional — for real-time updates)

Without a webhook, PR data is only synced when you click the **Sync** button or when a repository is first connected. Webhooks give you live updates.

### Setup (requires a publicly accessible URL)

For local development, use a tunnel tool like [ngrok](https://ngrok.com/):

```bash
ngrok http 8080
```

Then in your GitHub repository settings:

1. Go to **Settings** > **Webhooks** > **Add webhook**
2. Set **Payload URL** to: `https://your-ngrok-url.ngrok.io/api/webhooks/github`
3. Set **Content type** to: `application/json`
4. Set **Secret** to the same value as `GITHUB_CLIENT_SECRET` in your `.env`
5. Select events: **Pull requests**, **Pull request reviews**

---

## 7. Starting the Project

### Option A — Docker (recommended, runs everything together)

Make sure Docker Desktop is running, then:

```powershell
docker compose up --build
```

First run takes 2-3 minutes to pull images and build. After that:

| Service | URL |
|---------|-----|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| API Health | http://localhost:8080/api/health |
| PostgreSQL | localhost:5432 |

### Option B — Local (run services separately)

Start PostgreSQL only via Docker:

```powershell
docker compose up postgres -d
```

Start the Go backend:

```powershell
cd backend
go run cmd/server/main.go
```

Start the Next.js frontend:

```powershell
cd frontend
npm run dev
```

---

## 8. First-Run Walkthrough

Once everything is running:

1. Open http://localhost:3000
2. Click **Create an account** and register with an email
3. Or click **Continue with GitHub** to sign in with your GitHub account
4. After login you land on the **Dashboard** (empty until repos are connected)
5. Go to **Repositories** in the sidebar
6. Enter a repository you have access to in the format `owner/repo-name` and click **Connect**
7. The backend will start syncing pull requests in the background (takes 10-60 seconds depending on PR count)
8. Refresh the **Dashboard** and **Pull Requests** pages to see data
9. Go to **Analytics** to see metrics — use the date range buttons to filter

---

## 9. Summary of All Required Values

| Variable | Where to get it | Required |
|----------|----------------|----------|
| `DB_PASSWORD` | Make up any strong password | Yes |
| `JWT_SECRET` | Generate randomly (min 32 chars) | Yes |
| `GITHUB_CLIENT_ID` | GitHub Developer Settings > OAuth Apps | Yes |
| `GITHUB_CLIENT_SECRET` | GitHub Developer Settings > OAuth Apps | Yes |
| `GITHUB_REDIRECT_URL` | Must match what you set in the GitHub OAuth App | Yes |
| `FRONTEND_URL` | Your frontend origin — `http://localhost:3000` for local | Yes |
| `NEXT_PUBLIC_API_URL` | Your backend URL — `http://localhost:8080` for local | Yes |

All other variables have safe defaults for local development.

---

## 10. Troubleshooting

**"JWT_SECRET is required" on backend start**
You have not set `JWT_SECRET` in your `.env` file. See section 4.

**"repository not found or access denied" when connecting a repo**
Your GitHub account is not connected. Log out and log back in via **Continue with GitHub**.

**GitHub OAuth callback fails with "invalid oauth state"**
Your browser blocked the `oauth_state` cookie. Make sure you are using `http://localhost:3000` (not 127.0.0.1) and that cookies are enabled.

**Frontend shows blank page after GitHub login**
The callback redirects to `/dashboard?token=...`. The token is saved to localStorage automatically. If the page is blank, check the browser console for API errors.

**No data after connecting a repository**
The sync runs in the background. Wait 30-60 seconds and refresh. Check backend logs for sync errors: `docker compose logs backend -f`

**Database connection refused**
If running the backend outside Docker, ensure `DB_HOST=localhost` in `backend/.env`. If using Docker, ensure `DB_HOST=postgres`.
