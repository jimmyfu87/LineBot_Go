var user_db = db.getSiblingDB("linebot_db")
user_db.createCollection("user");
user_db.createCollection("message");