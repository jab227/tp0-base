import socket
import logging
import signal
import common.protocol as protocol
import common.utils as utils
from common.signal_handler import sigterm_handler_init, SignalSIGTERM
from common.rw import send_all, recv_all


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        sigterm_handler_init()
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

        # TODO: Modify this program to handle signal to graceful shutdown
        # the server
        while True:
            client_sock = None
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
                client_sock = None
            except SignalSIGTERM as name:
                logging.info(
                    f'action: signal | result: success | msg: received {name.signal}'
                )
                self._server_socket.close()
                logging.info(
                    f'action: close_socket | result: success | msg: "closed server socket"'
                )
                if client_sock:
                    client_sock.close()
                    logging.info(
                        f'action:_close socket | result: success | msg: "closed client socket"'
                    )
                return

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            bet = recv_request(client_sock)
            utils.store_bets([bet])
            send_acknowledge(bet.number, client_sock)
            logging.info(
                f"action: apuesta_almacenada | result: success | dni: {bet.document}| numero: {bet.number}"
            )
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
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


def __recv_header(sock) -> protocol.Header:
    header_bytes = recv_all(sock, protocol.Header.SIZE)
    return protocol.Header.decode(header_bytes)


def recv_request(sock) -> utils.Bet:
    header = __recv_header(sock)
    payload = recv_all(sock, header.payload_size)
    req = protocol.Request(header, payload)
    return req.parse_bet()


def send_acknowledge(bet_number: int, sock):
    ack = protocol.AcknowledgeResponse(bet_number)
    data = ack.encode()
    send_all(sock, data)
