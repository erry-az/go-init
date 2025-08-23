env "local" {
  src = "file://db/schema.sql"
  url = "postgres://postgres:postgres@postgres-main:5432/go_init_db?sslmode=disable"
  dev = "docker://postgres/16/dev?search_path=public"
  migration {
    dir = "file://db/migrations"
  }
}

lint {
  destructive {
    error = false
  }
}