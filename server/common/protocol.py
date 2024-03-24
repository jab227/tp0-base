from enum import Enum
import logging
from typing import Optional
from dataclasses import dataclass
import common.utils as utils

class MessageKind(Enum):
    BET = 0
    ACKNOWLEDGE = 1
    DONE = 2

SIZEOF_UINT32 = 4
PAYLOAD_FIELD_COUNT = 5

@dataclass
class Header:
    payload_size: int = 0
    count: int = 0
    agency_id: int = 0

    HEADER_SIZE = 12


@dataclass
class Acknowledge:
    bet_count: int
    bet_numbers: list[int]    
    kind: MessageKind = MessageKind.ACKNOWLEDGE


    
def decode_bet_message(b: bytes) -> Optional[Header]:
    payload_size = int.from_bytes(b[0:4], byteorder='little')
    count = int.from_bytes(b[4:8], byteorder='little')    
    agency_id = int.from_bytes(b[8:Header.HEADER_SIZE], byteorder='little')
    return Header(payload_size, count, agency_id)


def parse_payload(header: Header, payload: bytes) -> list[utils.Bet]:
    lines = payload.splitlines()
    bets = []
    for line in lines:
        fields = line.split(b',')        
        if len(fields) != PAYLOAD_FIELD_COUNT:
            raise RuntimeError(f"wrong number of fields in payload: {len(fields)}")
        fields = list(map(lambda f: f.decode("utf-8"), fields ))
        bet = utils.Bet(str(header.agency_id), fields[0], fields[1], fields[2], fields[3], fields[4])
        bets.append(bet)
    if len(bets) != header.count:
            raise RuntimeError(f"wrong batch size: {len(bets)}")
    return bets


def encode_ack(ack: Acknowledge) -> bytes:
    kind = int.to_bytes(ack.kind.value, 1, byteorder='little')
    bet_count = int.to_bytes(ack.bet_count, SIZEOF_UINT32, byteorder='little')
    ack_bytes = kind + bet_count
    for n in ack.bet_numbers:
        ack_bytes += int.to_bytes(n, SIZEOF_UINT32, byteorder='little')
    return ack_bytes
    


def decode_kind(b: bytes) -> Optional[MessageKind]:
    try:
        return MessageKind(b[0])
    except Exception:
        return None
