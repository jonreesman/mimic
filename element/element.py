from datetime import datetime
import json
import os
import sys

import mysql.connector
from mysql.connector.cursor import MySQLCursorPrepared
from mysql.connector import errorcode

sql_element_user_check_query = "INSERT INTO users(user_id, source) VALUES (%s, %s)"
sql_element_add_query = "INSERT INTO element_messages(message_id, user_id, channel_id, msg, time_stamp) VALUES (%s, %s, %s, %s, %s)"


class database:
    def __init__(self, port, name, user, password):
        self.port = port
        self.name = name
        self.user = user
        self.password = password
    def connect(self):
        try:
            self.connection = mysql.connector.connect(
                                        host='localhost',
                                        port=self.port, 
                                        database=self.name,
                                        user=self.user,
                                        password=self.password,
                                        charset='utf8'
                )
            self.element = self.connection.cursor(cursor_class=MySQLCursorPrepared)
        except mysql.connector.Error as err:
            if err.errno == errorcode.ER_ACCESS_DENIED_ERROR:
                print("Something is wrong with your user name or password")
            elif err.errno == errorcode.ER_BAD_DB_ERROR:
                print("Database does not exist")
            else:
                print(err)
        
    def check_users(self, users):
        for user in users:
            try:
                self.element.execute(sql_element_user_check_query, (user['user'],0))
            except:
                pass
    def add_statements(self, messages):
        i = 1
        for message in messages:
            self.element.execute(sql_element_add_query,
                    (str(i).encode('utf-8'),
                    message['user'].encode('utf-8'),
                    "0".encode("utf-8"),
                    message['message'].encode("utf-8"),
                    datetime(1970, 1, 1).timestamp())
                )
            print("Adding statement")

            i = i + 1
    def close(self):
        self.connection.commit()
        self.connection.close()
        self.element.close()


def get_data(filename):
    f = open(filename)
    data = json.load(f)
    for m in data:
        print(m['user'], ' -> ', m['message'])
    f.close()
    return data



def main():
    port = os.environ.get("DB_PORT")
    db_name = os.environ.get("DB_NAME")
    db_user = os.environ.get("DB_USER")
    db_pwd = os.environ.get("DB_PWD")
    db = database(port, db_name, db_user, db_pwd)
    data = get_data("./element/element_final.json")
    db.connect()
    db.check_users(data)
    db.add_statements(data)
    db.close()

if __name__ == "__main__":
    sys.exit(main())