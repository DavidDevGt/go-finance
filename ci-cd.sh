#!/bin/bash

SERVICE_NAME="finance-api"
BUILD_NAME="go-finance"
DEPLOY_PATH="/home/davidgt/finance"

echo "🔄 Pulling latest changes..."
git reset --hard
git pull origin main || { echo "❌ Error: git pull falló"; exit 1; }

echo "🔨 Building project..."
go mod tidy
go build -o "$BUILD_NAME" . || { echo "❌ Error: falló el build"; exit 1; }

echo "⛔ Stopping $SERVICE_NAME service..."
sudo systemctl stop "$SERVICE_NAME"

echo "🚚 Deploying binary to $DEPLOY_PATH/$SERVICE_NAME..."
sudo cp "./$BUILD_NAME" "$DEPLOY_PATH/$SERVICE_NAME" || { echo "❌ Error al copiar binario"; exit 1; }
sudo chmod +x "$DEPLOY_PATH/$SERVICE_NAME"

echo "✅ Starting $SERVICE_NAME service..."
sudo systemctl start "$SERVICE_NAME"
sudo systemctl status "$SERVICE_NAME" --no-pager

echo "📄 Showing last logs:"
journalctl -u "$SERVICE_NAME" -n 10 --no-pager
