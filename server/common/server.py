import socket
import logging
import signal
import common.protocol as protocol
import common.utils as utils

from typing import Optional

class SignalSIGTERM(Exception):
    def __init__(self, signum):
        self.signal = signal.Signals(signum).name

def sigterm_handler(signum, frame):
    _ = frame
    raise SignalSIGTERM(signum)
    
class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        signal.signal(signal.SIGTERM, sigterm_handler)
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        
    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while True:
            client_sock = None
            try:
                client_sock = self.__accept_new_connection()            
                self.__handle_client_connection(client_sock)
            except SignalSIGTERM as name:
                logging.info(f'action: signal | result: success | msg: received {name.signal}')                                

                self._server_socket.close()
                logging.info(f'action: close_socket | result: success | msg: "closed server socket"')

                if client_sock:
                    client_sock.close()
                    logging.info(f'action:_close socket | result: success | msg: "closed client socket"')                    
                return

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            header = read_header(client_sock)
            if header is None:
                logging.error(f"action: receive_header | result: fail | error: invalid header")
                return
            logging.info(f"action: receive_header | result: success | agency: {header.agency_id}")
            bet = read_payload(header, client_sock)
            utils.store_bets([bet])
            write_acknowledge(bet.number, client_sock)
            logging.info(f"action: apuesta_almacenada | result: success | dni: {bet.document}| numero: {bet.number}")
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        except RuntimeError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

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


def read_header(sock) -> Optional[protocol.Header]:
    header_bytes = recv_exact(sock, protocol.Header.HEADER_SIZE)
    return protocol.decode_header(header_bytes)


def read_payload(header: protocol.Header, sock) -> utils.Bet:
    payload_bytes = recv_exact(sock, header.payload_size)
    return protocol.parse_payload(str(header.agency_id), payload_bytes)


def write_acknowledge(bet_number: int, sock):
    ack = protocol.Ack(bet_number)
    data = protocol.encode_ack(ack)
    send_exact(sock, data)
    
    
