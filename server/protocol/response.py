import protocol.message as message
import common.utils as utils

from typing import Union

class Acknowledge:
    bet_count: int
    bet_numbers: list[int]    

    def __init__(self, bets: list[utils.Bet]):
        self.bet_count = len(bets)
        self.bet_numbers = [b.number for b in bets]
        
    def encode(self) -> bytes:
        kind = int.to_bytes(message.Kind.ACKNOWLEDGE.value, 1, byteorder='little')
        bet_count = int.to_bytes(self.bet_count, message.SIZEOF_UINT32, byteorder='little')
        ack_bytes = kind + bet_count
        for n in self.bet_numbers:
            ack_bytes += int.to_bytes(n, message.SIZEOF_UINT32, byteorder='little')
        return ack_bytes

    
class WinnersList:
    winners_count: int
    dnis: list[str]
    
    def __init__(self, dnis: list[str]) -> None:
        self.winners_count = len(dnis)
        self.dnis = dnis
        
    def encode(self) -> bytes:
        kind = int.to_bytes(message.Kind.WINNERSLIST.value, 1, byteorder='little')
        winners_count = int.to_bytes(self.winners_count,
                                     message.SIZEOF_UINT32,
                                     byteorder='little')
        payload = b''
        for dni in self.dnis:
            payload += dni.encode()
            payload += b','
        payload_size = int.to_bytes(len(payload), message.SIZEOF_UINT32, byteorder='little')
        return kind + winners_count + payload_size + payload

    
class WinnersUnavailable:
    def encode(self) -> bytes:
        return int.to_bytes(message.Kind.WINNERSUNAVAILABLE.value, 1, byteorder='little')

    
Response = Union[Acknowledge, WinnersUnavailable, WinnersList]


def encode_response(res: Response) -> bytes:
    return res.encode()
