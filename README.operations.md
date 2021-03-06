# Checklist before starting

* (done) В сервисах не используется dockerhub напрямую
* (done) Ресурсы в терраформе везде выставленны продовые
* (done) Деньги в Облако закинуты
* (done) В квоты облака не упираемся
* CI отключен
* Время выставленно в чексистеме
* (done) На машинках участников нет лишних файлов, которых не должны быть в сервисе
* Всё передеплоено
* Архивы для команд открываются и при помощи содержимого можно попасть на машины и в VPN
* README для команд дописан, актуален, и раскидан на машинки
* Бэкапы базы запущенны
* После деплоя сервисы запустились и работают (попробовать сходить на них через проксик)


## Процесс перед запуском
* остановить github worker'а (запущен в tmux), выключить CI наверное в целом в github
* сделать terraform destroy на команды
* поправить время в чексистем (правильное закомментировано)
* передеплоить её (можно cs-deploy.yml cs-stop cs-init cs-start)
* пересоздать команды
* запустить скрипт `teams/download_vpn_configs.sh`
* запустить `teams/make_archive.py`
* вкоммитить все артефакты
