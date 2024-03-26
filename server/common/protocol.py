from enum import Enum
from typing import Optional
from dataclasses import dataclass
import common.utils as utils

class MessageKind(Enum):
    POST_BET = 0


PAYLOAD_FIELD_COUNT = 5

@dataclass
class Header:
    kind: MessageKind
    agency_id: int
    payload_size: int

    HEADER_SIZE = 9

@dataclass
class Ack:
    bet_number: int

    ACK_SIZE = 4


def decode_header(b: bytes) -> Optional[Header]:
    try:
        kind = MessageKind(b[0])
    except Exception:
        return None
    payload_size = int.from_bytes(b[1:5], byteorder='little')    
    agency_id = int.from_bytes(b[5:9], byteorder='little')
    return Header(kind, agency_id, payload_size)



def parse_payload(agency_id: str, payload: bytes) -> utils.Bet:
    fields = payload.split(b',')
    if len(fields) != PAYLOAD_FIELD_COUNT:
        raise RuntimeError(f"wrong number of fields in payload: {len(fields)}")
    fields = list(map(lambda f: f.decode("utf-8"), fields ))
    return utils.Bet(agency_id, fields[0], fields[1], fields[2], fields[3], fields[4])


def encode_ack(ack: Ack) -> bytes:
    return int.to_bytes(ack.bet_number, ack.ACK_SIZE, byteorder='little')
