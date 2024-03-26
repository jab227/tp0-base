import common.utils as utils
import protocol.message as message
from dataclasses import dataclass
from typing import Optional, Union


class Bet:
    PAYLOAD_FIELD_COUNT = 5    
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

def decode(kind: message.Kind, b: bytes) -> Optional[Request]:
    if kind == message.Kind.BET:
        payload_size = int.from_bytes(b[0:4], byteorder='little')
        count = int.from_bytes(b[4:8], byteorder='little')    
        agency_id = int.from_bytes(b[8:Bet.SIZE], byteorder='little')
        return Bet(payload_size, count, agency_id)
    elif kind == message.Kind.DONE:
        agency_id = int.from_bytes(b[0:4], byteorder='little')        
        return Done(agency_id)
    elif kind == message.Kind.WINNERS:
        agency_id = int.from_bytes(b[0:4], byteorder='little')        
        return Winners(agency_id)
    else:
        return None

    
def parse_payload(req: Bet, payload: bytes) -> list[utils.Bet]:
    lines = payload.splitlines()
    bets = []
    for line in lines:
        fields = line.split(b',')        
        if len(fields) != Bet.PAYLOAD_FIELD_COUNT:
            raise RuntimeError(f"wrong number of fields in payload: {len(fields)}")
        fields = list(map(lambda f: f.decode("utf-8"), fields ))
        bet = utils.Bet(str(req.agency_id), fields[0], fields[1], fields[2], fields[3], fields[4])
        bets.append(bet)
    if len(bets) != req.count:
            raise RuntimeError(f"wrong batch size: {len(bets)}")
    return bets    
