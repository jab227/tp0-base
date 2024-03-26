import common.utils as utils
from multiprocessing.managers import BaseManager
from typing import Optional


class BetsStorage:
    winners: Optional[dict[int, list[str]]]
    agencies: dict[int, bool]
    
    def __init__(self, number_of_agencies: int):
        self.agencies = {i: False for i in range(1, int(number_of_agencies) + 1)}
        self.winners = None

    def store_bets(self, id: int, bets: list[utils.Bet], done: bool=False):
        if not done:
            utils.store_bets(bets)
            return
        self.agencies[id] = True
        if all(self.agencies.values()) and self.winners is None:
            self.winners = {i: [] for i in self.agencies.keys()}
            for bet in utils.load_bets():
                if utils.has_won(bet):
                        self.winners[bet.agency].append(bet.document)

    def get_winners(self, id: int) -> Optional[list[str]]:
            return self.winners[id] if self.winners is not None else None


class StorageManager(BaseManager):
    pass
