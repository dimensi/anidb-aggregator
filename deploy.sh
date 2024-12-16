#!/bin/bash

# Конфигурация
SERVER="root@193.222.62.199"
REMOTE_DIR="/root/db-server"
NGINX_CONF_DIR="/etc/nginx/conf.d"

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Начинаем деплой db-server...${NC}"

# Создаем архив с необходимыми файлами
echo "Создаем архив..."
cp go.mod go.sum db-server/
mkdir -p db-server/db-server
cp db-server/main.go db-server/db-server/
tar --no-xattrs -czf db-server.tar.gz \
    db-server/Dockerfile \
    db-server/docker-compose.yml \
    db-server/db-server/ \
    db-server/nginx.conf \
    db-server/go.mod \
    db-server/go.sum

rm -rf db-server/go.mod db-server/go.sum db-server/db-server/

# Копируем файлы на сервер
echo "Копируем файлы на сервер..."
scp db-server.tar.gz $SERVER:~/ || { echo -e "${RED}Ошибка при копировании файлов${NC}"; exit 1; }

# Выполняем команды на сервере
ssh $SERVER << 'EOF'
    # Останавливаем существующий контейнер если есть
    cd ~/db-server && docker compose down || true
    
    # Очищаем старые файлы
    rm -rf ~/db-server
    mkdir -p ~/db-server
    
    # Распаковываем новые файлы
    tar -xzf ~/db-server.tar.gz -C ~/
    rm ~/db-server.tar.gz
    
    # Копируем nginx конфиг
    sudo cp ~/db-server/nginx.conf /etc/nginx/sites-available/db.dimensi.dev.conf
    sudo ln -sf /etc/nginx/sites-available/db.dimensi.dev.conf /etc/nginx/sites-enabled/db.dimensi.dev.conf
    sudo nginx -t && sudo systemctl reload nginx
    
    # Создаем директорию для данных если её нет
    mkdir -p ~/db-server/data
    
    # Запускаем сервер
    cd ~/db-server && docker compose up -d --build
    
    echo "Проверяем статус..."
    docker compose ps
EOF

# Очищаем локальные временные файлы
rm db-server.tar.gz

echo -e "${GREEN}Деплой завершен!${NC}" 