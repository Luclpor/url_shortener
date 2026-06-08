# go-musthave-shortener-tpl



Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

## Бенчмарки и профилирование памяти

Бенчмарки запускаются с 30-секундной нагрузкой:

```bash
go test ./internal/service ./internal/service/audit -bench=. -benchtime=30s -benchmem -run=^$
```

Результат общего benchmark-прогона:

```text
BenchmarkGenerateRandomString-12    	984405860	        37.04 ns/op	       0 B/op	       0 allocs/op
BenchmarkCreateShortURL-12          	512988926	        69.87 ns/op	      21 B/op	       2 allocs/op
BenchmarkGetURL-12                  	1000000000	         7.721 ns/op	       0 B/op	       0 allocs/op
BenchmarkEventUpdate-12             	1000000000	         4.477 ns/op	       0 B/op	       0 allocs/op
```

Профили памяти снимались по всем benchmark'ам пакета `internal/service`: генерация короткого ключа, создание короткой ссылки и получение URL. Базовый профиль был снят командой:

```bash
go test ./internal/service -bench=. -benchtime=30s -benchmem -memprofile=profiles/base.pprof -run=^$
```

Результат базового прогона:

```text
BenchmarkGenerateRandomString-12    	623023068	        62.00 ns/op	       8 B/op	       1 allocs/op
BenchmarkCreateShortURL-12          	440170748	        82.04 ns/op	      24 B/op	       2 allocs/op
BenchmarkGetURL-12                  	1000000000	         7.019 ns/op	       0 B/op	       0 allocs/op
```

Команды `pprof top`, `pprof list`, `pprof peek` и граф `pprof web` показали, что значительная часть аллокаций приходится на `GenerateRandomString`: в общем service-профиле функция занимала `9.60GB` из `17.62GB`. Генератор использовал `[]rune` для ASCII-алфавита, поэтому был переписан на `[]byte`.

Повторный профиль снят той же нагрузкой:

```bash
go test ./internal/service -bench=. -benchtime=30s -benchmem -memprofile=profiles/result.pprof -run=^$
```

Результат после оптимизации:

```text
BenchmarkGenerateRandomString-12    	984405860	        37.04 ns/op	       0 B/op	       0 allocs/op
BenchmarkCreateShortURL-12          	512988926	        69.87 ns/op	      21 B/op	       2 allocs/op
BenchmarkGetURL-12                  	1000000000	         7.721 ns/op	       0 B/op	       0 allocs/op
```

Вывод команды:

```bash
pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

```text
File: service.test
Type: alloc_space
Time: 2026-06-08 12:32:19 MSK
Showing nodes accounting for -5385.08MB, 29.85% of 18042.29MB total
Dropped 33 nodes (cum <= 90.21MB)
      flat  flat%   sum%        cum   cum%
-6674.60MB 36.99% 36.99% -6674.60MB 36.99%  github.com/Luclpor/url_shortener.git/internal/service.GenerateRandomString
 1289.52MB  7.15% 29.85%      248MB  1.37%  github.com/Luclpor/url_shortener.git/internal/service.(*URLManager).CreateShortURL
         0     0% 29.85% -1041.52MB  5.77%  github.com/Luclpor/url_shortener.git/internal/service.(*URLManager).getUniqueKey
         0     0% 29.85%      248MB  1.37%  github.com/Luclpor/url_shortener.git/internal/service.BenchmarkCreateShortURL
         0     0% 29.85% -5633.09MB 31.22%  github.com/Luclpor/url_shortener.git/internal/service.BenchmarkGenerateRandomString
         0     0% 29.85% -5384.58MB 29.84%  testing.(*B).launch
         0     0% 29.85% -5384.58MB 29.84%  testing.(*B).runN
```
