FROM migrate/migrate
# FROM busybox

ADD db/migrations /migrations

ENTRYPOINT ["sh", "-c", "migrate -path=/migrations/ -database postgres://$USER:$PASSWORD@$HOST:$PORT/$DATABASE?sslmode=disable up"]
