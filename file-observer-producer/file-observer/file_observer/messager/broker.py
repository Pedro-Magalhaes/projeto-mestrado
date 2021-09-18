from abc import ABC, abstractclassmethod

class Broker(ABC):

    @abstractclassmethod
    def send(self, topic, msg, key) -> bool:
        pass