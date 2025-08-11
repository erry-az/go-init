env "local" {
  src = "file://db/schema.sql"
  dev = "docker://postgres/15/dev?search_path=public"
  migration {
    dir = "file://db/migrations"
  }
}

lint {
  destructive {
    error = false
  }
}