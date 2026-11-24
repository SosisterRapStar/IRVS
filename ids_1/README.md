# Tasks API

REST API для списка задач с PostgreSQL.

## ДЗ №1. Запуск в Docker Compose

Для `make up` нужны Docker Compose и Go. Команда запускает PostgreSQL, применяет миграции и запускает API. При запуске API на хосте используется `make run` с доступной базой, к которой уже применены миграции.

## ДЗ №2. Развёртывание в Kubernetes

- [k8s/app.yaml](k8s/app.yaml) описывает Deployment и внутренний Service `tasks-api` на порту `8080`. API подключается к базе по адресу `postgres:5432`.
- [k8s/postgres.yaml](k8s/postgres.yaml) описывает StatefulSet и headless Service `postgres`. Из `volumeClaimTemplates` StatefulSet создаёт PVC `data-postgres-0` на 1 ГиБ через `local-path`.
- [Dockerfile](Dockerfile) добавляет SQL-миграции и клиент PostgreSQL в образ API. Init-контейнер ждёт PostgreSQL и применяет миграцию из этого образа перед запуском API.

Из каталога `ids_1` примените обычные YAML-манифесты.

```sh
kubectl apply -f k8s/
```

### Проверка сохранности данных

В командах проверки используйте namespace развёртывания. В отдельном терминале откройте порт API через `kubectl port-forward svc/tasks-api 8080:8080`. Swagger UI будет доступен по адресу `http://127.0.0.1:8080/swagger/index.html`.

Создайте запись, запомните UID Pod и PVC, затем пересоздайте Pod базы. Для извлечения `id` из JSON нужен `python3`.

```sh
api=http://127.0.0.1:8080
created=$(curl -fsS -H 'Content-Type: application/json' \
  -d '{"title":"IRVS HW2 persistence check"}' "$api/tasks")
printf '%s\n' "$created"
task_id=$(printf '%s\n' "$created" | python3 -c 'import json, sys; print(json.load(sys.stdin)["id"])')

old_pod_uid=$(kubectl get pod postgres-0 -o jsonpath='{.metadata.uid}')
old_pvc_uid=$(kubectl get pvc data-postgres-0 -o jsonpath='{.metadata.uid}')
kubectl delete pod postgres-0
kubectl wait --for=create pod/postgres-0 --timeout=180s
kubectl wait --for=condition=Ready pod/postgres-0 --timeout=180s
kubectl wait --for=condition=Available deployment/tasks-api --timeout=180s

new_pod_uid=$(kubectl get pod postgres-0 -o jsonpath='{.metadata.uid}')
new_pvc_uid=$(kubectl get pvc data-postgres-0 -o jsonpath='{.metadata.uid}')
stored=$(curl -fsS --retry 10 --retry-delay 2 "$api/tasks/$task_id")
printf 'Pod: %s → %s\nPVC: %s → %s\n%s\n' \
  "$old_pod_uid" "$new_pod_uid" "$old_pvc_uid" "$new_pvc_uid" "$stored"
test "$old_pod_uid" != "$new_pod_uid" &&
  test "$old_pvc_uid" = "$new_pvc_uid" &&
  test "$created" = "$stored"
```

UID Pod меняется, UID PVC остаётся прежним, а `GET /tasks/{id}` возвращает ту же запись, что и `POST /tasks`.

## Swagger UI

![Swagger UI](images/swagger.png)

## Лицензия

[MIT](LICENSE)
