import common.utils as utils
from multiprocessing.managers import SyncManager
from typing import Optional


class BetsStorage:
    
    def __init__(self, number_of_agencies: int, barrier):
        self.barrier = barrier
        self.agencies = {i: False for i in range(1, int(number_of_agencies) + 1)}
        self.winners = None

    def signal(self, agencyid: int):
        self.agencies[agencyid] = True
        if self.winners is None and all(self.agencies.values()):
            self.winners = {i: [] for i in self.agencies.keys()}
            for bet in utils.load_bets():
                if utils.has_won(bet):
                        self.winners[bet.agency].append(bet)

    def store_bets(self, bets: list[utils.Bet]):
        utils.store_bets(bets)
        
    def get(self, agencyid: int) -> Optional[list[utils.Bet]]:
            return self.winners[agencyid] if self.winners is not None else None


class StorageManager(SyncManager):
    pass
