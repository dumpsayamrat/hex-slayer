# HexSlayer deployment
# Usage:
#   make build                    — build backend + frontend locally
#   make deploy HOST=<ip>         — ship to Lightsail
#   make setup HOST=<ip>          — one-time server provisioning

HOST ?= YOUR_LIGHTSAIL_IP
KEY  ?= ~/.ssh/hexslayer.pem
SSH  ?= ubuntu@$(HOST)
SSH_OPTS ?= -i $(KEY)
SCP  = scp $(SSH_OPTS)
SSSH = ssh $(SSH_OPTS)

# --- Local builds ---

.PHONY: build build-backend build-frontend

build: build-backend build-frontend

build-backend:
	cd backend && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o ../dist/server ./cmd/server

build-frontend:
	cd frontend && npm ci && npm run build
	rm -rf dist/frontend
	cp -r frontend/dist dist/frontend

# --- Remote deployment ---

.PHONY: setup deploy

setup:
	$(SCP) deploy/hexslayer.service $(SSH):/tmp/hexslayer.service
	$(SCP) deploy/hexslayer.nginx $(SSH):/tmp/hexslayer.nginx
	$(SCP) deploy/setup.sh $(SSH):/tmp/setup.sh
	$(SSSH) $(SSH) 'sudo bash /tmp/setup.sh'

deploy:
	$(SCP) dist/server $(SSH):/tmp/hexslayer-server
	$(SSSH) $(SSH) 'sudo mv /tmp/hexslayer-server /opt/hexslayer/server && sudo chmod +x /opt/hexslayer/server'
	rsync -az --delete -e "ssh $(SSH_OPTS)" dist/frontend/ $(SSH):/tmp/hexslayer-frontend/
	$(SSSH) $(SSH) 'sudo rsync -a --delete /tmp/hexslayer-frontend/ /opt/hexslayer/frontend/ && sudo rm -rf /tmp/hexslayer-frontend'
	$(SSSH) $(SSH) 'sudo systemctl restart hexslayer'
	@echo "Deployed! Check http://$(HOST)"
