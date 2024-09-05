from enum import Enum, IntEnum
import logging
from typing import Union
from dataclasses import dataclass
import common.utils as utils

PAYLOAD_FIELD_COUNT = 5

class BetParseError(Exception):
    def __init__(self, length):
        self.length = length

def parse_bet(bet_data: bytes, id: int) -> utils.Bet:
    fields = bet_data.split(b',')
    if len(fields) != PAYLOAD_FIELD_COUNT:
        raise RuntimeError(f"wrong number of fields in payload: {len(fields)}")
    fields = list(map(lambda f: f.decode("utf-8"), fields))
    return utils.Bet(str(id), fields[0], fields[1], fields[2], fields[3], fields[4])


class MessageKind(Enum):
    POST_BET = 0
    BET_BATCH = 1
    BET_BATCH_END = 2
    GET_WINNERS = 3

class BatchEnd:
    pass

class GetWinners:
    pass

@dataclass
class Header:
    kind: MessageKind  #  u8 sizeof == 1
    agency_id: int  # u32 sizeof == 4
    payload_size: int  # u32 sizeof == 4

    SIZE = 9

    @classmethod
    def decode(cls, p: bytes):
        kind = MessageKind(p[0])
        payload_size = int.from_bytes(p[1:5], byteorder='little')
        agency_id = int.from_bytes(p[5:9], byteorder='little')
        return cls(kind=kind, agency_id=agency_id, payload_size=payload_size)


RequestUnion = Union[list[utils.Bet], BatchEnd, GetWinners]


class ResponseKind(IntEnum):
    ACKNOWLEDGE = 0
    WINNERS_READY = 1
    BETTING_RESULTS = 2
    
    
class Request:
    header: Header
    payload: bytes

    def __init__(self, header: Header, payload: bytes):
        self.header = header
        self.payload = payload

    def parse(self) -> RequestUnion:
        bets = []
        if self.header.kind == MessageKind.BET_BATCH:
            try:
                i = 0
                p = self.payload
                while i < len(self.payload):
                    bet_size = int.from_bytes(p[i : i + 4], byteorder='little')
                    i += 4
                    bet_data = p[i : i + bet_size]
                    bet = parse_bet(bet_data, self.header.agency_id)
                    i += bet_size
                    bets.append(bet)
            except ValueError as e:
                logging.debug(f"error: {e}")
                raise BetParseError(len(bets))
        elif self.header.kind == MessageKind.BET_BATCH_END:
            return BatchEnd()
        elif self.header.kind == MessageKind.GET_WINNERS:
            return GetWinners()
        else:
            bets.append(parse_bet(self.payload, self.header.agency_id))
        return bets


@dataclass
class WinnersReady:
    SIZE = 1
    def encode(self) -> bytes:
        return b''

@dataclass
class BettingResults:
    bets: list[utils.Bet]

    def encode(self) -> bytes:
        payload_str = ""
        for i, bet in enumerate(self.bets):
            payload_str += f"{bet.document}"
            if i < len(self.bets) - 1:
                payload_str += ","
                
        return bytes(payload_str, 'utf-8')

@dataclass
class AcknowledgeResponse:
    bet_status: int  # u32 sizeof == 4
    SIZE = 1

    def encode(self) -> bytes:
        return int.to_bytes(self.bet_status, self.SIZE, byteorder='little')

Response = Union[AcknowledgeResponse, WinnersReady, BettingResults]

def encode(res: Response) -> bytes:
    result = b''
    if isinstance(res, AcknowledgeResponse):
        result += int.to_bytes(ResponseKind.ACKNOWLEDGE, 1, byteorder='little')
        result += int.to_bytes(1, 4, byteorder='little')
        result += res.encode()
    elif isinstance(res, WinnersReady):
        result += int.to_bytes(ResponseKind.WINNERS_READY, 1, byteorder='little')
        result += int.to_bytes(0, 4, byteorder='little')
        result += res.encode()
    elif isinstance(res, BettingResults):
        result += int.to_bytes(ResponseKind.BETTING_RESULTS, 1, byteorder='little')
        payload = res.encode()
        result += int.to_bytes(len(payload), 4, byteorder='little')
        result += payload
    else:
        assert False
    return result
