#!/bin/zsh

# Создаем директорию для бинарников если её нет
mkdir -p bin

# Список приложений
apps=(anime365-saver anidb-saver shikimori-saver)

# Список операционных систем и их расширений
typeset -A os_ext
os_ext=(
    linux "-linux"
    darwin "-mac"
)

# Сборка для каждого приложения и ОС
for app in $apps; do
    for os in ${(k)os_ext}; do
        echo "Building $app for $os..."
        GOOS=$os GOARCH=amd64 go build -o "bin/$app$os_ext[$os]" "./$app"
    done
done

echo "Build complete!"