Output demo.gif

Require echo
Require docker

Set Shell "bash"
Set FontSize 20
Set Width 1200
Set Height 800
Set Theme "GruvboxDark"
Set Padding 16
Set Margin 18
Set BorderRadius 10
Set MarginFill "#141414"
Set WindowBar Colorful

Type "docker pull postgres:16"
Enter
Sleep 3s

Type@20ms "docker run --name postgres-traversql-example -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_USER=myuser -e POSTGRES_DB=mydb -p 5432:5432 -d postgres:16"
Enter
Sleep 3s

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

Type@20ms "export POSTGRES_HOST=localhost"
Enter
Type@20ms "export POSTGRES_PORT=5432"
Enter
Type@20ms "export POSTGRES_USER=myuser"
Enter
Type@20ms "export POSTGRES_PASSWORD=mysecretpassword"
Enter
Type@20ms "export POSTGRES_DB=mydb"
Enter
Type "go run github.com/desprit-media/traversql-core/cmd/traversql@latest traverse --table=order_items --pk-fields=order_item_id --pk-values=1 --output=out.sql"
Enter
Sleep 3s

Type "cat out.sql"
Enter
Sleep 6s
