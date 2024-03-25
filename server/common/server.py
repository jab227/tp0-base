from dataclasses import dataclass
from enum import Enum
import multiprocessing
import socket
import logging
import signal
import common.protocol as protocol
import common.storage as storage
from typing import Optional


class SignalSIGTERM(Exception):
    def __init__(self, signum):
        self.signal = signal.Signals(signum).name

def sigterm_handler(signum, frame):
    _ = frame
    raise SignalSIGTERM(signum)


class Server:
    number_of_agencies: int
    def __init__(self, port, listen_backlog, number_of_agencies):
        # Initialize server socket
        signal.signal(signal.SIGTERM, sigterm_handler)
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.server_logger = logging.getLogger("Server")
        self.number_of_agencies = number_of_agencies
        
    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        storage.StorageManager().register('bets_store',storage.BetsStorage)            
        task_id = 0
        with storage.StorageManager() as manager:
            with multiprocessing.Pool() as pool:
                shared_store = manager.bets_store(self.number_of_agencies)                                    
                while True:
                    connections = []
                    try:
                        client_sock = self.__accept_new_connection()
                        logger = logging.getLogger(f"client-{task_id}")
                        connections.append(client_sock)
                        _ = pool.apply_async(handle_client_connection, (client_sock, logger, shared_store,))
                    except SignalSIGTERM as name:
                        self.server_logger.info(f'action: signal | result: success | msg: received {name.signal}')
                        self._server_socket.close()
                        self.server_logger.info(f'action: close_socket | result: success | msg: "closed server socket"')
                        for s in connections:
                            s.close()

                        self.server_logger.info(f'action:_close socket | result: success | msg: "closed clients sockets"')
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

    

def send_exact(sock, data: bytes):
    sent = 0
    while sent < len(data):
        n = sock.send(data[sent:])
        if n == 0:
            raise RuntimeError("connection was closed")
        sent += n

        
def recv_exact(sock, read_size: int) -> bytes:
    buf = []
    read = 0
    while read < read_size:
        data = sock.recv(read_size - read)
        if data == b'':
            raise RuntimeError("connection was closed")
        read += len(data)
        buf.append(data)
    return b''.join(buf)


def read_kind(sock) -> Optional[protocol.MessageKind]:
    b = recv_exact(sock, 1)
    return protocol.decode_kind(b)


def read_request(sock, kind: protocol.MessageKind) -> Optional[protocol.Request]:
    bs = b''
    if kind == protocol.MessageKind.BET:
        bs = recv_exact(sock, protocol.Bet.SIZE)
        bet = protocol.decode(kind, bs)
        if bet is None or not isinstance(bet, protocol.Bet):
            return None
        payload_bytes = recv_exact(sock, bet.payload_size)
        payload = protocol.parse_payload(bet, payload_bytes)
        bet.bets = payload
        return bet
    bs = recv_exact(sock, 4)    
    return protocol.decode(kind, bs)

    
def write_response(sock, response: protocol.Response):
    data = response.encode()
    send_exact(sock, data)
    
    

def handle_client_connection(client_sock, logger, store):
    """
        Read message from a specific client socket and closes the socket
    
        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
    try:
        while True:
            kind  =  read_kind(client_sock)
            if kind is None:
                logger.error(f"action: receive_request | result: fail | error: unknown message kind")
                return
            
            req = read_request(client_sock, kind)
            if req is None:
                logger.error(f"action: receive_request | result: fail | error: invalid req")
                return
            
            if isinstance(req, protocol.Bet):
                id = req.agency_id
                store.store_bets(id, req.bets)
                response = protocol.Acknowledge(req.bets)
                write_response(client_sock, response)
                continue
            elif isinstance(req, protocol.Done):
                store.store_bets(req.agency_id, [], done=True)
                logger.info(f"action: receive_request | result: success | agency: {req.agency_id} | type: done")
            elif isinstance(req, protocol.Winners):
                winner_count = store.get_winner_count(req.agency_id)
                if winner_count is None:
                    response = protocol.WinnersUnavailable()                    
                    write_response(client_sock, response)
                    logger.info(f"action: receive_request | result: fail | agency: {req.agency_id} | type: waiting for agencies to submit bets")
                else:
                    response = protocol.WinnersList(winner_count)
                    write_response(client_sock, response)
                    return
            else:
                logger.info(f"action: receive_request | result: fail | error: unknown message")
                return
    except OSError as e:
        logger.error(f"action: receive_message | result: fail | error: {e}")
    except RuntimeError as e:
        logger.error(f"action: receive_message | result: fail | error: {e}")
    finally:
        client_sock.close()            
