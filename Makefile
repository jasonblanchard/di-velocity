createdb:
	# createuser -e -d -P -E di
	createdb -U di -e -O di di_velocity

dropdb:
	dropdb di_velocity

dbmigrate:
	migrate -database postgres://di:di@localhost:5432/di_velocity?sslmode=disable -path db/migrations up
