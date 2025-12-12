firs change the const CACHE_VERSION = 3; in sw.js
then


git pull && docker build --no-cache -t apps_app:latest .
docker service update --force apps_app