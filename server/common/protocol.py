from enum import Enum
import logging
from typing import Optional, Union
from dataclasses import dataclass
import common.utils as utils


class MessageKind(Enum):
    BET = 0
    ACKNOWLEDGE = 1
    DONE = 2
    WINNERS = 3
    WINNERSUNAVAILABLE = 4
    WINNERSLIST = 5

SIZEOF_UINT32 = 4
PAYLOAD_FIELD_COUNT = 5


class Bet:
    payload_size: int
    count: int
    agency_id: int
    bets: list[utils.Bet]

    SIZE = 12

    def __init__(self, payload_size: int, count: int, agency_id: int) -> None:
        self.payload_size = payload_size
        self.count = count
        self.agency_id = agency_id
        self.bets = []

@dataclass
class Done:
    agency_id: int

@dataclass
class Winners:
    agency_id: int
    

Request = Union[Bet, Winners, Done]

def decode(kind: MessageKind, b: bytes) -> Optional[Request]:
    if kind == MessageKind.BET:
        payload_size = int.from_bytes(b[0:4], byteorder='little')
        count = int.from_bytes(b[4:8], byteorder='little')    
        agency_id = int.from_bytes(b[8:Bet.SIZE], byteorder='little')
        return Bet(payload_size, count, agency_id)
    elif kind == MessageKind.DONE:
        agency_id = int.from_bytes(b[0:4], byteorder='little')        
        return Done(agency_id)
    elif kind == MessageKind.WINNERS:
        agency_id = int.from_bytes(b[0:4], byteorder='little')        
        return Winners(agency_id)
    else:
        return None
        


class Acknowledge:
    bet_count: int
    bet_numbers: list[int]    

    def __init__(self, bets: list[utils.Bet]):
        self.bet_count = len(bets)
        self.bet_numbers = [b.number for b in bets]
        
    def encode(self) -> bytes:
        kind = int.to_bytes(MessageKind.ACKNOWLEDGE.value, 1, byteorder='little')
        bet_count = int.to_bytes(self.bet_count, SIZEOF_UINT32, byteorder='little')
        ack_bytes = kind + bet_count
        for n in self.bet_numbers:
            ack_bytes += int.to_bytes(n, SIZEOF_UINT32, byteorder='little')
        return ack_bytes

    
class WinnersList:
    winners_count: int
    dnis: list[str]
    
    def __init__(self, dnis: list[str]) -> None:
        self.winners_count = len(dnis)
        self.dnis = dnis
        
    def encode(self) -> bytes:
        kind = int.to_bytes(MessageKind.WINNERSLIST.value, 1, byteorder='little')
        winners_count = int.to_bytes(self.winners_count,
                                     SIZEOF_UINT32,
                                     byteorder='little')
        payload = b''
        for dni in self.dnis:
            payload += dni.encode()
            payload += b','
        payload_size = int.to_bytes(len(payload), SIZEOF_UINT32, byteorder='little')
        return kind + winners_count + payload_size + payload

    
class WinnersUnavailable:
    def encode(self) -> bytes:
        return int.to_bytes(MessageKind.WINNERSUNAVAILABLE.value, 1, byteorder='little')

    
Response = Union[Acknowledge, WinnersUnavailable, WinnersList]


def parse_payload(req: Bet, payload: bytes) -> list[utils.Bet]:
    lines = payload.splitlines()
    bets = []
    for line in lines:
        fields = line.split(b',')        
        if len(fields) != PAYLOAD_FIELD_COUNT:
            raise RuntimeError(f"wrong number of fields in payload: {len(fields)}")
        fields = list(map(lambda f: f.decode("utf-8"), fields ))
        bet = utils.Bet(str(req.agency_id), fields[0], fields[1], fields[2], fields[3], fields[4])
        bets.append(bet)
    if len(bets) != req.count:
            raise RuntimeError(f"wrong batch size: {len(bets)}")
    return bets


def encode_response(res: Response) -> bytes:
    return res.encode()
    

def decode_kind(b: bytes) -> Optional[MessageKind]:
    try:
        return MessageKind(b[0])
    except Exception:
        return None
