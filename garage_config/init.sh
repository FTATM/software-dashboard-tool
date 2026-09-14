#!/usr/bin/env bash
set -euo pipefail

SERVICE="garage"
ZONE="dc1"
CAPACITY="10G"
BUCKET_NAME="iot-images"
KEY_NAME="app-key"

echo "==> Starting $SERVICE container..."
docker compose up -d "$SERVICE"

echo "==> Waiting for Garage daemon to become ready..."
until docker compose exec -T "$SERVICE" /garage status >/dev/null 2>&1; do
  sleep 1
done

echo "==> Fetching Node ID..."
# Extract strictly the 64-character hex ID before '@'
RAW_NODE_OUTPUT=$(docker compose exec -T "$SERVICE" /garage node id)
NODE_ID=$(echo "$RAW_NODE_OUTPUT" | grep -oE '[0-9a-fA-F]{64}' | head -n 1)

if [[ -z "$NODE_ID" ]]; then
  echo "Error: Failed to parse Node ID."
  echo "$RAW_NODE_OUTPUT"
  exit 1
fi

echo "Node ID: $NODE_ID"

echo "==> Configuring cluster layout..."
docker compose exec -T "$SERVICE" /garage layout assign -z "$ZONE" -c "$CAPACITY" "$NODE_ID"

echo "==> Applying layout..."
LAYOUT_VER=$(docker compose exec -T "$SERVICE" /garage layout show 2>/dev/null | awk '/Current layout version:/ {print $4}')
TARGET_VER=$(( ${LAYOUT_VER:-0} + 1 ))
docker compose exec -T "$SERVICE" /garage layout apply --version "$TARGET_VER"

echo "==> Creating bucket: $BUCKET_NAME..."
docker compose exec -T "$SERVICE" /garage bucket create "$BUCKET_NAME" 2>/dev/null || true

echo "==> Handling access key: $KEY_NAME..."
if ! docker compose exec -T "$SERVICE" /garage key info "$KEY_NAME" >/dev/null 2>&1; then
  KEY_OUTPUT=$(docker compose exec -T "$SERVICE" /garage key create "$KEY_NAME")
  docker compose exec -T "$SERVICE" /garage bucket allow --read --write "$BUCKET_NAME" --key "$KEY_NAME" >/dev/null 2>&1

  echo -e "\n=========================================="
  echo " COPY YOUR CREDENTIALS (SECRET SHOWN ONCE)"
  echo "=========================================="
  echo "$KEY_OUTPUT"
  echo "=========================================="
else
  echo "==> Key '$KEY_NAME' already exists."
  echo "Note: Garage cannot re-print a secret key once created."
  docker compose exec -T "$SERVICE" /garage bucket allow --read --write "$BUCKET_NAME" --key "$KEY_NAME" >/dev/null 2>&1
fi