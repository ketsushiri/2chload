# 2chload
## Использовать
Синтаксис: `./2chload [board/thread]`
```bash
$ ./2chload b/23232323 math/123123 pr/123321
```
Для каждого треда будет создана своя директория с поддиректориями `pics` и `vids`.

## Собрать
```bash
$ git clone https://github.com/ketsushiri/2chload.git
$ cd 2chload
$ go build -o 2chload main.go
```

## License
MIT
