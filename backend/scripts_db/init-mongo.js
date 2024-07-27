db = db.getSiblingDB('admin');
db.createUser({
    user: "root",
    pwd: "Hialeah2024*-",
    roles: [{ role: "root", db: "admin" }]
  });

db = db.getSiblingDB('database_module_db');