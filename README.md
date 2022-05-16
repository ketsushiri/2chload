# 2chload
Утилита для загрузки всех файлов из треда или нескольких тредов на имиджборде 2ch. 

## Использовать
Скачайте последний релиз [отсюда.](https://github.com/ketsushiri/2chload/releases) Далее запустите `2chload(.exe)` с аргументами вида board/thread. Можно использовать произвольное число аргументов.
```bash
$ 2chload b/268037208 math/29047 pr/1008826
```
Для каждого треда будет создана своя директория с поддиректориями `pics` и `vids`.

## Собрать
Компилятор go версии >= 1.16
```bash
$ git clone https://github.com/ketsushiri/2chload.git
$ cd 2chload
$ go build
```
## License
MIT
