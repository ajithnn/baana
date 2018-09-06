# Baana (Arrow[sanskrit])

Baana is a rails-like MVC scaffolding command line tool, which also provides routing and db handling.
It is opinionated, uses github.com/gin-gonic/gin as the base framework and uses github.com/jinzhu/gorm as ORM.

# Installation

go get -u gitlab.com/ajithnn/baana

Ensure GOBIN is in your PATH

# Usage

1. CD to the folder where you need to init a MVC App.
2. Run `baana init [name-of-app]`
3. Command creates folder structure as shown below
    * models
      - Model files
    * controllers
      - Controller files
    * config
      - db.json 
      - routes.json
    * server
      - Main server using Gin, calls baana/service.LoadRoutes to load routes from routes.json.
    * route
      - route.go - Auto-generated do not edit.
    * migrations
      - migrations.go - All generated migrations will be added to this file.
4. Run `baana generate model [name-of-model]`
5. Above command does the following
    * Create a [name].go file in models with base code + swagger comments for documentation.
    * Create a [name].go file in controllers with base code + swagger comments for documentation.
    * Create CRUD routes in config/routes.json.
6. Run `baana generate migration [name-of-migration]`
7. Above command generates a migration function in migrations/migrations.go.
8. Add AutoMigrate calls in migration to create model tables. See: github.com/jinzhu/gorm for details on db functions.
9. Run dependency manager to load all dependency - If dep -- dep init
10. Build 
11. Run as ./[name-of-app] migrate up "" - to run migrations 
12. Run ./[name-of-app] server to start app.

 
