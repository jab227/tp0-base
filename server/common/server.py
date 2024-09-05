import socket
import logging
import multiprocessing
import common.protocol as protocol
import common.utils as utils
from common import storage
from common import connection
from common.signal_handler import sigterm_handler_init, SignalSIGTERM
from typing import Tuple
from common.rw import send_all, recv_all


class Server:
    MAX_AGENCIES = 5

    def __init__(self, port, listen_backlog):
        # Initialize server socket
        sigterm_handler_init()
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.logger = logging.getLogger("Server")

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        storage.StorageManager().register('bets_store', storage.BetsStorage)
        task_id = 0
        with storage.StorageManager() as manager:
            barrier = manager.Barrier(self.MAX_AGENCIES)
            with multiprocessing.Pool() as pool:
                shared_store = manager.bets_store(self.MAX_AGENCIES, barrier)
                connections = []
                error = False
                while True:
                    try:
                        client_sock = self.__accept_new_connection()
                        logger = logging.getLogger(f"client-{task_id}")
                        connections.append(client_sock)
                        handler = connection.Handler(client_sock, logger, shared_store, barrier)
                        _ = pool.apply_async(handler.handle)
                        task_id += 1
                    except SignalSIGTERM as name:
                        error = True
                        self.logger.info(
                            f'action: signal | result: success | msg: received {name.signal}'
                        )
                    finally:
                        if error:
                            self._server_socket.close()
                            self.logger.info(
                                f'action: close_socket | result: success | msg: "closed server socket"'
                            )
                            for s in connections:
                                s.close()
                                self.logger.info(f'action:_close socket | result: success | msg: "closed clients sockets"')
                                return
    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
