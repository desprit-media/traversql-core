Output demo.gif

Require echo
Require docker

Set Shell "bash"
Set FontSize 24
Set Width 1200
Set Height 600
Set Theme "GruvboxDark"
Set Padding 8
Set Margin 24
Set BorderRadius 10
Set MarginFill "#000040"
Set WindowBar Colorful

Type "docker pull postgres:16"
Enter
Sleep 1s

Type "docker run --name postgres-traversql-example -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_USER=myuser -e POSTGRES_DB=mydb -p 5432:5432 -d postgres:16"
Enter
Sleep 1s

Type "docker cp schema.sql postgres-traversql-example:/schema.sql"
Enter
Sleep 1s

Type "docker cp data.sql postgres-traversql-example:/data.sql"
Enter
Sleep 1s

Type "docker exec -it postgres-traversql-example psql -U myuser -d mydb -f /schema.sql"
Enter
Sleep 2s

Type "docker exec -it postgres-traversql-example psql -U myuser -d mydb -f /data.sql"
Enter
Sleep 2s

Type "POSTGRES_HOST=localhost POSTGRES_PORT=5432 POSTGRES_USER=myuser POSTGRES_PASSWORD=mysecretpassword POSTGRES_DB=mydb go run github.com/desprit-media/traversql-core/cmd/traversql@v0.0.3 traverse --table=order_items --pk-fields=order_item_id --pk-values=1 --output=out.sql"
Enter
Sleep 2s

Type "cat out.sql"
Enter
Sleep 3s
