# Что нужно настроить до деплоя
1. Удалить папочку с командами (`rm -r teams/1??`)
2. Перегенерировать все ключи (`cd teams && python gen_ssh_keys.py && python gen_serial_passwords.py`)
3. Изменить все пароли в `ansible/group_vars/cs`
4. Желательно ещё настроить время и поправить другие настройки
4. Изменить списки сервисов в
  * `terraform/03-teams/teams.tf`
  * `ansible/deploy-services.yml`
  * `ansible/group_vars/cs`
5. В `terraform/03-teams/*` выставить настройки размера инстансов.
6. Ещё поправить нужно везде количество команд на нужное, и сами команды прописать. Для checksystem'ы есть скриптик для помощи в генерации этого списка
7. В `ansible/roles/proxy/` поправить шаблоны `http_services.j2` и `stream_services.j2`, чтобы настроить L7 прокси.

# Процесс деплоя CTF

1. Сделать папку для ctf'а в Я.Облаке
2. Сделать сервисный аккаунт с названием `deployer`. Дать ему права на эту папку
3. Сделать файлик secrets.sh по образцу в `terraform/secrets.example.sh`, сделать `. secrets.sh`
4. `cd packer_images && packer build ubuntu-with-docker.pkr.hcl`
4. (на ноуте) `cd terraform/01-deployer-and-network && terraform init && terraform apply`
5. Зайти на получившуюся машину (её ip сохраниться в `teams/deployer_ip` и его покажет terraform)
6. Склонировать репозиторий
7. Опять же `secrets.sh` подготовить.
8. `cd 02-jury-infra && terraform init && terraform apply`
9. `cd 03-reams && terraform init && terraform apply`
10. `cd ansible && ./deploy_everything.sh`


