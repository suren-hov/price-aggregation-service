# Deploying to a server

This sets up: the Go backend running as a systemd service on `127.0.0.1:8080`,
and nginx on port 80 serving the built React dashboard as static files while
reverse-proxying `/price`, `/health`, `/config`, `/metrics` to the backend.
Since nginx serves both on the same origin, the dashboard's relative fetch
calls (`/price`, etc.) work with no CORS setup.

These are the same [`../HOWTO.md`](../HOWTO.md) build steps, just wired into
systemd + nginx instead of running in a terminal.

## 1. Build

```bash
cd /home/suren/Projects/Go/price-aggregation-service

# Backend
CGO_ENABLED=0 go build -o bin/price-aggregation-service .

# Frontend
cd web && npm install && npm run build && cd ..

# Config
cp -n .env.example .env   # edit if you want non-default settings
```

## 2. Install the systemd service

```bash
sudo cp deploy/systemd/price-aggregation-service.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now price-aggregation-service
sudo systemctl status price-aggregation-service
```

Check it's actually serving before moving on:

```bash
curl -s localhost:8080/health -o /dev/null -w '%{http_code}\n'   # 503 until the first poll, then 200
curl -s localhost:8080/price
```

## 3. Install nginx

```bash
sudo apt update
sudo apt install -y nginx

sudo cp deploy/nginx/price-aggregation-service.conf /etc/nginx/sites-available/price-aggregation-service
sudo ln -sf /etc/nginx/sites-available/price-aggregation-service /etc/nginx/sites-enabled/price-aggregation-service

# Remove the default site so it doesn't conflict on port 80
sudo rm -f /etc/nginx/sites-enabled/default

sudo nginx -t          # validates the config before touching the live server
sudo systemctl reload nginx
```

## 4. Firewall

The Go service binds `0.0.0.0:8080` (all interfaces), so it's directly
reachable too unless you block it - only nginx (port 80) is meant to be
public.

```bash
sudo ufw allow 80/tcp
sudo ufw deny 8080/tcp     # only if ufw is active (`sudo ufw status`)
```

## 5. Verify

From another machine (or `curl http://<server-ip>/`):

* `http://<server-ip>/` — the React dashboard
* `http://<server-ip>/price` — proxied JSON from the backend
* `http://<server-ip>/health` — proxied health check

## Redeploying after a code change

```bash
git pull
CGO_ENABLED=0 go build -o bin/price-aggregation-service .
sudo systemctl restart price-aggregation-service

cd web && npm install && npm run build && cd ..
# nginx serves web/dist directly - no reload needed for a static rebuild,
# but reload anyway if you changed deploy/nginx/price-aggregation-service.conf:
sudo nginx -t && sudo systemctl reload nginx
```

## Adding HTTPS later

Once you have a domain pointed at this server, swap `server_name _;` in
[`nginx/price-aggregation-service.conf`](nginx/price-aggregation-service.conf)
for the real domain, then:

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.example
```

Certbot rewrites the nginx config in place to add the TLS listener and
redirect HTTP → HTTPS.
