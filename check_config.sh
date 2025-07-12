#!/bin/sh

CONFIG_FILE="${XRAY_CONFIG_DIR}/config.json"

# Проверка существования файла
if [ ! -f "$CONFIG_FILE" ]; then
  echo "[*] $CONFIG_FILE not found. Creating empty JSON object..."
  mkdir -p "$(dirname "$CONFIG_FILE")"
  echo '{}' > "$CONFIG_FILE"
else
  echo "[*] $CONFIG_FILE already exists. Skipping."
fi
