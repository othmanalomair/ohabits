git pull && docker build --no-cache -t apps_app:latest .
docker service update --force apps_app