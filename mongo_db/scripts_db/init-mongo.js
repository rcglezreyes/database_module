db = db.getSiblingDB('admin');

if (db.system.users.find({ user: "root" }).count() === 0) {
  db.createUser({
    user: "root",
    pwd: "new123",
    roles: [{ role: "root", db: "admin" }]
  });
}

db = db.getSiblingDB('database_module_db');

if (db.system.users.find({ user: "dbuser" }).count() === 0) {
  db.createUser({
    user: "dbuser",
    pwd: "new123",
    roles: [{ role: "readWrite", db: "database_module_db" }]
  });
}

db.createCollection("predictions_assestments");
db.createCollection("predictions_vle");
db.createCollection("predictions_risks");
db.login.insertOne({ user: "test", password: "new123" });
