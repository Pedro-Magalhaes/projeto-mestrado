from abc import ABC, abstractmethod

class Messager(ABC):

    @abstractmethod
    def send(self, msg):
        pass