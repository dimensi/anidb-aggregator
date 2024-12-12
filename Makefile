# Определяем переменные
BINARY_DIR = bin
APPS = anime365-saver anidb-saver shikimori-saver db-mapper jikan-saver
GOOS ?= $(shell go env GOOS)
GOARCH = amd64

# Определяем суффиксы для разных ОС
ifeq ($(GOOS),windows)
    SUFFIX = .exe
else ifeq ($(GOOS),darwin)
    SUFFIX = -mac
else
    SUFFIX = -linux
endif

# Цель по умолчанию
.DEFAULT_GOAL := all

# Создание директории для бинарников
$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

# Сборка всех приложений для текущей ОС
.PHONY: all
all: $(BINARY_DIR) $(APPS)

# Сборка конкретного приложения
.PHONY: $(APPS)
$(APPS): %: $(BINARY_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_DIR)/$@$(SUFFIX) ./$@

# Сборка для всех поддерживаемых ОС
.PHONY: build-all
build-all: $(BINARY_DIR)
	@for app in $(APPS); do \
		GOOS=linux GOARCH=$(GOARCH) go build -o $(BINARY_DIR)/$$app-linux ./$$app; \
		GOOS=darwin GOARCH=$(GOARCH) go build -o $(BINARY_DIR)/$$app-mac ./$$app; \
	done

# Запуск программ
.PHONY: run-anime365
run-anime365:
	./$(BINARY_DIR)/anime365-saver$(SUFFIX)

.PHONY: run-anidb
run-anidb:
	./$(BINARY_DIR)/anidb-saver$(SUFFIX)

.PHONY: run-shikimori
run-shikimori:
	./$(BINARY_DIR)/shikimori-saver$(SUFFIX)

.PHONY: run-jikan
run-jikan:
	./$(BINARY_DIR)/jikan-saver$(SUFFIX)

.PHONY: run-db-mapper
run-db-mapper:
	./$(BINARY_DIR)/db-mapper$(SUFFIX)

# Очистка бинарников
.PHONY: clean
clean:
	rm -rf $(BINARY_DIR)

# Помощь
.PHONY: help
help:
	@echo "Доступные команды:"
	@echo "  make              - собрать все приложения для текущей ОС"
	@echo "  make build-all    - собрать все приложения для всех ОС"
	@echo "  make anime365-saver - собрать только anime365-saver"
	@echo "  make anidb-saver   - собрать только anidb-saver"
	@echo "  make shikimori-saver - собрать только shikimori-saver"
	@echo "  make db-mapper       - собрать только db-mapper"
	@echo "  make jikan-saver - собрать только jikan-saver"
	@echo "  make run-anime365   - запустить anime365-saver"
	@echo "  make run-anidb      - запустить anidb-saver"
	@echo "  make run-shikimori  - запустить shikimori-saver"
	@echo "  make run-jikan      - запустить jikan-saver"
	@echo "  make run-db-mapper  - запустить db-mapper"
	@echo "  make clean          - удалить все бинарники"
