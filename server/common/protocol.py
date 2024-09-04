from enum import Enum
import logging
from typing import Union
from typing import Optional
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


class BatchEnd:
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


class Request:
    header: Header
    payload: bytes

    def __init__(self, header: Header, payload: bytes):
        self.header = header
        self.payload = payload

    def parse(self) -> Union[list[utils.Bet], BatchEnd]:
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
        else:
            bets.append(parse_bet(self.payload, self.header.agency_id))
        return bets


@dataclass
class AcknowledgeResponse:
    bet_status: int  # u32 sizeof == 4
    SIZE = 1

    def encode(self) -> bytes:
        return int.to_bytes(self.bet_status, self.SIZE, byteorder='little')
