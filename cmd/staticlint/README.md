# Статический анализатор для проекта Metric Gathering

В этой директории находится инструмент статического анализа кода для проекта сбора метрик.

## Обзор

Инструмент `staticlint` представляет собой мультичекер, объединяющий:

1. Стандартные анализаторы из `golang.org/x/tools/go/analysis/passes`
2. Дополнительные публичные анализаторы: `bodyclose` и `errcheck`
3. Пользовательский анализатор (`noexit`), запрещающий прямые вызовы `os.Exit()` в функции `main` пакета `main`

## Использование

Запустите инструмент статического анализа для вашего кода:

```bash
go run cmd/staticlint/main.go ./...
```

Для проверки конкретного пакета или директории:

```bash
go run cmd/staticlint/main.go ./cmd/agent/...
```

## Включенные анализаторы

### Стандартные анализаторы Go

В инструмент включены следующие стандартные анализаторы из `golang.org/x/tools/go/analysis/passes`:

- `asmdecl`: Проверяет соответствие между файлами ассемблера и объявлениями Go
- `assign`: Проверяет бесполезные присваивания
- `atomic`: Проверяет распространенные ошибки при использовании пакета sync/atomic
- `bools`: Проверяет распространенные ошибки с логическими операторами
- `buildtag`: Проверяет теги сборки
- `cgocall`: Обнаруживает нарушения правил передачи указателей в cgo
- `composite`: Проверяет составные литералы без ключей
- `copylock`: Проверяет блокировки, ошибочно переданные по значению
- `errorsas`: Проверяет, что второй аргумент errors.As является указателем на тип, реализующий error
- `httpresponse`: Проверяет ошибки при использовании HTTP-ответов
- `loopclosure`: Проверяет ссылки на переменные цикла из вложенных функций
- `lostcancel`: Проверяет отсутствие вызова функции отмены контекста
- `nilfunc`: Проверяет бесполезные сравнения между функциями и nil
- `printf`: Проверяет соответствие вызовов в стиле printf их строкам формата
- `shift`: Проверяет сдвиги, равные или превышающие ширину целого числа
- `stdmethods`: Проверяет опечатки в методах известных интерфейсов
- `structtag`: Проверяет, что теги полей структуры соответствуют reflect.StructTag.Get
- `tests`: Проверяет распространенные ошибки использования тестов и примеров
- `unmarshal`: Проверяет передачу значений не-указателей или не-интерфейсов в unmarshal
- `unreachable`: Проверяет недостижимый код
- `unusedresult`: Проверяет неиспользуемые результаты вызовов функций

### Дополнительные публичные анализаторы

- `bodyclose`: Проверяет, правильно ли закрыты тела HTTP-ответов
- `errcheck`: Обеспечивает проверку ошибок

### Пользовательские анализаторы

#### Анализатор NoExit

Анализатор `noexit` гарантирует, что `os.Exit()` не вызывается напрямую в функции `main` пакета `main`. Это важно для правильной очистки ресурсов и плавного завершения работы.

Вместо использования `os.Exit()` рассмотрите одну из следующих альтернатив:

1. Возврат из функции main с соответствующей обработкой кодов выхода
2. Использование обработки сигналов и шаблонов плавного завершения
3. Настройка отложенных операций (defer) для очистки ресурсов

## Зависимости

Инструмент статического анализа требует следующих зависимостей:

```
golang.org/x/tools
github.com/kisielk/errcheck
github.com/timakin/bodyclose
```

Эти зависимости будут автоматически установлены при выполнении команды `go mod tidy`. 