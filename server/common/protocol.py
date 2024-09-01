from enum import Enum
from typing import Optional
from dataclasses import dataclass
import common.utils as utils

PAYLOAD_FIELD_COUNT = 5

class MessageKind(Enum):
    POST_BET = 0

@dataclass
class Header:
    kind: MessageKind #  u8 sizeof == 1
    agency_id: int    # u32 sizeof == 4
    payload_size: int # u32 sizeof == 4

    SIZE = 9

    @classmethod
    def decode(cls, p: bytes):
        kind = MessageKind(p[0])
        payload_size = int.from_bytes(p[1:5], byteorder='little')    
        agency_id = int.from_bytes(p[5:9], byteorder='little')
        return cls(kind=kind, agency_id=agency_id, payload_size=payload_size)
4
    
class Request:
    header: Header
    payload: bytes

    def __init__(self, header: Header, payload: bytes):
        self.header = header
        self.payload = payload

    def parse_bet(self):
        fields = self.payload.split(b',')
        if len(fields) != PAYLOAD_FIELD_COUNT:
            raise RuntimeError(f"wrong number of fields in payload: {len(fields)}")
        fields = list(map(lambda f: f.decode("utf-8"), fields ))
        return utils.Bet(self.header.agency_id, fields[0], fields[1], fields[2], fields[3], fields[4])

    
@dataclass
class AcknowledgeResponse:
    bet_number: int # u32 sizeof == 4
    SIZE = 4    

    def encode(self) -> bytes:
        return int.to_bytes(self.bet_number, SIZE, byteorder='little')

