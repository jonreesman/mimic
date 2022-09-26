import logging
import os
import grpc
import sys
import pickle as rick
import mysql.connector
from mysql.connector.cursor import MySQLCursorPrepared
from mysql.connector import errorcode

import markovify

from concurrent.futures import ThreadPoolExecutor
from mimic_pb2 import MsgResponse
from mimic_pb2_grpc import MessagesServicer, add_MessagesServicer_to_server

class EnvError(Exception):
    pass

class message:
    def __init__(self, timedata, msg):
        self.timedata = timedata
        self.msg = msg

class MessagesServer(MessagesServicer):
    def __init__(self, text_model):
        self.text_model = text_model
    def Detect(self, request, context):
        logging.info('detect request size: %d', len(request.signal))
        print(text_model)
        print(type(text_model))
        response = self.text_model.make_sentence()
        while response == None:
            response = self.text_model.make_sentence()
        resp = MsgResponse(msg=response)
        return resp

def load_model(user):
    db_url = os.environ.get("DB_URL")
    db_port = os.environ.get("DB_PORT")
    database = os.environ.get("DB_NAME")
    db_user = os.environ.get("DB_USER")
    db_pwd = os.environ.get("DB_PWD")
    text = ""
    try: 
        connection = mysql.connector.connect(
                                    host=db_url,
                                    port=db_port, 
                                    database=database,
                                    user=db_user,
                                    password=db_pwd,
                                    charset='utf8'
            )
    except mysql.connector.Error as err:
        if err.errno == errorcode.ER_ACCESS_DENIED_ERROR:
            logging.info("Something is wrong with your user name or password " + err)
        elif err.errno == errorcode.ER_BAD_DB_ERROR:
            logging.info("Something is wrong with the DB " + err)
        else:
            logging.info(err)
        return None
    except errorcode.Error as err:
        logging.info("Error: " + err)
        return None
    else:
        sql_element_select_query = """SELECT msg from element_messages WHERE user_id=%s"""
        sql_discord_select_query = """SELECT msg from discord_messages WHERE user_id=%s"""
        element = connection.cursor(buffered=True)
        discord = connection.cursor(buffered=True)
        element.execute(sql_element_select_query, (user,))
        discord.execute(sql_discord_select_query, (user,))
        for row in discord:
            text+=row[0]+'\n'
        for row in element:
            text+=row[0]+'\n'
        connection.close()
        element.close()
        discord.close()
        text_model = markovify.NewlineText(text, well_formed=False)
        text_model = text_model.compile()
        print(text_model.make_sentence())
        return text_model

if __name__ == "__main__":
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(levelname)s - %(message)s',
    )
    if "USER_TO_MIMIC" not in os.environ:
        raise EnvError("No ID for USER_TO_MIMIC. Try setting the env variable")

    user = os.getenv('USER_TO_MIMIC')
    port = os.getenv('PY_PORT',9999)
    text_model = load_model(user)
    server = grpc.server(ThreadPoolExecutor())
    add_MessagesServicer_to_server(MessagesServer(text_model), server)
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logging.info('server reads on port %r', port)
    server.wait_for_termination()