import socket
import logging
import common.protocol as protocol
import common.utils as utils
from common.signal_handler import sigterm_handler_init, SignalSIGTERM
from typing import Tuple
from common.rw import send_all, recv_all


class Server:
    MAX_AGENCIES = 5
    STATE_BETTING = 0
    STATE_DRAWING = 1
    STATE_READY = 2

    def __init__(self, port, listen_backlog):
        # Initialize server socket
        sigterm_handler_init()
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._clients = dict()
        self._bets = dict()
        self.state = self.STATE_BETTING

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        current_client = 1
        while True:
            client_sock = None
            try:
                if self.state == Server.STATE_BETTING:
                    client_sock = self.__accept_new_connection()
                    self.__handle_client_connection(client_sock)
                    client_sock = None
                    if len(self._clients) == self.MAX_AGENCIES:
                        self.state = self.STATE_DRAWING
                        for i in range(1, self.MAX_AGENCIES + 1):
                            self._bets[i] = []

                        bets = utils.load_bets()
                        for bet in bets:
                            self._bets[int(bet.agency)].append(bet)

                elif self.state == self.STATE_DRAWING:
                    if current_client <= self.MAX_AGENCIES:
                        s = self._clients[current_client]
                        current_client += 1
                        send_ready(s)
                    else:
                        self.state = self.STATE_READY
                        current_client = 1
                elif self.state == self.STATE_READY:
                    if current_client <= self.MAX_AGENCIES:
                        s = self._clients[current_client]
                        agency_id, req = recv_request(s)
                        assert agency_id == current_client
                        if isinstance(req, protocol.GetWinners):
                            winners = []
                            for bet in self._bets[current_client]:
                                if utils.has_won(bet):
                                    winners.append(bet)
                            send_betting_results(winners, s)
                            s = self._clients.pop(current_client)
                            s.close()
                            current_client += 1
                    else:
                        self.state = self.STATE_BETTING
                        current_client = 1
                        
            except SignalSIGTERM as name:
                logging.info(
                    f'action: signal | result: success | msg: received {name.signal}'
                )
                self._server_socket.close()
                logging.info(
                    f'action: close_socket | result: success | msg: "closed server socket"'
                )
                #handle depending on the state
                if client_sock:
                    client_sock.close()
                    logging.info(
                        f'action:_close socket | result: success | msg: "closed client socket"'
                    )
                for _, sock in self._clients:
                    sock.close()
                return

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        error = False
        try:
            while True:
                try:
                    agencyid, req = recv_request(client_sock)
                    if isinstance(req, protocol.BatchEnd):
                        logging.info(f"action: fin batch | result: success")
                        send_acknowledge(0, client_sock)
                        self._clients[agencyid] = client_sock
                        return
                    elif (
                        isinstance(req, list)
                    ):
                        self._clients[int(agencyid)] = client_sock
                        length = len(req)
                        utils.store_bets(req)
                        logging.info(
                            f"action: apuesta_recibida | result: success | cantidad: {length}"
                        )
                        send_acknowledge(0, client_sock)
                    else:
                        raise RuntimeError(f"unexpected type, {type(req)}")
                except protocol.BetParseError as e:
                    logging.error(
                        f"action: apuesta_recibida | result: fail | cantidad: {e}"
                    )
                    send_acknowledge(1, client_sock)
        except OSError as e:
            error = True
            logging.error(f"action: receive_message | result: fail | error: {e}")
        except RuntimeError as e:
            error = True            
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            if error:
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


def recv_request(sock) -> Tuple[int, protocol.RequestUnion]:
    header = __recv_header(sock)
    payload = recv_all(sock, header.payload_size)
    req = protocol.Request(header, payload)
    return (header.agency_id, req.parse())


def send_ready(sock):
    results = protocol.WinnersReady()
    data = protocol.encode(results)
    send_all(sock, data)


def send_betting_results(winners: list[utils.Bet], sock):
    results = protocol.BettingResults(winners)
    data = protocol.encode(results)
    send_all(sock, data)


def send_acknowledge(bet_number: int, sock):
    ack = protocol.AcknowledgeResponse(bet_number)
    data = protocol.encode(ack)
    send_all(sock, data)
