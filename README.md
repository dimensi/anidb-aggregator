# Ichime Saver

## Сборка и запуск приложений

### Доступные команды Make

#### Основные команды
- `make` - собрать все приложения для текущей ОС
- `make build-all` - собрать все приложения для всех ОС (Linux, MacOS, Windows)
- `make clean` - удалить все бинарники

#### Сборка отдельных приложений
- `make anime365-saver` - собрать только anime365-saver
- `make shikimori-saver` - собрать только shikimori-saver

#### Запуск приложений
- `make run-anime365` - запустить anime365-saver
- `make run-shikimori` - запустить shikimori-saver

#### Сборка для конкретной ОС
Вы можете указать конкретную ОС при сборке, используя переменную GOOS:
```bash
GOOS=darwin make shikimori-saver
GOOS=linux make shikimori-saver
```

### Структура проекта

Все скомпилированные бинарные файлы сохраняются в директории `bin/` со следующими суффиксами:
- Linux: `-linux`
- MacOS: `-mac`
- Windows: `.exe`

### Примеры использования

1. Сборка всех приложений для текущей ОС:
```bash
make
```

2. Сборка конкретного приложения для Windows:
```bash
GOOS=windows make anime365-saver
```

3. Запуск anime365-saver:
```bash
make run-anime365
```

4. Очистка всех бинарных файлов:
```bash
make clean
```

5. Показать список всех доступных команд:
```bash
make help
```

### Требования
- Go 1.21 или выше
- Make

### Примечание
Убедитесь, что у вас установлен Go и Make перед использованием этих команд.
