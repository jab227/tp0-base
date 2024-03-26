from enum import Enum
from typing import Optional

SIZEOF_UINT32 = 4

class Kind(Enum):
    BET = 0
    ACKNOWLEDGE = 1
    DONE = 2
    WINNERS = 3
    WINNERSUNAVAILABLE = 4
    WINNERSLIST = 5
    

def decode(b: bytes) -> Optional[Kind]:
    try:
        return Kind(b[0])
    except Exception:
        return None    
