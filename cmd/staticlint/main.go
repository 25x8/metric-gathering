// Package main предоставляет инструмент статического анализа кода.
//
// # Использование
//
// Запустите инструмент для анализа вашего кода:
//
//	go run cmd/staticlint/main.go ./...
//
// Это запустит все настроенные анализаторы.
//
// # Включенные анализаторы
//
// Стандартные анализаторы Go из golang.org/x/tools/go/analysis/passes:
//   - asmdecl: проверяет соответствие между файлами ассемблера и объявлениями Go
//   - assign: проверяет бесполезные присваивания
//   - atomic: проверяет распространенные ошибки при использовании пакета sync/atomic
//   - bools: проверяет распространенные ошибки с логическими операторами
//   - buildtag: проверяет теги сборки
//   - cgocall: обнаруживает нарушения правил передачи указателей в cgo
//   - composite: проверяет составные литералы без ключей
//   - copylock: проверяет блокировки, ошибочно переданные по значению
//   - errorsas: проверяет, что второй аргумент errors.As является указателем на тип, реализующий error
//   - httpresponse: проверяет ошибки при использовании HTTP-ответов
//   - loopclosure: проверяет ссылки на переменные цикла из вложенных функций
//   - lostcancel: проверяет отсутствие вызова функции отмены контекста
//   - nilfunc: проверяет бесполезные сравнения между функциями и nil
//   - printf: проверяет соответствие вызовов в стиле printf их строкам формата
//   - shift: проверяет сдвиги, равные или превышающие ширину целого числа
//   - stdmethods: проверяет опечатки в методах известных интерфейсов
//   - structtag: проверяет, что теги полей структуры соответствуют reflect.StructTag.Get
//   - tests: проверяет распространенные ошибки использования тестов и примеров
//   - unmarshal: проверяет передачу значений не-указателей или не-интерфейсов в unmarshal
//   - unreachable: проверяет недостижимый код
//   - unusedresult: проверяет неиспользуемые результаты вызовов функций
//
// Дополнительные публичные анализаторы:
//   - bodyclose: проверяет, правильно ли закрыты тела HTTP-ответов
//   - errcheck: обеспечивает проверку ошибок
//
// Пользовательские анализаторы:
//   - noexit: запрещает прямые вызовы os.Exit в функции main пакета main
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"

	"github.com/25x8/metric-gathering/cmd/staticlint/noexit"
)

func main() {
	// Стандартные анализаторы Go
	standardAnalyzers := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,
	}

	publicAnalyzers := []*analysis.Analyzer{
		bodyclose.Analyzer, 
		errcheck.Analyzer,
	}

	customAnalyzers := []*analysis.Analyzer{
		noexit.Analyzer,
	}

	var analyzers []*analysis.Analyzer
	analyzers = append(analyzers, standardAnalyzers...)
	analyzers = append(analyzers, publicAnalyzers...)
	analyzers = append(analyzers, customAnalyzers...)

	multichecker.Main(analyzers...)
}
