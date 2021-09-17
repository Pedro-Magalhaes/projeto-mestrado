from abc import ABC, abstractclassmethod

class Broker(ABC):

    @abstractclassmethod
    def send(self, msg) -> bool:
        pass