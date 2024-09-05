import time
import common.utils as utils
import common.protocol as protocol
from enum import Enum
from typing import Tuple
from common.rw import send_all, recv_all


class State(Enum):
    BETTING = "BETTING"
    DRAWING = "DRAWING"
    READY = "READY"
    EXIT = "EXIT"

class Handler:
    def __init__(self, sock, logger, shared_store, barrier):
        self.sock = sock
        self.barrier = barrier
        self.logger = logger
        self.shared_store = shared_store
        self.state = State.BETTING

    def betting(self):
        try:
            agencyid, req = recv_request(self.sock)
            if isinstance(req, protocol.BatchEnd):
                self.logger.info(f"action: fin batch | result: success")
                send_acknowledge(0, self.sock)
                self.state = State.DRAWING
                self.shared_store.signal(agencyid)
                return
            elif isinstance(req, list):
                self.logger.debug("received batch")
                length = len(req)
                self.shared_store.store_bets(req)
                self.logger.info(
                    f"action: apuesta_recibida | result: success | cantidad: {length}"
                )
                send_acknowledge(0, self.sock)
                self.logger.debug("send ack batch")                
            else:
                raise RuntimeError(f"unexpected type, {type(req)}")

        except protocol.BetParseError as e:
            self.logger.error(
                f"action: apuesta_recibida | result: fail | cantidad: {e}"
            )
            send_acknowledge(1, self.sock)
            
    def ready(self):
        agencyid, req = recv_request(self.sock)
        if isinstance(req, protocol.GetWinners):
            winners = None
            self.barrier.wait()
            winners = self.shared_store.get(agencyid)
            send_betting_results(winners, self.sock)
            self.state = State.EXIT
        return
    
    def handle(self):
        try:
            while True:
                if self.state == State.BETTING:
                    self.betting()
                elif self.state == State.DRAWING:
                    send_ready(self.sock)
                    self.state = State.READY
                elif self.state == State.READY:
                    self.ready()
                elif self.state == State.EXIT:
                    return
        except OSError as e:
            self.logger.error(f"action: receive_message | result: fail | error: {e}")
        except RuntimeError as e:
            self.logger.error(f"action: receive_message | result: fail | error: {e}")
        

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
