# Split-repository deployment runbook

CryptoCheck is deployed from three sibling directories on the server:

```text
~/cryptocheck-api          # Go API and worker source
~/cryptocheck-client       # Next.js source
~/cryptocheck-deployment   # compose file, Caddyfile and private .env
```

The deployment directory is intentionally not a Git repository: it contains
server-specific configuration and secrets. Do not commit its `.env` file.

## One-time server setup

Clone the two source repositories and keep the compose configuration in its own
directory:

```bash
git clone git@github.com:0Hoag/cryptocheck-api.git ~/cryptocheck-api
git clone git@github.com:0Hoag/cryptocheck-client.git ~/cryptocheck-client
mkdir -p ~/cryptocheck-deployment
```

Create `~/cryptocheck-deployment/.env` from a local secure copy. It must set
the production secrets, scanner keys and `NEXT_PUBLIC_API_URL` to the HTTPS
public site URL. Keep MongoDB/RabbitMQ data in named Docker volumes.

The compose build contexts must point to the sibling source folders, for
example:

```yaml
backend-api:
  build:
    context: /home/ubuntu/cryptocheck-api
    dockerfile: deployment/Dockerfile.backend
frontend:
  build:
    context: /home/ubuntu/cryptocheck-client
    dockerfile: deployment/Dockerfile.frontend
```

## Deploy a new release

Run these commands on the server after a commit has been pushed to `main`:

```bash
cd ~/cryptocheck-api && git pull --ff-only origin main
cd ~/cryptocheck-client && git pull --ff-only origin main

cd ~/cryptocheck-deployment
docker compose --project-name deployment --env-file .env config >/dev/null
docker compose --project-name deployment --env-file .env up -d --build --force-recreate
```

`up -d --build --force-recreate` rebuilds the API, worker and frontend from the
two latest Git checkouts while preserving named MongoDB/RabbitMQ volumes.
Never use `docker compose down -v` for a normal application update.

## Verify and diagnose

```bash
cd ~/cryptocheck-deployment
docker compose --project-name deployment ps
docker compose --project-name deployment logs --tail=100 backend-api
docker compose --project-name deployment logs --tail=100 frontend
curl -fsS http://127.0.0.1:8080/swagger/index.html >/dev/null
```

Use the configured HTTPS domain for browser testing. Do not test the frontend
through an IP address if `NEXT_PUBLIC_API_URL` points at the domain; browsers
will otherwise make API requests to the wrong origin. Confirm the domain has an
`A` record to the server IP before asking Caddy to obtain or renew certificates.

## Rollback

Find the previously working SHA in either source repository, check out that
commit in both repositories as appropriate, and run the same deploy command.
The database volumes are retained; source rollback does not remove data.
