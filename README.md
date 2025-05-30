# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Пользовательский статический анализатор

В проекте реализован пользовательский статический анализатор кода, расположенный в директории `cmd/staticlint`. Анализатор включает:

1. Стандартные анализаторы из `golang.org/x/tools/go/analysis/passes`
2. Дополнительные публичные анализаторы: `bodyclose` и `errcheck` 
3. Пользовательский анализатор, запрещающий прямые вызовы `os.Exit()` в функции `main` пакета `main`

### Использование статического анализатора

Для запуска статического анализатора на всей кодовой базе:

```bash
go run cmd/staticlint/main.go ./...
```

Для проверки конкретного пакета или директории:

```bash
go run cmd/staticlint/main.go ./cmd/agent/...
```

### Анализатор NoExit

Анализатор `noexit` гарантирует, что `os.Exit()` не вызывается напрямую в функции `main` пакета `main`. Это важно для правильной очистки ресурсов и плавного завершения работы.

Вместо использования `os.Exit()` рассмотрите одну из следующих альтернатив:

1. Возврат из функции main с соответствующей обработкой кодов выхода
2. Использование обработки сигналов и шаблонов плавного завершения
3. Настройка отложенных операций (defer) для очистки ресурсов

Примеры правильной обработки завершения программы можно найти в файлах `examples/exit_test/main.go` и `examples/exit_pattern/main.go`.
